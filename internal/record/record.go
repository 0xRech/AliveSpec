package record

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/0xRech/AliveSpec/internal/observe"
	"github.com/0xRech/AliveSpec/internal/spec"
	"github.com/0xRech/AliveSpec/internal/ui"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func Run(args []string) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	name := fs.String("name", "journey", "name of the successful journey being recorded")
	out := fs.String("out", "", "output contract path (default: <name>.alivespec.yaml)")
	duration := fs.Duration("duration", 0, "recording window; 0 records until Ctrl+C")
	color := fs.String("color", "auto", "terminal color: auto, always, never")
	allFiles := fs.Bool("all-files", false, "include all opened absolute files instead of likely configuration files only")
	var processes multiFlag
	fs.Var(&processes, "comm", "Linux process name to observe (repeatable; empty records system-wide)")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		if *name == "journey" && fs.NArg() == 1 {
			*name = fs.Arg(0)
		} else {
			return fmt.Errorf("usage: alivespec record [journey] [flags]")
		}
	}
	if strings.TrimSpace(*name) == "" {
		return fmt.Errorf("journey name cannot be empty")
	}
	if *color != "auto" && *color != "always" && *color != "never" {
		return fmt.Errorf("invalid --color %q; use auto, always or never", *color)
	}
	if *duration < 0 {
		return fmt.Errorf("--duration cannot be negative")
	}
	if *out == "" {
		filename := slug(*name)
		if filename == "" {
			filename = "journey"
		}
		*out = filename + ".alivespec.yaml"
	}

	observer, err := observe.NewBPFTrace(processes)
	if err != nil {
		return err
	}

	printer := ui.New(os.Stdout, *color)
	printer.RecordHeader(*name, observer.Name(), *duration, processes, *out)
	if len(processes) == 0 {
		printer.Warning("No --comm filter set. System-wide recording can contain unrelated runtime noise.")
	}

	baseCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	ctx := baseCtx
	cancel := func() {}
	if *duration > 0 {
		ctx, cancel = context.WithTimeout(baseCtx, *duration)
	}
	defer cancel()

	events, errs, err := observer.Start(ctx)
	if err != nil {
		return err
	}

	collector := NewCollector(*allFiles)
	started := time.Now()
	for events != nil || errs != nil {
		select {
		case event, ok := <-events:
			if !ok {
				events = nil
				continue
			}
			show, confidence := collector.Add(event)
			if show {
				printer.Event(event, confidence)
			}
		case observerErr, ok := <-errs:
			if !ok {
				errs = nil
				continue
			}
			if observerErr != nil {
				return observerErr
			}
		}
	}

	contract := collector.Contract(*name)
	if err := spec.Save(*out, contract); err != nil {
		return err
	}
	eventCount, processCount, connectionCount, fileCount := collector.Counts()
	printer.Summary(*name, time.Since(started), eventCount, processCount, connectionCount, collector.DNSCount(), fileCount, *out)
	return nil
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
