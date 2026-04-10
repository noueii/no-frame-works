package identity

import "context"

// TestIdentityClient is a configurable test double for identity.Client.
type TestIdentityClient struct {
	ResSession  *SessionResult
	ResMeDetail *UserDetail
	Err         error
}

func (c *TestIdentityClient) Login(_ context.Context, _, _ string) (*SessionResult, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResSession, nil
}

func (c *TestIdentityClient) Register(_ context.Context, _, _ string) (*SessionResult, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResSession, nil
}

func (c *TestIdentityClient) Logout(_ context.Context, _ string) error {
	return c.Err
}

func (c *TestIdentityClient) GetSession(_ context.Context, _ string) (*UserDetail, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResMeDetail, nil
}

// GetDefaultTestIdentityClient returns a TestIdentityClient with sensible defaults.
func GetDefaultTestIdentityClient() *TestIdentityClient {
	return &TestIdentityClient{
		ResSession: &SessionResult{SessionToken: "test-session-token"},
		ResMeDetail: &UserDetail{
			IdentityID: "00000000-0000-0000-0000-000000000001",
			Email:      "test@example.com",
		},
	}
}
