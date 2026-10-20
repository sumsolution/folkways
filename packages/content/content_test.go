package content

import (
	"encoding/json"
	"io/fs"
	"testing"
)

func TestFS(t *testing.T) {
	for _, name := range []string{"instructions.md", "settings.json", "skills", "prompts", "workflows", "templates/folkways.yaml"} {
		if _, err := fs.Stat(FS, name); err != nil {
			t.Errorf("%s not embedded: %v", name, err)
		}
	}
	b, err := fs.ReadFile(FS, "settings.json")
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(b, &settings); err != nil {
		t.Errorf("settings.json is not a JSON object: %v", err)
	}
}
