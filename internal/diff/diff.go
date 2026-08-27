package diff

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/0xRech/AliveSpec/internal/spec"
	"github.com/0xRech/AliveSpec/internal/ui"
)

type change struct {
	kind  string
	value string
}

func Run(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	color := fs.String("color", "auto", "terminal color: auto, always, never")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: alivespec diff <before.yaml> <after.yaml>")
	}
	if *color != "auto" && *color != "always" && *color != "never" {
		return fmt.Errorf("invalid --color %q; use auto, always or never", *color)
	}

	beforePath, afterPath := fs.Arg(0), fs.Arg(1)
	before, err := spec.Load(beforePath)
	if err != nil {
		return err
	}
	after, err := spec.Load(afterPath)
	if err != nil {
		return err
	}

	oldSet, newSet := flatten(before), flatten(after)
	var removed, added []change
	for key, item := range oldSet {
		if _, ok := newSet[key]; !ok {
			removed = append(removed, item)
		}
	}
	for key, item := range newSet {
		if _, ok := oldSet[key]; !ok {
			added = append(added, item)
		}
	}
	orderChanges(removed)
	orderChanges(added)

	printer := ui.New(os.Stdout, *color)
	printer.DiffHeader(before.Metadata.Name, after.Metadata.Name, beforePath, afterPath)
	if len(removed) == 0 && len(added) == 0 {
		printer.DiffNoChanges()
	} else {
		for _, item := range removed {
			printer.DiffChange(false, item.kind, item.value)
		}
		if len(removed) > 0 && len(added) > 0 {
			fmt.Fprintln(os.Stdout)
		}
		for _, item := range added {
			printer.DiffChange(true, item.kind, item.value)
		}
	}
	printer.DiffSummary(len(added), len(removed))
	return nil
}

func orderChanges(items []change) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].kind == items[j].kind {
			return items[i].value < items[j].value
		}
		return items[i].kind < items[j].kind
	})
}

func flatten(c *spec.Contract) map[string]change {
	out := map[string]change{}
	for _, r := range c.Requires.Processes {
		key := fmt.Sprintf("process:%s:running=%t", r.Name, r.Running)
		value := r.Name
		if r.Executable != "" {
			value += " · " + r.Executable
		}
		out[key] = change{kind: "process", value: value}
	}
	for _, r := range c.Requires.Services {
		key := fmt.Sprintf("service:%s:active=%t", r.Name, r.Active)
		out[key] = change{kind: "service", value: r.Name}
	}
	for _, r := range c.Requires.Listeners {
		key := fmt.Sprintf("listener:%s:%d", r.Protocol, r.Port)
		out[key] = change{kind: "listener", value: fmt.Sprintf("%s/%d", strings.ToUpper(r.Protocol), r.Port)}
	}
	for _, r := range c.Requires.Connections {
		key := fmt.Sprintf("connection:%s:%s:%d", r.Protocol, r.Host, r.Port)
		out[key] = change{kind: "connection", value: fmt.Sprintf("%s %s", strings.ToUpper(r.Protocol), net.JoinHostPort(r.Host, fmt.Sprint(r.Port)))}
	}
	for _, r := range c.Requires.DNS {
		key := fmt.Sprintf("dns:%s:resolves=%t", r.Name, r.Resolves)
		out[key] = change{kind: "dns", value: r.Name}
	}
	for _, r := range c.Requires.TLS {
		key := fmt.Sprintf("tls:%s:%d:minValidityDays=%d", r.Host, r.Port, r.MinValidityDays)
		out[key] = change{kind: "tls", value: fmt.Sprintf("%s · minimum %d days", net.JoinHostPort(r.Host, fmt.Sprint(r.Port)), r.MinValidityDays)}
	}
	for _, r := range c.Requires.Files {
		key := fmt.Sprintf("file:%s:sha256=%s", r.Path, r.SHA256)
		value := r.Path
		if r.SHA256 != "" {
			value += " · SHA-256 " + shortHash(r.SHA256)
		}
		out[key] = change{kind: "file", value: value}
	}
	return out
}

func shortHash(value string) string {
	if len(value) <= 12 {
		return value
	}
	return value[:12] + "…"
}
