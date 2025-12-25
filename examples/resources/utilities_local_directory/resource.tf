resource "utilities_local_directory" "example" {
  force       = true
  path        = "/tmp/test"
  permissions = "0750"
}
