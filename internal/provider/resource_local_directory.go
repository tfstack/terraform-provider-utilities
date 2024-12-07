package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = (*resourceUtilitiesLocalDirectory)(nil)
)

type resourceUtilitiesLocalDirectory struct{}

type LocalDirectory struct {
	Force       types.Bool   `tfsdk:"force"`
	Group       types.String `tfsdk:"group"`
	Managed     types.Bool   `tfsdk:"managed"`
	Path        types.String `tfsdk:"path"`
	Permissions types.String `tfsdk:"permissions"`
	User        types.String `tfsdk:"user"`
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

// func getUserID(userName string) (int, error) {
// 	cmd := exec.Command("id", "-u", userName)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return -1, fmt.Errorf("failed to execute id command for user '%s': %v", userName, err)
// 	}

// 	uid, err := strconv.Atoi(strings.TrimSpace(string(output)))
// 	if err != nil {
// 		return -1, fmt.Errorf("failed to convert UID to integer: %v", err)
// 	}

// 	return uid, nil
// }

// func getGroupID(groupName string) (int, error) {
// 	cmd := exec.Command("id", "-g", groupName)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return -1, fmt.Errorf("failed to execute id command for group '%s': %v", groupName, err)
// 	}

// 	gid, err := strconv.Atoi(strings.TrimSpace(string(output)))
// 	if err != nil {
// 		return -1, fmt.Errorf("failed to convert GID to integer: %v", err)
// 	}

// 	return gid, nil
// }

func isProtectedPath(path string) bool {
	normalizedPath := strings.TrimRight(path, "/")

	for _, protected := range protectedPaths {
		normalizedProtected := strings.TrimRight(protected, "/")
		if normalizedPath == normalizedProtected {
			return true
		}
	}

	return false
}

func getCurrentGroupName() (string, error) {
	cmd := exec.Command("id", "-gn")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute 'id -gn' command: %v", err)
	}

	groupName := strings.TrimSpace(string(output))
	return groupName, nil
}

func (r *resourceUtilitiesLocalDirectory) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *resourceUtilitiesLocalDirectory) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "utilities_local_directory"
}

