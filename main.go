package main

import (
	"fmt"
	"os"

	"github.com/mark-c-hall/gator/internal/command"
	"github.com/mark-c-hall/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v", err)
		os.Exit(1)
	}

	state := command.NewState(cfg)
	commands := command.NewCommands()
	commands.Register("login", command.HandlerLogin)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Not enough arguments given\n")
		os.Exit(1)
	}
	cmd := command.NewCommand(os.Args[1:])
	err = commands.Run(state, cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command '%s': %v\n", cmd.Name, err)
		os.Exit(1)
	}
}
