package content

import (
	"io/fs"
	"testing"
)

func TestFS(t *testing.T) {
	for _, name := range []string{"instructions.md", "settings.json", "skills", "prompts", "workflows", "templates/folkways.yaml"} {
		if _, err := fs.Stat(FS, name); err != nil {
			t.Errorf("%s not embedded: %v", name, err)
		}
	}
}
