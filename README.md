# Terraform FileSystem Provider

This is a [Terraform][terraform] provider for managing the local filesystem with
Terraform. It enables you to treat "files as code" the same way you already
treat infrastructure as code!


## Installation

1. Download the latest compiled binary from [GitHub releases][releases].

1. Untar the archive.

1. Move it into `$HOME/.terraform.d/plugins`:

    ```sh
    $ mkdir -p $HOME/.terraform.d/plugins
    $ mv terraform-provider-filesystem $HOME/.terraform.d/plugins/terraform-provider-filesystem
    ```

1. Create your Terraform configurations as normal, and run `terraform init`:

    ```sh
    $ terraform init
    ```

    This will find the plugin locally.


## Usage

1. Create a Terraform configuration file:

    ```hcl
    resource "filesystem_file_writer" "example" {
      path     = "file.txt"
      contents = "hello world"
    }

    resource "filesystem_file_reader" "example" {
      path = "${filesystem_file_writer.example.path}"
    }
    ```

1. Run `terraform init` to pull in the provider:

    ```sh
    $ terraform init
    ```

1. Run `terraform plan` and `terraform apply` to interact with the filesystem:

    ```sh
    $ terraform plan

    $ terraform apply
    ```

## Examples

For more examples, please see the [examples][examples] folder in this
repository.

## Reference

### Filesystem Reader

#### Usage

```hcl
resource "filesystem_file_reader" "read" {
  path = "my-file.txt"
}
```

#### Arguments

Arguments are provided as inputs to the resource, in the `*.tf` file.

- `path` `(string, required)` - the path to the file on disk.

- `root` `(string: $CWD)` - the root of the Terraform configurations. By
  default, this will be the current working directory. If you're running
  Terraform against configurations outside of the working directory (like
  `terraform apply ../../foo`), set this value to `${path.module}`.

#### Attributes

Attributes are values that are only known after creation.

- `base64contents` `(string)` - the contents of the file as a base64 encoded
  string. Useful for binary files.

- `contents` `(string)` - the contents of the file as a string. Contents are
  converted to a string, so it is not recommended you use this resource on
  binary files.

- `name` `(string)` - the name of the file.

- `size` `(int)` - the size of the file in bytes.

- `mode` `(int)` - the permissions on the file in octal.


### Filesystem Writer

#### Usage

```hcl
resource "filesystem_file_writer" "write" {
  path     = "my-file.txt"
  contents = "hello world!"
}
```

#### Arguments

- `path` `(string, required)` - the path to the file on disk.

- `contents` `(string, required)` - the contents of the file as a string.

- `root` `(string: $CWD)` - the root of the Terraform configurations. By
  default, this will be the current working directory. If you're running
  Terraform against configurations outside of the working directory (like
  `terraform apply ../../foo`), set this value to `${path.module}`.

- `create_parent_dirs` `(bool: true)` - create parent directories if they do not
  exist. By default, this is true. If set to false, the parent directories of
  the file must exist or this resource will error.

- `delete_on_destroy` `(bool: true)` - delete this file on destroy. Set this to
  false and Terraform will leave the file on disk on `terraform destroy`.

- `mode` `(int)` - the permissions on the file in octal.

#### Attributes

- `name` `(string)` - the name of the file.

- `size` `(int)` - the size of the file in bytes.

## FAQ

**Q: How is this different than the built-in `${file()}` function?**<br>
A: The built-in `file` function resolves paths and files at compile time. This
means the file must exist before Terraform can begin executing. In some
situations, the Terraform run itself may create files, but they will not exist
at start time. This Terraform provider enables you to treat files just like
other cloud resources, resolving them at runtime. This allows you to read and
write files from other sources without worrying about dependency ordering.

**Q: How is this different than [terraform-provider-local][terraform-provider-local]?**<br>
A: There are quite a few differences:

1. The equivalent "reader" is a data source. Data sources are resolved before
resources run, meaning it is not possible to use the data source to read a file
that is created _during_ the terraform run. Terraform will fail early that it
could not read the file. This provider specifically addresses that challenge by
using a resource instead of a data source.

1. The equivalent "reader" does not expose all the fields of the stat file (like
mode and owner permissions).

1. The equivalent "writer" does not allow setting file permissions, controlling
parent directory creation, or controlling deletion behavior. Additionally, as a
**super ultra bad thing**, the file permissions are written as 0777 (globally
executable), leaving a large security loophole.

1. The equivalent "writer" does not use an atomic file write. For large file
chunks, this can result in a partially committed file and/or improper
permissions that compromise security.

1. Neither the equivalent "reader" nor the "writer" limit the size of the file
being read/written. This poses a security threat as an attacker could overflow
the process (think about Terraform running arbitrary configuration as a hosted
service).

1. The terraform-provider-local stores the full path of the file in the state,
rendering the configurations un-portable. This provider calculates the filepath
relative to the Terraform module, allowing for more flexibility.

**Q: Is it secure?**<br>
A: The contents of files written and read are stored **in plain text** in the
statefile. They are marked as sensitive in the output, but they will still be
stored in the state. This is required in order for other resources to be able to
read the values. If you are using these resources with sensitive data, you
should encrypt your state using [remote state][remote-state].

## License & Author

```
Copyright 2018 Google, Inc.
Copyright 2018 Seth Vargo

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

[terraform]: https://www.terraform.io/
[releases]: https://github.com/sethvargo/terraform-provider-filesystem/releases
[examples]: https://github.com/sethvargo/terraform-provider-filesystem/tree/master/examples
[remote-state]: https://www.terraform.io/docs/state/remote.html
[terraform-provider-local]: https://github.com/terraform-providers/terraform-provider-local
