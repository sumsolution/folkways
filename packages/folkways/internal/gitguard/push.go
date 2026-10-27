package gitguard

import (
	"fmt"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/team"
)

// verdict is git-guard's answer for a command line: deny wins over ask; both empty allows.
type verdict struct {
	deny, ask string
}

// pushTarget is where a push goes: the remote and each source → destination branch. known is false when a
// word isn't statically known or the push names no single set of branches (--all, --mirror).
type pushTarget struct {
	remote string
	refs   [][2]string
	known  bool
}

// Push options whose operand is the next word.
var pushOptWithArg = map[string]bool{"--repo": true, "-o": true, "--push-option": true, "--receive-pack": true, "--exec": true}

// parsePush reads `git push [options] [remote [refspec...]]` run in dir. With no refspec it uses the
// branch's upstream, else the branch of the same name.
func parsePush(dir string, args []arg) pushTarget {
	var pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !a.ok {
			return pushTarget{}
		}
		switch v := a.val; {
		case v == "--all" || v == "--mirror" || v == "--branches":
			return pushTarget{}
		case pushOptWithArg[v]:
			i++
		case strings.HasPrefix(v, "-"):
		default:
			pos = append(pos, v)
		}
	}
	branch, _ := gitOut(dir, "symbolic-ref", "--short", "-q", "HEAD")
	t := pushTarget{remote: "origin", known: true}
	if len(pos) > 0 {
		t.remote = pos[0]
	}
	if len(pos) <= 1 {
		dst := branch
		if up, err := gitOut(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"); err == nil {
			remote, b, _ := strings.Cut(up, "/")
			if len(pos) == 0 {
				t.remote = remote
			}
			if remote == t.remote {
				dst = b
			}
		}
		if dst == "" {
			return pushTarget{}
		}
		t.refs = [][2]string{{"HEAD", dst}}
		return t
	}
	for _, spec := range pos[1:] {
		src, dst, hasDst := strings.Cut(strings.TrimPrefix(spec, "+"), ":")
		if !hasDst {
			dst = src
			if src == "HEAD" {
				dst = branch
			}
		}
		if src == "" {
			src = "HEAD" // a deletion (`:branch`) is checked like a push of HEAD: approval still applies
		}
		dst = strings.TrimPrefix(dst, "refs/heads/")
		if dst == "" {
			return pushTarget{}
		}
		t.refs = append(t.refs, [2]string{src, dst})
	}
	return t
}

// defaultBranch is the remote's default branch, from refs/remotes/<remote>/HEAD, else "main".
func defaultBranch(dir, remote string) string {
	if ref, err := gitOut(dir, "symbolic-ref", "--short", "refs/remotes/"+remote+"/HEAD"); err == nil {
		if b, ok := strings.CutPrefix(ref, remote+"/"); ok && b != "" {
			return b
		}
	}
	return "main"
}

// isBrain reports whether dir is in the brain repo or one of its worktrees: they share a git common dir.
func isBrain(dir string, cfg team.Config) bool {
	a, err1 := gitOut(dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	b, err2 := gitOut(cfg.Dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	return err1 == nil && err2 == nil && sameDir(a, b)
}

// checkPushVerdict is the push rule set (ADR 0014): outside the brain, the review gate; in the brain, no
// knowledge/ change goes straight to the default branch; then push.brain or push.app decides whether to ask.
func checkPushVerdict(dir string, args []arg, cfg team.Config) verdict {
	brain := isBrain(dir, cfg)
	if !brain {
		if reason := checkPush(dir, cfg); reason != "" {
			return verdict{deny: reason}
		}
	}
	t := parsePush(dir, args)
	def := defaultBranch(dir, t.remote)
	toDefault := !t.known
	for _, r := range t.refs {
		if r[1] != def {
			continue
		}
		toDefault = true
		if brain {
			if reason := checkKnowledge(dir, t.remote, r[0], def); reason != "" {
				return verdict{deny: reason}
			}
		}
	}
	kind := "app"
	if brain {
		kind = "brain"
	}
	switch approval := cfg.PushApproval(kind); {
	case approval == "always", approval == "default-branch" && toDefault:
		return verdict{ask: fmt.Sprintf("folkways.yaml push.%s is %s: confirm this push with the user.", kind, approval)}
	}
	return verdict{}
}

// checkKnowledge denies a push to the default branch whose new commits change knowledge/: knowledge lands by
// PR. It fails open when the remote branch can't be read.
func checkKnowledge(dir, remote, src, def string) string {
	out, err := gitOut(dir, "log", "--format=%h", remote+"/"+def+".."+src, "--", "knowledge/")
	if err != nil || out == "" {
		return ""
	}
	return fmt.Sprintf("Push blocked: commits for %s change knowledge/, which lands by PR. Push the branch and open a PR instead (the pr skill).", def)
}

// unknownPush is the verdict for a push whose directory isn't known: ask unless no push ever asks.
func unknownPush(cfg team.Config) verdict {
	if cfg.PushApproval("brain") == "never" && cfg.PushApproval("app") == "never" {
		return verdict{}
	}
	return verdict{ask: "git-guard can't tell which repo this push is in: confirm it with the user."}
}
