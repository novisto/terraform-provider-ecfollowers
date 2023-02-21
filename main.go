package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"terraform-provider-ecfollowers/ecfollowers"
)

func main() {
	providerserver.Serve(
		context.Background(),
		ecfollowers.New,
		providerserver.ServeOpts{Address: "registry.terraform.io/novisto/ecfollowers"},
	)
}
