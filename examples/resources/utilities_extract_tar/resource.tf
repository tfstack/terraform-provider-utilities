# resource "utilities_extract_tar" "local_source" {
#   destination  = "/tmp/test"
#   source       = "./external/jq_1.7.zip"
# }

resource "utilities_extract_tar" "url_source" {
  destination  = "/tmp/test2"
  url          = "https://github.com/tfstack/terraform-provider-utilities/blob/21-new-extract-tar-resource/examples/resources/utilities_extract_tar/sample.tar"
}
