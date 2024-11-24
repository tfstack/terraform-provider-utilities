
resource "utilities_local_directory" "example" {
  force = true
  group = "root"
  path  = "/tmp/test"
  permissions = "0755"
  user = "root"
}
