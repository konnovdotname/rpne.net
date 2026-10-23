package main

import (
	"sort"
	"time"
)

func ownReplies(root *Thread) []Post {
	var posts []Post
	seen := make(map[string]bool)

	var walk func(*Thread)

	walk = func(parent *Thread) {
		for i := range parent.Replies {
			child := &parent.Replies[i]

			if child.Post.URI == "" ||
				child.Post.Author.DID != root.Post.Author.DID ||
				child.Post.Record.Reply == nil ||
				child.Post.Record.Reply.Parent.URI != parent.Post.URI ||
				seen[child.Post.URI] {
				continue
			}

			seen[child.Post.URI] = true
			posts = append(posts, child.Post)
			walk(child)
		}
	}

	walk(root)

	sort.SliceStable(posts, func(i, j int) bool {
		a, errA := time.Parse(
			time.RFC3339Nano,
			posts[i].Record.CreatedAt,
		)
		b, errB := time.Parse(
			time.RFC3339Nano,
			posts[j].Record.CreatedAt,
		)

		if errA == nil && errB == nil && !a.Equal(b) {
			return a.Before(b)
		}

		return posts[i].URI < posts[j].URI
	})

	return posts
}
