package renderer

import (
	"fmt"
	"strings"
)

func renderFlags(flags []string, config SVGConfig) string {
	if len(flags) == 0 {
		return ""
	}

	var sb strings.Builder
	x := 0.0

	for _, flag := range flags {
		displayFlag := flag
		needsBox := false

		switch flag {
		case "S":
			displayFlag = "\u03A3"
		case "?!":
			displayFlag = "?!\u03A3"
		case "I":
			displayFlag = "I"
		case "TU", "N":
			needsBox = true
		}

		if needsBox {
			boxWidth := float64(len(displayFlag))*FlagCharWidth + FlagBoxPadding
			sb.WriteString(fmt.Sprintf(`<rect x="%.0f" y="-8" width="%.0f" height="14" fill="none" stroke="%s" rx="2"/>`,
				x, boxWidth, config.BorderColor))
			sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="2" class="flag-box">%s</text>`,
				x+FlagBoxTextOffset, escapeXML(displayFlag)))
			x += boxWidth + FlagGap
		} else {
			sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="2" class="flag-box">%s</text>`,
				x, escapeXML(displayFlag)))
			x += float64(len(displayFlag))*FlagCharWidth + FlagGap
		}
	}

	return sb.String()
}
