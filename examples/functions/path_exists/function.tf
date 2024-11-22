# example: existing path
output "path_exists_existing_dir" {
  value = (provider::utilities::path_exists("/tmp")).exists
}

# example: non-Existing path
output "path_exists_non_existing_path" {
  value = (provider::utilities::path_exists("/tmp/nonexistent/path")).exists
}

# example: empty path
output "path_exists_empty_path" {
  value = (provider::utilities::path_exists("")).exists
}