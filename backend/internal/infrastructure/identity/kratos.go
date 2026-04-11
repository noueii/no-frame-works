package identity

import (
	"context"
	"net/http"

	"github.com/go-errors/errors"
	ory "github.com/ory/kratos-client-go"
)

type KratosClient struct {
	client *ory.APIClient
}

func NewKratosClient(client *ory.APIClient) *KratosClient {
	return &KratosClient{client: client}
}

func (c *KratosClient) Login(ctx context.Context, email, password string) (*SessionResult, error) {
	flow, resp, err := c.client.FrontendAPI.CreateNativeLoginFlow(ctx).Execute()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, errors.Errorf("failed to create login flow: %w", err)
	}

	updateBody := ory.UpdateLoginFlowBody{
		UpdateLoginFlowWithPasswordMethod: &ory.UpdateLoginFlowWithPasswordMethod{
			Method:     "password",
			Identifier: email,
			Password:   password,
		},
	}

	login, loginResp, err := c.client.FrontendAPI.UpdateLoginFlow(ctx).
		Flow(flow.GetId()).
		UpdateLoginFlowBody(updateBody).
		Execute()
	if loginResp != nil && loginResp.Body != nil {
		defer loginResp.Body.Close()
	}
	if err != nil {
		if loginResp != nil &&
			(loginResp.StatusCode == http.StatusBadRequest || loginResp.StatusCode == http.StatusUnauthorized) {
			return nil, errors.Errorf("invalid credentials")
		}
		return nil, errors.Errorf("login failed: %w", err)
	}

	return &SessionResult{SessionToken: login.GetSessionToken()}, nil
}

func (c *KratosClient) Register(
	ctx context.Context,
	email, password string,
) (*SessionResult, error) {
	flow, resp, err := c.client.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, errors.Errorf("failed to create registration flow: %w", err)
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

	reg, regResp, err := c.client.FrontendAPI.UpdateRegistrationFlow(ctx).
		Flow(flow.GetId()).
		UpdateRegistrationFlowBody(updateBody).
		Execute()
	if regResp != nil && regResp.Body != nil {
		defer regResp.Body.Close()
	}
	if err != nil {
		if regResp != nil && regResp.StatusCode == http.StatusBadRequest {
			return nil, errors.Errorf("registration failed — check email/password requirements")
		}
		return nil, errors.Errorf("registration failed: %w", err)
	}

	return &SessionResult{SessionToken: reg.GetSessionToken()}, nil
}

func (c *KratosClient) Logout(ctx context.Context, sessionToken string) error {
	session, sessionResp, err := c.client.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).Execute()
	if sessionResp != nil && sessionResp.Body != nil {
		defer sessionResp.Body.Close()
	}
	if err != nil {
		return nil // session already invalid
	}

	disableResp, err := c.client.IdentityAPI.DisableSession(ctx, session.GetId()).Execute()
	if disableResp != nil && disableResp.Body != nil {
		defer disableResp.Body.Close()
	}
	if err != nil {
		return errors.Errorf("failed to disable session: %w", err)
	}

	return nil
}

func (c *KratosClient) identityToDetail(ident *ory.Identity) *UserDetail {
	detail := &UserDetail{
		IdentityID: ident.GetId(),
	}

	traits, hasTraits := ident.GetTraitsOk()
	if !hasTraits || traits == nil {
		return detail
	}

	traitsMap, isMap := (*traits).(map[string]interface{})
	if !isMap {
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
	session, resp, err := c.client.FrontendAPI.ToSession(ctx).
		XSessionToken(sessionToken).Execute()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, errors.Errorf("kratos session check failed: %w", err)
	}

	identity := session.GetIdentity()
	return c.identityToDetail(&identity), nil
}

func (c *KratosClient) GetIdentity(ctx context.Context, id string) (*UserDetail, error) {
	ident, resp, err := c.client.IdentityAPI.GetIdentity(ctx, id).Execute()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, nil //nolint:nilnil // not found is not an error
		}
		return nil, errors.Errorf("failed to get identity: %w", err)
	}

	return c.identityToDetail(ident), nil
}

func (c *KratosClient) UpdateTraits(
	ctx context.Context,
	id string,
	traits map[string]interface{},
) (*UserDetail, error) {
	current, getResp, err := c.client.IdentityAPI.GetIdentity(ctx, id).Execute()
	if getResp != nil && getResp.Body != nil {
		defer getResp.Body.Close()
	}
	if err != nil {
		return nil, errors.Errorf("failed to get identity for update: %w", err)
	}

	body := ory.UpdateIdentityBody{
		SchemaId: current.GetSchemaId(),
		State:    current.GetState(),
		Traits:   traits,
	}

	updated, updateResp, err := c.client.IdentityAPI.UpdateIdentity(ctx, id).
		UpdateIdentityBody(body).Execute()
	if updateResp != nil && updateResp.Body != nil {
		defer updateResp.Body.Close()
	}
	if err != nil {
		return nil, errors.Errorf("failed to update identity traits: %w", err)
	}

	return c.identityToDetail(updated), nil
}

func (c *KratosClient) ListIdentities(ctx context.Context) ([]UserDetail, error) {
	identities, resp, err := c.client.IdentityAPI.ListIdentities(ctx).Execute()
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, errors.Errorf("failed to list identities: %w", err)
	}

	details := make([]UserDetail, 0, len(identities))
	for _, ident := range identities {
		details = append(details, *c.identityToDetail(&ident))
	}
	return details, nil
}
