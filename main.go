package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"fhir_renderer/handlers"
)

func main() {
	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create gin router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Enable CORS
	router.Use(corsMiddleware())

	// Routes
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/editor")
	})
	router.GET("/health", handlers.HealthHandler)
	router.GET("/help", handlers.HelpHandler)
	router.GET("/render", handlers.RenderHandler)
	router.POST("/render", handlers.RenderPOSTHandler)
	router.GET("/example", handlers.ExampleHandler)
	router.GET("/editor", handlers.EditorHandler)

	// Start server
	log.Printf("FHIR Renderer starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  GET  /health  - Health check")
	log.Printf("  GET  /help    - API documentation (markdown)")
	log.Printf("  GET  /render?resource={url-encoded-json}  - Render SVG from query param")
	log.Printf("  POST /render  - Render SVG from JSON body")
	log.Printf("  GET  /example - Get example JSON schema")
	log.Printf("  GET  /editor  - Interactive editor page")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
