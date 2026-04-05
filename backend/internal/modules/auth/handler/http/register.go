package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	ory "github.com/ory/kratos-client-go"
)

type registerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

type registerResponse struct {
	SessionToken string `json:"sessionToken"`
}

// Register handles POST /auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Email == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	flow, _, err := h.kratos.FrontendAPI.CreateNativeRegistrationFlow(r.Context()).Execute()
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create registration flow: %v", err))
		return
	}

	traits := map[string]interface{}{
		"email": body.Email,
	}
	if body.FirstName != "" {
		traits["first_name"] = body.FirstName
	}
	if body.LastName != "" {
		traits["last_name"] = body.LastName
	}

	updateBody := ory.UpdateRegistrationFlowBody{
		UpdateRegistrationFlowWithPasswordMethod: &ory.UpdateRegistrationFlowWithPasswordMethod{
			Method:   "password",
			Password: body.Password,
			Traits:   traits,
		},
	}

	reg, resp, err := h.kratos.FrontendAPI.UpdateRegistrationFlow(r.Context()).
		Flow(flow.GetId()).
		UpdateRegistrationFlowBody(updateBody).
		Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusBadRequest {
			writeError(w, http.StatusBadRequest, "registration failed — check email/password requirements")
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("registration failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, registerResponse{
		SessionToken: reg.GetSessionToken(),
	})
}
