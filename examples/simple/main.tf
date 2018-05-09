provider "filesystem" {}

resource "filesystem_file_writer" "demo" {
  path     = "file.txt"
  contents = "hello world!"
  mode     = "0644"
}
