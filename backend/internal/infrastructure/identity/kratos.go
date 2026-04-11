package identity

import (
	"context"
	"fmt"
	"net/http"

	ory "github.com/ory/kratos-client-go"
)

type KratosClient struct {
	client *ory.APIClient
}

func NewKratosClient(client *ory.APIClient) *KratosClient {
	return &KratosClient{client: client}
}

func (c *KratosClient) Login(ctx context.Context, email, password string) (*SessionResult, error) {
	flow, _, err := c.client.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create login flow: %w", err)
	}

	updateBody := ory.UpdateLoginFlowBody{
		UpdateLoginFlowWithPasswordMethod: &ory.UpdateLoginFlowWithPasswordMethod{
			Method:     "password",
			Identifier: email,
			Password:   password,
		},
	}

	login, resp, err := c.client.FrontendAPI.UpdateLoginFlow(ctx).
		Flow(flow.GetId()).
		UpdateLoginFlowBody(updateBody).
		Execute()
	if err != nil {
		if resp != nil &&
			(resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized) {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return &SessionResult{SessionToken: login.GetSessionToken()}, nil
}

func (c *KratosClient) Register(ctx context.Context, email, password string) (*SessionResult, error) {
	flow, _, err := c.client.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create registration flow: %w", err)
	}

	traits := map[string]interface{}{
		"email": email,
	}

	updateBody := ory.UpdateRegistrationFlowBody{
		UpdateRegistrationFlowWithPasswordMethod: &ory.UpdateRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: password,
			Traits:   traits,
		},
	}

	reg, resp, err := c.client.FrontendAPI.UpdateRegistrationFlow(ctx).
		Flow(flow.GetId()).
		UpdateRegistrationFlowBody(updateBody).
		Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusBadRequest {
			return nil, fmt.Errorf("registration failed — check email/password requirements")
		}
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	return &SessionResult{SessionToken: reg.GetSessionToken()}, nil
}

func (c *KratosClient) Logout(ctx context.Context, sessionToken string) error {
	session, _, err := c.client.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).Execute()
	if err != nil {
		return nil // session already invalid
	}

	_, err = c.client.IdentityAPI.DisableSession(ctx, session.GetId()).Execute()
	if err != nil {
		return fmt.Errorf("failed to disable session: %w", err)
	}

	return nil
}

func (c *KratosClient) identityToDetail(ident *ory.Identity) *UserDetail {
	detail := &UserDetail{
		IdentityID: ident.GetId(),
	}

	traits, ok := ident.GetTraitsOk()
	if !ok || traits == nil {
		return detail
	}

	traitsMap, ok := (*traits).(map[string]interface{})
	if !ok {
		return detail
	}

	if email, ok := traitsMap["email"].(string); ok {
		detail.Email = email
	}
	if username, ok := traitsMap["username"].(string); ok {
		detail.Username = username
	}

	return detail
}

func (c *KratosClient) GetSession(ctx context.Context, sessionToken string) (*UserDetail, error) {
	session, _, err := c.client.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).Execute()
	if err != nil {
		return nil, fmt.Errorf("kratos session check failed: %w", err)
	}

	identity := session.GetIdentity()
	return c.identityToDetail(&identity), nil
}

func (c *KratosClient) GetIdentity(ctx context.Context, id string) (*UserDetail, error) {
	ident, _, err := c.client.IdentityAPI.GetIdentity(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	return c.identityToDetail(ident), nil
}

func (c *KratosClient) UpdateTraits(ctx context.Context, id string, traits map[string]interface{}) (*UserDetail, error) {
	current, _, err := c.client.IdentityAPI.GetIdentity(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity for update: %w", err)
	}

	body := ory.UpdateIdentityBody{
		SchemaId: current.GetSchemaId(),
		State:    string(current.GetState()),
		Traits:   traits,
	}

	updated, _, err := c.client.IdentityAPI.UpdateIdentity(ctx, id).
		UpdateIdentityBody(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update identity traits: %w", err)
	}

	return c.identityToDetail(updated), nil
}

func (c *KratosClient) ListIdentities(ctx context.Context) ([]UserDetail, error) {
	identities, _, err := c.client.IdentityAPI.ListIdentities(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list identities: %w", err)
	}

	details := make([]UserDetail, 0, len(identities))
	for _, ident := range identities {
		details = append(details, *c.identityToDetail(&ident))
	}
	return details, nil
}
