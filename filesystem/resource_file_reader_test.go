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

func TestFileSystemFileReader(t *testing.T) {
	t.Run("local", func(t *testing.T) {
		t.Parallel()

		contents := "This is some content!"

		f, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())

		if _, err := f.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
		if err := f.Sync(); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(f.Name(), 0644); err != nil {
			t.Fatal(err)
		}

		config := fmt.Sprintf(`
			resource "filesystem_file_reader" "file" {
				path = "%s"
			}
		`, f.Name())

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest: true,
			Providers:  testProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: func(s *terraform.State) error {
						attrs := s.RootModule().Resources["filesystem_file_reader.file"].Primary.Attributes

						name := filepath.Base(f.Name())
						if act, exp := attrs["name"], name; act != exp {
							t.Errorf("expected %q to be %q", act, exp)
						}
						if act, exp := attrs["mode"], "0644"; act != exp {
							t.Errorf("expected %q to be %q", act, exp)
						}
						if act, exp := attrs["contents"], contents; act != exp {
							t.Errorf("expected %q to be %q", act, exp)
						}
						return nil
					},
				},
			},
			CheckDestroy: func(*terraform.State) error {
				// The reader should NOT destroy the file.
				if _, err := os.Stat(f.Name()); os.IsNotExist(err) {
					t.Errorf("file should not have been deleted")
				}
				return nil
			},
		})
	})
}
