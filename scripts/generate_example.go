// Script to regenerate example.svg from handlers/example.json
//
// Usage: go run scripts/generate_example.go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"fhir_renderer/models"
	"fhir_renderer/renderer"
)

func main() {
	// Read the example JSON
	data, err := os.ReadFile("handlers/example.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading example JSON: %v\n", err)
		os.Exit(1)
	}

	// Parse the resource
	var resource models.ResourceDefinition
	if err := json.Unmarshal(data, &resource); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Render to SVG
	config := renderer.DefaultConfig()
	svg := renderer.Render(&resource, config)

	// Write to example.svg
	if err := os.WriteFile("example.svg", []byte(svg), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing SVG: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated example.svg")
}
