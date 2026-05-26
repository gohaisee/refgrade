package service

import "entgo.io/ent"

type User struct {
	Edges struct {
		Posts []Post
	}
}

type Post struct{}

func Walk(users []User) {
	for _, u := range users {
		_ = u.Edges
	}
}
