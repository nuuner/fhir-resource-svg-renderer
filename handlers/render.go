package handlers

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
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

// compressBrotliBase64URL compresses JSON bytes to Brotli and encodes as Base64URL
func compressBrotliBase64URL(jsonBytes []byte) (string, error) {
	var buf bytes.Buffer
	w := brotli.NewWriterLevel(&buf, brotli.BestCompression)
	if _, err := w.Write(jsonBytes); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf.Bytes()), nil
}

// decompressBrotliBase64URL decodes Base64URL and decompresses Brotli
func decompressBrotliBase64URL(encoded string) ([]byte, error) {
	compressed, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	r := brotli.NewReader(bytes.NewReader(compressed))
	return io.ReadAll(r)
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
// GET /render?resource={brotli-base64url-json}
func RenderHandler(c *gin.Context) {
	resourceParam := c.Query("resource")
	if resourceParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing 'resource' query parameter",
			"usage": "GET /render?resource={brotli-base64url-json}",
		})
		return
	}

	decodedJSON, err := decompressBrotliBase64URL(resourceParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid encoding (expected Brotli + Base64URL)",
			"details": err.Error(),
		})
		return
	}

	var resource models.ResourceDefinition
	if err := json.Unmarshal(decodedJSON, &resource); err != nil {
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

// CompressHandler compresses JSON to Brotli+Base64URL
// POST /compress with JSON body → returns {"compressed": "..."}
func CompressHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	if !json.Valid(body) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	compressed, err := compressBrotliBase64URL(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Compression failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"compressed": compressed})
}

// DecompressHandler decompresses Brotli+Base64URL to JSON
// POST /decompress with {"data": "compressed-string"} → returns JSON
func DecompressHandler(c *gin.Context) {
	var req struct {
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	decompressed, err := decompressBrotliBase64URL(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Decompression failed", "details": err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, string(decompressed))
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
| GET | /render?resource={compressed} | Render Brotli+Base64URL compressed JSON to SVG |
| POST | /render | Render JSON body to SVG |
| POST | /compress | Compress JSON → {"compressed": "..."} |
| POST | /decompress | Decompress {"data": "..."} → JSON |

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

### Compress JSON
` + "```bash" + `
curl -X POST http://localhost:8080/compress \
  -H "Content-Type: application/json" \
  -d '{"name":"Patient","type":"DomainResource"}'
# Returns: {"compressed":"G8gBgJOgxIm..."}
` + "```" + `

### GET Request (with compressed data)
` + "```" + `
GET /render?resource=G8gBgJOgxIm...
` + "```" + `

### POST Request (raw JSON)
` + "```bash" + `
curl -X POST http://localhost:8080/render \
  -H "Content-Type: application/json" \
  -d '{"name":"Patient","type":"DomainResource","elements":[{"name":"id","type":"id","cardinality":"0..1"}]}'
` + "```" + `

### Decompress
` + "```bash" + `
curl -X POST http://localhost:8080/decompress \
  -H "Content-Type: application/json" \
  -d '{"data":"G8gBgJOgxIm..."}'
# Returns: {"name":"Patient","type":"DomainResource"}
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

## URL Compression

The GET /render endpoint uses Brotli compression + Base64URL encoding for ~60-70% size reduction.

**Format:** Brotli compress → Base64URL encode (no padding)

Use the /compress endpoint to create compressed strings, or use the interactive editor.

## Notes

- GET /render: Requires Brotli+Base64URL compressed JSON (use /compress or the editor)
- POST /render: Send raw JSON with Content-Type: application/json
- CORS enabled (Access-Control-Allow-Origin: *)
- Responses cached 1 hour (Cache-Control: public, max-age=3600)
`

	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, helpText)
}
