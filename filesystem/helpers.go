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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
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
		return "", errors.Wrap(err, "failed to expand homedir")
	}

	p, err = filepath.Abs(p)
	if err != nil {
		return "", errors.Wrap(err, "failed to expand")
	}

	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "failed to get working directory")
		}
		root = cwd
	}

	p, err = filepath.Rel(root, p)
	if err != nil {
		return "", errors.Wrap(err, "failed to get path relative to module")
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
		return nil, errors.Wrap(err, "failed to open")
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "failed to stat")
	}

	if stat.Size() > fileSizeLimit {
		return nil, errors.New("file is too large (> 1GiB)")
	}

	if stat.IsDir() {
		return nil, errors.New("is a directory")
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read")
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
	d, err := ioutil.TempDir("", "terraform-provider-filesystem")
	if err != nil {
		return errors.Wrap(err, "failed to create temp dir")
	}
	defer os.RemoveAll(d)

	b := filepath.Base(i.dest)
	f, err := ioutil.TempFile(d, b)
	if err != nil {
		return errors.Wrap(err, "failed to create temp file")
	}

	if _, err := f.Write([]byte(i.contents)); err != nil {
		return errors.Wrap(err, "failed to write")
	}

	if err := f.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync")
	}

	if err := f.Close(); err != nil {
		return errors.Wrap(err, "failed to close")
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
				return errors.Wrap(err, "failed to stat file")
			}
		} else {
			perms = stat.Mode()
		}
	}

	parent := filepath.Dir(i.dest)
	if _, err := os.Stat(parent); os.IsNotExist(err) {
		if i.createDirs {
			if err := os.MkdirAll(parent, 0700); err != nil {
				return errors.Wrap(err, "failed to make parent directory")
			}
		} else {
			return errors.New("no parent directory")
		}
	}

	if err := MoveFile(f.Name(), i.dest); err != nil {
		return errors.Wrap(err, "failed to rename")
	}

	if err := os.Chmod(i.dest, perms); err != nil {
		return errors.Wrap(err, "failed to chmod")
	}

	return nil
}

// ref: https://gist.github.com/var23rav/23ae5d0d4d830aff886c3c970b8f6c6b
func MoveFile(sourcePath, destPath string) error {
    inputFile, err := os.Open(sourcePath)
    if err != nil {
        return fmt.Errorf("Couldn't open source file: %s", err)
    }
    outputFile, err := os.Create(destPath)
    if err != nil {
        inputFile.Close()
        return fmt.Errorf("Couldn't open dest file: %s", err)
    }
    defer outputFile.Close()
    _, err = io.Copy(outputFile, inputFile)
    inputFile.Close()
    if err != nil {
        return fmt.Errorf("Writing to output file failed: %s", err)
    }
    // The copy was successful, so now delete the original file
    err = os.Remove(sourcePath)
    if err != nil {
        return fmt.Errorf("Failed removing original file: %s", err)
    }
    return nil
}

func parseFileMode(s string) (os.FileMode, error) {
	mode, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse mode")
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
