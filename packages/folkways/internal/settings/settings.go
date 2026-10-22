// Package settings merges the framework settings fragment into a team's .claude/settings.json. A port of
// packages/cli/src/settings.ts. JSON objects keep their key order (Object), so the merged file matches
// what JSON.stringify writes.
package settings

import (
	"slices"
	"strings"

	"github.com/sumsolution/folkways/packages/folkways/internal/clierr"
)

// FrameworkHookPrefix starts the command of every framework-owned hook entry: the guard that turns the
// hook off when `folkways` isn't on PATH (ADR 0012). packages/content/settings.json uses it.
const FrameworkHookPrefix = "command -v folkways >/dev/null 2>&1 || "

// oldHookPrefixes are earlier command forms, still owned so sync removes them.
var oldHookPrefixes = []string{
	`"${CLAUDE_PROJECT_DIR:-.}"/node_modules/.bin/folkways-hook `,
	"node_modules/.bin/folkways hook ",
}

// IsFrameworkHook reports whether a hook command is framework-owned.
func IsFrameworkHook(command string) bool {
	return strings.HasPrefix(command, FrameworkHookPrefix) ||
		slices.ContainsFunc(oldHookPrefixes, func(p string) bool { return strings.HasPrefix(command, p) })
}

func isOurs(h any) bool {
	o, ok := h.(*Object)
	if !ok {
		return false
	}
	c, ok := o.Get("command")
	s, isString := c.(string)
	return ok && isString && IsFrameworkHook(s)
}

// withoutOurs drops framework entries from a group. Groups of unexpected shape stay as they are.
func withoutOurs(group any) any {
	o, ok := group.(*Object)
	if !ok {
		return group
	}
	hooks, ok := o.Get("hooks")
	list, isArray := hooks.([]any)
	if !ok || !isArray {
		return group
	}
	kept := []any{}
	for _, h := range list {
		if !isOurs(h) {
			kept = append(kept, h)
		}
	}
	out := o.clone()
	out.Set("hooks", kept)
	return out
}

func isEmptyGroup(g any) bool {
	o, ok := g.(*Object)
	if !ok {
		return false
	}
	hooks, _ := o.Get("hooks")
	list, isArray := hooks.([]any)
	return isArray && len(list) == 0
}

// Merge merges the framework settings fragment into a team's settings: it replaces framework hook entries,
// adds missing framework `ask` rules, sets `disableAllHooks` from the fragment (ADR 0001 M11) and keeps
// everything else. It fails when the team settings have a shape it can't merge into. Neither input changes.
func Merge(current any, fragment *Object) (*Object, error) {
	cur, ok := current.(*Object)
	if !ok {
		return nil, clierr.New(".claude/settings.json must be a JSON object")
	}
	curHooks, hasHooks := cur.Get("hooks")
	if _, isObj := curHooks.(*Object); hasHooks && !isObj {
		return nil, clierr.New(".claude/settings.json: hooks must be a JSON object")
	}
	perms, hasPerms := cur.Get("permissions")
	permsObj, isObj := perms.(*Object)
	if hasPerms && !isObj {
		return nil, clierr.New(".claude/settings.json: permissions must be a JSON object")
	}
	if isObj {
		if ask, has := permsObj.Get("ask"); has {
			if _, isArray := ask.([]any); !isArray {
				return nil, clierr.New(".claude/settings.json: permissions.ask must be an array")
			}
		}
	}
	out := cur.clone()

	hooks := &Object{}
	if h, ok := out.Get("hooks"); ok {
		events := h.(*Object)
		for _, event := range events.Keys {
			groups, _ := events.Get(event)
			list, isArray := groups.([]any)
			if !isArray {
				hooks.Set(event, groups)
				continue
			}
			kept := []any{}
			for _, g := range list {
				if g = withoutOurs(g); !isEmptyGroup(g) {
					kept = append(kept, g)
				}
			}
			if len(kept) > 0 {
				hooks.Set(event, kept)
			}
		}
	}
	if f, ok := fragment.Get("hooks"); ok {
		if events, ok := f.(*Object); ok {
			for _, event := range events.Keys {
				groups, _ := events.Get(event)
				existing, _ := hooks.Get(event) // `?? []`: a missing or null event starts empty
				if existing == nil {
					existing = []any{}
				}
				list, isArray := existing.([]any)
				if !isArray {
					return nil, clierr.New(".claude/settings.json: hooks.%s must be an array", event)
				}
				fragGroups, _ := clone(groups).([]any)
				hooks.Set(event, append(slices.Clone(list), fragGroups...))
			}
		}
	}
	if len(hooks.Keys) > 0 {
		out.Set("hooks", hooks)
	} else {
		out.Delete("hooks")
	}

	if v, ok := fragment.Get("disableAllHooks"); ok {
		out.Set("disableAllHooks", v)
	}

	var ask []any
	if p, ok := fragment.Get("permissions"); ok {
		if po, ok := p.(*Object); ok {
			a, _ := po.Get("ask")
			ask, _ = a.([]any)
		}
	}
	if len(ask) > 0 {
		p, _ := out.Get("permissions")
		permissions, ok := p.(*Object)
		if !ok {
			permissions = &Object{}
		}
		e, _ := permissions.Get("ask")
		existing, _ := e.([]any)
		merged := slices.Clone(existing)
		for _, rule := range ask {
			if !slices.ContainsFunc(existing, func(x any) bool { return sameScalar(x, rule) }) {
				merged = append(merged, rule)
			}
		}
		if merged == nil {
			merged = []any{}
		}
		permissions.Set("ask", merged)
		out.Set("permissions", permissions)
	}
	return out, nil
}

// sameScalar is Array.includes for the scalars an ask rule can be: objects and arrays never match.
func sameScalar(a, b any) bool {
	switch a.(type) {
	case *Object, []any:
		return false
	}
	switch b.(type) {
	case *Object, []any:
		return false
	}
	return a == b
}