func (r *resourceUtilitiesLocalDirectory) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LocalDirectory

	// Retrieve plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if it's a valid directory
	directoryPath := data.Path.ValueString()
	info, err := os.Stat(directoryPath)
	if err == nil {
		// Path exists, check if it's a directory
		if info != nil && !info.IsDir() {
			resp.Diagnostics.AddError(
				"Invalid Directory Path",
				fmt.Sprintf("The path '%s' exists but is not a directory.", directoryPath),
			)
			return
		}
	} else if !os.IsNotExist(err) {
		// Unexpected error while checking the path
		resp.Diagnostics.AddError(
			"Error Retrieving Path Info",
			fmt.Sprintf("Failed to check the path '%s': %v", directoryPath, err),
		)
		return
	}

	// Compute default user if not set
	userName := data.User.ValueString()
	if userName == "" {
		currentUserInfo, err := user.Current()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Retrieving Current User",
				fmt.Sprintf("Failed to retrieve the current user: %v", err),
			)
			return
		}
		userName = currentUserInfo.Username
		data.User = types.StringValue(userName)
	}

	// Compute default group if not set
	groupName := data.Group.ValueString()
	if groupName == "" {
		var err error

		groupName, err = getCurrentGroupName()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Retrieving Current Group Name",
				err.Error(),
			)
			return
		}

		data.Group = types.StringValue(groupName)
	}

	// Convert user name to UID
	userInfo, err := user.Lookup(userName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid User",
			fmt.Sprintf("Failed to lookup user '%s': %v", userName, err),
		)
		return
	}

	// Convert UID to integer
	uid, err := strconv.Atoi(userInfo.Uid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid User ID",
			fmt.Sprintf("Failed to convert user ID '%s' to integer: %v", userInfo.Uid, err),
		)
		return
	}

	// Convert group name to GID
	groupInfo, err := user.LookupGroup(groupName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Looking Up Group",
			fmt.Sprintf("Failed to lookup group '%s': %v", groupName, err),
		)
		return
	}

	// Convert GID to integer
	gid, err := strconv.Atoi(groupInfo.Gid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Group ID",
			fmt.Sprintf("Failed to convert GID '%s' to integer: %v", groupInfo.Gid, err),
		)
		return
	}

	// For Windows, fallback to default values
	if runtime.GOOS == "windows" {
		uid = -1 // No valid UID on Windows
		gid = -1 // No valid GID on Windows
	}

	// Check if the directory exists otherwise create
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		// Create the directory since it doesn't exist
		if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
			resp.Diagnostics.AddError(
				"Error Creating Directory",
				fmt.Sprintf("Failed to create directory '%s': %v", directoryPath, err),
			)
			return
		}
		data.Managed = types.BoolValue(true)
	} else if err != nil {
		// Handle unexpected errors
		resp.Diagnostics.AddError(
			"Error Checking Directory",
			fmt.Sprintf("Failed to check directory '%s': %v", directoryPath, err),
		)
		return
	}

	// Handle file permission and ownership
	if isProtectedPath(directoryPath) {
		tflog.Warn(ctx, "Skipping ownership modification for protected OS path", map[string]interface{}{
			"path":   directoryPath,
			"reason": "The specified path is considered critical to the operating system and should not have its ownership modified to avoid potential system instability or security risks.",
		})
	} else {
		// Apply ownership to the directory
		if err := os.Chown(directoryPath, uid, gid); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Ownership",
				fmt.Sprintf("Failed to set ownership for path '%s': %s", directoryPath, err.Error()),
			)
			return
		}

		// Apply permissions if provided
		if perms := data.Permissions.ValueString(); perms != "" {
			mode, err := strconv.ParseUint(perms, 8, 32)
			if err != nil {
				resp.Diagnostics.AddError("Invalid Permissions", fmt.Sprintf("Failed to convert permissions '%s': %v", perms, err))
				return
			}
			if err := os.Chmod(directoryPath, os.FileMode(mode)); err != nil {
				resp.Diagnostics.AddError("Error Setting Permissions", err.Error())
				return
			}
		}
	}

	// Retrieve the current directory info
	info, err = os.Stat(directoryPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving Directory Info",
			err.Error(),
		)
		return
	}

	// Get the current permissions (mode) of the directory
	currentPermissions := info.Mode().Perm()

	// Set data.Permissions to the current permissions
	data.Permissions = types.StringValue(fmt.Sprintf("0%o", currentPermissions))

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceUtilitiesLocalDirectory) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LocalDirectory

	// Retrieve state data
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	directoryPath := data.Path.ValueString()

	// Check if directory exists
	info, err := os.Stat(directoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, no deletion needed
			tflog.Info(ctx, "Directory does not exist, skipping deletion", map[string]interface{}{"path": directoryPath})
			resp.State.RemoveResource(ctx)
			return
		}

		// Unexpected error accessing directory
		resp.Diagnostics.AddError("Error Accessing Directory", fmt.Sprintf("Failed to access directory for deletion: %v", err))
		return
	}

	// Ensure it is a directory
	if !info.IsDir() {
		resp.Diagnostics.AddError("Invalid Directory Path", fmt.Sprintf("The path '%s' exists but is not a directory.", directoryPath))
		return
	}

	// Check if the path is protected and prevent deletion if true
	if isProtectedPath(directoryPath) {
		tflog.Warn(ctx, "Attempted to delete a protected path, skipping deletion", map[string]interface{}{"path": directoryPath})
		return
	}

	// Decide the action based on force and managed flags
	if data.Force.ValueBool() {
		// Force deletion
		tflog.Info(ctx, "Force deletion enabled, removing directory", map[string]interface{}{"path": directoryPath})
		if err := os.RemoveAll(directoryPath); err != nil {
			resp.Diagnostics.AddError("Error Deleting Directory", fmt.Sprintf("Failed to delete directory: %v", err))
			return
		}
	} else if data.Managed.ValueBool() {
		// Managed directory, proceed with deletion
		tflog.Info(ctx, "Managed directory, proceeding with deletion", map[string]interface{}{"path": directoryPath})
		if err := os.RemoveAll(directoryPath); err != nil {
			resp.Diagnostics.AddError("Error Deleting Directory", fmt.Sprintf("Failed to delete directory: %v", err))
			return
		}
	} else {
		// Unmanaged directory, skipping deletion
		tflog.Info(ctx, "Directory is unmanaged, skipping deletion", map[string]interface{}{"path": directoryPath})
		return
	}

	// Successfully removed the directory, update state
	resp.State.RemoveResource(ctx)
	tflog.Info(ctx, "Directory successfully deleted", map[string]interface{}{"path": directoryPath})
}

