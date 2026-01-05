package main

import (
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
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", handlerGetFeeds)
	cmds.register("follow", handlerFollow)
	cmds.register("following", handlerFollowing)
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
