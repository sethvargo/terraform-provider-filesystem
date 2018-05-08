provider "filesystem" {}

resource "filesystem_file_writer" "demo" {
  path     = "file.txt"
  contents = "hello world!"
  mode     = "0644"
}

# resource "filesystem_file_reader" "demo" {
#   path = "${filesystem_file_writer.demo.path}"
# }
#
# output "contents" {
#   value = "${filesystem_file_reader.demo.contents}"
# }

