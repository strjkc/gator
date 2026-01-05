package main

import (
	"context"
	"fmt"
)

func handlerClearUsers(s *state, cmd command) error {
	err := s.db.RemoveUsers(context.Background())
	if err != nil {
		fmt.Printf("Error removing users: %s\n", err.Error())
		return err
	}
	return nil
}
