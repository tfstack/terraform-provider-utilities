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
