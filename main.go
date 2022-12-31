package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/sethvargo/terraform-provider-filesystem/internal/filesystem"
)

//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	Version string = "dev"
	Commit  string = ""
)

func main() {
	debugMode := flag.Bool("debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        *debugMode,
		ProviderAddr: "github.com/sethvargo/terraform-provider-filesystem",
		ProviderFunc: filesystem.New(Version),
	}

	plugin.Serve(opts)
}
