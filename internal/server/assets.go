package server

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed web/templates/*.html
var templatesFS embed.FS

// Note: web/static is currently empty, so we don't embed it
// When static files are added, uncomment the following line:
// //go:embed web/static
// var staticFS embed.FS

// EmbeddedAssets handles embedded web assets
type EmbeddedAssets struct {
	templates *template.Template
}

// NewEmbeddedAssets creates a new embedded assets handler
func NewEmbeddedAssets() (*EmbeddedAssets, error) {
	// Parse templates from embedded filesystem
	tmpl, err := template.ParseFS(templatesFS, "web/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &EmbeddedAssets{
		templates: tmpl,
	}, nil
}

// GetTemplates returns the parsed templates
func (e *EmbeddedAssets) GetTemplates() *template.Template {
	return e.templates
}

// SetupStaticRoutes sets up static file serving with embedded assets
func (e *EmbeddedAssets) SetupStaticRoutes(router *gin.Engine) {
	// Currently no static files are embedded, serve empty filesystem
	// When static files are added and embedded, update this method
	router.StaticFS("/static", http.FS(emptyFS{}))
}

// HTML renders HTML templates with embedded assets
func (e *EmbeddedAssets) HTML(c *gin.Context, name string, data any) {
	// Execute the template
	err := e.templates.ExecuteTemplate(c.Writer, name, data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
}

// emptyFS is an empty filesystem for cases where static files don't exist
type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) {
	return &emptyFile{}, nil
}

type emptyFile struct{}

func (f *emptyFile) Stat() (fs.FileInfo, error) { return &emptyFileInfo{}, nil }
func (f *emptyFile) Read([]byte) (int, error)    { return 0, nil }
func (f *emptyFile) Close() error                { return nil }

type emptyFileInfo struct{}

func (fi *emptyFileInfo) Name() string       { return "" }
func (fi *emptyFileInfo) Size() int64        { return 0 }
func (fi *emptyFileInfo) Mode() fs.FileMode  { return 0444 }
func (fi *emptyFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *emptyFileInfo) IsDir() bool        { return false }
func (fi *emptyFileInfo) Sys() interface{}   { return nil }

// isTemplateFile checks if a file is a template file
func isTemplateFile(name string) bool {
	return strings.HasSuffix(name, ".html")
}

// templateFileHelper helps extract template filename from full path
func templateFileHelper(path string) string {
	// Extract just the filename from the full path
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}