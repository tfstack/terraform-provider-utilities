package provider

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceLocalDirectory struct{}

func NewDataSourceLocalDirectory() datasource.DataSource {
	return &dataSourceLocalDirectory{}
}

func (d *dataSourceLocalDirectory) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// No configuration required for this data source
}

func (d *dataSourceLocalDirectory) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "utilities_local_directory"
}

func (d *dataSourceLocalDirectory) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var pathAttribute types.String
	diags := req.Config.GetAttribute(ctx, path.Root("path"), &pathAttribute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	directoryPath := pathAttribute.ValueString()

	exists := false
	permissions := ""
	userName := ""
	groupName := ""
	id := directoryPath

	info, err := os.Stat(directoryPath)
	if err == nil && info.IsDir() {
		exists = true

		permissions = fmt.Sprintf("%04o", info.Mode().Perm())

		if sysInfo, ok := info.Sys().(*syscall.Stat_t); ok {
			userObj, userErr := user.LookupId(strconv.Itoa(int(sysInfo.Uid)))
			if userErr == nil {
				userName = userObj.Username
			} else {
				userName = strconv.Itoa(int(sysInfo.Uid))
			}

			groupObj, groupErr := user.LookupGroupId(strconv.Itoa(int(sysInfo.Gid)))
			if groupErr == nil {
				groupName = groupObj.Name
			} else {
				groupName = strconv.Itoa(int(sysInfo.Gid))
			}
		}
	}

	state := struct {
		Id          types.String `tfsdk:"id"`
		Exists      types.Bool   `tfsdk:"exists"`
		Path        types.String `tfsdk:"path"`
		Permissions types.String `tfsdk:"permissions"`
		User        types.String `tfsdk:"user"`
		Group       types.String `tfsdk:"group"`
	}{
		Id:          types.StringValue(id),
		Exists:      types.BoolValue(exists),
		Path:        types.StringValue(directoryPath),
		Permissions: types.StringValue(permissions),
		User:        types.StringValue(userName),
		Group:       types.StringValue(groupName),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *dataSourceLocalDirectory) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides information about a local directory, including its metadata and permissions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the local directory, which is the same as the path.",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path to the local directory.",
				Required:            true,
			},
			"exists": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the directory exists.",
				Computed:            true,
			},
			"permissions": schema.StringAttribute{
				MarkdownDescription: "Permissions of the directory in octal format (e.g., 0755).",
				Computed:            true,
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "The name of the user owning the directory.",
				Computed:            true,
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "The name of the group owning the directory.",
				Computed:            true,
			},
		},
	}
}