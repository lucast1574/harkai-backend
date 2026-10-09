package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := check(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("structure OK: production files <= 200 lines; functions return values")
}

func check() error {
	problems := []string{}
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.Contains(filepath.ToSlash(path), "internal/testutil/") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if lines := len(strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")); lines > 200 {
			problems = append(problems, fmt.Sprintf("%s: %d lines exceeds 200", path, lines))
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, data, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.FuncDecl:
				required := (n.Name.Name == "main" && file.Name.Name == "main" && n.Recv == nil) || (n.Name.Name == "ServeHTTP" && n.Recv != nil)
				if !required && (n.Type.Results == nil || len(n.Type.Results.List) == 0) {
					problems = append(problems, fmt.Sprintf("%s: function %s must return a value", fset.Position(n.Pos()), n.Name.Name))
				}
			case *ast.FuncLit:
				if n.Type.Results == nil || len(n.Type.Results.List) == 0 {
					problems = append(problems, fmt.Sprintf("%s: anonymous function must return a value", fset.Position(n.Pos())))
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		return err
	}
	if len(problems) > 0 {
		return fmt.Errorf("%s", strings.Join(problems, "\n"))
	}
	return nil
}
