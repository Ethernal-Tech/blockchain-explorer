package config

import (
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	RPCUrl       string
	DbUser       string
	DbPassword   string
	DbHost       string
	DbPort       string
	DbName       string
	WorkersCount int
	Step         int
}

func LoadConfig() (Config, error) {
	configFile, err := filepath.Abs(".env")

	if err != nil {
		return Config{}, err
	}

	err = read(configFile)

	if err != nil {
		return Config{}, err
	}

	return Config{
		RPCUrl:       viper.GetString("RPCUrl"),
		DbUser:       viper.GetString("DB_USER"),
		DbPassword:   viper.GetString("DB_PASSWORD"),
		DbHost:       viper.GetString("DB_HOST"),
		DbPort:       viper.GetString("DB_PORT"),
		DbName:       viper.GetString("DB_NAME"),
		WorkersCount: viper.GetInt("WORKERS_COUNT"),
		Step:         viper.GetInt("STEP"),
	}, nil
}

// Read - Reading .env file content, during application start up
func read(file string) error {
	viper.SetConfigFile(file)

	return viper.ReadInConfig()
}
