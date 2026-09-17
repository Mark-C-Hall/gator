package main

import (
	"fmt"
	"os"

	"github.com/mark-c-hall/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v", err)
		os.Exit(1)
	}

	err = cfg.SetUser("mark")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting username: %v", err)
	}

	cfg, err = config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v", err)
		os.Exit(1)
	}
	fmt.Printf("DB_URL: %s, Username: %s\n", cfg.DB_URL, cfg.CurrentUserName)
}
