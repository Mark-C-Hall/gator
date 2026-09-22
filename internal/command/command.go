package command

import (
	"errors"
	"fmt"

	"github.com/mark-c-hall/gator/internal/config"
)

type State struct {
	Config *config.Config
}

func NewState(cfg *config.Config) *State {
	return &State{Config: cfg}
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
	err := s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("login handler: error setting username: %w", err)
	}
	fmt.Printf("User '%s' has been set\n", username)
	return nil
}
