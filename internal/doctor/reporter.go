package doctor

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// RenderReport prints the diagnostic report to the terminal.
func RenderReport(report *Report) {
	cyan := color.New(color.FgCyan, color.Bold)
	white := color.New(color.FgWhite, color.Bold)
	dim := color.New(color.FgHiBlack)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	red := color.New(color.FgRed, color.Bold)

	sep := "  ══════════════════════════════════════════════════════"

	fmt.Println()
	cyan.Println("  🏥 PORTMAN DOCTOR — System Health Report")
	dim.Println(sep)

	// Score
	fmt.Println()
	scoreColor := green
	scoreEmoji := "🟢"
	scoreLabel := "HEALTHY"
	if report.Score < 80 {
		scoreColor = yellow
		scoreEmoji = "🟡"
		scoreLabel = "NEEDS ATTENTION"
	}
	if report.Score < 50 {
		scoreColor = red
		scoreEmoji = "🔴"
		scoreLabel = "UNHEALTHY"
	}

	white.Printf("  Score: ")
	scoreColor.Printf("%d/100  %s %s\n", report.Score, scoreEmoji, scoreLabel)

	// Port overview
	dim.Printf("  Scanning %d port(s) across %d process(es)\n", report.PortCount, report.ProcessCount)
	fmt.Println()

	if len(report.Findings) == 0 {
		green.Println("  ✓ No issues found! Your port configuration looks great.")
		fmt.Println()
		return
	}

	// Group findings by severity
	var errors, warnings, infos []Finding
	for _, f := range report.Findings {
		switch f.Severity {
		case SeverityError:
			errors = append(errors, f)
		case SeverityWarning:
			warnings = append(warnings, f)
		case SeverityInfo:
			infos = append(infos, f)
		}
	}

	// Render errors
	if len(errors) > 0 {
		red.Printf("  ✗ %d Error(s)\n\n", len(errors))
		for _, f := range errors {
			renderFinding(f, red, dim)
		}
		fmt.Println()
	}

	// Render warnings
	if len(warnings) > 0 {
		yellow.Printf("  ⚠ %d Warning(s)\n\n", len(warnings))
		for _, f := range warnings {
			renderFinding(f, yellow, dim)
		}
		fmt.Println()
	}

	// Render info
	if len(infos) > 0 {
		infoCyan := color.New(color.FgCyan)
		infoCyan.Printf("  ℹ %d Info\n\n", len(infos))
		for _, f := range infos {
			renderFinding(f, infoCyan, dim)
		}
		fmt.Println()
	}

	// Fixable summary
	fixable := GetFixablePorts(report.Findings)
	if len(fixable) > 0 {
		dim.Println(sep)
		yellow.Printf("  💡 %d issue(s) can be auto-fixed.\n", len(fixable))
		dim.Println("  Run: portman doctor --fix")
		fmt.Println()
	}
}

func renderFinding(f Finding, severity *color.Color, dim *color.Color) {
	icon := "•"
	if f.Fixable {
		icon = "🔧"
	}

	severity.Printf("    %s %s\n", icon, f.Title)

	// Wrap long descriptions
	desc := f.Description
	if len(desc) > 70 {
		words := strings.Fields(desc)
		var lines []string
		line := ""
		for _, w := range words {
			if len(line)+len(w)+1 > 68 {
				lines = append(lines, line)
				line = w
			} else {
				if line != "" {
					line += " "
				}
				line += w
			}
		}
		if line != "" {
			lines = append(lines, line)
		}
		for _, l := range lines {
			dim.Printf("      %s\n", l)
		}
	} else {
		dim.Printf("      %s\n", desc)
	}

	if f.Fixable && f.FixAction != "" {
		fix := color.New(color.FgGreen)
		fix.Printf("      → %s\n", f.FixAction)
	}
}
