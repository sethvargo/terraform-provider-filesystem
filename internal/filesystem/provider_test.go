package filesystem

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testProviders = map[string]func() (*schema.Provider, error){
	"filesystem": func() (*schema.Provider, error) {
		return New("test")(), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New("test")().InternalValidate(); err != nil {
		t.Fatal(err)
	}
}
