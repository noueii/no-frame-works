package identity

import (
	"context"
	"fmt"
	"net/http"

	ory "github.com/ory/kratos-client-go"
)

// KratosClient implements Client using the Ory Kratos SDK.
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

func (c *KratosClient) GetSession(ctx context.Context, sessionToken string) (*UserDetail, error) {
	session, _, err := c.client.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).Execute()
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

	return detail, nil
}
