package provider

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"

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

	// Get the path from the arguments
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
	_, err := os.Stat(inputs.Path)
	if err != nil {
		resp.Error = function.NewArgumentFuncError(1, "Error retrieving path information")
		tflog.Error(ctx, "Error retrieving path information", map[string]interface{}{"path": inputs.Path, "error": err.Error()})
		return
	}

	var owner string

	// On Unix-based systems, use system calls to get UID and GID without Sys()
	if runtime.GOOS != "windows" {
		// Retrieve the user and group via command-line tools (like `stat` or `ls -l`)
		// Using "stat" command (POSIX) to get user and group information
		cmd := exec.Command("stat", "-c", "%U:%G", inputs.Path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			resp.Error = function.NewArgumentFuncError(1, "Error retrieving file owner information")
			tflog.Error(ctx, "Error retrieving file owner information", map[string]interface{}{"path": inputs.Path, "error": err.Error()})
			return
		}

		// Parse the output of the stat command
		parts := strings.Split(string(output), ":")
		if len(parts) != 2 {
			resp.Error = function.NewArgumentFuncError(1, "Invalid stat output")
			tflog.Error(ctx, "Invalid stat output", map[string]interface{}{"path": inputs.Path, "output": string(output)})
			return
		}

		// Assign the parsed user and group
		owner = parts[0]
	} else {
		// For Windows systems, set the owner to "unknown"
		owner = "unknown"
	}

	tflog.Debug(ctx, "File owner retrieved", map[string]interface{}{
		"path":  inputs.Path,
		"owner": owner,
	})

	// Set the result (owner name)
	if err := resp.Result.Set(ctx, types.StringValue(owner)); err != nil {
		resp.Error = function.NewArgumentFuncError(1, "Failed to set result")
		tflog.Error(ctx, "Failed to set result", map[string]interface{}{"error": err.Error()})
	}
}
