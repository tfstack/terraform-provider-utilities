// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ function.Function = &PathPermissionFunction{}

type PathPermissionFunction struct{}

func NewFunctionPathPermission() function.Function {
	return &PathPermissionFunction{}
}

func (f *PathPermissionFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "path_permission"
}

func (f *PathPermissionFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Checks the permissions of a given file or directory path.",
		MarkdownDescription: `Returns the permissions (read, write, execute) of the specified file or directory path for different user types: owner, group, and others.

### Parameters
- **path**: The file or directory path to check.

### Returns
- **string**: The file or directory permissions in octal format (e.g., "0755").`,
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "path",
				MarkdownDescription: "Path to the file or directory.",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *PathPermissionFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var inputs struct {
		Path string `tfsdk:"path"`
	}

	// Retrieve the path argument
	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &inputs.Path))
	if resp.Error != nil {
		return
	}

	// Check if the path is empty and return an error
	if inputs.Path == "" {
		resp.Error = function.NewArgumentFuncError(1, "Path cannot be empty")
		tflog.Error(ctx, "Path is empty", nil)
		return
	}

	tflog.Debug(ctx, "Checking path permissions", map[string]interface{}{"path": inputs.Path})

	// Get file information
	fileInfo, err := os.Stat(inputs.Path)
	if err != nil {
		resp.Error = function.NewArgumentFuncError(1, fmt.Sprintf("Error retrieving path information: %s", err.Error()))
		tflog.Error(ctx, "Error retrieving path information", map[string]interface{}{"path": inputs.Path, "error": err.Error()})
		return
	}

	// Get file permissions in octal format
	permissions := fmt.Sprintf("%04o", fileInfo.Mode().Perm())

	// Log permissions for debugging purposes
	tflog.Debug(ctx, "File permissions retrieved", map[string]interface{}{"path": inputs.Path, "permissions": permissions})

	// Set the result as the permission string
	if err := resp.Result.Set(ctx, permissions); err != nil {
		resp.Error = function.NewArgumentFuncError(1, "Failed to set result")
		tflog.Error(ctx, "Failed to set result", map[string]interface{}{"error": err.Error()})
	}
}
