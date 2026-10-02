// Package skill names framework skills the way the running agent sees them.
package skill

import "os"

// namespaceEnv holds the namespace plugin skills live in; the Claude Code plugin's runner sets it to `folkways`.
const namespaceEnv = "FOLKWAYS_SKILL_NAMESPACE"

// Name returns the framework skill's name: `start` in a team repo, `folkways:start` in the plugin.
func Name(s string) string {
	if ns := os.Getenv(namespaceEnv); ns != "" {
		return ns + ":" + s
	}
	return s
}
