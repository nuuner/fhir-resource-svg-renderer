package renderer

import (
	"fmt"
	"strings"

	"fhir_renderer/models"
)

// Render generates SVG for a resource definition
func Render(resource *models.ResourceDefinition, config SVGConfig) string {
	// Initialize text measurer
	tm, err := NewTextMeasurer(config.FontSize)
	if err != nil {
		return renderFallback()
	}
	defer tm.Close()
	config.textMeasurer = tm

	// Flatten the resource hierarchy
	flatElements := resource.Flatten()

	// Calculate name column width based on actual content
	maxNameWidth := tm.MeasureString(resource.Name)
	for _, fe := range flatElements {
		indentWidth := float64(fe.Depth) * config.TreeStyle.IndentPx
		nameWidth := indentWidth + config.IconSize + 12 + tm.MeasureString(fe.Element.Name)
		if nameWidth > maxNameWidth {
			maxNameWidth = nameWidth
		}
	}
	config.NameColWidth = maxNameWidth + config.Padding*2
	if config.NameColWidth < MinNameColWidth {
		config.NameColWidth = MinNameColWidth
	}
	if config.NameColWidth > MaxNameColWidth {
		config.NameColWidth = MaxNameColWidth
	}

	// Pre-calculate row data with text wrapping
	rows := make([]RowData, len(flatElements))
	totalContentHeight := 0.0

	for i, fe := range flatElements {
		row := RowData{
			Element: fe,
			IsRoot:  i == 0,
			IsAlt:   i%2 == 1,
		}

		// Calculate available widths for each column (with buffer for font rendering differences)
		nameIndent := float64(fe.Depth)*config.TreeStyle.IndentPx + config.IconSize + 8
		availableNameWidth := config.NameColWidth - nameIndent - config.Padding - 15
		availableTypeWidth := config.TypeColWidth - config.Padding*2 - 15
		availableDescWidth := config.DescriptionColWidth - config.Padding*2 - 15

		// Wrap text for each column
		row.NameLines = []string{fe.Element.Name} // Name usually fits, just use single line
		if tm.MeasureString(fe.Element.Name) > availableNameWidth {
			row.NameLines = tm.WrapText(fe.Element.Name, availableNameWidth)
		}

		row.TypeLines = tm.WrapText(fe.Element.Type, availableTypeWidth)

		// Build description text
		descText := fe.Element.Description
		isBold := false
		if fe.Element.Usage == "not-used" {
			if descText == "" {
				descText = UnusedElementLabel
			}
		} else if fe.Element.Usage == "todo" {
			isBold = true
			if !strings.HasPrefix(descText, "TODO") {
				descText = "TODO: " + descText
			}
		}
		if fe.Element.Notes != "" && fe.Element.Usage != "not-used" {
			if descText != "" {
				descText += " - "
			}
			descText += fe.Element.Notes
		}
		// Bold text is ~10% wider, reduce available width accordingly
		descWidth := availableDescWidth
		if isBold {
			descWidth = availableDescWidth * 0.90
		}
		row.DescLines = tm.WrapText(descText, descWidth)

		// Calculate row height based on max lines
		maxLines := len(row.NameLines)
		if len(row.TypeLines) > maxLines {
			maxLines = len(row.TypeLines)
		}
		if len(row.DescLines) > maxLines {
			maxLines = len(row.DescLines)
		}

		// Row height: top margin + content + bottom margin
		row.RowHeight = RowTopMargin + float64(maxLines)*config.LineHeight + RowBottomMargin
		if row.RowHeight < config.MinRowHeight {
			row.RowHeight = config.MinRowHeight
		}

		rows[i] = row
		totalContentHeight += row.RowHeight
	}

	// Calculate total dimensions
	totalWidth := config.NameColWidth + config.FlagsColWidth + config.CardinalityColWidth +
		config.TypeColWidth + config.DescriptionColWidth
	totalHeight := config.TitleHeight + config.HeaderHeight + totalContentHeight + 2

	// Build SVG
	var sb strings.Builder

	// SVG header
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
     width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f">
<defs>
    <style>
        .header-text { font-family: %s; font-size: %.0fpx; font-weight: bold; fill: %s; }
        .cell-text { font-family: %s; font-size: %.0fpx; fill: %s; }
        .link-text { font-family: %s; font-size: %.0fpx; fill: %s; cursor: pointer; }
        .not-used { font-family: %s; font-size: %.0fpx; fill: %s; font-style: italic; }
        .todo { font-family: %s; font-size: %.0fpx; fill: %s; font-weight: bold; }
        .flag-box { font-family: %s; font-size: 10px; fill: %s; }
        .title-text { font-family: %s; font-size: 14px; font-weight: bold; fill: %s; }
    </style>
`,
		totalWidth, totalHeight, totalWidth, totalHeight,
		config.FontFamily, config.HeaderFontSize, config.HeaderTextColor,
		config.FontFamily, config.FontSize, config.TextColor,
		config.FontFamily, config.FontSize, config.LinkColor,
		config.FontFamily, config.FontSize, config.NotUsedColor,
		config.FontFamily, config.FontSize, config.TodoColor,
		config.FontFamily, config.TextColor,
		config.FontFamily, config.HeaderTextColor))

	// Add clipPath definitions for each column
	colStarts := []float64{
		0,
		config.NameColWidth,
		config.NameColWidth + config.FlagsColWidth,
		config.NameColWidth + config.FlagsColWidth + config.CardinalityColWidth,
		config.NameColWidth + config.FlagsColWidth + config.CardinalityColWidth + config.TypeColWidth,
	}
	colWidths := []float64{
		config.NameColWidth,
		config.FlagsColWidth,
		config.CardinalityColWidth,
		config.TypeColWidth,
		config.DescriptionColWidth,
	}
	colNames := []string{"name", "flags", "card", "type", "desc"}

	for i, name := range colNames {
		sb.WriteString(fmt.Sprintf(`    <clipPath id="clip-%s"><rect x="%.0f" y="0" width="%.0f" height="%.0f"/></clipPath>
`,
			name, colStarts[i], colWidths[i], totalHeight))
	}
	sb.WriteString("</defs>\n")

	// Title bar
	sb.WriteString(fmt.Sprintf(`<rect x="0" y="0" width="%.0f" height="%.0f" fill="%s" stroke="%s"/>
<text x="%.0f" y="%.0f" class="title-text">Structure</text>
`,
		totalWidth, config.TitleHeight, config.HeaderBgColor, config.BorderColor,
		config.Padding, config.TitleHeight/2+5))

	// Column headers
	headerY := config.TitleHeight
	sb.WriteString(renderHeaderRow(config, headerY, totalWidth))

	// Data rows with variable heights
	currentY := headerY + config.HeaderHeight
	for _, row := range rows {
		sb.WriteString(renderDataRowWrapped(row, config, currentY, totalWidth))
		currentY += row.RowHeight
	}

	sb.WriteString("</svg>")
	return sb.String()
}

func renderFallback() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="400" height="100">
<text x="10" y="50" font-family="Arial" font-size="14">Error: Could not load font for text measurement</text>
</svg>`
}
