//go:build generate

package tools

import (
	// tfplugindocs generates provider documentation from schema descriptions and examples/
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir ..
