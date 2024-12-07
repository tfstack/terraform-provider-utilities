package provider

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"syscall"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ function.Function = &PathOwnerFunction{}

type PathOwnerFunction struct{}

func NewFunctionPathOwner() function.Function {
	return &PathOwnerFunction{}
}

func (f *PathOwnerFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "path_owner"
}

func (f *PathOwnerFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Retrieves the owner of a given file or directory path.",
		MarkdownDescription: `
Returns the owner (username) of the specified file or directory path.

### Parameters
- **path**: The file or directory path to check.

### Returns
- **string**: The username of the file or directory owner.
`,
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "path",
				MarkdownDescription: "Path to the file or directory.",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *PathOwnerFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var inputs struct {
		Path string `tfsdk:"path"`
	}

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &inputs.Path))
	if resp.Error != nil {
		return
	}

	tflog.Debug(ctx, "Checking path owner", map[string]interface{}{
		"path": inputs.Path,
	})

	// Check if the path is empty and return an error
	if inputs.Path == "" {
		resp.Error = function.NewArgumentFuncError(1, "Path cannot be empty")
		tflog.Error(ctx, "Path is empty", nil)
		return
	}

	// Get file information
	fileInfo, err := os.Stat(inputs.Path)
	if err != nil {
		resp.Error = function.NewArgumentFuncError(1, "Error retrieving path information")
		tflog.Error(ctx, "Error retrieving path information", map[string]interface{}{"path": inputs.Path, "error": err.Error()})
		return
	}

	// Retrieve the system-specific file info
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		resp.Error = function.NewArgumentFuncError(1, "Failed to retrieve file system information")
		tflog.Error(ctx, "Failed to retrieve file system information", map[string]interface{}{"path": inputs.Path})
		return
	}

	// Convert UID to a string
	uid := fmt.Sprintf("%d", stat.Uid)

	// Lookup user by UID
	usr, err := user.LookupId(uid)
	if err != nil {
		resp.Error = function.NewArgumentFuncError(1, "Error retrieving file owner information")
		tflog.Error(ctx, "Error retrieving file owner information", map[string]interface{}{"path": inputs.Path, "uid": uid, "error": err.Error()})
		return
	}

	tflog.Debug(ctx, "File owner retrieved", map[string]interface{}{
		"path":  inputs.Path,
		"owner": usr.Username,
	})

	// Set the result (owner name)
	if err := resp.Result.Set(ctx, types.StringValue(usr.Username)); err != nil {
		resp.Error = function.NewArgumentFuncError(1, "Failed to set result")
		tflog.Error(ctx, "Failed to set result", map[string]interface{}{"error": err.Error()})
	}
}
