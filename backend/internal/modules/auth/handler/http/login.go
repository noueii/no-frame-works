package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	ory "github.com/ory/kratos-client-go"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	SessionToken string `json:"sessionToken"`
}

// Login handles POST /auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Email == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	flow, _, err := h.kratos.FrontendAPI.CreateNativeLoginFlow(r.Context()).Execute()
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to create login flow: %v", err),
		)
		return
	}

	updateBody := ory.UpdateLoginFlowBody{
		UpdateLoginFlowWithPasswordMethod: &ory.UpdateLoginFlowWithPasswordMethod{
			Method:     "password",
			Identifier: body.Email,
			Password:   body.Password,
		},
	}

	login, resp, err := h.kratos.FrontendAPI.UpdateLoginFlow(r.Context()).
		Flow(flow.GetId()).
		UpdateLoginFlowBody(updateBody).
		Execute()
	if err != nil {
		if resp != nil &&
			(resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("login failed: %v", err))
		return
	}

	token := login.GetSessionToken()

	http.SetCookie(w, &http.Cookie{
		Name:     "ory_kratos_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, loginResponse{
		SessionToken: token,
	})
}
