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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFileReader() *schema.Resource {
	return &schema.Resource{
		Description: "Reads a file on disk as a resource.",

		CreateContext: resourceFileReaderCreate,
		ReadContext:   resourceFileReaderRead,
		DeleteContext: resourceFileReaderDelete,

		Schema: map[string]*schema.Schema{
			"path": {
				Type:             schema.TypeString,
				Description:      "Path to read the file on disk",
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

			//
			// Computed values
			//
			"contents": {
				Type:        schema.TypeString,
				Description: "Raw file contents",
				Computed:    true,
				Sensitive:   true,
			},

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

			"mode": {
				Type:        schema.TypeString,
				Description: "File mode bits",
				Computed:    true,
			},
		},
	}
}

// resourceFileReaderCreate expands the file path and calls Read.
func resourceFileReaderCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	path := d.Get("path").(string)
	root := d.Get("root").(string)

	p, err := expandRelativePath(path, root)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(p)

	return resourceFileReaderRead(ctx, d, meta)
}

// resourceFileReaderRead reads the file contents from disk. It returns an error if
// it fails to read the contents. The entire file contents are read into memory
// because Terraform cannot pass around an io.Reader.
func resourceFileReaderRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	p := d.Id()

	stat, err := readFileAndStats(p)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", stat.name)
	d.Set("contents", stat.contents)
	d.Set("size", stat.size)
	d.Set("mode", stat.mode)

	return nil
}

// resourceFileReaderDelete deletes our tracking of that file. It is basically a
// no-op. It does not delete the file on disk.
func resourceFileReaderDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId("")
	return nil
}
