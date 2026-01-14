package main

import (
	"context"
	"database/sql"
	"log"
	"os"

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

func (c *commands) run(s *state, cmd command) error {
	fun := c.handlers[cmd.name]
	return fun(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func registerHandlers(cmds *commands) {
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerClearUsers)
	cmds.register("users", handlerListUsers)
	cmds.register("agg", handlerFetch)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerGetFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", handlerFollowing)
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		currUser := s.conf.User
		userData, err := s.db.GetUser(context.Background(), currUser)
		if err != nil {
			return err
		}
		return handler(s, cmd, userData)
	}
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
	registerHandlers(cmds)
	args := os.Args[1:]
	cmdName := args[0]
	cmdArgs := args[1:]
	cmds.run(stat, command{name: cmdName, args: cmdArgs})
}
