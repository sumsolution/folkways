// Package content is the framework content the folkways binary embeds (ADR 0011): skills, instructions,
// prompts, the settings fragment, workflows and the init templates.
package content

import "embed"

// FS holds the content at the paths it has in this directory. `all:` keeps dotfiles such as .gitkeep.
//
//go:embed instructions.md settings.json all:skills all:prompts all:workflows all:templates
var FS embed.FS
