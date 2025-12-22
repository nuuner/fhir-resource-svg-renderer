package renderer

import (
	"fmt"
	"strings"
)

// TreeLineStyle contains styling parameters for tree lines
type TreeLineStyle struct {
	Color    string
	Width    float64
	IndentPx float64 // Pixels per indent level
}

// DefaultTreeStyle returns the default tree line styling
func DefaultTreeStyle() TreeLineStyle {
	return TreeLineStyle{
		Color:    "#CCCCCC",
		Width:    1.0,
		IndentPx: 20.0,
	}
}

// RenderTreeLines generates SVG for tree structure lines
// parentLasts indicates whether each ancestor was the last child at its level
// isLast indicates whether this element is the last child at its level
// depth is the current nesting depth (0 = root, 1 = first level children, etc.)
// firstLineY is the Y position of the first line of text (for horizontal connector alignment)
func RenderTreeLines(x, y, rowHeight, firstLineY float64, depth int, parentLasts []bool, isLast bool, style TreeLineStyle) string {
	if depth == 0 {
		return "" // No tree lines for root
	}

	var sb strings.Builder

	// Draw vertical continuation lines for ancestors that weren't last
	for i := 0; i < depth-1; i++ {
		if i < len(parentLasts) && !parentLasts[i] {
			lineX := x + float64(i)*style.IndentPx + style.IndentPx/2
			sb.WriteString(fmt.Sprintf(
				`<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="%s" stroke-width="%f"/>`,
				lineX, y, lineX, y+rowHeight, style.Color, style.Width))
		}
	}

	// Draw the connector for current element
	connectorX := x + float64(depth-1)*style.IndentPx + style.IndentPx/2

	if isLast {
		// L-shaped connector (└──)
		// Vertical part (from top to first line position)
		sb.WriteString(fmt.Sprintf(
			`<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="%s" stroke-width="%f"/>`,
			connectorX, y, connectorX, firstLineY, style.Color, style.Width))
	} else {
		// T-shaped connector (├──)
		// Vertical part (full height to continue for siblings)
		sb.WriteString(fmt.Sprintf(
			`<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="%s" stroke-width="%f"/>`,
			connectorX, y, connectorX, y+rowHeight, style.Color, style.Width))
	}

	// Horizontal part (from connector to icon) - aligned with first line
	horizontalEndX := x + float64(depth)*style.IndentPx - TreeHorizontalGap
	sb.WriteString(fmt.Sprintf(
		`<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="%s" stroke-width="%f"/>`,
		connectorX, firstLineY, horizontalEndX, firstLineY, style.Color, style.Width))

	return sb.String()
}
