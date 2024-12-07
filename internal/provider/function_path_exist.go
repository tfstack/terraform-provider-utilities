package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ function.Function = &PathExistsFunction{}

type PathExistsFunction struct{}

func NewFunctionPathExists() function.Function {
	return &PathExistsFunction{}
}

func (f *PathExistsFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "path_exists"
}

func (f *PathExistsFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Checks if a given file or directory path exists.",
		MarkdownDescription: `
Validates whether the specified file or directory exists on the system.

### Parameters
- **path**: The file or directory path to check.

### Returns
- **bool**: A boolean indicating if the path exists (true) or not (false).
`,
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "path",
				MarkdownDescription: "Path to the file or directory.",
			},
		},
		Return: function.BoolReturn{},
	}
}

func (f *PathExistsFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
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

	resp.Error = resp.Result.Set(ctx, types.BoolValue(exists))
}
