package handlers

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"

	"fhir_renderer/models"
	"fhir_renderer/renderer"
)

//go:embed example.json
var exampleJSON []byte

//go:embed help.md
var helpMarkdown string

// SVGCacheTTLSeconds is the cache duration for rendered SVGs
const SVGCacheTTLSeconds = 3600

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
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", SVGCacheTTLSeconds))
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
	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.String(http.StatusOK, helpMarkdown)
}
