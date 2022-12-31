---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "filesystem_file_writer Resource - terraform-provider-filesystem"
subcategory: ""
description: |-
  Creates and manages a file on disk.
---

# filesystem_file_writer (Resource)

Creates and manages a file on disk.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `path` (String) Path to write the file on disk

### Optional

- `contents` (String, Sensitive) Raw file contents
- `create_parent_dirs` (Boolean) Create parent directories if they do not exist
- `delete_on_destroy` (Boolean) Delete the created file on destroy
- `mode` (String) File mode bits
- `root` (String) Path to the root of the module

### Read-Only

- `id` (String) The ID of this resource.
- `name` (String) Basename of the file
- `size` (Number) Size of the file in bytes

