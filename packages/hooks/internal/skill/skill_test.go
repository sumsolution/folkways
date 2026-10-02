package skill

import "testing"

func TestName(t *testing.T) {
	t.Setenv(namespaceEnv, "")
	if got := Name("start"); got != "start" {
		t.Errorf("Name(start) without a namespace = %q, want start", got)
	}
	t.Setenv(namespaceEnv, "folkways")
	if got := Name("start"); got != "folkways:start" {
		t.Errorf("Name(start) in the folkways namespace = %q, want folkways:start", got)
	}
}
