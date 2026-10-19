package agentcore

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// A seam is named pkg.Name for a function or a type and pkg.Type.Name for a
// method or a struct field. The Seams section of doc.go lists them one per
// line as "//   - `name`: what it lets the app supply", and these tests hold
// the list to the code in both directions.
var seamLine = regexp.MustCompile("^//\\s+- `([A-Za-z0-9_.]+)`")

// listedSeams reads the names doc.go lists under its Seams heading.
func listedSeams(t *testing.T) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile("doc.go")
	if err != nil {
		t.Fatal(err)
	}
	listed := map[string]bool{}
	in := false
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "// # Seams") {
			in = true
			continue
		}
		if !in {
			continue
		}
		if m := seamLine.FindStringSubmatch(line); m != nil {
			listed[m[1]] = true
		}
	}
	if len(listed) == 0 {
		t.Fatal("doc.go lists no seams under a Seams heading")
	}
	return listed
}

// declared returns every exported name the module declares, and the subset
// that is an extension point by shape: an exported Set function or method, and
// an exported struct field whose type is a function.
func declared(t *testing.T) (all, shaped map[string]bool) {
	t.Helper()
	all, shaped = map[string]bool{}, map[string]bool{}
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != "." && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		pkg := file.Name.Name
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				name := pkg + "." + d.Name.Name
				if d.Recv != nil {
					name = pkg + "." + recvName(d.Recv.List[0].Type) + "." + d.Name.Name
				}
				all[name] = true
				if strings.HasPrefix(d.Name.Name, "Set") {
					shaped[name] = true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || !ts.Name.IsExported() {
						continue
					}
					all[pkg+"."+ts.Name.Name] = true
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}
					for _, f := range st.Fields.List {
						_, isFunc := f.Type.(*ast.FuncType)
						for _, n := range f.Names {
							if !n.IsExported() {
								continue
							}
							name := pkg + "." + ts.Name.Name + "." + n.Name
							all[name] = true
							if isFunc {
								shaped[name] = true
							}
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return all, shaped
}

func recvName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.StarExpr:
		return recvName(x.X)
	case *ast.IndexExpr:
		return recvName(x.X)
	case *ast.Ident:
		return x.Name
	}
	return "?"
}

func TestEverySeamInTheCodeIsListed(t *testing.T) {
	listed := listedSeams(t)
	_, shaped := declared(t)
	var missing []string
	for name := range shaped {
		if !listed[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("seams in the code that doc.go does not list:\n%s", strings.Join(missing, "\n"))
	}
}

func TestEverySeamListedExists(t *testing.T) {
	listed := listedSeams(t)
	all, _ := declared(t)
	var gone []string
	for name := range listed {
		if !all[name] {
			gone = append(gone, name)
		}
	}
	sort.Strings(gone)
	if len(gone) > 0 {
		t.Errorf("seams doc.go lists that the code does not declare:\n%s", strings.Join(gone, "\n"))
	}
}
