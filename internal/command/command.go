package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"time"

	"github.com/mark-c-hall/gator/internal/config"
	"github.com/mark-c-hall/gator/internal/database"
	"github.com/mark-c-hall/gator/internal/feed"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

type State struct {
	Config *config.Config
	DB     *database.Queries
}

func NewState(cfg *config.Config, db *database.Queries) *State {
	return &State{Config: cfg, DB: db}
}

type Command struct {
	Name string
	Args []string
}

func NewCommand(args []string) Command {
	cmd := Command{Name: args[0]}
	if len(args) > 1 {
		cmd.Args = args[1:]
	}
	return cmd
}

type Commands struct {
	m map[string]func(*State, Command) error
}

func NewCommands() *Commands {
	return &Commands{m: make(map[string]func(*State, Command) error)}
}

func (c *Commands) Run(s *State, cmd Command) error {
	if handler, ok := c.m[cmd.Name]; ok {
		return handler(s, cmd)
	}
	return errors.New("command not found in registered command")
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.m[name] = f
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("login handler expects a single argument: username")
	}
	username := cmd.Args[0]

	_, err := s.DB.GetUser(context.Background(), username)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("user is not registered in database")
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("login handler: error setting username: %w", err)
	}
	fmt.Printf("User '%s' has been set\n", username)
	return nil
}

func HandlderRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("register handler expects a signel argument: username")
	}
	username := cmd.Args[0]

	_, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		UserID:    uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == pqerror.UniqueViolation {
		return errors.New("username already exists")
	}
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("error setting username: %w", err)
	}

	fmt.Printf("User '%s' was created\n", username)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.DB.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting users: %w", err)
	}
	fmt.Println("All users deleted")
	return nil
}

func HandlerGetUsers(s *State, cmd Command) error {
	users, err := s.DB.GetAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all users: %w", err)
	}
	for _, user := range users {
		if user.Name == s.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	const feedURL = "https://www.wagslane.dev/index.xml"
	feed, err := feed.FetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	fmt.Println(feed)
	return nil
}
