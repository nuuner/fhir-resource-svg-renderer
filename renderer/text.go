package renderer

import (
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

// TextMeasurer handles text measurement and wrapping
type TextMeasurer struct {
	face     font.Face
	fontSize float64
}

// NewTextMeasurer creates a new text measurer with the specified font size
func NewTextMeasurer(fontSize float64) (*TextMeasurer, error) {
	// Parse the embedded Go font
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	// Create a font face with the specified size
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	return &TextMeasurer{
		face:     face,
		fontSize: fontSize,
	}, nil
}

// MeasureString returns the width of a string in pixels
func (tm *TextMeasurer) MeasureString(s string) float64 {
	advance := font.MeasureString(tm.face, s)
	return fixedToFloat(advance)
}

// WrapText wraps text to fit within maxWidth, returning multiple lines
func (tm *TextMeasurer) WrapText(text string, maxWidth float64) []string {
	if text == "" {
		return []string{""}
	}

	// Check if text fits without wrapping
	if tm.MeasureString(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		testLine := currentLine + " " + word
		if tm.MeasureString(testLine) <= maxWidth {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return lines
}

// TruncateText truncates text to fit within maxWidth, adding ellipsis if needed
func (tm *TextMeasurer) TruncateText(text string, maxWidth float64) string {
	if text == "" {
		return ""
	}

	if tm.MeasureString(text) <= maxWidth {
		return text
	}

	ellipsis := "..."
	ellipsisWidth := tm.MeasureString(ellipsis)
	availableWidth := maxWidth - ellipsisWidth

	if availableWidth <= 0 {
		return ellipsis
	}

	// Binary search for the right length
	runes := []rune(text)
	low, high := 0, len(runes)

	for low < high {
		mid := (low + high + 1) / 2
		if tm.MeasureString(string(runes[:mid])) <= availableWidth {
			low = mid
		} else {
			high = mid - 1
		}
	}

	if low == 0 {
		return ellipsis
	}

	return string(runes[:low]) + ellipsis
}

// LineHeight returns the recommended line height for the font
func (tm *TextMeasurer) LineHeight() float64 {
	metrics := tm.face.Metrics()
	return fixedToFloat(metrics.Height)
}

// Ascent returns the ascent of the font (height above baseline)
func (tm *TextMeasurer) Ascent() float64 {
	metrics := tm.face.Metrics()
	return fixedToFloat(metrics.Ascent)
}

func fixedToFloat(f fixed.Int26_6) float64 {
	return float64(f) / 64.0
}

// Close releases resources
func (tm *TextMeasurer) Close() {
	if closer, ok := tm.face.(interface{ Close() error }); ok {
		closer.Close()
	}
}
