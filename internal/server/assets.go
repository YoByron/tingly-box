package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var webDistFS embed.FS

// EmbeddedAssets handles embedded web assets
type EmbeddedAssets struct{}

// NewEmbeddedAssets creates a new embedded assets handler
func NewEmbeddedAssets() (*EmbeddedAssets, error) {
	return &EmbeddedAssets{}, nil
}

// SetupStaticRoutes sets up static file serving with embedded assets
func (e *EmbeddedAssets) SetupStaticRoutes(router *gin.Engine) {

	// Serve static assets from embedded filesystem
	st, _ := fs.Sub(webDistFS, "web/dist/assets")
	router.StaticFS("/assets", http.FS(st))

	router.StaticFile("/vite.svg", "web/dist/vite.svg")

	router.NoRoute(func(c *gin.Context) {
		data, err := webDistFS.ReadFile("web/dist/index.html")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
}

// HTML renders HTML templates with embedded assets
func (e *EmbeddedAssets) HTML(c *gin.Context, name string, data any) {
	// For SPA, just serve the index.html file directly
	// Ignore the name parameter since we only have one index.html
	c.FileFromFS("web/dist/index.html", http.FS(webDistFS))
}
