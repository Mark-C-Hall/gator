package command

import (
	"context"
	"fmt"

	"github.com/mark-c-hall/gator/internal/database"
)

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(s *State, cmd Command) error {
	return func(s *State, cmd Command) error {
		loggedInUser, err := s.DB.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("error getting username: %w", err)
		}
		return handler(s, cmd, loggedInUser)
	}
}
