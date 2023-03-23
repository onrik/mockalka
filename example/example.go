package example

import "context"

type User struct {
	ID   int
	Name string
}

type DB interface {
	User(ctx context.Context, id int) (*User, error)
	Users(ctx context.Context) ([]User, error)
	UserByName(context.Context, string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, id int, fields map[string]map[int]string) error
}
