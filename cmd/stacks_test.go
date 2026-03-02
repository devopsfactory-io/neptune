package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWriteStacksListOutput_Text(t *testing.T) {
	var buf bytes.Buffer
	paths := []string{"stack-a", "stack-b"}
	err := writeStacksListOutput(&buf, formatText, paths)
	if err != nil {
		t.Fatalf("writeStacksListOutput: %v", err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), out)
	}
	if lines[0] != "stack-a" || lines[1] != "stack-b" {
		t.Errorf("expected stack-a and stack-b, got %q", out)
	}
}

func TestWriteStacksListOutput_JSON(t *testing.T) {
	var buf bytes.Buffer
	paths := []string{"stack-a", "stack-b"}
	err := writeStacksListOutput(&buf, formatJSON, paths)
	if err != nil {
		t.Fatalf("writeStacksListOutput: %v", err)
	}
	var got stacksListOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(got.Stacks) != 2 || got.Stacks[0] != "stack-a" || got.Stacks[1] != "stack-b" {
		t.Errorf("expected stacks [stack-a stack-b], got %v", got.Stacks)
	}
}

func TestWriteStacksListOutput_YAML(t *testing.T) {
	var buf bytes.Buffer
	paths := []string{"stack-a", "stack-b"}
	err := writeStacksListOutput(&buf, formatYAML, paths)
	if err != nil {
		t.Fatalf("writeStacksListOutput: %v", err)
	}
	var got stacksListOutput
	if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if len(got.Stacks) != 2 || got.Stacks[0] != "stack-a" || got.Stacks[1] != "stack-b" {
		t.Errorf("expected stacks [stack-a stack-b], got %v", got.Stacks)
	}
}

func TestWriteStacksListOutput_Formatted(t *testing.T) {
	var buf bytes.Buffer
	paths := []string{"stack-a", "stack-b"}
	err := writeStacksListOutput(&buf, formatFormatted, paths)
	if err != nil {
		t.Fatalf("writeStacksListOutput: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Neptune Stacks") {
		t.Errorf("output should contain title, got %q", out)
	}
	if !strings.Contains(out, "stack-a") || !strings.Contains(out, "stack-b") {
		t.Errorf("output should contain stack paths, got %q", out)
	}
}

func TestWriteStacksListOutput_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	err := writeStacksListOutput(&buf, "invalid", []string{"a"})
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestWriteStackCreateOutput_Text(t *testing.T) {
	var buf bytes.Buffer
	err := writeStackCreateOutput(&buf, formatText, "my-stack")
	if err != nil {
		t.Fatalf("writeStackCreateOutput: %v", err)
	}
	out := strings.TrimSuffix(buf.String(), "\n")
	if out != "my-stack" {
		t.Errorf("expected my-stack, got %q", out)
	}
}

func TestWriteStackCreateOutput_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := writeStackCreateOutput(&buf, formatJSON, "my-stack")
	if err != nil {
		t.Fatalf("writeStackCreateOutput: %v", err)
	}
	var got stackCreateOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if got.Path != "my-stack" {
		t.Errorf("expected path my-stack, got %q", got.Path)
	}
}

func TestWriteStackCreateOutput_YAML(t *testing.T) {
	var buf bytes.Buffer
	err := writeStackCreateOutput(&buf, formatYAML, "my-stack")
	if err != nil {
		t.Fatalf("writeStackCreateOutput: %v", err)
	}
	var got stackCreateOutput
	if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if got.Path != "my-stack" {
		t.Errorf("expected path my-stack, got %q", got.Path)
	}
}

func TestWriteStackCreateOutput_Formatted(t *testing.T) {
	var buf bytes.Buffer
	err := writeStackCreateOutput(&buf, formatFormatted, "my-stack")
	if err != nil {
		t.Fatalf("writeStackCreateOutput: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Neptune Stack Created") {
		t.Errorf("output should contain title, got %q", out)
	}
	if !strings.Contains(out, "my-stack") {
		t.Errorf("output should contain path, got %q", out)
	}
}

func TestWriteStackCreateOutput_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	err := writeStackCreateOutput(&buf, "invalid", "a")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestIsValidFormat(t *testing.T) {
	for _, f := range validFormats {
		if !isValidFormat(f) {
			t.Errorf("isValidFormat(%q) = false, want true", f)
		}
	}
	if isValidFormat("invalid") {
		t.Error("isValidFormat(\"invalid\") = true, want false")
	}
}

func TestStackHclContent_NoDependsOn(t *testing.T) {
	got := stackHclContent("my-stack", nil)
	if !strings.Contains(got, "name") || !strings.Contains(got, "\"my-stack\"") {
		t.Errorf("expected name and value in output, got %q", got)
	}
	if strings.Contains(got, "\n  depends_on = [") {
		t.Errorf("expected no depends_on attribute line when nil, got %q", got)
	}
	if !strings.Contains(got, "Optional: depends_on") {
		t.Errorf("expected optional comment when no deps, got %q", got)
	}
}

func TestStackHclContent_WithDependsOn(t *testing.T) {
	got := stackHclContent("foundation/network", []string{"foundation/base", "foundation/iam"})
	if !strings.Contains(got, "foundation/network") {
		t.Errorf("expected name in output, got %q", got)
	}
	if !strings.Contains(got, "depends_on = [\"foundation/base\", \"foundation/iam\"]") {
		t.Errorf("expected depends_on line, got %q", got)
	}
	if strings.Contains(got, "Optional: depends_on") {
		t.Errorf("expected no optional comment when deps set, got %q", got)
	}
}
