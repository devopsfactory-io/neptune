package stacks

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/devopsfactory-io/neptune/internal/domain"
)

func TestTerramateChangedStacks(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		changes    []string
		want       []string
		wantErr    string
		rootConfig string
	}{
		{
			name:  "unchanged stacks are excluded",
			files: map[string]string{"app/stack.tm.hcl": `stack { name = "app" }`},
		},
		{
			name: "after and before determine run order",
			files: map[string]string{
				"a/stack.tm.hcl": "stack {\n name = \"a\"\n after = [\"/z\"]\n}",
				"m/stack.tm.hcl": "stack {\n name = \"m\"\n before = [\"/z\"]\n}",
				"z/stack.tm.hcl": `stack { name = "z" }`,
			},
			changes: []string{"a/terragrunt.hcl", "m/terragrunt.hcl", "z/terragrunt.hcl"},
			want:    []string{"m", "z", "a"},
		},
		{
			name: "shared module and watched file select their stacks",
			files: map[string]string{
				"module-user/stack.tm.hcl": `stack { name = "module-user" }`,
				"module-user/main.tf":      "module \"shared\" {\n source = \"../modules/shared\"\n}",
				"modules/shared/main.tf":   "# shared module\n",
				"watcher/stack.tm.hcl":     "stack {\n name = \"watcher\"\n watch = [\"/shared.txt\"]\n}",
				"shared.txt":               "shared input\n",
				"unrelated/stack.tm.hcl":   `stack { name = "unrelated" }`,
			},
			changes: []string{"modules/shared/main.tf", "shared.txt"},
			want:    []string{"module-user", "watcher"},
		},
		{
			name: "wants and wanted_by do not expand Neptune changed stack selection",
			files: map[string]string{
				"app/stack.tm.hcl":  "stack {\n name = \"app\"\n wants = [\"/base\"]\n}",
				"base/stack.tm.hcl": "stack {\n name = \"base\"\n wanted_by = [\"/app\"]\n}",
			},
			changes: []string{"app/main.tf"},
			want:    []string{"app"},
		},
		{
			name:       "Terragrunt shared source changes select both consumers",
			rootConfig: "terramate {\n required_version = \">= 0.1.0\"\n config {}\n}\n",
			files: map[string]string{
				"a/stack.tm.hcl":           `stack { name = "a" }`,
				"a/terragrunt.hcl":         "terraform {\n source = \"../_modules/minimal\"\n}\n",
				"b/stack.tm.hcl":           `stack { name = "b" }`,
				"b/terragrunt.hcl":         "terraform {\n source = \"../_modules/minimal\"\n}\n",
				"_modules/minimal/main.tf": "# module\n",
			},
			changes: []string{"_modules/minimal/main.tf"},
			want:    []string{"a", "b"},
		},
		{
			name:       "Terragrunt include changes select the consuming stack",
			rootConfig: "terramate {\n required_version = \">= 0.1.0\"\n config {}\n}\n",
			files: map[string]string{
				"a/stack.tm.hcl":           `stack { name = "a" }`,
				"a/terragrunt.hcl":         "include \"root\" {\n path = find_in_parent_folders(\"root.hcl\")\n}\nterraform {\n source = \"../_modules/minimal\"\n}\n",
				"root.hcl":                 "locals { value = \"initial\" }\n",
				"_modules/minimal/main.tf": "# module\n",
			},
			changes: []string{"root.hcl"},
			want:    []string{"a"},
		},
		{
			name: "dotfile changes select the owning stack",
			files: map[string]string{
				"app/stack.tm.hcl":        `stack { name = "app" }`,
				"app/.terraform.lock.hcl": "# provider lock\n",
			},
			changes: []string{"app/.terraform.lock.hcl"},
			want:    []string{"app"},
		},
		{
			name: "ordering cycles fail",
			files: map[string]string{
				"a/stack.tm.hcl": "stack {\n name = \"a\"\n after = [\"/b\"]\n}",
				"b/stack.tm.hcl": "stack {\n name = \"b\"\n after = [\"/a\"]\n}",
			},
			changes: []string{"a/main.tf", "b/main.tf"},
			wantErr: "computing run order",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			gitCmd := func(args ...string) {
				t.Helper()
				cmd := exec.Command("git", args...)
				cmd.Dir = root
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, output)
				}
			}
			write := func(name, content string) {
				t.Helper()
				path := filepath.Join(root, name)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			gitCmd("init", "-b", "main")
			gitCmd("config", "user.name", "Neptune Test")
			gitCmd("config", "user.email", "test@neptune.invalid")
			gitCmd("config", "commit.gpgsign", "false")
			rootConfig := tt.rootConfig
			if rootConfig == "" {
				rootConfig = "terramate { required_version = \">= 0.1.0\" }\n"
			}
			write("terramate.tm.hcl", rootConfig)
			for name, content := range tt.files {
				write(name, content)
			}
			gitCmd("add", ".")
			gitCmd("commit", "-m", "base")
			gitCmd("update-ref", "refs/remotes/origin/main", "HEAD")
			for _, name := range tt.changes {
				write(name, tt.files[name]+"\n# changed\n")
			}
			if len(tt.changes) > 0 {
				gitCmd("add", ".")
				gitCmd("commit", "-m", "change")
			}
			t.Setenv("NEPTUNE_CONFIG_PATH", filepath.Join(root, ".neptune.yaml"))
			cfg := &domain.NeptuneConfig{Repository: &domain.RepositoryConfig{Branch: "main"}}
			got, err := (&TerramateProvider{}).ChangedStacks(context.Background(), cfg)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Stacks, tt.want) {
				t.Fatalf("stacks = %v, want %v", got.Stacks, tt.want)
			}
		})
	}
}
