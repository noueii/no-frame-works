package identity

import "context"

// TestIdentityClient is a configurable test double for identity.Client.
type TestIdentityClient struct {
	ResMeDetail *UserDetail
	Err         error
}

func (c *TestIdentityClient) GetMeDetail(_ context.Context, _ string) (*UserDetail, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	return c.ResMeDetail, nil
}

// GetDefaultTestIdentityClient returns a TestIdentityClient with sensible defaults.
func GetDefaultTestIdentityClient() *TestIdentityClient {
	return &TestIdentityClient{
		ResMeDetail: &UserDetail{
			IdentityID: "00000000-0000-0000-0000-000000000001",
			Email:      "test@example.com",
			FirstName:  "Test",
			LastName:   "User",
		},
	}
}
