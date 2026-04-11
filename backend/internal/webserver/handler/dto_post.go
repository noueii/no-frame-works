package handler

import (
	"github.com/google/uuid"

	"github.com/noueii/no-frame-works/generated/oapi"
	"github.com/noueii/no-frame-works/internal/modules/post"
)

func toOAPIPost(v post.View) oapi.Post {
	return oapi.Post{
		Id:       uuid.MustParse(v.ID),
		Title:    v.Title,
		Content:  v.Content,
		AuthorId: v.AuthorID,
	}
}

func toOAPIPosts(views []post.View) []oapi.Post {
	posts := make([]oapi.Post, len(views))
	for i, v := range views {
		posts[i] = toOAPIPost(v)
	}
	return posts
}
