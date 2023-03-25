package config

import (
	"flag"
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
	CheckPointDistance   uint
	EthLogs              bool
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
	config := Config{}
	config.fillConfigurations()
	config.fillDefaults()

	return config, nil
}

// Read - Reading .env file content, during application start up
func read(file string) error {
	viper.SetConfigFile(file)
	return viper.ReadInConfig()
}

func (cfg *Config) fillConfigurations() {
	flag.StringVar(&cfg.HTTPUrl, "http.addr", viper.GetString("HTTPUrl"), "Blockchain node HTTP address")
	flag.StringVar(&cfg.WebSocketUrl, "ws.addr", viper.GetString("WebSocketUrl"), "Blockchain node WebSocket address")
	flag.StringVar(&cfg.DbUser, "db.user", viper.GetString("DB_USER"), "Database user")
	flag.StringVar(&cfg.DbPassword, "db.password", viper.GetString("DB_PASSWORD"), "Database user password")
	flag.StringVar(&cfg.DbHost, "db.host", viper.GetString("DB_HOST"), "Database server host")
	flag.StringVar(&cfg.DbPort, "db.port", viper.GetString("DB_PORT"), "Database server port")
	flag.StringVar(&cfg.DbName, "db.name", viper.GetString("DB_NAME"), "Database name")
	flag.StringVar(&cfg.DbSSL, "db.ssl", viper.GetString("DB_SSL"), "Enable (verify-full) or disable TLS")
	flag.StringVar(&cfg.Mode, "mode", viper.GetString("MODE"), "Manual or automatic mode of application")
	flag.UintVar(&cfg.WorkersCount, "workers value", viper.GetUint("WORKERS_COUNT"), "Number of goroutines to use for fetching data from blockchain")
	flag.UintVar(&cfg.Step, "step value", viper.GetUint("STEP"), "Number of requests in one batch sent to the blockchain")
	flag.UintVar(&cfg.CallTimeoutInSeconds, "timeout value", viper.GetUint("CALL_TIMEOUT_IN_SECONDS"), "Sets a timeout used for requests sent to the blockchain")
	flag.UintVar(&cfg.CheckPointWindow, "checkpoint.window value", viper.GetUint("CHECKPOINT_WINDOW"), "Sets after how many created blocks the checkpoint is determined")
	flag.UintVar(&cfg.CheckPointDistance, "checkpoint.distance value", viper.GetUint("CHECKPOINT_DISTANCE"), "Sets the checkpoint distance from the latest block on the blockchain")
	flag.BoolVar(&cfg.EthLogs, "eth.logs", viper.GetBool("INCLUDE_ETH_LOGS"), "Include Ethereum Logs")
	flag.Parse()
}

func (cfg *Config) fillDefaults() {
	if cfg.Step == 0 {
		cfg.Step = 1000
	}

	if cfg.WorkersCount == 0 {
		cfg.WorkersCount = 32
	}
}
