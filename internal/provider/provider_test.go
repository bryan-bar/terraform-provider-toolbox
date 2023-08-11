package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func protoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"toolbox": providerserver.NewProtocol6WithError(New()),
	}
}

func providerVersion() map[string]resource.ExternalProvider {
	return map[string]resource.ExternalProvider{
		"toolbox": {
			VersionConstraint: "0.1.2",
			Source:            "EnterpriseDB/toolbox",
		},
	}
}
