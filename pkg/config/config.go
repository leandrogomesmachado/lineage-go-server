package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	MaxConnections int    `yaml:"max_connections"`
	MaxPlayers     int    `yaml:"max_players"`
	ID             int    `yaml:"id"`
	Name           string `yaml:"name"`
}

type DatabaseConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
	Timeout  int    `yaml:"timeout"`
}

type SecurityConfig struct {
	AutoCreateAccounts bool `yaml:"auto_create_accounts"`
	MaxLoginAttempts   int  `yaml:"max_login_attempts"`
	BanDurationMinutes int  `yaml:"ban_duration_minutes"`
}

type GameServerInfo struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LoginServerInfo struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatapackConfig struct {
	Path string `yaml:"path"`
}

type GeodataConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type WorldConfig struct {
	GridSize int `yaml:"grid_size"`
	RegionsX int `yaml:"regions_x"`
	RegionsY int `yaml:"regions_y"`
}

type RatesConfig struct {
	XP    float64 `yaml:"xp"`
	SP    float64 `yaml:"sp"`
	Adena float64 `yaml:"adena"`
	Drop  float64 `yaml:"drop"`
	Spoil float64 `yaml:"spoil"`
}

type LoggingConfig struct {
	Level   string `yaml:"level"`
	File    string `yaml:"file"`
	Console bool   `yaml:"console"`
}

type LoginServerConfig struct {
	Server      ServerConfig     `yaml:"server"`
	LoginServer LoginServerInfo  `yaml:"loginserver"`
	Database    DatabaseConfig   `yaml:"database"`
	Security    SecurityConfig   `yaml:"security"`
	GameServers []GameServerInfo `yaml:"gameservers"`
	Logging     LoggingConfig    `yaml:"logging"`
}

type GameServerConfig struct {
	Server      ServerConfig    `yaml:"server"`
	LoginServer LoginServerInfo `yaml:"loginserver"`
	Database    DatabaseConfig  `yaml:"database"`
	Datapack    DatapackConfig  `yaml:"datapack"`
	Geodata     GeodataConfig   `yaml:"geodata"`
	World       WorldConfig     `yaml:"world"`
	Rates       RatesConfig     `yaml:"rates"`
	Logging     LoggingConfig   `yaml:"logging"`
}

func LoadLoginServerConfig(path string) (*LoginServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config LoginServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadGameServerConfig(path string) (*GameServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config GameServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
