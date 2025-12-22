package renderer

// Font and text rendering constants
const (
	// FontRenderingBuffer is extra buffer for font rendering variations
	FontRenderingBuffer = 15.0

	// BoldTextWidthFactor is width multiplier for estimating bold text width
	BoldTextWidthFactor = 0.90

	// HeaderTextMarginY is vertical margin for header text positioning
	HeaderTextMarginY = 6.0

	// TextVerticalOffset is Y offset for text baseline alignment
	TextVerticalOffset = 4.0

	// TitleVerticalOffset is Y offset for title text positioning
	TitleVerticalOffset = 5.0

	// BorderStrokeWidth is the width of cell borders
	BorderStrokeWidth = 0.5
)

// Icon and spacing constants
const (
	// IconTextGap is space between icon and text
	IconTextGap = 4.0

	// IconPaddingRight is padding after icon before text measurement
	IconPaddingRight = 8.0

	// IconSpaceInMeasurement is space reserved for icon in name column width calculation
	IconSpaceInMeasurement = 12.0
)

// Flag rendering constants
const (
	// FlagCharWidth is estimated width per character in flag text
	FlagCharWidth = 7.0

	// FlagBoxPadding is horizontal padding inside flag boxes
	FlagBoxPadding = 6.0

	// FlagGap is space between flags
	FlagGap = 4.0

	// FlagBoxTextOffset is X offset for text inside flag box
	FlagBoxTextOffset = 3.0
)

// Tree line constants
const (
	// TreeHorizontalGap is gap between horizontal tree line and icon
	TreeHorizontalGap = 2.0

	// IconLineVerticalOffset is the vertical offset for icons and horizontal tree lines
	IconLineVerticalOffset = 2.0
)

// SVG output constants
const (
	// SVGHeightPadding is extra padding at bottom of SVG
	SVGHeightPadding = 2.0
)
