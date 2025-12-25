resource "utilities_extract_zip" "local_source" {
  destination = "/tmp/test"
  source      = "./external/jq_1.7.zip"
}

resource "utilities_extract_zip" "url_source" {
  destination = "/tmp/test"
  url         = "https://github.com/platformfuzz/rpm-builder/archive/refs/tags/jq_1.7.zip"
}
