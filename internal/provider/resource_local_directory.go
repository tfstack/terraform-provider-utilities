package provider

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = (*resourceUtilitiesLocalDirectory)(nil)
)

type resourceUtilitiesLocalDirectory struct{}

type LocalDirectory struct {
	Force       types.Bool   `tfsdk:"force"`
	Group       types.String `tfsdk:"user"`
	ID          types.String `tfsdk:"id"`
	Managed     types.Bool   `tfsdk:"managed"`
	Path        types.String `tfsdk:"path"`
	Permissions types.String `tfsdk:"permissions"`
	User        types.String `tfsdk:"group"`
}

var protectedPaths = []string{
	"/",          // Root directory
	"/etc",       // System configuration files
	"/usr",       // User programs
	"/var/lib",   // Application data, databases
	"/bin",       // Essential binaries
	"/sbin",      // System binaries
	"/boot",      // Boot loader files
	"/proc",      // Kernel and process information
	"/sys",       // System files
	"/dev",       // Device files
	"/lib",       // Libraries
	"/opt",       // Optional software
	"/tmp",       // Temporary files
	"/var/run",   // Runtime files, PID files
	"/var/lock",  // Lock files
	"/var/cache", // Cache data
	"/var/log",   // Log files
	"/home",      // User home directories (though some may be non-critical depending on use)
	"/root",      // Root user's home directory
	"/mnt",       // Mount points (should not be deleted unless intentionally unmounted)
	"/media",     // Removable media mounts (e.g., USB drives)
	"/srv",       // Data for services
	"/var/spool", // Spool directories (e.g., mail, cron jobs)
	"/var/tmp",   // Temporary files that persist between reboots
	"/libexec",   // Helper programs typically used by system services
}

func NewResourceUtilitiesLocalDirectory() resource.Resource {
	return &resourceUtilitiesLocalDirectory{}
}

func (r *resourceUtilitiesLocalDirectory) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *resourceUtilitiesLocalDirectory) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "utilities_local_directory"
}

