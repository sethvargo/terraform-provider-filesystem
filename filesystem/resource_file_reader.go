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
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceFileReader() *schema.Resource {
	return &schema.Resource{
		Create: resourceFileReaderCreate,
		Read:   resourceFileReaderRead,
		Delete: resourceFileReaderDelete,

		Schema: map[string]*schema.Schema{
			"path": &schema.Schema{
				Type:             schema.TypeString,
				Description:      "Path to read the file on disk",
				ForceNew:         true,
				Required:         true,
				DiffSuppressFunc: diffSuppressRelativePath,
			},

			"root": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Path to the root of the module",
				ForceNew:    true,
				Optional:    true,
			},

			//
			// Computed values
			//
			"contents": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Raw file contents",
				Computed:    true,
				Sensitive:   true,
			},

			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Basename of the file",
				Computed:    true,
			},

			"size": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Size of the file in bytes",
				Computed:    true,
			},

			"mode": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "File mode bits",
				Computed:    true,
			},
		},
	}
}

// resourceFileReaderCreate expands the file path and calls Read.
func resourceFileReaderCreate(d *schema.ResourceData, meta interface{}) error {
	path := d.Get("path").(string)
	root := d.Get("root").(string)

	p, err := expandRelativePath(path, root)
	if err != nil {
		return err
	}

	d.SetId(p)

	return resourceFileReaderRead(d, meta)
}

// resourceFileReaderRead reads the file contents from disk. It returns an error if
// it fails to read the contents. The entire file contents are read into memory
// because Terraform cannot pass around an io.Reader.
func resourceFileReaderRead(d *schema.ResourceData, meta interface{}) error {
	p := d.Id()

	stat, err := readFileAndStats(p)
	if err != nil {
		return err
	}

	d.Set("name", stat.name)
	d.Set("contents", stat.contents)
	d.Set("size", stat.size)
	d.Set("mode", stat.mode)

	return nil
}

// resourceFileReaderDelete deletes our tracking of that file. It is basically a
// no-op. It does not delete the file on disk.
func resourceFileReaderDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}
