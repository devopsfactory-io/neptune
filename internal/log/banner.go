package log

import (
	"fmt"
	"os"
	"strings"
)

const bannerWidth = 76

// Banner prints a bordered box with title and body lines to stderr.
// Long lines are truncated or wrapped to bannerWidth to keep layout consistent.
func Banner(title string, lines []string) {
	pad := func(s string, width int) string {
		if len(s) > width {
			return s[:width-3] + "..."
		}
		return s + strings.Repeat(" ", width-len(s))
	}
	lineLen := bannerWidth - 2
	titleLine := " " + title + " "
	if len(titleLine) > lineLen {
		titleLine = " " + title[:lineLen-5] + "... "
	}
	dashCount := lineLen - len(titleLine)
	if dashCount < 0 {
		dashCount = 0
	}
	leftDash := dashCount / 2
	rightDash := dashCount - leftDash
	top := "╭" + strings.Repeat("─", leftDash) + titleLine + strings.Repeat("─", rightDash) + "╮"
	fmt.Fprintln(os.Stderr, top)
	fmt.Fprintln(os.Stderr, "│"+strings.Repeat(" ", lineLen)+"│")
	innerWidth := bannerWidth - 4
	for _, line := range lines {
		for _, part := range wrap(line, innerWidth) {
			fmt.Fprintln(os.Stderr, "│ "+pad(part, innerWidth)+" │")
		}
	}
	fmt.Fprintln(os.Stderr, "│"+strings.Repeat(" ", lineLen)+"│")
	fmt.Fprintln(os.Stderr, "╰"+strings.Repeat("─", lineLen)+"╯")
}

func wrap(s string, width int) []string {
	if width <= 0 || len(s) <= width {
		if s != "" {
			return []string{s}
		}
		return nil
	}
	var out []string
	for len(s) > width {
		out = append(out, s[:width])
		s = strings.TrimLeft(s[width:], " ")
	}
	if s != "" {
		out = append(out, s)
	}
	return out
}
