package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Dist returns sub filesystem for the compiled assets
func Dist() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
