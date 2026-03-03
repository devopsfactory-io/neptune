package stacks

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/devopsfactory-io/neptune/internal/domain"
)

func TestParseStackHcl(t *testing.T) {
	dir := t.TempDir()

	t.Run("name only", func(t *testing.T) {
		path := filepath.Join(dir, "name_only.hcl")
		if err := os.WriteFile(path, []byte(`stack { name = "my-stack" }`), 0644); err != nil {
			t.Fatal(err)
		}
		name, deps, err := ParseStackHcl(path)
		if err != nil {
			t.Fatal(err)
		}
		if name != "my-stack" {
			t.Errorf("name = %q, want my-stack", name)
		}
		if deps != nil {
			t.Errorf("dependsOn = %v, want nil", deps)
		}
	})

	t.Run("name and depends_on", func(t *testing.T) {
		path := filepath.Join(dir, "with_deps.hcl")
		content := `stack {
  name       = "stack-1"
  depends_on = ["stack-2", "foundation"]
}
`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		name, deps, err := ParseStackHcl(path)
		if err != nil {
			t.Fatal(err)
		}
		if name != "stack-1" {
			t.Errorf("name = %q, want stack-1", name)
		}
		want := []string{"stack-2", "foundation"}
		if !reflect.DeepEqual(deps, want) {
			t.Errorf("dependsOn = %v, want %v", deps, want)
		}
	})

	t.Run("invalid HCL", func(t *testing.T) {
		path := filepath.Join(dir, "bad.hcl")
		if err := os.WriteFile(path, []byte(`stack { name = `), 0644); err != nil {
			t.Fatal(err)
		}
		_, _, err := ParseStackHcl(path)
		if err == nil {
			t.Fatal("expected error for invalid HCL")
		}
	})

	t.Run("no stack block", func(t *testing.T) {
		path := filepath.Join(dir, "no_block.hcl")
		if err := os.WriteFile(path, []byte(`foo { name = "x" }`), 0644); err != nil {
			t.Fatal(err)
		}
		_, _, err := ParseStackHcl(path)
		if err == nil {
			t.Fatal("expected error when no stack block")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		path := filepath.Join(dir, "no_name.hcl")
		if err := os.WriteFile(path, []byte(`stack { depends_on = [] }`), 0644); err != nil {
			t.Fatal(err)
		}
		_, _, err := ParseStackHcl(path)
		if err == nil {
			t.Fatal("expected error when name missing")
		}
	})
}

func TestResolveDepPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "repo")
	// stack at app/stack-a; ../foundation -> foundation
	t.Run("relative parent", func(t *testing.T) {
		// From app/stack-a, ../foundation resolves to app/foundation (one level up from stack-a).
		got := resolveDepPath(root, "app/stack-a", "../foundation")
		want := "app/foundation"
		if got != want {
			t.Errorf("resolveDepPath(..., \"app/stack-a\", \"../foundation\") = %q, want %q", got, want)
		}
	})
	t.Run("relative parent to repo root", func(t *testing.T) {
		// From stack-a at repo root, ../foundation would be outside repo; from stack-a (one level), ../ = root.
		got := resolveDepPath(root, "stack-a", "../foundation")
		want := "foundation"
		if got != want {
			t.Errorf("resolveDepPath(..., \"stack-a\", \"../foundation\") = %q, want %q", got, want)
		}
	})
	t.Run("repo-root-relative", func(t *testing.T) {
		got := resolveDepPath(root, "stack-a", "stack-b")
		want := "stack-b"
		if got != want {
			t.Errorf("resolveDepPath(..., \"stack-a\", \"stack-b\") = %q, want %q", got, want)
		}
	})
	t.Run("relative same dir", func(t *testing.T) {
		got := resolveDepPath(root, "app/stack-a", "./other")
		want := "app/stack-a/other"
		if got != want {
			t.Errorf("resolveDepPath(..., \"app/stack-a\", \"./other\") = %q, want %q", got, want)
		}
	})
}

