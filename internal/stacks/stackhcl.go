package stacks

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

const stackBlockType = "stack"

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
	block := blocks[0]
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return "", nil, fmt.Errorf("stack.hcl attributes: %s", diags.Error())
	}
	nameAttr, ok := attrs["name"]
	if !ok {
		return "", nil, fmt.Errorf("stack.hcl: missing required attribute %q", "name")
	}
	nameVal, diags := nameAttr.Expr.Value(nil)
	if diags.HasErrors() {
		return "", nil, fmt.Errorf("stack.hcl name: %s", diags.Error())
	}
	if nameVal.Type() != cty.String {
		return "", nil, fmt.Errorf("stack.hcl: attribute %q must be a string", "name")
	}
	name = nameVal.AsString()
	if name == "" {
		return "", nil, fmt.Errorf("stack.hcl: attribute %q must be non-empty", "name")
	}
	dependsOn = nil
	if depAttr, ok := attrs["depends_on"]; ok {
		depVal, diags := depAttr.Expr.Value(nil)
		if diags.HasErrors() {
			return "", nil, fmt.Errorf("stack.hcl depends_on: %s", diags.Error())
		}
		if !depVal.Type().IsListType() && !depVal.Type().IsTupleType() {
			return "", nil, fmt.Errorf("stack.hcl: attribute %q must be a list of strings", "depends_on")
		}
		it := depVal.ElementIterator()
		for it.Next() {
			_, v := it.Element()
			if v.Type() != cty.String {
				return "", nil, fmt.Errorf("stack.hcl: %q must be a list of strings", "depends_on")
			}
			s := v.AsString()
			if s != "" {
				dependsOn = append(dependsOn, s)
			}
		}
	}
	return name, dependsOn, nil
}
