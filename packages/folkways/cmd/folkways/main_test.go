package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	cases := []struct {
		args   []string
		code   int
		stdout string
		stderr string
	}{
		{[]string{"version"}, 0, "dev\n", ""},
		{[]string{"--version"}, 0, "dev\n", ""},
		{[]string{"help"}, 0, usage + "\n", ""},
		{nil, 1, "", usage + "\n"},
		{[]string{"nope"}, 1, "", usage + "\n"},
	}
	for _, c := range cases {
		var out, errOut bytes.Buffer
		if code := run(c.args, &out, &errOut); code != c.code || out.String() != c.stdout || errOut.String() != c.stderr {
			t.Errorf("run(%q) = %d, %q, %q; want %d, %q, %q", c.args, code, out.String(), errOut.String(), c.code, c.stdout, c.stderr)
		}
	}
}

// The hook path exits 0 whatever it gets: Copilot CLI denies the tool call otherwise.
func TestHookAlwaysExitsZero(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "folkways")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	for _, args := range [][]string{{"hook"}, {"hook", "nope"}, {"hook", "git-guard"}, {"hook", "git-guard", "extra"}} {
		cmd := exec.Command(bin, args...)
		cmd.Dir = t.TempDir() // no folkways.yaml
		cmd.Stdin = strings.NewReader("not json")
		if err := cmd.Run(); err != nil {
			t.Errorf("folkways %s: %v, want exit 0", strings.Join(args, " "), err)
		}
	}
}

// TestNoPackageLevelWork keeps start-up cheap and safe for hooks (ADR 0011). Go runs every linked package's
// initialization before main, on every hook call, and a panic there exits nonzero before any recover.
// So no `func init`, and a package-level var may hold only literals, references, constant expressions,
// func literals, errors.New, or regexp.MustCompile of a string literal. This test compiles each of those
// patterns, so a bad one fails here even in a package with no tests of its own.
func TestNoPackageLevelWork(t *testing.T) {
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == "npm" || d.Name() == "bin" || d.Name() == "testdata") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		f, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			return err
		}
		for _, decl := range f.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Recv == nil && decl.Name.Name == "init" {
					t.Errorf("%s: func init runs on every hook call; do the work where it's needed", fset.Position(decl.Pos()))
				}
			case *ast.GenDecl:
				if decl.Tok != token.VAR {
					continue
				}
				for _, spec := range decl.Specs {
					for _, v := range spec.(*ast.ValueSpec).Values {
						if !cheap(v) {
							t.Errorf("%s: package-level var does work at start-up; compute it where it's used", fset.Position(v.Pos()))
						}
						checkPatterns(t, fset, v)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// cheap reports whether a package-level initializer is allowed by TestNoPackageLevelWork.
func cheap(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.BasicLit, *ast.Ident, *ast.FuncLit:
		return true
	case *ast.SelectorExpr:
		return cheap(e.X) // pkg.Name is cheap; f().Name runs f
	case *ast.UnaryExpr:
		return e.Op != token.ARROW && cheap(e.X) // a channel receive blocks
	case *ast.ParenExpr:
		return cheap(e.X)
	case *ast.BinaryExpr:
		return cheap(e.X) && cheap(e.Y)
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			if !cheap(elt) {
				return false
			}
		}
		return true
	case *ast.KeyValueExpr:
		return cheap(e.Key) && cheap(e.Value)
	case *ast.CallExpr:
		fn, ok := e.Fun.(*ast.SelectorExpr)
		if !ok || len(e.Args) != 1 {
			return false
		}
		pkg, ok := fn.X.(*ast.Ident)
		if !ok {
			return false
		}
		_, lit := e.Args[0].(*ast.BasicLit)
		call := pkg.Name + "." + fn.Sel.Name
		return lit && (call == "errors.New" || call == "regexp.MustCompile")
	}
	return false
}

func TestCheap(t *testing.T) {
	cases := map[string]bool{
		`"x"`:                         true,
		`os.Args`:                     true,
		`&T{A: 1, B: pkg.C}`:          true,
		`-1 + 2`:                      true,
		`func() {}`:                   true,
		`errors.New("x")`:             true,
		`regexp.MustCompile("x")`:     true,
		`team.Load().Configured`:      false,
		`<-ready`:                     false,
		`!(<-ready)`:                  false,
		`T{A: f()}`:                   false,
		`regexp.MustCompile(pattern)`: false,
		`os.Getenv("HOME")`:           false,
	}
	for src, want := range cases {
		e, err := parser.ParseExpr(src)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		if got := cheap(e); got != want {
			t.Errorf("cheap(%s) = %v, want %v", src, got, want)
		}
	}
}

// checkPatterns compiles every regexp.MustCompile string literal in a package-level initializer.
func checkPatterns(t *testing.T, fset *token.FileSet, e ast.Expr) {
	ast.Inspect(e, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return true
		}
		fn, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || fn.Sel.Name != "MustCompile" {
			return true
		}
		if pkg, ok := fn.X.(*ast.Ident); !ok || pkg.Name != "regexp" {
			return true
		}
		if lit, ok := call.Args[0].(*ast.BasicLit); ok {
			if pattern, err := strconv.Unquote(lit.Value); err == nil {
				if _, err := regexp.Compile(pattern); err != nil {
					t.Errorf("%s: regexp.MustCompile would panic at start-up: %v", fset.Position(lit.Pos()), err)
				}
			}
		}
		return true
	})
}
