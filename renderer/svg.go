package renderer

import (
	"fmt"
	"strings"

	"fhir_renderer/models"
)

// ColumnWidths holds the calculated widths for each column
type ColumnWidths struct {
	Name        float64
	Flags       float64
	Cardinality float64
	Type        float64
	Description float64
}

// Total returns the sum of all column widths
func (cw ColumnWidths) Total() float64 {
	return cw.Name + cw.Flags + cw.Cardinality + cw.Type + cw.Description
}

// Render generates SVG for a resource definition
func Render(resource *models.ResourceDefinition, config SVGConfig) string {
	tm, err := NewTextMeasurer(config.FontSize)
	if err != nil {
		return renderFallback()
	}
	defer tm.Close()
	config.textMeasurer = tm

	config.NameColWidth = calculateNameColumnWidth(resource, tm, config)
	rows := prepareRows(resource.Flatten(), tm, config)
	colWidths := ColumnWidths{
		Name:        config.NameColWidth,
		Flags:       config.FlagsColWidth,
		Cardinality: config.CardinalityColWidth,
		Type:        config.TypeColWidth,
		Description: config.DescriptionColWidth,
	}

	totalHeight := calculateTotalHeight(rows, config)
	return buildSVG(rows, colWidths, totalHeight, config)
}

// calculateNameColumnWidth determines the optimal name column width based on content
func calculateNameColumnWidth(resource *models.ResourceDefinition, tm *TextMeasurer, config SVGConfig) float64 {
	flatElements := resource.Flatten()
	maxNameWidth := tm.MeasureString(resource.Name)

	for _, fe := range flatElements {
		indentWidth := float64(fe.Depth) * config.TreeStyle.IndentPx
		nameWidth := indentWidth + config.IconSize + IconSpaceInMeasurement + tm.MeasureString(fe.Element.Name)
		if nameWidth > maxNameWidth {
			maxNameWidth = nameWidth
		}
	}

	width := maxNameWidth + config.Padding*2
	if width < MinNameColWidth {
		width = MinNameColWidth
	}
	if width > MaxNameColWidth {
		width = MaxNameColWidth
	}
	return width
}

// prepareRows creates RowData for each flattened element with text wrapping
func prepareRows(flatElements []models.FlatElement, tm *TextMeasurer, config SVGConfig) []RowData {
	rows := make([]RowData, len(flatElements))

	for i, fe := range flatElements {
		rows[i] = prepareRow(fe, i, tm, config)
	}

	return rows
}

// prepareRow creates a single RowData with wrapped text and calculated height
func prepareRow(fe models.FlatElement, index int, tm *TextMeasurer, config SVGConfig) RowData {
	row := RowData{
		Element: fe,
		IsRoot:  index == 0,
		IsAlt:   index%2 == 1,
	}

	// Calculate available widths for each column
	nameIndent := float64(fe.Depth)*config.TreeStyle.IndentPx + config.IconSize + IconPaddingRight
	availableNameWidth := config.NameColWidth - nameIndent - config.Padding - FontRenderingBuffer
	availableTypeWidth := config.TypeColWidth - config.Padding*2 - FontRenderingBuffer
	availableDescWidth := config.DescriptionColWidth - config.Padding*2 - FontRenderingBuffer

	// Wrap name text
	row.NameLines = []string{fe.Element.Name}
	if tm.MeasureString(fe.Element.Name) > availableNameWidth {
		row.NameLines = tm.WrapText(fe.Element.Name, availableNameWidth)
	}

	// Wrap type text
	row.TypeLines = tm.WrapText(fe.Element.Type, availableTypeWidth)

	// Build and wrap description text
	descText, isBold := buildDescriptionText(fe)
	descWidth := availableDescWidth
	if isBold {
		descWidth = availableDescWidth * BoldTextWidthFactor
	}
	row.DescLines = tm.WrapText(descText, descWidth)

	// Calculate row height
	row.RowHeight = calculateRowHeight(row, config)

	return row
}

