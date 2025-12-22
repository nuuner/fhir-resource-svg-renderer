package handlers

import (
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"fhir_renderer/models"
	"fhir_renderer/renderer"
)

//go:embed example.json
var exampleJSON []byte

// validateResource checks that required fields are present
func validateResource(resource *models.ResourceDefinition) error {
	if resource.Name == "" {
		return errors.New("missing required field 'name'")
	}
	if resource.Type == "" {
		return errors.New("missing required field 'type'")
	}
	return nil
}

// renderAndRespond renders the resource to SVG and writes the response
func renderAndRespond(c *gin.Context, resource *models.ResourceDefinition) {
	config := renderer.DefaultConfig()
	svg := renderer.Render(resource, config)

	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", "public, max-age=3600")
	c.String(http.StatusOK, svg)
}

// RenderHandler handles the /render endpoint
// GET /render?resource={url-encoded-json}
func RenderHandler(c *gin.Context) {
	resourceParam := c.Query("resource")
	if resourceParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing 'resource' query parameter",
			"usage": "GET /render?resource={url-encoded-json}",
		})
		return
	}

	decodedJSON, err := url.QueryUnescape(resourceParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid URL encoding",
			"details": err.Error(),
		})
		return
	}

	var resource models.ResourceDefinition
	if err := json.Unmarshal([]byte(decodedJSON), &resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}

	if err := validateResource(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	renderAndRespond(c, &resource)
}

// RenderPOSTHandler handles POST requests with JSON body
// POST /render with JSON body
func RenderPOSTHandler(c *gin.Context) {
	var resource models.ResourceDefinition

	if err := c.ShouldBindJSON(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON body",
			"details": err.Error(),
		})
		return
	}

	if err := validateResource(&resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	renderAndRespond(c, &resource)
}

// HealthHandler returns health status
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"service": "fhir-renderer",
	})
}

// ExampleHandler returns an example JSON schema
func ExampleHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, string(exampleJSON))
}

// HelpHandler returns API documentation in markdown format
func HelpHandler(c *gin.Context) {
	helpText := `# FHIR Renderer API

Renders FHIR ResourceDefinition structures as SVG diagrams.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check → {"status":"ok"} |
| GET | /help | This documentation |
| GET | /example | Example ResourceDefinition JSON |
| GET | /render?resource={json} | Render URL-encoded JSON to SVG |
| POST | /render | Render JSON body to SVG |

## JSON Schema

### ResourceDefinition (root)
` + "```json" + `
{
  "resourceType": "ResourceDefinition",  // optional, identifier
  "name": "...",                          // REQUIRED: resource name
  "type": "...",                          // REQUIRED: e.g. "DomainResource"
  "flags": ["TU", "N"],                   // optional: metadata flags
  "description": "...",                   // optional: human description
  "elements": [...],                      // optional: child elements
  "extensions": [...]                     // optional: FHIR extensions
}
` + "```" + `

### Element (nested)
` + "```json" + `
{
  "name": "...",             // REQUIRED: field name
  "type": "...",             // REQUIRED: data type
  "cardinality": "0..*",     // optional: "0..1", "1..1", "0..*", "1..*"
  "flags": ["S", "?!"],      // optional: FHIR flags
  "typeRef": "https://...",  // optional: link to type docs
  "description": "...",      // optional: field description
  "usage": "used",           // optional: implementation status
  "notes": "...",            // optional: custom notes
  "binding": {...},          // optional: value set binding
  "elements": [...],         // optional: nested children (BackboneElement)
  "extensions": [...]        // optional: extensions on this element
}
` + "```" + `

### Binding
` + "```json" + `
{
  "strength": "required",              // "required"|"extensible"|"preferred"|"example"
  "valueSet": "val1 | val2 | val3",    // allowed values or reference URL
  "url": "https://..."                 // optional: value set docs link
}
` + "```" + `

### Extension
` + "```json" + `
{
  "name": "...",           // REQUIRED: extension name
  "url": "https://...",    // REQUIRED: extension URL
  "type": "...",           // REQUIRED: data type
  "cardinality": "0..1",   // optional: cardinality
  "context": "...",        // optional: where extension applies (root-level only)
  "description": "..."     // optional: description
}
` + "```" + `

## Flags

| Flag | Symbol | Meaning |
|------|--------|---------|
| S | Σ | Summary element |
| ?! | ?!Σ | Modifier element |
| I | I | Has constraint |
| TU | [TU] | Trial use (boxed) |
| N | [N] | Normative (boxed) |

## Usage Values

| Value | Rendering |
|-------|-----------|
| used | Normal style |
| not-used | Grayed out (#999) |
| todo | Bold orange, "TODO:" prefix |
| optional | Default style |

## Icons (auto-selected by type)

- **Folder (yellow)**: Root resource
- **Folder+dot**: BackboneElement (nested structure)
- **Diamond (blue)**: Simple element
- **Circle "E" (orange)**: Extension
- **Circle+line (green)**: Choice type [x]
- **Arrow (blue)**: Reference type

## Examples

### GET Request
` + "```" + `
GET /render?resource=%7B%22name%22%3A%22Patient%22%2C%22type%22%3A%22DomainResource%22%7D
` + "```" + `

### POST Request
` + "```bash" + `
curl -X POST http://localhost:8080/render \
  -H "Content-Type: application/json" \
  -d '{"name":"Patient","type":"DomainResource","elements":[{"name":"id","type":"id","cardinality":"0..1"}]}'
` + "```" + `

### Minimal Valid JSON
` + "```json" + `
{"name":"MyResource","type":"DomainResource"}
` + "```" + `

## Response

- **Success**: SVG/XML (Content-Type: image/svg+xml)
- **Error**: JSON with "error" and optional "details" fields

## Errors

| Code | Cause |
|------|-------|
| 400 | Missing 'resource' param, invalid JSON, missing 'name'/'type' |

## Notes

- GET: URL-encode the JSON (use encodeURIComponent or similar)
- POST: Send raw JSON with Content-Type: application/json
- CORS enabled (Access-Control-Allow-Origin: *)
- Responses cached 1 hour (Cache-Control: public, max-age=3600)
`

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, helpText)
}
