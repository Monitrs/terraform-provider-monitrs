// Package main is the terraform-provider-monitrs entrypoint.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/Monitrs/terraform-provider-monitrs/internal/provider"
)

// Set by GoReleaser via ldflags on release builds; "dev" during local dev.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider in debug mode")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/Monitrs/monitrs",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
