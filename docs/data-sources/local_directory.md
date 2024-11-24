---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "utilities_local_directory Data Source - terraform-provider-utilities"
subcategory: ""
description: |-
  Provides information about a local directory, including its metadata and permissions.
---

# utilities_local_directory (Data Source)

Provides information about a local directory, including its metadata and permissions.

## Example Usage

```terraform
data "utilities_local_directory" "example" {
  path = "/tmp"
}

output "utilities_local_directory" {
  value = {
    exists      = data.utilities_local_directory.example.exists,
    group       = data.utilities_local_directory.example.group,
    id          = data.utilities_local_directory.example.id,
    path        = data.utilities_local_directory.example.path,
    permissions = data.utilities_local_directory.example.permissions,
    user        = data.utilities_local_directory.example.user,
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `path` (String) The path to the local directory.

### Read-Only

- `exists` (Boolean) Indicates if the directory exists.
- `group` (String) The name of the group owning the directory.
- `id` (String) The unique identifier for the local directory, which is the same as the path.
- `permissions` (String) Permissions of the directory in octal format (e.g., 0755).
- `user` (String) The name of the user owning the directory.