func TestExpandDirDeps(t *testing.T) {
	allPaths := []string{"foundation/base", "foundation/networking", "app/frontend", "standalone"}
	t.Run("directory dep", func(t *testing.T) {
		got := expandDirDeps(allPaths, []string{"foundation"})
		want := []string{"foundation/base", "foundation/networking"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expandDirDeps(foundation) = %v, want %v", got, want)
		}
	})
	t.Run("exact stack dep", func(t *testing.T) {
		got := expandDirDeps(allPaths, []string{"standalone"})
		want := []string{"standalone"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expandDirDeps(standalone) = %v, want %v", got, want)
		}
	})
	t.Run("multiple deps", func(t *testing.T) {
		got := expandDirDeps(allPaths, []string{"foundation", "standalone"})
		want := []string{"foundation/base", "foundation/networking", "standalone"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expandDirDeps = %v, want %v", got, want)
		}
	})
}

func TestTopologicalOrder(t *testing.T) {
	t.Run("simple order", func(t *testing.T) {
		entries := []domain.StackEntry{
			{Path: "stack-a", DependsOn: []string{"stack-b"}},
			{Path: "stack-b", DependsOn: nil},
		}
		order, err := topologicalOrder(entries)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"stack-b", "stack-a"}
		if !reflect.DeepEqual(order, want) {
			t.Errorf("order = %v, want %v", order, want)
		}
	})
	t.Run("cycle", func(t *testing.T) {
		entries := []domain.StackEntry{
			{Path: "stack-a", DependsOn: []string{"stack-b"}},
			{Path: "stack-b", DependsOn: []string{"stack-a"}},
		}
		_, err := topologicalOrder(entries)
		if err == nil {
			t.Fatal("expected error for cycle")
		}
	})
}

func TestDiscoverStackHclOrdered(t *testing.T) {
	dir := t.TempDir()
	mkStack := func(path, content string) {
		full := filepath.Join(dir, filepath.FromSlash(path), "stack.hcl")
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("two stacks with depends_on", func(t *testing.T) {
		mkStack("stack-b", `stack { name = "stack-b" }`)
		mkStack("stack-a", "stack {\n  name       = \"stack-a\"\n  depends_on = [\"stack-b\"]\n}\n")
		order, err := discoverStackHclOrdered(dir)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"stack-b", "stack-a"}
		if !reflect.DeepEqual(order, want) {
			t.Errorf("order = %v, want %v", order, want)
		}
	})

	t.Run("directory dep", func(t *testing.T) {
		// app/frontend depends on ../foundation which resolves to app/foundation (stacks under that dir).
		mkStack("app/foundation/base", `stack { name = "base" }`)
		mkStack("app/foundation/networking", `stack { name = "networking" }`)
		mkStack("app/frontend", "stack {\n  name       = \"frontend\"\n  depends_on = [\"../foundation\"]\n}\n")
		order, err := discoverStackHclOrdered(dir)
		if err != nil {
			t.Fatal(err)
		}
		// app/foundation/base and app/foundation/networking must come before app/frontend
		idx := func(s string) int {
			for i, p := range order {
				if p == s {
					return i
				}
			}
			return -1
		}
		if idx("app/frontend") <= idx("app/foundation/base") || idx("app/frontend") <= idx("app/foundation/networking") {
			t.Errorf("app/frontend should come after foundation stacks; order = %v", order)
		}
	})

	t.Run("cycle returns error", func(t *testing.T) {
		sub := t.TempDir()
		mkStack2 := func(path, content string) {
			full := filepath.Join(sub, filepath.FromSlash(path), "stack.hcl")
			os.MkdirAll(filepath.Dir(full), 0755)
			os.WriteFile(full, []byte(content), 0644)
		}
		mkStack2("x", "stack {\n  name       = \"x\"\n  depends_on = [\"y\"]\n}\n")
		mkStack2("y", "stack {\n  name       = \"y\"\n  depends_on = [\"x\"]\n}\n")
		_, err := discoverStackHclOrdered(sub)
		if err == nil {
			t.Fatal("expected error for cycle")
		}
	})
}
