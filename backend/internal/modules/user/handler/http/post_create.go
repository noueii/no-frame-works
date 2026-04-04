package http

import (
	"encoding/json"
	"net/http"

	"github.com/noueii/no-frame-works/internal/modules/user"
)

type createUserRequestBody struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func toUserResponse(v user.UserView) userResponse {
	return userResponse{
		ID:    v.ID,
		Name:  v.Name,
		Email: v.Email,
	}
}

// PostCreate handles POST /users.
func (h *Handler) PostCreate(w http.ResponseWriter, r *http.Request) {
	var body createUserRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req := user.CreateUserRequest{
		Name:  body.Name,
		Email: body.Email,
	}

	result, err := h.api.CreateUser(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(result))
}
