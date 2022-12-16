//go:build tools
// +build tools

package tools

import (
	// document generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform-provider-scaffolding/internal/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		Debug: true,
		ProviderFunc: func() *schema.Provider {
			return provider.New("2.1.1")()
		},
	})
}
