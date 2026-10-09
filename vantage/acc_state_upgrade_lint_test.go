package vantage

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestAccStepsPersistCurrentSchema rejects acceptance tests that apply with a
// published provider and never apply again with this provider. PlanOnly runs
// the state upgrader in memory and leaves the old schema version on disk.
// Core CI checks out releases/latest and runs the suite; terraform show then
// fails because the v1 provider cannot read schema version 0.
func TestAccStepsPersistCurrentSchema(t *testing.T) {
	t.Parallel()

	violations, seen := scanAccStateUpgradeTests(t)
	if len(violations) > 0 {
		t.Fatalf("schema-upgrade acceptance tests must apply with ProtoV6ProviderFactories after ExternalProviders, or terraform show fails on the old schema version:\n%s", strings.Join(violations, "\n"))
	}

	const known = "TestAccTeam_StateUpgradeV0toV1"
	if !seen[known] {
		t.Fatalf("did not inspect %s; the schema-upgrade gate is not checking the known case", known)
	}
}

func TestAccStepsPersistCurrentSchema_examples(t *testing.T) {
	t.Parallel()

	const bad = `package vantage
func TestAccExample(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{},
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				PlanOnly:                 true,
			},
		},
	})
}`
	violations := violationsInSource("example_test.go", bad)
	if len(violations) != 1 || !strings.Contains(violations[0], "TestAccExample") {
		t.Fatalf("bad upgrade test: got %v", violations)
	}

	const good = `package vantage
func TestAccExample(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{},
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				PlanOnly:                 true,
			},
		},
	})
}`
	if violations := violationsInSource("example_test.go", good); len(violations) != 0 {
		t.Fatalf("persisted upgrade test: got %v", violations)
	}
}

func scanAccStateUpgradeTests(t *testing.T) ([]string, map[string]bool) {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	var violations []string
	seen := map[string]bool{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, "_test.go") || name == "acc_state_upgrade_lint_test.go" {
			continue
		}
		path := filepath.Join(dir, name)
		parsed, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		fileViolations, fileSeen := violationsInFile(name, parsed)
		violations = append(violations, fileViolations...)
		for fn := range fileSeen {
			seen[fn] = true
		}
	}
	return violations, seen
}

func violationsInSource(filename, src string) []string {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return []string{filename + ": " + err.Error()}
	}
	violations, _ := violationsInFile(filename, parsed)
	return violations
}

func violationsInFile(filename string, file *ast.File) ([]string, map[string]bool) {
	var violations []string
	seen := map[string]bool{}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isResourceTestCall(call) {
				return true
			}
			if !callHasIdent(call, "ExternalProviders") {
				return true
			}
			seen[fn.Name.Name] = true
			violations = append(violations, stepViolations(filename, fn.Name.Name, call)...)
			return true
		})
	}
	return violations, seen
}

func stepViolations(filename, fnName string, call *ast.CallExpr) []string {
	if len(call.Args) < 2 {
		return []string{filename + ": " + fnName + " resource.Test call is missing arguments"}
	}
	testCase := compositeLit(call.Args[1])
	if testCase == nil {
		return []string{filename + ": " + fnName + " resource.Test case is not a composite literal"}
	}
	stepsExpr, ok := compositeKey(testCase, "Steps")
	if !ok {
		return []string{filename + ": " + fnName + " resource.Test case has ExternalProviders but no Steps"}
	}
	steps := compositeLit(stepsExpr)
	if steps == nil {
		return []string{filename + ": " + fnName + " Steps is not a composite literal"}
	}

	type stepFlags struct {
		external bool
		proto    bool
		planOnly bool
	}
	parsed := make([]stepFlags, 0, len(steps.Elts))
	for _, elt := range steps.Elts {
		lit := compositeLit(elt)
		if lit == nil {
			return []string{filename + ": " + fnName + " has a Steps element that is not a composite literal"}
		}
		_, external := compositeKey(lit, "ExternalProviders")
		_, proto := compositeKey(lit, "ProtoV6ProviderFactories")
		planOnlyExpr, hasPlanOnly := compositeKey(lit, "PlanOnly")
		parsed = append(parsed, stepFlags{
			external: external,
			proto:    proto,
			planOnly: hasPlanOnly && isTrue(planOnlyExpr),
		})
	}

	var violations []string
	for i, step := range parsed {
		if !step.external || step.planOnly {
			continue
		}
		persisted := false
		for _, later := range parsed[i+1:] {
			if later.proto && !later.planOnly {
				persisted = true
				break
			}
		}
		if !persisted {
			violations = append(violations, filename+": "+fnName+" applies with ExternalProviders and never applies with ProtoV6ProviderFactories afterward")
		}
	}
	return violations
}

func isResourceTestCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Test" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "resource"
}

func callHasIdent(n ast.Node, name string) bool {
	found := false
	ast.Inspect(n, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if ok && id.Name == name {
			found = true
			return false
		}
		return !found
	})
	return found
}

func compositeLit(expr ast.Expr) *ast.CompositeLit {
	switch n := expr.(type) {
	case *ast.CompositeLit:
		return n
	case *ast.UnaryExpr:
		return compositeLit(n.X)
	case *ast.ParenExpr:
		return compositeLit(n.X)
	default:
		return nil
	}
}

func compositeKey(lit *ast.CompositeLit, name string) (ast.Expr, bool) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		id, ok := kv.Key.(*ast.Ident)
		if ok && id.Name == name {
			return kv.Value, true
		}
	}
	return nil, false
}

func isTrue(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "true"
}
