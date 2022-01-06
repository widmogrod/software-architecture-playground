package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	db "github.com/widmogrod/software-architecture-playground/hypo/prisma-go/sdk"
	"testing"
)

func TestScenario(t *testing.T) {
	client := db.NewClient()
	err := client.Prisma.Connect()
	assert.NoError(t, err)

	defer func() {
		err := client.Prisma.Disconnect()
		assert.NoError(t, err)
	}()

	ctx := context.Background()

	postID := uuid.Must(uuid.NewUUID()).String()

	t.Run("create post", func(t *testing.T) {
		post, err := client.Post.CreateOne(
			db.Post.Title.Set("My new post"),
			db.Post.Published.Set(true),
			db.Post.Desc.Set("Hi there."),
			db.Post.ID.Set(postID),
		).Exec(ctx)
		assert.NoError(t, err)

		t.Logf("post: %+v", post)
	})

	t.Run("create comment and link with post", func(t *testing.T) {
		// then create a comment
		comments, err := client.Comment.CreateOne(
			db.Comment.Content.Set("my description"),
			// link the post we created before
			db.Comment.Post.Link(
				db.Post.ID.Equals(postID),
			),
		).Exec(ctx)
		assert.NoError(t, err)

		t.Logf("post: %+v", comments)
	})

	t.Run("find post without comment", func(t *testing.T) {
		post, err := client.Post.
			FindUnique(db.Post.ID.Equals(postID)).
			Exec(ctx)

		assert.NoError(t, err)
		assert.Equal(t, postID, post.ID)

		niceJSON, err := json.MarshalIndent(post, "", "")
		assert.NoError(t, err)
		assert.Panics(t, func() {
			_ = post.Comments()
		})
		t.Logf("post: %s\n", niceJSON)

	})
	t.Run("find post with comments", func(t *testing.T) {
		post, err := client.Post.
			FindUnique(db.Post.ID.Equals(postID)).
			With(db.Post.Comments.Fetch()).
			Exec(ctx)

		assert.NoError(t, err)
		assert.Equal(t, postID, post.ID)

		niceJSON, err := json.MarshalIndent(post, "", "")
		assert.NoError(t, err)
		assert.NotPanics(t, func() {
			_ = post.Comments()
		})
		t.Logf("post: %s\n", niceJSON)
	})
}
