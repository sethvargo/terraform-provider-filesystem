// Generate an SSH key
resource "tls_private_key" "ssh" {
  algorithm = "RSA"
  rsa_bits  = "4096"
}

// Save the SSH keys to disk
resource "filesystem_file_writer" "save-private-key" {
  path     = "${path.module}/.ssh/id_rsa"
  contents = tls_private_key.ssh.private_key_pem
  mode     = "0600"
}

resource "filesystem_file_writer" "save-public-key" {
  path     = "${path.module}/.ssh/id_rsa.pub"
  contents = tls_private_key.ssh.public_key_openssh
  mode     = "0644"
}
