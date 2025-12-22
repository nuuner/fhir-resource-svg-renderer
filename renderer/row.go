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
	textY := y + config.HeaderHeight/2 + TitleVerticalOffset
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
`, x+HeaderTextMarginY, textY, escapeXML(h.name)))
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

	sb.WriteString(renderRowBackground(row, y, totalWidth, config))
	sb.WriteString(renderRowBorder(y, row.RowHeight, totalWidth, config))

	x := config.Padding
	baseTextY := y + RowTopMargin + config.FontSize
	firstLineCenterY := y + RowTopMargin + config.FontSize/2

	sb.WriteString(renderTreeAndIcon(row, x, y, firstLineCenterY, config))
	sb.WriteString(renderNameColumn(row, x, baseTextY, config))

	x += config.NameColWidth
	sb.WriteString(renderColumnSeparator(x, y, row.RowHeight, config))

	sb.WriteString(renderFlagsColumn(row, x, y, config))
	x += config.FlagsColWidth
	sb.WriteString(renderColumnSeparator(x, y, row.RowHeight, config))

	sb.WriteString(renderCardinalityColumn(row, x, y, config))
	x += config.CardinalityColWidth
	sb.WriteString(renderColumnSeparator(x, y, row.RowHeight, config))

	sb.WriteString(renderTypeColumn(row, x, baseTextY, config))
	x += config.TypeColWidth
	sb.WriteString(renderColumnSeparator(x, y, row.RowHeight, config))

	sb.WriteString(renderDescriptionColumn(row, x, baseTextY, config))

	return sb.String()
}

// renderRowBackground renders the background rectangle for a row
func renderRowBackground(row RowData, y, totalWidth float64, config SVGConfig) string {
	bgColor := config.RowBgColor
	if row.IsAlt {
		bgColor = config.AltRowBgColor
	}
	return fmt.Sprintf(`<rect x="0" y="%.0f" width="%.0f" height="%.0f" fill="%s"/>
`,
		y, totalWidth, row.RowHeight, bgColor)
}

// renderRowBorder renders the bottom border of a row
func renderRowBorder(y, rowHeight, totalWidth float64, config SVGConfig) string {
	return fmt.Sprintf(`<line x1="0" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="%.1f"/>
`,
		y+rowHeight, totalWidth, y+rowHeight, config.BorderColor, BorderStrokeWidth)
}

// renderColumnSeparator renders a vertical column separator line
func renderColumnSeparator(x, y, rowHeight float64, config SVGConfig) string {
	return fmt.Sprintf(`<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s"/>
`,
		x, y, x, y+rowHeight, config.BorderColor)
}

// renderTreeAndIcon renders tree lines and the element icon
func renderTreeAndIcon(row RowData, x, y, firstLineCenterY float64, config SVGConfig) string {
	var sb strings.Builder
	fe := row.Element

	// Tree lines
	treeLines := RenderTreeLines(x, y, row.RowHeight, firstLineCenterY, fe.Depth, fe.ParentLasts, fe.IsLast, config.TreeStyle)
	sb.WriteString(treeLines)

	// Icon
	iconX := x + float64(fe.Depth)*config.TreeStyle.IndentPx
	iconY := firstLineCenterY - config.IconSize/2
	hasChildren := len(fe.Element.Elements) > 0
	iconType := GetIconTypeForElement(fe.Element.Type, row.IsRoot, hasChildren)
	sb.WriteString(RenderIcon(iconType, iconX, iconY, config.IconSize))

	return sb.String()
}

// renderNameColumn renders the name column with multi-line support
func renderNameColumn(row RowData, x, baseTextY float64, config SVGConfig) string {
	var sb strings.Builder
	fe := row.Element

	nameX := x + float64(fe.Depth)*config.TreeStyle.IndentPx + config.IconSize + IconTextGap
	textClass := "link-text"
	if fe.Element.Usage == "not-used" {
		textClass = "not-used"
	}

	sb.WriteString(`<g clip-path="url(#clip-name)">
`)
	for i, line := range row.NameLines {
		lineY := baseTextY + float64(i)*config.LineHeight
		sb.WriteString(fmt.Sprintf(`<text x="%.0f" y="%.0f" class="%s">%s</text>
`,
			nameX, lineY, textClass, escapeXML(line)))
	}
	sb.WriteString("</g>\n")

	return sb.String()
}

// renderFlagsColumn renders the flags column
func renderFlagsColumn(row RowData, x, y float64, config SVGConfig) string {
	flagsStr := renderFlags(row.Element.Element.Flags, config)
	flagsY := y + row.RowHeight/2
	return fmt.Sprintf(`<g clip-path="url(#clip-flags)" transform="translate(%.0f, %.0f)">%s</g>
`, x+config.Padding, flagsY, flagsStr)
}

// renderCardinalityColumn renders the cardinality column
func renderCardinalityColumn(row RowData, x, y float64, config SVGConfig) string {
	cardY := y + row.RowHeight/2 + TextVerticalOffset
	return fmt.Sprintf(`<g clip-path="url(#clip-card)"><text x="%.0f" y="%.0f" class="cell-text">%s</text></g>
`,
		x+config.Padding, cardY, escapeXML(row.Element.Element.Cardinality))
}

// renderTypeColumn renders the type column with multi-line and link support
func renderTypeColumn(row RowData, x, baseTextY float64, config SVGConfig) string {
	var sb strings.Builder
	fe := row.Element

	sb.WriteString(`<g clip-path="url(#clip-type)">
`)
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

	return sb.String()
}

// renderDescriptionColumn renders the description column with multi-line support
func renderDescriptionColumn(row RowData, x, baseTextY float64, config SVGConfig) string {
	var sb strings.Builder
	fe := row.Element

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
