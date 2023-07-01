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
	Checkpoint           uint64
	CheckpointWindow     uint
	CheckpointDistance   uint
	EthLogs              bool
	NFTs                 bool
	IPFSGatewayUrl       string
}

func LoadConfig() (*Config, error) {
	configFile, err := filepath.Abs(".env")

	if err != nil {
		return &Config{}, err
	}

	err = read(configFile)

	if err != nil {
		return &Config{}, err
	}

	config := &Config{}
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
	flag.UintVar(&cfg.WorkersCount, "workers", viper.GetUint("WORKERS_COUNT"), "Number of goroutines to use for fetching data from blockchain")
	flag.UintVar(&cfg.Step, "step", viper.GetUint("STEP"), "Number of requests in one batch sent to the blockchain")
	flag.UintVar(&cfg.CallTimeoutInSeconds, "timeout", viper.GetUint("CALL_TIMEOUT_IN_SECONDS"), "Sets a timeout used for requests sent to the blockchain")
	flag.Uint64Var(&cfg.Checkpoint, "checkpoint", viper.GetUint64("CHECKPOINT"), "Sets the number of the starting block for synchronization and validation")
	flag.UintVar(&cfg.CheckpointWindow, "checkpoint.window", viper.GetUint("CHECKPOINT_WINDOW"), "Sets after how many created blocks the checkpoint is determined")
	flag.UintVar(&cfg.CheckpointDistance, "checkpoint.distance", viper.GetUint("CHECKPOINT_DISTANCE"), "Sets the checkpoint distance from the latest block on the blockchain")
	flag.BoolVar(&cfg.EthLogs, "eth.logs", viper.GetBool("INCLUDE_ETH_LOGS"), "Include Ethereum Logs")
	flag.BoolVar(&cfg.NFTs, "nfts", viper.GetBool("INCLUDE_NFTS"), "Include NFTs (to be included, logs must be included as well)")
	flag.StringVar(&cfg.IPFSGatewayUrl, "ipfs.gateway", viper.GetString("IPFS_GATEWAY_URL"), "IPFS Gateway address")
	flag.Parse()
}

func (cfg *Config) fillDefaults() {
	if cfg.Step == 0 {
		cfg.Step = 1000
	}

	if cfg.WorkersCount == 0 {
		cfg.WorkersCount = 32
	}

	if cfg.Checkpoint == 0 {
		cfg.Checkpoint = 1
	}
}
