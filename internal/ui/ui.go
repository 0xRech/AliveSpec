package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/0xRech/AliveSpec/internal/observe"
)

type Printer struct {
	out   io.Writer
	color bool
}

func New(out io.Writer, mode string) *Printer {
	color := false
	switch mode {
	case "always":
		color = true
	case "auto", "":
		if f, ok := out.(*os.File); ok {
			if info, err := f.Stat(); err == nil {
				color = info.Mode()&os.ModeCharDevice != 0
			}
		}
	}
	return &Printer{out: out, color: color}
}

func (p *Printer) c(code, value string) string {
	if !p.color {
		return value
	}
	return "\x1b[" + code + "m" + value + "\x1b[0m"
}

func (p *Printer) RecordHeader(name, backend string, duration time.Duration, processes []string, output string) {
	fmt.Fprintln(p.out)
	fmt.Fprintf(p.out, "%s\n", p.c("1;36", " ALIVESPEC  /  RUNTIME LEARN "))
	fmt.Fprintln(p.out, "────────────────────────────────────────────────────────────")
	fmt.Fprintf(p.out, "  journey     %s\n", p.c("1", name))
	fmt.Fprintf(p.out, "  observer    %s\n", backend)
	if duration > 0 {
		fmt.Fprintf(p.out, "  window      %s\n", duration)
	} else {
		fmt.Fprintln(p.out, "  window      until Ctrl+C")
	}
	if len(processes) > 0 {
		fmt.Fprintf(p.out, "  processes   %s\n", strings.Join(processes, ", "))
	} else {
		fmt.Fprintf(p.out, "  processes   %s\n", p.c("33", "all (noisy)"))
	}
	fmt.Fprintf(p.out, "  contract    %s\n", output)
	fmt.Fprintln(p.out, "────────────────────────────────────────────────────────────")
	fmt.Fprintf(p.out, "%s recording runtime evidence…\n\n", p.c("32", "●"))
}

func (p *Printer) Event(e observe.Event) {
	ts := e.Time.Format("15:04:05")
	process := truncate(e.Process, 15)
	switch e.Kind {
	case observe.KindProcess:
		fmt.Fprintf(p.out, "  %s  %s  %-15s  %s\n", p.c("90", ts), p.c("35", "PROC"), process, e.Path)
	case observe.KindTCP:
		fmt.Fprintf(p.out, "  %s  %s   %-15s  %s:%d\n", p.c("90", ts), p.c("36", "TCP"), process, e.Host, e.Port)
	case observe.KindFile:
		fmt.Fprintf(p.out, "  %s  %s  %-15s  %s\n", p.c("90", ts), p.c("34", "FILE"), process, e.Path)
	}
}

func (p *Printer) Warning(message string) {
	fmt.Fprintf(p.out, "\n%s %s\n", p.c("33", "!"), message)
}

func (p *Printer) Summary(name string, elapsed time.Duration, events, processes, connections, files int, output string) {
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, "────────────────────────────────────────────────────────────")
	fmt.Fprintf(p.out, "%s %s\n\n", p.c("1;32", "✓"), p.c("1", "Operational contract compiled"))
	fmt.Fprintf(p.out, "  journey       %s\n", name)
	fmt.Fprintf(p.out, "  observed      %d events in %s\n", events, prettyDuration(elapsed))
	fmt.Fprintf(p.out, "  processes     %d\n", processes)
	fmt.Fprintf(p.out, "  connections   %d\n", connections)
	fmt.Fprintf(p.out, "  config files  %d\n", files)
	fmt.Fprintln(p.out)
	fmt.Fprintf(p.out, "  %s  %s\n", p.c("32", "CONTRACT"), output)
	fmt.Fprintln(p.out, "────────────────────────────────────────────────────────────")
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	if max <= 1 {
		return value[:max]
	}
	return value[:max-1] + "…"
}

func prettyDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(100 * time.Millisecond).String()
}
