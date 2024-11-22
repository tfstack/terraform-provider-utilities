# Example: Existing Path
output "path_exists_existing_dir" {
  value = (provider::utilities::path_exists("/tmp")).exists
}

# Example: Non-Existing Path
output "path_exists_non_existing_path" {
  value = (provider::utilities::path_exists("/tmp/nonexistent/path")).exists
}

# Example: Empty Path
output "path_exists_empty_path" {
  value = (provider::utilities::path_exists("")).exists
}
