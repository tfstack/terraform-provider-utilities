---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "path_owner function - terraform-provider-utilities"
subcategory: ""
description: |-
  Retrieves the owner of a given file or directory path.
---

# function: path_owner

Returns the owner (username) of the specified file or directory path.

### Parameters
- **path**: The file or directory path to check.

### Returns
- **string**: The username of the file or directory owner.

## Example Usage

```terraform
# example: existing path
output "path_owner_existing_dir" {
  value = provider::utilities::path_owner("/tmp")
}

# example: non-Existing path
output "path_owner_non_existing_path" {
  value = provider::utilities::path_owner("/tmp/nonexistent/path")
}

# example: empty path
output "path_owner_empty_path" {
  value = provider::utilities::path_owner("")
}
```

## Signature

<!-- signature generated by tfplugindocs -->
```text
path_owner(path string) string
```

## Arguments

<!-- arguments generated by tfplugindocs -->
1. `path` (String) Path to the file or directory.

