package users

import (
	"strings"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
)

type User struct {
	ID     string
	Name   string
	Active bool
}

func New(id, name string, active bool) (*User, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errorsx.ErrUserName
	}
	if strings.TrimSpace(id) == "" {
		return nil, errorsx.ErrUserID
	}

	return &User{
		ID:     id,
		Name:   name,
		Active: active,
	}, nil
}
