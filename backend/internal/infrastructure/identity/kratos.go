package identity

import (
	"context"
	"fmt"

	ory "github.com/ory/kratos-client-go"
)

// KratosClient implements Client using the Ory Kratos SDK.
type KratosClient struct {
	client *ory.APIClient
}

func NewKratosClient(client *ory.APIClient) *KratosClient {
	return &KratosClient{client: client}
}

func (c *KratosClient) GetMeDetail(ctx context.Context, sessionToken string) (*UserDetail, error) {
	session, _, err := c.client.FrontendAPI.ToSession(ctx).XSessionToken(sessionToken).Execute()
	if err != nil {
		return nil, fmt.Errorf("kratos session check failed: %w", err)
	}

	identity := session.GetIdentity()
	traits, ok := identity.GetTraitsOk()
	if !ok || traits == nil {
		return nil, fmt.Errorf("kratos identity has no traits")
	}

	traitsMap, ok := (*traits).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("kratos traits are not a map")
	}

	detail := &UserDetail{
		IdentityID: identity.GetId(),
	}

	if email, ok := traitsMap["email"].(string); ok {
		detail.Email = email
	}
	if firstName, ok := traitsMap["first_name"].(string); ok {
		detail.FirstName = firstName
	}
	if lastName, ok := traitsMap["last_name"].(string); ok {
		detail.LastName = lastName
	}

	return detail, nil
}
