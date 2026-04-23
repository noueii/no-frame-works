package handler

import (
	"github.com/google/uuid"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/app/domain"
)

// toOAPIPost converts a *domain.Post into an oapi.Post for HTTP responses.
// Lives in the handler layer because it's the place where domain types are
// projected into the HTTP contract — services themselves return *domain.Post
// unchanged, and this function is the only place that knows about oapi.Post.
func toOAPIPost(p *domain.Post) oapi.Post {
	return oapi.Post{
		Id:       uuid.MustParse(p.ID),
		Title:    p.Title,
		Content:  p.Content,
		AuthorId: p.AuthorID,
	}
}

// toOAPIPosts converts a slice of domain.Post into a slice of oapi.Post.
func toOAPIPosts(posts []domain.Post) []oapi.Post {
	result := make([]oapi.Post, len(posts))
	for i := range posts {
		result[i] = toOAPIPost(&posts[i])
	}
	return result
}
