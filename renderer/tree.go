package renderer

import (
	"fmt"
	"strings"
)

// TreeLineStyle contains styling parameters for tree lines
type TreeLineStyle struct {
	Color     string
	Width     float64
	IndentPx  float64 // Pixels per indent level
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
	horizontalEndX := x + float64(depth)*style.IndentPx - 2
	sb.WriteString(fmt.Sprintf(
		`<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="%s" stroke-width="%f"/>`,
		connectorX, firstLineY, horizontalEndX, firstLineY, style.Color, style.Width))

	return sb.String()
}

// RenderTreeLinesForRows generates all tree lines for a list of rows
// This function draws all vertical lines at once for better visual consistency
// Note: This is a simplified version that uses row center for firstLineY
func RenderTreeLinesForRows(rows []TreeRow, startX, startY, rowHeight float64, style TreeLineStyle) string {
	if len(rows) == 0 {
		return ""
	}

	var sb strings.Builder

	for i, row := range rows {
		y := startY + float64(i)*rowHeight
		firstLineY := y + rowHeight/2 // Use row center as approximation
		lines := RenderTreeLines(startX, y, rowHeight, firstLineY, row.Depth, row.ParentLasts, row.IsLast, style)
		sb.WriteString(lines)
	}

	return sb.String()
}

// TreeRow contains the information needed to render tree lines for a single row
type TreeRow struct {
	Depth       int
	ParentLasts []bool
	IsLast      bool
}

// CalculateNameColumnWidth calculates the width needed for the Name column
// based on maximum depth and element name lengths
func CalculateNameColumnWidth(maxDepth int, maxNameLen int, style TreeLineStyle, iconSize, charWidth float64) float64 {
	indentWidth := float64(maxDepth) * style.IndentPx
	iconWidth := iconSize + 8 // icon + padding
	textWidth := float64(maxNameLen) * charWidth
	padding := 20.0 // left and right padding

	return indentWidth + iconWidth + textWidth + padding
}
