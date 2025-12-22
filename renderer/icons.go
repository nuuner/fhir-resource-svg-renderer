package renderer

import (
	"fmt"
	"strings"
)

// Icon types matching HL7 FHIR visual style
const (
	IconResource        = "resource"        // Yellow folder - for root resource
	IconBackboneElement = "backbone"        // Yellow folder with dot - for backbone elements
	IconElement         = "element"         // Blue diamond - for regular elements
	IconExtension       = "extension"       // Orange circle with E - for extensions
	IconChoice          = "choice"          // Green circle - for choice types
	IconReference       = "reference"       // Blue arrow - for references
)

// RenderIcon returns SVG markup for the specified icon type at the given position
func RenderIcon(iconType string, x, y float64, size float64) string {
	switch iconType {
	case IconResource:
		return renderFolderIcon(x, y, size, "#FDB813", true) // Yellow folder
	case IconBackboneElement:
		return renderFolderIcon(x, y, size, "#FDB813", false) // Yellow folder with inner mark
	case IconElement:
		return renderDiamondIcon(x, y, size, "#005EB8") // Blue diamond
	case IconExtension:
		return renderExtensionIcon(x, y, size, "#FF8C00") // Orange extension
	case IconChoice:
		return renderChoiceIcon(x, y, size, "#28A745") // Green choice
	case IconReference:
		return renderReferenceIcon(x, y, size, "#005EB8") // Blue reference
	default:
		return renderDiamondIcon(x, y, size, "#005EB8") // Default to element
	}
}

// renderFolderIcon draws a folder icon (for resources and backbone elements)
func renderFolderIcon(x, y, size float64, color string, filled bool) string {
	// Folder shape
	w := size * 0.9
	h := size * 0.7
	tabW := w * 0.4
	tabH := h * 0.2

	fillColor := color
	if !filled {
		fillColor = "#FFFFFF"
	}

	svg := fmt.Sprintf(`<g transform="translate(%f,%f)">
    <path d="M0,%f L0,%f L%f,%f L%f,0 L%f,0 L%f,%f L0,%f Z"
          fill="%s" stroke="%s" stroke-width="1"/>`,
		x, y,
		tabH, h, w, h, w, tabW, tabW, tabH, tabH,
		fillColor, color)

	if !filled {
		// Add inner dot for backbone element
		svg += fmt.Sprintf(`<circle cx="%f" cy="%f" r="%f" fill="%s"/>`,
			w/2, h*0.6, size*0.12, color)
	}

	svg += "</g>"
	return svg
}

// renderDiamondIcon draws a diamond icon (for regular elements)
func renderDiamondIcon(x, y, size float64, color string) string {
	half := size / 2
	return fmt.Sprintf(`<polygon points="%f,%f %f,%f %f,%f %f,%f"
        fill="%s" stroke="%s" stroke-width="0.5"/>`,
		x+half, y,        // top
		x+size, y+half,   // right
		x+half, y+size,   // bottom
		x, y+half,        // left
		color, color)
}

// renderExtensionIcon draws an extension icon (circle with E)
func renderExtensionIcon(x, y, size float64, color string) string {
	cx := x + size/2
	cy := y + size/2
	r := size / 2

	return fmt.Sprintf(`<g>
    <circle cx="%f" cy="%f" r="%f" fill="%s"/>
    <text x="%f" y="%f" fill="white" font-family="Arial" font-size="%f"
          text-anchor="middle" dominant-baseline="central" font-weight="bold">E</text>
</g>`,
		cx, cy, r, color,
		cx, cy, size*0.6)
}

// renderChoiceIcon draws a choice type icon (green circle with split)
func renderChoiceIcon(x, y, size float64, color string) string {
	cx := x + size/2
	cy := y + size/2
	r := size / 2

	return fmt.Sprintf(`<g>
    <circle cx="%f" cy="%f" r="%f" fill="%s"/>
    <line x1="%f" y1="%f" x2="%f" y2="%f" stroke="white" stroke-width="1.5"/>
</g>`,
		cx, cy, r, color,
		cx, cy-r*0.5, cx, cy+r*0.5)
}

// renderReferenceIcon draws a reference icon (arrow pointing right)
func renderReferenceIcon(x, y, size float64, color string) string {
	// Arrow pointing right
	arrowSize := size * 0.8
	startX := x + size*0.1
	midY := y + size/2

	return fmt.Sprintf(`<g>
    <line x1="%f" y1="%f" x2="%f" y2="%f" stroke="%s" stroke-width="2"/>
    <polygon points="%f,%f %f,%f %f,%f" fill="%s"/>
</g>`,
		startX, midY, startX+arrowSize*0.6, midY, color,
		startX+arrowSize*0.5, midY-arrowSize*0.3,
		startX+arrowSize, midY,
		startX+arrowSize*0.5, midY+arrowSize*0.3,
		color)
}

// GetIconTypeForElement determines the appropriate icon type based on element properties
func GetIconTypeForElement(elementType string, isRoot bool, hasChildren bool) string {
	if isRoot {
		return IconResource
	}

	switch elementType {
	case "BackboneElement":
		return IconBackboneElement
	case "Extension":
		return IconExtension
	default:
		// Check if it's a choice type (ends with [x])
		if strings.HasSuffix(elementType, "[x]") {
			return IconChoice
		}
		// Check if it's a reference
		if strings.HasPrefix(elementType, "Reference") {
			return IconReference
		}
		// Check if it has children (backbone-like)
		if hasChildren {
			return IconBackboneElement
		}
		return IconElement
	}
}

// RenderGitHubIcon renders a GitHub mark icon at the given position
func RenderGitHubIcon(x, y, size float64, color string) string {
	// GitHub mark path scaled to fit within size
	scale := size / 16.0
	return fmt.Sprintf(`<g transform="translate(%f,%f) scale(%f)">
    <path fill="%s" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
</g>`, x, y, scale, color)
}
