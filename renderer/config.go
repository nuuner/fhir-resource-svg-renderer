package renderer

// Layout constants
const (
	// Row margins
	RowTopMargin    = 4.0
	RowBottomMargin = 6.0

	// Name column constraints
	MinNameColWidth = 180.0
	MaxNameColWidth = 300.0

	// Footer
	FooterHeight = 24.0

	// Labels
	UnusedElementLabel = "Not used"
)

// SVGConfig contains configuration for SVG rendering
type SVGConfig struct {
	FontFamily       string
	FontSize         float64
	HeaderFontSize   float64
	MinRowHeight     float64 // Minimum row height
	LineHeight       float64 // Height per line of text
	HeaderHeight     float64
	TitleHeight      float64
	IconSize         float64
	Padding          float64
	TreeStyle        TreeLineStyle

	// Column widths
	NameColWidth        float64
	FlagsColWidth       float64
	CardinalityColWidth float64
	TypeColWidth        float64
	DescriptionColWidth float64

	// Colors
	HeaderBgColor   string
	HeaderTextColor string
	RowBgColor      string
	AltRowBgColor   string
	BorderColor     string
	LinkColor       string
	TextColor       string
	NotUsedColor    string
	TodoColor       string

	// Text measurer (initialized during render)
	textMeasurer *TextMeasurer

	// CompressedResource is the Brotli+Base64URL encoded resource for footer links
	CompressedResource string
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() SVGConfig {
	return SVGConfig{
		FontFamily:          "Arial, sans-serif",
		FontSize:            12,
		HeaderFontSize:      13,
		MinRowHeight:        26, // topMargin(4) + lineHeight(16) + bottomMargin(6)
		LineHeight:          16,
		HeaderHeight:        28,
		TitleHeight:         32,
		IconSize:            14,
		Padding:             8,
		TreeStyle:           DefaultTreeStyle(),
		NameColWidth:        180,
		FlagsColWidth:       50,
		CardinalityColWidth: 55,
		TypeColWidth:        220,
		DescriptionColWidth: 400,
		HeaderBgColor:       "#F0F0F0",
		HeaderTextColor:     "#333333",
		RowBgColor:          "#FFFFFF",
		AltRowBgColor:       "#F8F8F8",
		BorderColor:         "#CCCCCC",
		LinkColor:           "#005EB8",
		TextColor:           "#333333",
		NotUsedColor:        "#999999",
		TodoColor:           "#FF6600",
	}
}