// buildDescriptionText constructs the description text and returns whether it should be bold
func buildDescriptionText(fe models.FlatElement) (string, bool) {
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

	return descText, isBold
}

// calculateRowHeight determines the height of a row based on its content
func calculateRowHeight(row RowData, config SVGConfig) float64 {
	maxLines := len(row.NameLines)
	if len(row.TypeLines) > maxLines {
		maxLines = len(row.TypeLines)
	}
	if len(row.DescLines) > maxLines {
		maxLines = len(row.DescLines)
	}

	height := RowTopMargin + float64(maxLines)*config.LineHeight + RowBottomMargin
	if height < config.MinRowHeight {
		height = config.MinRowHeight
	}
	return height
}

// calculateTotalHeight computes the total SVG height
func calculateTotalHeight(rows []RowData, config SVGConfig) float64 {
	contentHeight := 0.0
	for _, row := range rows {
		contentHeight += row.RowHeight
	}
	return config.TitleHeight + config.HeaderHeight + contentHeight + SVGHeightPadding
}

// buildSVG constructs the complete SVG string
func buildSVG(rows []RowData, colWidths ColumnWidths, totalHeight float64, config SVGConfig) string {
	var sb strings.Builder
	totalWidth := colWidths.Total()

	sb.WriteString(buildSVGHeader(totalWidth, totalHeight, config))
	sb.WriteString(buildClipPaths(colWidths, totalHeight, config))
	sb.WriteString("</defs>\n")
	sb.WriteString(buildTitleBar(totalWidth, config))
	sb.WriteString(renderHeaderRow(config, config.TitleHeight, totalWidth))
	sb.WriteString(buildDataRows(rows, totalWidth, config))
	sb.WriteString("</svg>")

	return sb.String()
}

// buildSVGHeader creates the SVG header with styles
func buildSVGHeader(totalWidth, totalHeight float64, config SVGConfig) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
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
		config.FontFamily, config.HeaderTextColor)
}

// buildClipPaths creates clip path definitions for each column
func buildClipPaths(colWidths ColumnWidths, totalHeight float64, config SVGConfig) string {
	var sb strings.Builder

	colStarts := []float64{
		0,
		colWidths.Name,
		colWidths.Name + colWidths.Flags,
		colWidths.Name + colWidths.Flags + colWidths.Cardinality,
		colWidths.Name + colWidths.Flags + colWidths.Cardinality + colWidths.Type,
	}
	widths := []float64{
		colWidths.Name,
		colWidths.Flags,
		colWidths.Cardinality,
		colWidths.Type,
		colWidths.Description,
	}
	names := []string{"name", "flags", "card", "type", "desc"}

	for i, name := range names {
		sb.WriteString(fmt.Sprintf(`    <clipPath id="clip-%s"><rect x="%.0f" y="0" width="%.0f" height="%.0f"/></clipPath>
`,
			name, colStarts[i], widths[i], totalHeight))
	}

	return sb.String()
}

// buildTitleBar creates the title bar section
func buildTitleBar(totalWidth float64, config SVGConfig) string {
	return fmt.Sprintf(`<rect x="0" y="0" width="%.0f" height="%.0f" fill="%s" stroke="%s"/>
<text x="%.0f" y="%.0f" class="title-text">Structure</text>
`,
		totalWidth, config.TitleHeight, config.HeaderBgColor, config.BorderColor,
		config.Padding, config.TitleHeight/2+TitleVerticalOffset)
}

// buildDataRows renders all data rows
func buildDataRows(rows []RowData, totalWidth float64, config SVGConfig) string {
	var sb strings.Builder
	currentY := config.TitleHeight + config.HeaderHeight

	for _, row := range rows {
		sb.WriteString(renderDataRowWrapped(row, config, currentY, totalWidth))
		currentY += row.RowHeight
	}

	return sb.String()
}

func renderFallback() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="400" height="100">
<text x="10" y="50" font-family="Arial" font-size="14">Error: Could not load font for text measurement</text>
</svg>`
}
