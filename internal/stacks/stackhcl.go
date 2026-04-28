package stacks

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

const stackBlockType = "stack"

// extractHclStringAttribute extracts a required string attribute from HCL attributes.
func extractHclStringAttribute(attrs hcl.Attributes, name string) (string, error) {
	attr, ok := attrs[name]
	if !ok {
		return "", fmt.Errorf("stack.hcl: missing required attribute %q", name)
	}
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return "", fmt.Errorf("stack.hcl %s: %s", name, diags.Error())
	}
	if val.Type() != cty.String {
		return "", fmt.Errorf("stack.hcl: attribute %q must be a string", name)
	}
	s := val.AsString()
	if s == "" {
		return "", fmt.Errorf("stack.hcl: attribute %q must be non-empty", name)
	}
	return s, nil
}

// extractHclListAttribute extracts an optional list-of-strings attribute from HCL attributes.
func extractHclListAttribute(attrs hcl.Attributes, name string) ([]string, error) {
	attr, ok := attrs[name]
	if !ok {
		return nil, nil
	}
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return nil, fmt.Errorf("stack.hcl %s: %s", name, diags.Error())
	}
	if !val.Type().IsListType() && !val.Type().IsTupleType() {
		return nil, fmt.Errorf("stack.hcl: attribute %q must be a list of strings", name)
	}
	var result []string
	it := val.ElementIterator()
	for it.Next() {
		_, v := it.Element()
		if v.Type() != cty.String {
			return nil, fmt.Errorf("stack.hcl: %q must be a list of strings", name)
		}
		if s := v.AsString(); s != "" {
			result = append(result, s)
		}
	}
	return result, nil
}

// ParseStackHcl reads a stack.hcl file and returns the stack block's name and optional depends_on list.
// name is required; depends_on is optional (nil if missing). Returns an error if the file is invalid
// or has no stack block or no name.
func ParseStackHcl(path string) (name string, dependsOn []string, err error) {
	src, err := os.ReadFile(path) //nolint:gosec // G304: path from controlled stack discovery
	if err != nil {
		return "", nil, fmt.Errorf("reading stack.hcl: %w", err)
	}
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(src, path)
	if diags.HasErrors() {
		return "", nil, fmt.Errorf("parsing stack.hcl: %s", diags.Error())
	}
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: stackBlockType}},
	}
	content, diags := file.Body.Content(schema)
	if diags.HasErrors() {
		return "", nil, fmt.Errorf("reading stack.hcl body: %s", diags.Error())
	}
	blocks := content.Blocks.OfType(stackBlockType)
	if len(blocks) == 0 {
		return "", nil, fmt.Errorf("stack.hcl: no %q block found", stackBlockType)
	}
	attrs, diags := blocks[0].Body.JustAttributes()
	if diags.HasErrors() {
		return "", nil, fmt.Errorf("stack.hcl attributes: %s", diags.Error())
	}
	name, err = extractHclStringAttribute(attrs, "name")
	if err != nil {
		return "", nil, err
	}
	dependsOn, err = extractHclListAttribute(attrs, "depends_on")
	if err != nil {
		return "", nil, err
	}
	return name, dependsOn, nil
}
