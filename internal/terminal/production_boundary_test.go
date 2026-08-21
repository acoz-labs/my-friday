package terminal

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestProductionNetworkAndSubprocessBoundary(t *testing.T) {
	root := filepath.Join("..", "..")
	allowedImports := map[string]bool{
		"bufio": true, "bytes": true, "crypto/rand": true, "crypto/sha256": true, "encoding/hex": true, "encoding/json": true,
		"errors": true, "fmt": true, "io": true, "io/fs": true, "os": true, "os/exec": true, "path/filepath": true,
		"runtime": true, "slices": true, "sort": true, "strconv": true, "strings": true, "syscall": true,
		"unicode": true, "unicode/utf8": true, "unsafe": true,
		"github.com/acoz-labs/my-friday/internal/environment": true,
		"github.com/acoz-labs/my-friday/internal/codexhome":   true,
		"github.com/acoz-labs/my-friday/internal/gitexec":     true,
		"github.com/acoz-labs/my-friday/internal/plan":        true,
		"github.com/acoz-labs/my-friday/internal/profile":     true,
		"github.com/acoz-labs/my-friday/internal/repository":  true,
		"github.com/acoz-labs/my-friday/internal/terminal":    true,
		"github.com/acoz-labs/my-friday/internal/transaction": true,
		"github.com/rivo/uniseg":                              true, "github.com/santhosh-tekuri/jsonschema/v6": true,
		"golang.org/x/text/unicode/norm": true,
		"golang.org/x/sys/unix":          true,
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && path != root && !strings.HasPrefix(path, filepath.Join(root, "cmd")) && !strings.HasPrefix(path, filepath.Join(root, "internal")) {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			name, _ := strconv.Unquote(spec.Path.Value)
			if !allowedImports[name] {
				t.Errorf("production import %q is not allowlisted in %s", name, path)
			}
			if name == "os/exec" && spec.Name != nil {
				t.Errorf("aliased or dot-imported os/exec in %s", path)
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, pkgOK := sel.X.(*ast.Ident)
			if !pkgOK || pkg.Name != "exec" {
				return true
			}
			if sel.Sel.Name != "Command" {
				t.Errorf("non-allowlisted os/exec API %s in %s", sel.Sel.Name, path)
				return true
			}
			literal, literalOK := call.Args[0].(*ast.BasicLit)
			if !literalOK {
				t.Errorf("dynamic subprocess in %s", path)
				return true
			}
			command, _ := strconv.Unquote(literal.Value)
			if command != "git" {
				t.Errorf("non-Git subprocess in %s", path)
			}
			if len(call.Args) < 2 {
				t.Errorf("Git invocation without explicit argv in %s", path)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