func (r *resourceUtilitiesLocalDirectory) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LocalDirectory

	// Retrieve plan data
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve the current directory info
	directoryPath := data.Path.ValueString()
	info, err := os.Stat(directoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Log a warning if the path doesn't exist
			tflog.Warn(ctx, "Directory does not exist; proceeding without error.", map[string]interface{}{
				"path": directoryPath,
			})
			data.User = types.StringNull()
			data.Group = types.StringNull()
			data.Permissions = types.StringNull()
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		} else {
			// Unexpected error
			resp.Diagnostics.AddError(
				"Error Retrieving Directory Info",
				fmt.Sprintf("Failed to check the path '%s': %v", directoryPath, err),
			)
			return
		}
	}

	// If the path exists, check if it’s a directory
	if info != nil && !info.IsDir() {
		resp.Diagnostics.AddError(
			"Invalid Directory Path",
			fmt.Sprintf("The path '%s' exists but is not a directory.", directoryPath),
		)
		return
	}

	var userName, groupName string

	if runtime.GOOS != "windows" {
		// Retrieve system-specific metadata
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			resp.Diagnostics.AddError(
				"Error Retrieving File Metadata",
				"Failed to retrieve system metadata for the directory.",
			)
			return
		}

		// Get UID and GID
		uid := stat.Uid
		gid := stat.Gid

		// Lookup user name
		userInfo, err := user.LookupId(fmt.Sprint(uid))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Retrieving User Info",
				fmt.Sprintf("Failed to retrieve user information for UID %d: %v", uid, err),
			)
			return
		}
		userName = userInfo.Username

		// Lookup group name
		groupInfo, err := user.LookupGroupId(fmt.Sprint(gid))
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Error Retrieving Group Info",
				fmt.Sprintf("Failed to retrieve group information for GID %d: %v", gid, err),
			)
			groupName = "" // Reset group name if unresolved
		} else {
			groupName = groupInfo.Name
		}
	} else {
		// Defaults for Windows
		userName = "unknown"
		groupName = "unknown"
	}

	// Retrieve and set directory permissions
	mode := info.Mode().Perm()
	data.Permissions = types.StringValue(fmt.Sprintf("%04o", mode))
	data.User = types.StringValue(userName)
	data.Group = types.StringValue(groupName)

	// Log read operation
	tflog.Info(ctx, "Read local directory", map[string]interface{}{
		"path":        directoryPath,
		"user":        userName,
		"group":       groupName,
		"permissions": data.Permissions.ValueString(),
	})

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceUtilitiesLocalDirectory) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	permissionsRegex := regexp.MustCompile(`^0[0-7]{3}$`)

	resp.Schema = schema.Schema{
		MarkdownDescription: `
The **utilities_local_directory** resource manages a local directory on the filesystem, ensuring it exists with specified attributes like permissions, ownership, and management status.

- **Managed vs Unmanaged**: Directories created by this resource are considered _managed_. Pre-existing directories are automatically marked as _unmanaged_.
- **Force Deletion**: The **force** attribute can be set to true to remove unmanaged directories during the destroy phase.
- **Permissions and Ownership**: The resource allows setting file permissions in octal format (e.g., **0755**) and specifying the user and group ownership.

**Note**: This resource is currently **not supported** on Windows systems.
`,
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path of the directory to manage. The directory must exist or be created based on other attributes.",
			},
			"user": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "User to own the directory. Defaults to the current system user if not specified.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"group": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Group to own the directory. Defaults to the current user's group if not specified.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"permissions": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Permissions to set on the directory, in octal format (e.g., 0755).",
				Validators: []validator.String{
					stringvalidator.RegexMatches(permissionsRegex, "must be a valid octal permission (e.g., 0755)"),
				},
			},
			"force": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to force creation of the directory, even if it already exists. Default is false.",
			},
			"managed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Indicates whether the directory is managed by this provider. Defaults to false for existing directories.",
			},
		},
	}
}

