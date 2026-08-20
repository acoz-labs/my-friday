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
	prohibited := map[string]bool{"net": true, "net/http": true, "net/rpc": true, "crypto/tls": true}
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
			if prohibited[name] {
				t.Errorf("production network import %q in %s", name, path)
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
			if !pkgOK || pkg.Name != "exec" || sel.Sel.Name != "Command" {
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
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
