package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/provider"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/we-work-in-the-cloud/nullcloud",
	}
	if err := providerserver.Serve(context.Background(), provider.New, opts); err != nil {
		log.Fatal(err)
	}
}
