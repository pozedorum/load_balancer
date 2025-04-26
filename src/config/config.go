package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

type ServerConfig struct {
	ID   int `json:"id"`
	Port int `json:"port"`
}

func LoadServerConfigList(path string) ([]ServerConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var configs []ServerConfig
	if err := json.Unmarshal(file, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

func FindPortInConfig(path, serverID string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	serverId, err := strconv.Atoi(serverID)
	if err != nil {
		return "", err
	}

	var configs []ServerConfig
	if err := json.Unmarshal(file, &configs); err != nil {
		return "", err
	}

	for _, config := range configs {
		if config.ID == serverId {
			return strconv.Itoa(config.Port), nil
		}
	}

	return "", errors.New("no such serverId in config")
}
