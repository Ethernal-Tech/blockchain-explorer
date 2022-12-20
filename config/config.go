package config

import (
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	HTTPUrl              string
	WebSocketUrl         string
	DbUser               string
	DbPassword           string
	DbHost               string
	DbPort               string
	DbName               string
	DbSSL                string
	WorkersCount         uint
	Step                 uint
	CallTimeoutInSeconds uint
	Mode                 string
	CheckPointWindow     uint
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

	config := Config{
		HTTPUrl:              viper.GetString("HTTPUrl"),
		WebSocketUrl:         viper.GetString("WebSocketUrl"),
		DbUser:               viper.GetString("DB_USER"),
		DbPassword:           viper.GetString("DB_PASSWORD"),
		DbHost:               viper.GetString("DB_HOST"),
		DbPort:               viper.GetString("DB_PORT"),
		DbName:               viper.GetString("DB_NAME"),
		DbSSL:                viper.GetString("DB_SSL"),
		WorkersCount:         viper.GetUint("WORKERS_COUNT"),
		Step:                 viper.GetUint("STEP"),
		CallTimeoutInSeconds: viper.GetUint("CALL_TIMEOUT_IN_SECONDS"),
		Mode:                 viper.GetString("MODE"),
		CheckPointWindow:     viper.GetUint("CHECKPOINT_WINDOW"),
	}

	config.fillDefaults()

	return config, nil
}

// Read - Reading .env file content, during application start up
func read(file string) error {
	viper.SetConfigFile(file)
	return viper.ReadInConfig()
}

func (cfg *Config) fillDefaults() {
	if cfg.Step == 0 {
		cfg.Step = 1000
	}

	if cfg.WorkersCount == 0 {
		cfg.WorkersCount = 32
	}
}
