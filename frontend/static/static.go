package static

import (
	"embed"
	"io/fs"
)

//go:embed *
var static embed.FS

// Get filesystem for accessing static files
func Get() fs.FS {
	return &static
}
