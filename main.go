package main

import (
	"fmt"
	config "github.com/strjkc/gator/internal/config"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		fmt.Errorf("Error reading config: %v", err)
	}
	fmt.Printf("%+v\n", conf)
	err = conf.SetUser("Majka")
	if err != nil {
		fmt.Errorf("Error setting user: %v", err)
	}
	conf, err = config.Read()
	if err != nil {
		fmt.Errorf("Error reading config: %v", err)
	}
	fmt.Printf("%+v\n", conf)
}
