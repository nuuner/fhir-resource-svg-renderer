package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// EditorHandler serves the interactive editor page
func EditorHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "editor.html", nil)
}
