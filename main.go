package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/strjkc/gator/internal/config"
	"github.com/strjkc/gator/internal/database"
)

type state struct {
	conf *config.Config
	db   *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
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

func (c *commands) run(s *state, cmd command) error {
	fun := c.handlers[cmd.name]
	return fun(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	stat := &state{}
	stat.conf = &conf
	dbURL := stat.conf.Dburl
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	stat.db = dbQueries
	cmds := &commands{handlers: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatal("Usage: gator login <username> <password>")
	}
	cmdName := args[0]
	cmdArgs := args[1:]
	cmds.run(stat, command{name: cmdName, args: cmdArgs})
}
