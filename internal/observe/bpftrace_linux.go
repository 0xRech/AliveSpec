//go:build linux

package observe

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BPFTrace struct {
	Processes []string
	DNSProbe  bool
}

var safeComm = regexp.MustCompile(`^[A-Za-z0-9_.:+-]{1,32}$`)

func NewBPFTrace(processes []string) (Observer, error) {
	if _, err := exec.LookPath("bpftrace"); err != nil {
		return nil, fmt.Errorf("bpftrace was not found in PATH; install bpftrace to use runtime recording")
	}
	for _, process := range processes {
		if !safeComm.MatchString(process) {
			return nil, fmt.Errorf("invalid --comm value %q", process)
		}
	}
	return &BPFTrace{Processes: processes, DNSProbe: supportsDNSProbe()}, nil
}

func (b *BPFTrace) Name() string {
	if b.DNSProbe {
		return "eBPF / bpftrace · DNS enabled"
	}
	return "eBPF / bpftrace"
}

func (b *BPFTrace) Start(ctx context.Context) (<-chan Event, <-chan error, error) {
	cmd := exec.CommandContext(ctx, "bpftrace", "-q", "-e", b.script())
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	events := make(chan Event, 256)
	errs := make(chan error, 2)
	stdoutDone := make(chan error, 1)
	stderrDone := make(chan string, 1)

	go func() {
		defer close(events)
		stdoutDone <- scanEvents(stdout, events)
	}()
	go func() {
		stderrDone <- scanStderr(stderr)
	}()

	go func() {
		scanErr := <-stdoutDone
		stderrText := <-stderrDone
		waitErr := cmd.Wait()
		if scanErr != nil {
			errs <- scanErr
		}
		if waitErr != nil && ctx.Err() == nil {
			if stderrText != "" {
				errs <- fmt.Errorf("bpftrace exited: %w: %s", waitErr, stderrText)
			} else {
				errs <- fmt.Errorf("bpftrace exited: %w", waitErr)
			}
		}
		close(errs)
	}()

	return events, errs, nil
}

func (b *BPFTrace) script() string {
	filter := b.filter()
	dnsProbe := ""
	if b.DNSProbe {
		dnsProbe = fmt.Sprintf(`
uprobe:libc.so.6:getaddrinfo %s
{
  printf("AS|DNS|%%d|%%s|%%s\n", pid, comm, str(arg0));
}
`, filter)
	}

	return fmt.Sprintf(`#include <linux/in.h>
#include <linux/in6.h>
#include <linux/socket.h>

tracepoint:sched:sched_process_exec %s
{
  printf("AS|EXEC|%%d|%%s|%%s\n", pid, comm, str(args->filename));
}

tracepoint:syscalls:sys_enter_openat %s
{
  printf("AS|FILE|%%d|%%s|%%s\n", pid, comm, str(args->filename));
}

tracepoint:syscalls:sys_enter_connect %s
{
  $sa = (struct sockaddr *)args->uservaddr;
  if ($sa->sa_family == 2) {
    $sa4 = (struct sockaddr_in *)args->uservaddr;
    printf("AS|TCP4|%%d|%%s|%%s|%%d\n", pid, comm, ntop($sa4->sin_addr.s_addr), bswap($sa4->sin_port));
  }
  if ($sa->sa_family == 10) {
    $sa6 = (struct sockaddr_in6 *)args->uservaddr;
    printf("AS|TCP6|%%d|%%s|%%s|%%d\n", pid, comm, ntop($sa6->sin6_addr.in6_u.u6_addr8), bswap($sa6->sin6_port));
  }
}
%s`, filter, filter, filter, dnsProbe)
}

func (b *BPFTrace) filter() string {
	if len(b.Processes) == 0 {
		return ""
	}
	parts := make([]string, 0, len(b.Processes))
	for _, process := range b.Processes {
		parts = append(parts, fmt.Sprintf(`comm == "%s"`, process))
	}
	return "/ " + strings.Join(parts, " || ") + " /"
}

func supportsDNSProbe() bool {
	cmd := exec.Command("bpftrace", "-l", "uprobe:libc.so.6:getaddrinfo")
	out, err := cmd.Output()
	return err == nil && strings.Contains(string(out), "getaddrinfo")
}

func scanEvents(r io.Reader, out chan<- Event) error {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "AS|") {
			continue
		}
		event, err := parseLine(line)
		if err != nil {
			return err
		}
		out <- event
	}
	return scanner.Err()
}

func scanStderr(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "; ")
}

func parseLine(line string) (Event, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 5 {
		return Event{}, fmt.Errorf("invalid observer event %q", line)
	}
	pid, err := strconv.Atoi(parts[2])
	if err != nil {
		return Event{}, fmt.Errorf("invalid observer pid in %q", line)
	}
	e := Event{Time: time.Now(), PID: pid, Process: parts[3]}
	switch parts[1] {
	case "EXEC":
		e.Kind = KindProcess
		e.Path = strings.Join(parts[4:], "|")
	case "FILE":
		e.Kind = KindFile
		e.Path = strings.Join(parts[4:], "|")
	case "DNS":
		e.Kind = KindDNS
		e.Name = strings.Join(parts[4:], "|")
	case "TCP4", "TCP6":
		if len(parts) != 6 {
			return Event{}, fmt.Errorf("invalid TCP observer event %q", line)
		}
		port, err := strconv.Atoi(parts[5])
		if err != nil {
			return Event{}, fmt.Errorf("invalid TCP port in %q", line)
		}
		e.Kind = KindTCP
		e.Host = parts[4]
		e.Port = port
	default:
		return Event{}, fmt.Errorf("unknown observer event %q", line)
	}
	return e, nil
}
