package provider

import (
	ory "github.com/ory/kratos-client-go"
)

func NewKratosProvider(env *EnvProvider) *ory.APIClient {
	cfg := ory.NewConfiguration()
	cfg.Servers = ory.ServerConfigurations{
		{URL: env.KratosPublicURL()},
	}
	return ory.NewAPIClient(cfg)
}
