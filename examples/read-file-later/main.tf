// Assume that "file.txt" does not exist at the start of the run. Perhaps
// there's another resource that creates the file as part of the Terraform run.
// As such, `${file("...")}` would not work as it resolves at the start of the
// run.

resource "null_resource" "pretend" {
  provisioner "local-exec" {
    command = "echo 'hello' > ${path.module}/file.txt"
  }
}

resource "filesystem_file_reader" "read" {
  path = "${path.module}/file.txt"

  depends_on = ["null_resource.pretend"]
}

output "contents" {
  value = "${filesystem_file_reader.read.contents}"
}
