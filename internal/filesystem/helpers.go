// Copyright 2018 Google, Inc.
// Copyright 2018 Seth Vargo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	defaultFilePerms = 0644

	// fileSizeLimit is the limit on the size of the file to read.
	fileSizeLimit = 1 * 1024 * 1024 * 1024 // 1 GiB
)

// expandRelativePath expands the given file path taking into account home directory and
// relative paths to the CWD.
func expandRelativePath(p, root string) (string, error) {
	p, err := homedir.Expand(p)
	if err != nil {
		return "", fmt.Errorf("failed to expand homedir: %w", err)
	}

	p, err = filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("failed to expand path: %w", err)
	}

	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		root = cwd
	}

	p, err = filepath.Rel(root, p)
	if err != nil {
		return "", fmt.Errorf("failed to get path relative to module: %w", err)
	}

	return p, nil
}

type fileStat struct {
	name     string
	contents string
	size     int64
	mode     string
}

func readFileAndStats(p string) (*fileStat, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat: %w", err)
	}

	if stat.Size() > fileSizeLimit {
		return nil, fmt.Errorf("file is too large (> 1GiB)")
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("is a directory")
	}

	contents, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}

	return &fileStat{
		name:     stat.Name(),
		contents: string(contents),
		size:     stat.Size(),
		mode:     fmt.Sprintf("%#o", stat.Mode()),
	}, nil
}

// atomicWriteInput is used as input to the atomicWrite function.
type atomicWriteInput struct {
	dest       string
	contents   string
	createDirs bool
	perms      os.FileMode
}

// atomicWrite atomically writes a file with the given contents and permissions.
func atomicWrite(i *atomicWriteInput) error {
	d, err := os.MkdirTemp("", "terraform-provider-filesystem")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(d)

	b := filepath.Base(i.dest)
	f, err := os.CreateTemp(d, b)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := f.Write([]byte(i.contents)); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close: %w", err)
	}

	// If the user did not explicitly set permissions, attempt to lookup the
	// current permissions on the file. If the file does not exist, fall back to
	// the default. Otherwise, inherit the current permissions.
	perms := i.perms
	if perms == 0 {
		stat, err := os.Stat(i.dest)
		if err != nil {
			if os.IsNotExist(err) {
				perms = defaultFilePerms
			} else {
				return fmt.Errorf("failed to stat file: %w", err)
			}
		} else {
			perms = stat.Mode()
		}
	}

	parent := filepath.Dir(i.dest)
	if _, err := os.Stat(parent); os.IsNotExist(err) {
		if i.createDirs {
			if err := os.MkdirAll(parent, 0700); err != nil {
				return fmt.Errorf("failed to make parent directory: %w", err)
			}
		} else {
			return fmt.Errorf("no parent directory")
		}
	}

	if err := os.Rename(f.Name(), i.dest); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	if err := os.Chmod(i.dest, perms); err != nil {
		return fmt.Errorf("failed to chmod: %w", err)
	}

	return nil
}

func parseFileMode(s string) (os.FileMode, error) {
	mode, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse mode: %w", err)
	}
	return os.FileMode(mode), nil
}

func diffSuppressRelativePath(_, old, new string, d *schema.ResourceData) bool {
	root := d.Get("root").(string)

	p, err := expandRelativePath(new, root)
	if err != nil {
		return false
	}

	return p == old
}
