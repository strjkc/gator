package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/strjkc/gator/internal/config"
)

type state struct {
	conf *config.Config
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
		return errors.New("Invalid command")
	}
	if err := s.conf.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Println("User set to", cmd.args[0])
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
	cmds := &commands{handlers: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatal("Usage: gator login <username> <password>")
	}
	cmdName := args[0]
	cmdArgs := args[1:]
	cmds.run(stat, command{name: cmdName, args: cmdArgs})
}
