package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	DEBUG        = "debug"
	LOG_LEVEL    = "log_level"
	HTTP_HOST    = "http_host"
	HTTP_PORT    = "http_port"
	SCANNER_ADDR = "scanner_addr"
)

type Config struct {
	HttpHost       string `mapstructure:"http_host"`
	HttpPort       uint   `mapstructure:"http_port"`
	ScannerAddress string `mapstructure:"scanner_addr"`
}

func LoadConfiguration() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Bind environment variables
	viper.BindEnv(DEBUG, "DEBUG")
	viper.BindEnv(LOG_LEVEL, "LOG_LEVEL")
	viper.BindEnv(HTTP_HOST, "HTTP_HOST")
	viper.BindEnv(HTTP_PORT, "HTTP_PORT")
	viper.BindEnv(SCANNER_ADDR, "SCANNER_ADDR")

	// Set the default configuration
	viper.SetDefault(DEBUG, false)
	viper.SetDefault(LOG_LEVEL, "info")
	viper.SetDefault(HTTP_HOST, "localhost")
	viper.SetDefault(HTTP_PORT, 8080)
	viper.SetDefault(SCANNER_ADDR, "")

	var cnf Config
	err := viper.Unmarshal(&cnf)

	return &cnf, err

}
