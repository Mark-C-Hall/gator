package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/mark-c-hall/gator/internal/command"
	"github.com/mark-c-hall/gator/internal/config"
	"github.com/mark-c-hall/gator/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v", err)
	}
	dbQueries := database.New(db)

	state := command.NewState(cfg, dbQueries)
	commands := command.NewCommands()
	commands.Register("login", command.HandlerLogin)
	commands.Register("register", command.HandlderRegister)
	commands.Register("reset", command.HandlerReset)
	commands.Register("users", command.HandlerGetUsers)
	commands.Register("agg", command.HandlerAgg)
	commands.Register("addfeed", command.HandlerAddFeed)
	commands.Register("feeds", command.HandlerFeeds)
	commands.Register("follow", command.HandlerFollow)
	commands.Register("following", command.HandlerFollowing)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Not enough arguments given\n")
		os.Exit(1)
	}
	cmd := command.NewCommand(os.Args[1:])
	err = commands.Run(state, cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running %s command: %v\n", cmd.Name, err)
		os.Exit(1)
	}
}