func (r *resourceUtilitiesLocalDirectory) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if runtime.GOOS == "windows" {
		resp.Diagnostics.AddError(
			"Incompatible Platform",
			"This resource cannot be used on Windows systems.",
		)
		return
	}

	var data LocalDirectory

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	directoryPath := data.Path.ValueString()
	info, err := os.Stat(directoryPath)

	if err == nil && info.IsDir() {
		data.ID = types.StringValue(directoryPath)
		data.Managed = types.BoolValue(false)

		tflog.Info(ctx, "Directory already exists", map[string]interface{}{
			"path":    directoryPath,
			"managed": false,
		})
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
			resp.Diagnostics.AddError(
				"Error Creating Directory",
				err.Error(),
			)
			return
		}

		if data.Permissions.ValueString() != "" {
			var perm int
			_, err := fmt.Sscanf(data.Permissions.ValueString(), "%o", &perm)
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid Permissions",
					fmt.Sprintf("Failed to parse permissions: %v", err),
				)
				return
			}

			if err := os.Chmod(directoryPath, os.FileMode(perm)); err != nil {
				resp.Diagnostics.AddError(
					"Error Setting Permissions",
					fmt.Sprintf("Failed to set directory permissions: %v", err),
				)
				return
			}
		}

		userName := data.User.ValueString()
		groupName := data.Group.ValueString()
		if userName != "" || groupName != "" {
			var uid, gid int

			if userName != "" {
				userInfo, err := user.Lookup(userName)
				if err != nil {
					resp.Diagnostics.AddError(
						"Invalid User",
						fmt.Sprintf("Failed to lookup user: %v", err),
					)
					return
				}
				uid, err = strconv.Atoi(userInfo.Uid)
				if err != nil {
					resp.Diagnostics.AddError(
						"Invalid User ID",
						fmt.Sprintf("Failed to convert user ID to integer: %v", err),
					)
					return
				}
			}

			if groupName != "" {
				groupInfo, err := user.LookupGroup(groupName)
				if err != nil {
					resp.Diagnostics.AddError(
						"Invalid Group",
						fmt.Sprintf("Failed to lookup group: %v", err),
					)
					return
				}
				gid, err = strconv.Atoi(groupInfo.Gid)
				if err != nil {
					resp.Diagnostics.AddError(
						"Invalid Group ID",
						fmt.Sprintf("Failed to convert group ID to integer: %v", err),
					)
					return
				}
			}

			if err := os.Chown(directoryPath, uid, gid); err != nil {
				resp.Diagnostics.AddError(
					"Error Setting Owner",
					fmt.Sprintf("Failed to set directory owner: %v", err),
				)
				return
			}
		}

		data.ID = types.StringValue(directoryPath)
		data.Managed = types.BoolValue(true)

		tflog.Info(ctx, "Created local directory", map[string]interface{}{
			"path":    directoryPath,
			"managed": true,
		})
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Checking Directory",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceUtilitiesLocalDirectory) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if runtime.GOOS == "windows" {
		resp.Diagnostics.AddError(
			"Incompatible Platform",
			"This resource cannot be used on Windows systems.",
		)
		return
	}

	var data LocalDirectory

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	directoryPath := data.Path.ValueString()
	info, err := os.Stat(directoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Directory",
			err.Error(),
		)
		return
	}

	if !info.IsDir() {
		resp.Diagnostics.AddError(
			"Invalid Directory Path",
			"The path exists but is not a directory.",
		)
		return
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		resp.Diagnostics.AddError(
			"Error Reading Directory Metadata",
			"Unable to retrieve directory metadata for ownership and permissions.",
		)
		return
	}

	uid := stat.Uid
	gid := stat.Gid

	userInfo, err := user.LookupId(fmt.Sprintf("%d", uid))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving User Info",
			fmt.Sprintf("Failed to lookup user by UID %d: %v", uid, err),
		)
		return
	}
	groupInfo, err := user.LookupGroupId(fmt.Sprintf("%d", gid))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving Group Info",
			fmt.Sprintf("Failed to lookup group by GID %d: %v", gid, err),
		)
		return
	}

	data.User = types.StringValue(userInfo.Username)
	data.Group = types.StringValue(groupInfo.Name)

	mode := info.Mode().Perm()
	data.Permissions = types.StringValue(fmt.Sprintf("%04o", mode))

	tflog.Info(ctx, "Read local directory", map[string]interface{}{
		"path":        directoryPath,
		"managed":     data.Managed.ValueBool(),
		"force":       data.Force.ValueBool(),
		"user":        userInfo.Username,
		"group":       groupInfo.Name,
		"permissions": data.Permissions.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceUtilitiesLocalDirectory) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if runtime.GOOS == "windows" {
		resp.Diagnostics.AddError(
			"Incompatible Platform",
			"This resource cannot be used on Windows systems.",
		)
		return
	}

	var data LocalDirectory

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	directoryPath := data.Path.ValueString()
	info, err := os.Stat(directoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
				resp.Diagnostics.AddError(
					"Error Creating Directory",
					err.Error(),
				)
				return
			}
			data.Managed = types.BoolValue(true)
		} else {
			resp.Diagnostics.AddError(
				"Error Accessing Directory",
				err.Error(),
			)
			return
		}
	}

	if !info.IsDir() {
		resp.Diagnostics.AddError(
			"Invalid Directory Path",
			"The path exists but is not a directory.",
		)
		return
	}

	if data.Permissions.ValueString() != "" {
		var perm int
		_, err := fmt.Sscanf(data.Permissions.ValueString(), "%o", &perm)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Permissions",
				fmt.Sprintf("Failed to parse permissions: %v", err),
			)
			return
		}

		if err := os.Chmod(directoryPath, os.FileMode(perm)); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Permissions",
				fmt.Sprintf("Failed to set directory permissions: %v", err),
			)
			return
		}
	}

	userName := data.User.ValueString()
	groupName := data.Group.ValueString()

	var uid, gid int

	if userName != "" {
		userInfo, err := user.Lookup(userName)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid User",
				fmt.Sprintf("Failed to lookup user '%s': %v", userName, err),
			)
			return
		}
		uid, err = strconv.Atoi(userInfo.Uid)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid User ID",
				fmt.Sprintf("Failed to convert user ID '%s' to integer: %v", userInfo.Uid, err),
			)
			return
		}
	} else {
		infoSys := info.Sys()
		stat, ok := infoSys.(*syscall.Stat_t)
		if !ok {
			resp.Diagnostics.AddError(
				"Invalid File Information",
				"Unable to retrieve system-specific file statistics for current user.",
			)
			return
		}
		uid = int(stat.Uid)
	}

	if groupName != "" {
		groupInfo, err := user.LookupGroup(groupName)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Group",
				fmt.Sprintf("Failed to lookup group '%s': %v", groupName, err),
			)
			return
		}
		gid, err = strconv.Atoi(groupInfo.Gid)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Group ID",
				fmt.Sprintf("Failed to convert group ID '%s' to integer: %v", groupInfo.Gid, err),
			)
			return
		}
	} else {
		infoSys := info.Sys()
		stat, ok := infoSys.(*syscall.Stat_t)
		if !ok {
			resp.Diagnostics.AddError(
				"Invalid File Information",
				"Unable to retrieve system-specific file statistics for current group.",
			)
			return
		}
		gid = int(stat.Gid)
	}

	if err := os.Chown(directoryPath, uid, gid); err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Ownership",
			fmt.Sprintf("Failed to set directory ownership for '%s': %v", directoryPath, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func isProtectedPath(path string) bool {
	for _, protected := range protectedPaths {
		if path == protected || strings.HasPrefix(path+"/", protected+"/") {
			return true
		}
	}

	return false
}

func (r *resourceUtilitiesLocalDirectory) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if runtime.GOOS == "windows" {
		resp.Diagnostics.AddError(
			"Incompatible Platform",
			"This resource cannot be used on Windows systems.",
		)
		return
	}

	var data LocalDirectory

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	directoryPath := data.Path.ValueString()

	info, err := os.Stat(directoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			tflog.Info(ctx, "Directory does not exist, skipping deletion", map[string]interface{}{
				"path": directoryPath,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Accessing Directory",
			fmt.Sprintf("Failed to access directory for deletion: %v", err),
		)
		return
	}

	if !info.IsDir() {
		resp.Diagnostics.AddError(
			"Invalid Directory Path",
			"The specified path exists but is not a directory, so it cannot be deleted.",
		)
		return
	}

	if isProtectedPath(directoryPath) {
		tflog.Warn(ctx, "Attempted to delete a protected path, skipping deletion", map[string]interface{}{
			"path": directoryPath,
		})
		return
	}

	if data.Force.ValueBool() {
		tflog.Info(ctx, "Force deletion enabled, removing directory", map[string]interface{}{
			"path": directoryPath,
		})
		if err := os.RemoveAll(directoryPath); err != nil {
			resp.Diagnostics.AddError(
				"Error Deleting Directory",
				fmt.Sprintf("Failed to delete directory: %v", err),
			)
			return
		}
	} else {
		if data.Managed.ValueBool() {
			tflog.Info(ctx, "Managed directory, proceeding with deletion", map[string]interface{}{
				"path": directoryPath,
			})
			if err := os.RemoveAll(directoryPath); err != nil {
				resp.Diagnostics.AddError(
					"Error Deleting Directory",
					fmt.Sprintf("Failed to delete directory: %v", err),
				)
				return
			}
		} else {
			tflog.Info(ctx, "Directory is unmanaged, skipping deletion", map[string]interface{}{
				"path": directoryPath,
			})
		}
	}

	resp.State.RemoveResource(ctx)
	tflog.Info(ctx, "Directory successfully deleted", map[string]interface{}{
		"path": directoryPath,
	})
}

func (r *resourceUtilitiesLocalDirectory) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The **utilities_local_directory** resource manages a local directory on the filesystem, ensuring it exists with specified attributes like permissions, ownership, and management status.

- **Managed vs Unmanaged**: Directories created by this resource are considered _managed_. Pre-existing directories are automatically marked as _unmanaged_.
- **Force Deletion**: The **force** attribute can be set to true to remove unmanaged directories during the destroy phase.
- **Permissions and Ownership**: The resource allows setting file permissions in octal format (e.g., **0755**) and specifying the user and group ownership.

**Note**: This resource is currently **not supported** on Windows systems.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the resource, representing the full path to the directory.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The absolute path to the directory to be managed.",
				Required:            true,
			},
			"managed": schema.BoolAttribute{
				MarkdownDescription: `
Indicates whether the directory is managed by this resource:
- **true**: The directory was created by this resource.
- **false**: The directory pre-existed and is considered unmanaged.`,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: `
Controls the behavior during the destroy phase:
- **true**: The directory will be forcibly deleted even if it is unmanaged.
- **false**: Unmanaged directories will not be deleted.`,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Optional: true,
			},
			"user": schema.StringAttribute{
				MarkdownDescription: `
Specifies the user owner of the directory. Defaults to "root".
- **Format**: Provide the username (e.g., "user1").`,
				Computed: true,
				Default:  stringdefault.StaticString("root"),
				Optional: true,
			},
			"group": schema.StringAttribute{
				MarkdownDescription: `
Specifies the group owner of the directory. Defaults to "root".
- **Format**: Provide the group name (e.g., "group1").`,
				Computed: true,
				Default:  stringdefault.StaticString("root"),
				Optional: true,
			},
			"permissions": schema.StringAttribute{
				MarkdownDescription: `
Defines the permissions for the directory in octal format (e.g., "0755"). Defaults to "0755".
- **Usage**: This can be used to ensure specific read, write, and execute permissions for the directory.`,
				Computed: true,
				Default:  stringdefault.StaticString("0755"),
				Optional: true,
			},
		},
	}
}