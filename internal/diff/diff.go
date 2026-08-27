package diff

import (
	"flag"
	"fmt"
	"sort"

	"github.com/0xRech/AliveSpec/internal/spec"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: alivespec diff <before.yaml> <after.yaml>")
	}
	before, err := spec.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	after, err := spec.Load(fs.Arg(1))
	if err != nil {
		return err
	}
	oldSet, newSet := flatten(before), flatten(after)
	var removed, added []string
	for item := range oldSet {
		if _, ok := newSet[item]; !ok {
			removed = append(removed, item)
		}
	}
	for item := range newSet {
		if _, ok := oldSet[item]; !ok {
			added = append(added, item)
		}
	}
	sort.Strings(removed)
	sort.Strings(added)
	fmt.Printf("AliveSpec diff: %s -> %s\n\n", before.Metadata.Name, after.Metadata.Name)
	if len(removed) == 0 && len(added) == 0 {
		fmt.Println("No requirement changes.")
		return nil
	}
	for _, item := range removed {
		fmt.Printf("- %s\n", item)
	}
	for _, item := range added {
		fmt.Printf("+ %s\n", item)
	}
	return nil
}

func flatten(c *spec.Contract) map[string]struct{} {
	out := map[string]struct{}{}
	for _, r := range c.Requires.Processes {
		out[fmt.Sprintf("process:%s:running=%t", r.Name, r.Running)] = struct{}{}
	}
	for _, r := range c.Requires.Services {
		out[fmt.Sprintf("service:%s:active=%t", r.Name, r.Active)] = struct{}{}
	}
	for _, r := range c.Requires.Listeners {
		out[fmt.Sprintf("listener:%s:%d", r.Protocol, r.Port)] = struct{}{}
	}
	for _, r := range c.Requires.Connections {
		out[fmt.Sprintf("connection:%s:%s:%d", r.Protocol, r.Host, r.Port)] = struct{}{}
	}
	for _, r := range c.Requires.DNS {
		out[fmt.Sprintf("dns:%s:resolves=%t", r.Name, r.Resolves)] = struct{}{}
	}
	for _, r := range c.Requires.TLS {
		out[fmt.Sprintf("tls:%s:%d:minValidityDays=%d", r.Host, r.Port, r.MinValidityDays)] = struct{}{}
	}
	for _, r := range c.Requires.Files {
		out[fmt.Sprintf("file:%s:sha256=%s", r.Path, r.SHA256)] = struct{}{}
	}
	return out
}
