# example: existing path
output "path_exists_existing_dir" {
  value = provider::utilities::path_exists("/tmp")
}

# example: non-Existing path
output "path_exists_non_existing_path" {
  value = provider::utilities::path_exists("/tmp/nonexistent/path")
}

# example: empty path
output "path_exists_empty_path" {
  value = provider::utilities::path_exists("")
}
