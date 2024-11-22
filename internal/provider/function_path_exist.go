package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var pathExistReturnAttrTypes = map[string]attr.Type{
	"exists": types.BoolType,
}

var _ function.Function = &PathExistFunction{}

type PathExistFunction struct{}

func NewFunctionPathExist() function.Function {
	return &PathExistFunction{}
}

func (f *PathExistFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "path_exists"
}

func (f *PathExistFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Checks if a given file or directory path exists.",
		MarkdownDescription: `
Validates whether the specified file or directory exists on the system.

### Parameters
- **path**: The file or directory path to check.

### Returns
- **exists**: A boolean indicating if the path exists (true) or not (false).
`,
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "path",
				MarkdownDescription: "Path to the file or directory.",
			},
		},
		Return: function.ObjectReturn{
			AttributeTypes: pathExistReturnAttrTypes,
		},
	}
}

func (f *PathExistFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var inputs struct {
		Path string `tfsdk:"path"`
	}

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &inputs.Path))
	if resp.Error != nil {
		return
	}

	tflog.Debug(ctx, "Checking if path exists", map[string]interface{}{
		"path": inputs.Path,
	})

	_, err := os.Stat(inputs.Path)
	exists := err == nil || !os.IsNotExist(err)

	pathExistsObj, diags := types.ObjectValue(
		pathExistReturnAttrTypes,
		map[string]attr.Value{
			"exists": types.BoolValue(exists),
		},
	)

	resp.Error = function.FuncErrorFromDiags(ctx, diags)
	if resp.Error != nil {
		return
	}

	resp.Error = resp.Result.Set(ctx, &pathExistsObj)
}
