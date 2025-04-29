package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
)

// структура для парсинга конфигов из файла
type ServerConfig struct {
	ID   int `json:"id"`
	Port int `json:"port"`
}

// Загрузка списка конфигов из файла
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

// Загрузка конфига для определённого сервера
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

type ClientConfig struct {
	Capacity int `json:"capacity"`
	Rate     int `json:"rate"`
}

type RateLimitConfig struct {
	Default ClientConfig            `json:"default"`
	Clients map[string]ClientConfig `json:"clients"`
}

var (
	rateLimitConfig     *RateLimitConfig
	rateLimitConfigOnce sync.Once
)

func LoadRateLimitConfig(RateLimitsConfigPath string) (*RateLimitConfig, error) {
	var loadErr error
	rateLimitConfigOnce.Do(func() {
		file, err := os.Open(RateLimitsConfigPath)
		if err != nil {
			loadErr = err
			return
		}
		defer file.Close()

		var cfg RateLimitConfig
		if err = json.NewDecoder(file).Decode(&cfg); err != nil {
			loadErr = err
			return
		}
		rateLimitConfig = &cfg
	})
	return rateLimitConfig, loadErr
}

func GetConfig() *RateLimitConfig {
	return rateLimitConfig
}
