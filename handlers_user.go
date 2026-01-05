package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/strjkc/gator/internal/database"
)

func handlerListUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Unable to retrieve users: %s", err)
		return err
	}
	currUser := s.conf.User
	if len(users) < 1 {
		fmt.Println("No users found")
		return nil
	}
	for _, user := range users {
		if user == currUser {
			fmt.Printf("* %s (current)\n", user)
		} else {

			fmt.Printf("* %s\n", user)
		}
	}
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("invalid command")
	}
	args := database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Name: cmd.args[0]}
	user, err := s.db.CreateUser(context.Background(), args)
	if err != nil {
		fmt.Printf("An error occured, unable to create user: %s", err)
		os.Exit(1)
	}
	fmt.Printf("User created: Id: %s, Created At: %s, UpdatedAt: %s, Name: %s", user.ID, user.CreatedAt, user.UpdatedAt, user.Name)
	handlerLogin(s, command{name: "login", args: []string{user.Name}})
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("invalid command")
	}

	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Printf("User %s doesn't exists", user.Name)
		os.Exit(1)
	}
	fmt.Printf("User from db: %s", user)
	if err := s.conf.SetUser(user.Name); err != nil {
		return err
	}
	fmt.Println("User set to", cmd.args[0])
	return nil
}
