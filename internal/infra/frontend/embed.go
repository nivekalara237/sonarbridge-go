package frontend

import (
	"embed"
	"io/fs"
)

var embeddedFiles embed.FS

func Files() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "dist")
}
