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
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFileWriter() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages a file on disk.",

		CreateContext: resourceFileWriterCreate,
		ReadContext:   resourceFileWriterRead,
		UpdateContext: resourceFileWriterUpdate,
		DeleteContext: resourceFileWriterDelete,

		Schema: map[string]*schema.Schema{
			"path": {
				Type:             schema.TypeString,
				Description:      "Path to write the file on disk",
				ForceNew:         true,
				Required:         true,
				DiffSuppressFunc: diffSuppressRelativePath,
			},

			"root": {
				Type:        schema.TypeString,
				Description: "Path to the root of the module",
				ForceNew:    true,
				Optional:    true,
			},

			"contents": {
				Type:        schema.TypeString,
				Description: "Raw file contents",
				Optional:    true,
				Sensitive:   true,
			},

			"mode": {
				Type:        schema.TypeString,
				Description: "File mode bits",
				Default:     "0644",
				Optional:    true,
			},

			"create_parent_dirs": {
				Type:        schema.TypeBool,
				Description: "Create parent directories if they do not exist",
				Default:     true,
				Optional:    true,
			},

			"delete_on_destroy": {
				Type:        schema.TypeBool,
				Description: "Delete the created file on destroy",
				Default:     true,
				Optional:    true,
			},

			//
			// Computed
			//
			"name": {
				Type:        schema.TypeString,
				Description: "Basename of the file",
				Computed:    true,
			},

			"size": {
				Type:        schema.TypeInt,
				Description: "Size of the file in bytes",
				Computed:    true,
			},
		},
	}
}

// resourceFileWriterCreate expands the file path and writes the file to disk.
func resourceFileWriterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	path := d.Get("path").(string)
	root := d.Get("root").(string)

	p, err := expandRelativePath(path, root)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("path", p)

	mode, err := parseFileMode(d.Get("mode").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := atomicWrite(&atomicWriteInput{
		dest:       p,
		contents:   d.Get("contents").(string),
		createDirs: d.Get("create_parent_dirs").(bool),
		perms:      os.FileMode(mode),
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(p)

	return resourceFileWriterRead(ctx, d, meta)
}

// resourceFileWriter reads the file contents from disk. It returns an error if
// it fails to read the contents. The entire file contents are read into memory
// because Terraform cannot pass around an io.Reader.
func resourceFileWriterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	p := d.Id()

	stat, err := readFileAndStats(p)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("path", p)
	d.Set("name", stat.name)
	d.Set("contents", stat.contents)
	d.Set("size", stat.size)
	d.Set("mode", stat.mode)

	return nil
}

// resourceFileWriterUpdate updates the file contents.
func resourceFileWriterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	p := d.Id()

	// If the contents have changed, everything else will chagen too on the atomic
	// write, so just delegate.
	if d.HasChange("contents") {
		return resourceFileWriterCreate(ctx, d, meta)
	}

	if d.HasChange("mode") {
		mode, err := parseFileMode(d.Get("mode").(string))
		if err != nil {
			return diag.FromErr(err)
		}

		if err := os.Chmod(p, mode); err != nil {
			return diag.Errorf("failed to chmod: %s", err)
		}

		d.Set("mode", fmt.Sprintf("%#o", mode))
	}

	return nil
}

// resourceFileWriterDelete deletes the file if the user specified to delete it.
func resourceFileWriterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	p := d.Id()

	if d.Get("delete_on_destroy").(bool) {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return diag.Errorf("failed to delete: %s", err)
		}
	}

	d.SetId("")

	return nil
}
