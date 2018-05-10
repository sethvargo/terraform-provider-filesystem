package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestFilesystemFileWriter(t *testing.T) {
	var cases = []struct {
		name   string
		config string
	}{
		{
			"basic",
			`
				resource "filesystem_file_writer" "file" {
					path     = "%s"
					contents = "%s"
					mode     = "%s"
				}
			`,
		},
		{
			"delete_on_destroy",
			`
				resource "filesystem_file_writer" "file" {
					path     = "%s"
					contents = "%s"
					mode     = "%s"

					delete_on_destroy = "false"
				}
			`,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := ioutil.TempFile("", "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}

			contents := "This is some content!"
			perms := "0755"
			config := fmt.Sprintf(tc.config, f.Name(), contents, perms)

			resource.UnitTest(t, resource.TestCase{
				Providers: testProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: func(s *terraform.State) error {
							attrs := s.RootModule().Resources["filesystem_file_writer.file"].Primary.Attributes

							name := filepath.Base(f.Name())
							if act, exp := attrs["name"], name; act != exp {
								t.Errorf("expected %q to be %q", act, exp)
							}
							if act, exp := attrs["mode"], perms; act != exp {
								t.Errorf("expected %q to be %q", act, exp)
							}
							if act, exp := attrs["contents"], contents; act != exp {
								t.Errorf("expected %q to be %q", act, exp)
							}

							b, err := ioutil.ReadFile(f.Name())
							if err != nil {
								t.Fatal(err)
							}

							if act, exp := string(b), contents; act != exp {
								t.Errorf("expected %q to be %q", act, exp)
							}

							return nil
						},
					},
				},
				CheckDestroy: func(s *terraform.State) error {
					attrs := s.RootModule().Resources["filesystem_file_writer.file"].Primary.Attributes

					_, err := os.Stat(f.Name())

					if attrs["delete_on_destroy"] == "true" {
						if !os.IsNotExist(err) {
							t.Errorf("expected file to be deleted")
						}
					} else {
						if os.IsNotExist(err) {
							t.Errorf("expected file to not be deleted")
						}
					}

					return nil
				},
			})
		})
	}
}
