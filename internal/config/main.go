package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Dburl string `json:"db_url"`
	User  string `json:"current_user_name"`
}

func Read() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	file, err := os.ReadFile(homeDir + "/gatorconfig.json")
	if err != nil {
		return Config{}, err
	}

	data := Config{}
	if err := json.Unmarshal(file, &data); err != nil {
		return Config{}, err
	}
	return data, nil
}

func (c Config) SetUser(uname string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	c.User = uname
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(homeDir+"/gatorconfig.json", data, 0644)
	if err != nil {
		return err
	}
	return nil
}
