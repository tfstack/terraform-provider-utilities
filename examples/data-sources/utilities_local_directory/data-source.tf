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
