package log

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const bannerWidth = 76

// Banner prints a bordered box with title and body lines to stderr.
// Long lines are truncated or wrapped to bannerWidth to keep layout consistent.
func Banner(title string, lines []string) {
	if err := BannerTo(os.Stderr, title, lines); err != nil {
		Error("banner write failed", "error", err)
	}
}

// BannerTo writes a bordered box with title and body lines to w.
// Long lines are truncated or wrapped to bannerWidth to keep layout consistent.
func BannerTo(w io.Writer, title string, lines []string) error {
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
	if _, err := fmt.Fprintln(w, top); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "│"+strings.Repeat(" ", lineLen)+"│"); err != nil {
		return err
	}
	innerWidth := bannerWidth - 4
	for _, line := range lines {
		for _, part := range wrap(line, innerWidth) {
			if _, err := fmt.Fprintln(w, "│ "+pad(part, innerWidth)+" │"); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(w, "│"+strings.Repeat(" ", lineLen)+"│"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "╰"+strings.Repeat("─", lineLen)+"╯"); err != nil {
		return err
	}
	return nil
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
