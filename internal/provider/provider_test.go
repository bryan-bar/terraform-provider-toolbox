package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func protoV5ProviderFactories() map[string]func() (tfprotov5.ProviderServer, error) {
	return map[string]func() (tfprotov5.ProviderServer, error){
		"toolbox": providerserver.NewProtocol5WithError(New()),
	}
}

func providerVersion() map[string]resource.ExternalProvider {
	return map[string]resource.ExternalProvider{
		"toolbox": {
			VersionConstraint: "0.0.13",
			Source:            "bryan-bar/toolbox",
		},
	}
}
