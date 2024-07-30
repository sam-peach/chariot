package userservice

import (
	e "chariottakehome/api/errors"
	"chariottakehome/internal/users"
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
)

// not tested exhaustively, but catches most cases
const emailRegexPattern string = "^[a-zA-Z0-9\\._-]+@[a-zA-Z0-9\\._-]+\\.\\w+$"

var emailRegex *regexp.Regexp

func init() {
	emailRegex = regexp.MustCompile(emailRegexPattern)
}

type UserService struct {
	UnimplementedUserServiceServer
	Repo users.UserRepository
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	reqEmail := req.GetEmail()

	if !emailRegex.MatchString(reqEmail) {
		return nil, e.RequestError{Err: errors.New("invalid email address")}
	}

	user, err := s.Repo.CreateUser(ctx, req.Email)
	if err != nil {
		fmt.Println(err)
		// TODO: Better error handling and returning more useful error messages to the client.
		// Example: duplicate emails.
		return nil, e.ApiError{Err: e.Internal}
	}

	return toProtoUser(user), nil
}

func toProtoUser(user *users.User) *User {
	return &User{
		Id:        user.Id.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