func (r *resourceUtilitiesLocalDirectory) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LocalDirectory

	// Retrieve plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if it's a valid directory
	directoryPath := data.Path.ValueString()
	info, err := os.Stat(directoryPath)
	if err == nil {
		// Path exists, check if it's a directory
		if info != nil && !info.IsDir() {
			resp.Diagnostics.AddError(
				"Invalid Directory Path",
				fmt.Sprintf("The path '%s' exists but is not a directory.", directoryPath),
			)
			return
		}
	} else if !os.IsNotExist(err) {
		// Unexpected error while checking the path
		resp.Diagnostics.AddError(
			"Error Retrieving Path Info",
			fmt.Sprintf("Failed to check the path '%s': %v", directoryPath, err),
		)
		return
	}

	// Compute default user if not set
	userName := data.User.ValueString()
	if userName == "" {
		currentUserInfo, err := user.Current()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Retrieving Current User",
				fmt.Sprintf("Failed to retrieve the current user: %v", err),
			)
			return
		}
		userName = currentUserInfo.Username
		data.User = types.StringValue(userName)
	}

	// Compute default group if not set
	groupName := data.Group.ValueString()
	if groupName == "" {
		var err error

		groupName, err = getCurrentGroupName()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Retrieving Current Group Name",
				err.Error(),
			)
			return
		}

		data.Group = types.StringValue(groupName)
	}

	// Convert user name to UID
	userInfo, err := user.Lookup(userName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid User",
			fmt.Sprintf("Failed to lookup user '%s': %v", userName, err),
		)
		return
	}

	// Convert UID to integer
	uid, err := strconv.Atoi(userInfo.Uid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid User ID",
			fmt.Sprintf("Failed to convert user ID '%s' to integer: %v", userInfo.Uid, err),
		)
		return
	}

	// Convert group name to GID
	groupInfo, err := user.LookupGroup(groupName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Looking Up Group",
			fmt.Sprintf("Failed to lookup group '%s': %v", groupName, err),
		)
		return
	}

	// Convert GID to integer
	gid, err := strconv.Atoi(groupInfo.Gid)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Group ID",
			fmt.Sprintf("Failed to convert GID '%s' to integer: %v", groupInfo.Gid, err),
		)
		return
	}

	// For Windows, fallback to default values
	if runtime.GOOS == "windows" {
		uid = -1 // No valid UID on Windows
		gid = -1 // No valid GID on Windows
	}

	// Check if the directory exists otherwise create
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		// Create the directory since it doesn't exist
		if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
			resp.Diagnostics.AddError(
				"Error Creating Directory",
				fmt.Sprintf("Failed to create directory '%s': %v", directoryPath, err),
			)
			return
		}
		data.Managed = types.BoolValue(true)
	} else if err != nil {
		// Handle unexpected errors
		resp.Diagnostics.AddError(
			"Error Checking Directory",
			fmt.Sprintf("Failed to check directory '%s': %v", directoryPath, err),
		)
		return
	}

	// Handle file permission and ownership
	if isProtectedPath(directoryPath) {
		tflog.Warn(ctx, "Skipping ownership modification for protected OS path", map[string]interface{}{
			"path":   directoryPath,
			"reason": "The specified path is considered critical to the operating system and should not have its ownership modified to avoid potential system instability or security risks.",
		})
	} else {
		// Apply ownership to the directory
		if err := os.Chown(directoryPath, uid, gid); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Ownership",
				fmt.Sprintf("Failed to set ownership for path '%s': %s", directoryPath, err.Error()),
			)
			return
		}

		// Apply permissions if provided
		if perms := data.Permissions.ValueString(); perms != "" {
			mode, err := strconv.ParseUint(perms, 8, 32)
			if err != nil {
				resp.Diagnostics.AddError("Invalid Permissions", fmt.Sprintf("Failed to convert permissions '%s': %v", perms, err))
				return
			}
			if err := os.Chmod(directoryPath, os.FileMode(mode)); err != nil {
				resp.Diagnostics.AddError("Error Setting Permissions", err.Error())
				return
			}
			data.Managed = types.BoolValue(true)
		}
	}

	// Retrieve the current directory info
	info, err = os.Stat(directoryPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Retrieving Directory Info",
			err.Error(),
		)
		return
	}

	// Get the current permissions (mode) of the directory
	currentPermissions := info.Mode().Perm()

	// Set data.Permissions to the current permissions
	data.Permissions = types.StringValue(fmt.Sprintf("0%o", currentPermissions))

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
