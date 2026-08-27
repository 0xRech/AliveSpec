package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/0xRech/AliveSpec/internal/observe"
)

const boxWidth = 72

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
	p.boxTop("AliveSpec · Runtime Learning")
	p.boxStatus("● RECORDING", "36")
	p.boxBlank()
	p.boxRow("Journey", name)
	if duration > 0 {
		p.boxRow("Window", duration.String())
	} else {
		p.boxRow("Window", "until Ctrl+C")
	}
	p.boxRow("Observer", backend)
	if len(processes) > 0 {
		p.boxRow("Processes", strings.Join(processes, ", "))
	} else {
		p.boxRow("Processes", "all (system-wide)")
	}
	p.boxRow("Contract", output)
	p.boxBottom()
	p.section("DISCOVERED")
}

func (p *Printer) Event(e observe.Event) {
	process := e.Process
	if process == "" {
		process = "unknown"
	}
	stamp := p.c("90", e.Time.Format("15:04:05"))

	switch e.Kind {
	case observe.KindProcess:
		fmt.Fprintf(p.out, "  %s  %s  %s\n", stamp, p.c("35", "◈ PROCESS   "), p.c("1", process))
		fmt.Fprintf(p.out, "            └─ %s\n\n", e.Path)
	case observe.KindTCP:
		fmt.Fprintf(p.out, "  %s  %s  %s\n", stamp, p.c("36", "↗ CONNECTION"), p.c("1", process))
		fmt.Fprintf(p.out, "            └─ %s:%d\n\n", e.Host, e.Port)
	case observe.KindFile:
		fmt.Fprintf(p.out, "  %s  %s  %s\n", stamp, p.c("34", "◫ FILE      "), p.c("1", process))
		fmt.Fprintf(p.out, "            └─ %s\n\n", e.Path)
	}
}

func (p *Printer) Warning(message string) {
	fmt.Fprintf(p.out, "  %s  %s\n\n", p.c("33", "! WARNING"), message)
}

func (p *Printer) Summary(name string, elapsed time.Duration, events, processes, connections, files int, output string) {
	p.boxTop("AliveSpec · Contract Compiled")
	p.boxStatus("✓ READY", "32")
	p.boxBlank()
	p.boxRow("Journey", name)
	p.boxRow("Duration", prettyDuration(elapsed))
	p.boxRow("Events", fmt.Sprint(events))
	p.boxRow("Processes", fmt.Sprint(processes))
	p.boxRow("Connections", fmt.Sprint(connections))
	p.boxRow("Config files", fmt.Sprint(files))
	p.boxBlank()
	p.boxRow("Contract", output)
	p.boxBottom()
}

func (p *Printer) VerifyHeader(name, contract, host string) {
	fmt.Fprintln(p.out)
	p.boxTop("AliveSpec · Contract Verification")
	p.boxStatus("● VERIFYING", "36")
	p.boxBlank()
	p.boxRow("Journey", name)
	if host != "" {
		p.boxRow("Learned host", host)
	}
	p.boxRow("Contract", contract)
	p.boxBottom()
	p.section("CHECKS")
}

func (p *Printer) Check(ok bool, kind, target, detail string) {
	status := p.c("31", "✕")
	if ok {
		status = p.c("32", "✓")
	}
	kind = strings.ToUpper(kind)
	fmt.Fprintf(p.out, "  %s  %-11s %s\n", status, p.c("1", padRight(kind, 11)), target)
	if detail != "" {
		fmt.Fprintf(p.out, "     %-11s └─ %s\n", "", detail)
	}
	fmt.Fprintln(p.out)
}

func (p *Printer) VerifySummary(passed, total int, elapsed time.Duration) {
	healthy := passed == total
	title := "AliveSpec · Contract Health"
	p.boxTop(title)
	if healthy {
		p.boxStatus("✓ HEALTHY", "32")
	} else {
		p.boxStatus("✕ DEGRADED", "31")
	}
	p.boxBlank()
	p.boxRow("Checks", fmt.Sprintf("%d/%d passed", passed, total))
	p.boxRow("Duration", prettyDuration(elapsed))
	p.boxBottom()
}

func (p *Printer) DiffHeader(before, after, beforePath, afterPath string) {
	fmt.Fprintln(p.out)
	p.boxTop("AliveSpec · Contract Diff")
	p.boxStatus("◆ COMPARING", "36")
	p.boxBlank()
	p.boxRow("Before", before)
	p.boxRow("After", after)
	p.boxRow("From", beforePath)
	p.boxRow("To", afterPath)
	p.boxBottom()
	p.section("CHANGES")
}

func (p *Printer) DiffChange(added bool, kind, value string) {
	symbol := p.c("31", "−")
	if added {
		symbol = p.c("32", "+")
	}
	fmt.Fprintf(p.out, "  %s  %-11s %s\n", symbol, p.c("1", padRight(strings.ToUpper(kind), 11)), value)
}

func (p *Printer) DiffSummary(added, removed int) {
	fmt.Fprintln(p.out)
	p.boxTop("AliveSpec · Diff Summary")
	if added == 0 && removed == 0 {
		p.boxStatus("✓ NO CHANGES", "32")
	} else {
		p.boxStatus("◆ CHANGED", "33")
	}
	p.boxBlank()
	p.boxRow("Added", fmt.Sprint(added))
	p.boxRow("Removed", fmt.Sprint(removed))
	p.boxBottom()
}

func (p *Printer) section(title string) {
	fmt.Fprintf(p.out, "\n  %s\n\n", p.c("1;90", title))
}

func (p *Printer) boxTop(title string) {
	prefix := "╭─ " + title + " "
	remaining := boxWidth - runeLen(prefix) - 1
	if remaining < 1 {
		remaining = 1
	}
	fmt.Fprintf(p.out, "%s%s╮\n", prefix, strings.Repeat("─", remaining))
}

func (p *Printer) boxBottom() {
	fmt.Fprintf(p.out, "╰%s╯\n", strings.Repeat("─", boxWidth-2))
}

func (p *Printer) boxBlank() {
	fmt.Fprintf(p.out, "│%s│\n", strings.Repeat(" ", boxWidth-2))
}

func (p *Printer) boxStatus(status, color string) {
	plain := "  " + status
	line := padRight(plain, boxWidth-2)
	if p.color {
		coloredStatus := p.c(color+";1", status)
		line = "  " + coloredStatus + strings.Repeat(" ", boxWidth-2-runeLen(plain))
	}
	fmt.Fprintf(p.out, "│%s│\n", line)
}

func (p *Printer) boxRow(label, value string) {
	labelWidth := 13
	maxValue := boxWidth - 2 - 2 - labelWidth - 1
	value = truncate(value, maxValue)
	plain := "  " + padRight(label, labelWidth) + " " + value
	fmt.Fprintf(p.out, "│%s│\n", padRight(plain, boxWidth-2))
}

func padRight(value string, width int) string {
	length := runeLen(value)
	if length >= width {
		return value
	}
	return value + strings.Repeat(" ", width-length)
}

func truncate(value string, max int) string {
	if max <= 0 {
		return ""
	}
	if runeLen(value) <= max {
		return value
	}
	if max == 1 {
		return "…"
	}
	runes := []rune(value)
	return string(runes[:max-1]) + "…"
}

func runeLen(value string) int {
	return utf8.RuneCountInString(value)
}

func prettyDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(100 * time.Millisecond).String()
}
