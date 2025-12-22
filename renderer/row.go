package renderer

import (
	"fmt"
	"strings"

	"fhir_renderer/models"
)

// RowData contains pre-calculated data for a row including wrapped text
type RowData struct {
	Element   models.FlatElement
	NameLines []string
	TypeLines []string
	DescLines []string
	RowHeight float64
	IsRoot    bool
	IsAlt     bool
}

func renderHeaderRow(config SVGConfig, y, totalWidth float64) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<rect x="0" y="%.0f" width="%.0f" height="%.0f" fill="%s" stroke="%s"/>
`,
		y, totalWidth, config.HeaderHeight, config.HeaderBgColor, config.BorderColor))

	x := config.Padding
	textY := y + config.HeaderHeight/2 + 5
	headerTextMargin := 6.0 // Extra left margin for header text
	headers := []struct {
		name  string
		width float64
	}{
		{"Name", config.NameColWidth},
		{"Flags", config.FlagsColWidth},
		{"Card.", config.CardinalityColWidth},
		{"Type", config.TypeColWidth},
		{"Description & Constraints", config.DescriptionColWidth},
	}

	for i, h := range headers {
		sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="header-text">%s</text>
`, x+headerTextMargin, textY, escapeXML(h.name)))
		x += h.width
		if i < len(headers)-1 {
			sb.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s"/>
`, x, y, x, y+config.HeaderHeight, config.BorderColor))
		}
	}

	return sb.String()
}

func renderDataRowWrapped(row RowData, config SVGConfig, y, totalWidth float64) string {
	var sb strings.Builder
	fe := row.Element

	// Background
	bgColor := config.RowBgColor
	if row.IsAlt {
		bgColor = config.AltRowBgColor
	}
	sb.WriteString(fmt.Sprintf(`<rect x="0" y="%.0f" width="%.0f" height="%.0f" fill="%s"/>
`,
		y, totalWidth, row.RowHeight, bgColor))

	// Bottom border
	sb.WriteString(fmt.Sprintf(`<line x1="0" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="0.5"/>
`,
		y+row.RowHeight, totalWidth, y+row.RowHeight, config.BorderColor))

	x := config.Padding
	baseTextY := y + RowTopMargin + config.FontSize          // First line baseline
	firstLineCenterY := y + RowTopMargin + config.FontSize/2 // Vertical center of first line

	// Tree lines - aligned with first line
	treeLines := RenderTreeLines(x, y, row.RowHeight, firstLineCenterY, fe.Depth, fe.ParentLasts, fe.IsLast, config.TreeStyle)
	sb.WriteString(treeLines)

	// Icon - aligned with first line of text
	iconX := x + float64(fe.Depth)*config.TreeStyle.IndentPx
	iconY := firstLineCenterY - config.IconSize/2 // Center icon on first line
	hasChildren := len(fe.Element.Elements) > 0
	iconType := GetIconTypeForElement(fe.Element.Type, row.IsRoot, hasChildren)
	sb.WriteString(RenderIcon(iconType, iconX, iconY, config.IconSize))

	// Name column - multi-line support
	nameX := iconX + config.IconSize + 4
	textClass := "link-text"
	if fe.Element.Usage == "not-used" {
		textClass = "not-used"
	}
	sb.WriteString(fmt.Sprintf(`<g clip-path="url(#clip-name)">
`))
	for i, line := range row.NameLines {
		lineY := baseTextY + float64(i)*config.LineHeight
		sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="%s">%s</text>
`,
			nameX, lineY, textClass, escapeXML(line)))
	}
	sb.WriteString("</g>\n")

	x += config.NameColWidth

	// Vertical line
	sb.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s"/>
`,
		x, y, x, y+row.RowHeight, config.BorderColor))

	// Flags
	flagsStr := renderFlags(fe.Element.Flags, config)
	flagsY := y + row.RowHeight/2
	sb.WriteString(fmt.Sprintf(`<g clip-path="url(#clip-flags)" transform="translate(%.0f, %.0f)">%s</g>
`, x+config.Padding, flagsY, flagsStr))
	x += config.FlagsColWidth

	sb.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s"/>
`,
		x, y, x, y+row.RowHeight, config.BorderColor))

	// Cardinality
	cardY := y + row.RowHeight/2 + 4
	sb.WriteString(fmt.Sprintf(`<g clip-path="url(#clip-card)"><text x="%.0f" y="%.0f" class="cell-text">%s</text></g>
`,
		x+config.Padding, cardY, escapeXML(fe.Element.Cardinality)))
	x += config.CardinalityColWidth

	sb.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s"/>
`,
		x, y, x, y+row.RowHeight, config.BorderColor))

	// Type column - multi-line support
	sb.WriteString(fmt.Sprintf(`<g clip-path="url(#clip-type)">
`))
	for i, line := range row.TypeLines {
		lineY := baseTextY + float64(i)*config.LineHeight
		if fe.Element.TypeRef != "" && i == 0 {
			sb.WriteString(fmt.Sprintf(`<a xlink:href="%s" target="_blank"><text x="%.0f" y="%.0f" class="link-text">%s</text></a>
`,
				escapeXML(fe.Element.TypeRef), x+config.Padding, lineY, escapeXML(line)))
		} else {
			sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="link-text">%s</text>
`,
				x+config.Padding, lineY, escapeXML(line)))
		}
	}
	sb.WriteString("</g>\n")
	x += config.TypeColWidth

	sb.WriteString(fmt.Sprintf(`<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s"/>
`,
		x, y, x, y+row.RowHeight, config.BorderColor))

	// Description column - multi-line support (no clipPath needed for last column)
	descClass := "cell-text"
	if fe.Element.Usage == "not-used" {
		descClass = "not-used"
	} else if fe.Element.Usage == "todo" {
		descClass = "todo"
	}
	for i, line := range row.DescLines {
		lineY := baseTextY + float64(i)*config.LineHeight
		sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="%s">%s</text>
`,
			x+config.Padding, lineY, descClass, escapeXML(line)))
	}

	return sb.String()
}
