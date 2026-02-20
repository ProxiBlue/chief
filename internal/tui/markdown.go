package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

// progressStyle is a customized dark style with no document margin,
// so rendered markdown fits flush within our panel padding.
var progressStyle ansi.StyleConfig

func init() {
	progressStyle = styles.DarkStyleConfig
	zero := uint(0)
	progressStyle.Document.Margin = &zero
	progressStyle.Document.StylePrimitive.BlockPrefix = ""
	progressStyle.Document.StylePrimitive.BlockSuffix = ""
}

// renderGlamour renders a markdown string as styled terminal output.
func renderGlamour(markdown string, width int) string {
	if width <= 0 || strings.TrimSpace(markdown) == "" {
		return ""
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(progressStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return markdown
	}

	rendered, err := r.Render(markdown)
	if err != nil {
		return markdown
	}

	// Trim leading/trailing blank lines that glamour adds
	return strings.TrimSpace(rendered)
}

// ansiStripRegex matches ANSI escape codes for stripping in tests.
var ansiStripRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape codes from a string. Exported for tests.
func stripANSI(s string) string {
	return ansiStripRegex.ReplaceAllString(s, "")
}
