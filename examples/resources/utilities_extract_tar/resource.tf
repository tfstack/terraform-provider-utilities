# resource "utilities_extract_tar" "local_source" {
#   destination  = "/tmp/test"
#   source       = "./external/jq_1.7.zip"
# }

resource "utilities_extract_tar" "url_source" {
  destination  = "/tmp/test"
  url          = "https://github.com/tfstack/terraform-provider-utilities/archive/refs/tags/v1.0.12.tar.gz"
}
