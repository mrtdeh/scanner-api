package apiconfig

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	DEBUG           = "debug"
	LOG_LEVEL       = "log_level"
	HTTP_HOST       = "http_host"
	HTTP_PORT       = "http_port"
	YARA_RULES_PATH = "yara_rules_path"
)

type Config struct {
	HttpHost string `mapstructure:"http_host"`
	HttpPort uint   `mapstructure:"http_port"`
}

func LoadConfiguration() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Bind environment variables
	viper.BindEnv(DEBUG, "DEBUG")
	viper.BindEnv(LOG_LEVEL, "LOG_LEVEL")
	viper.BindEnv(HTTP_HOST, "HTTP_HOST")
	viper.BindEnv(HTTP_PORT, "HTTP_PORT")
	viper.BindEnv(YARA_RULES_PATH, "YARA_RULES_PATH")

	// Set the default configuration
	viper.SetDefault(DEBUG, false)
	viper.SetDefault(LOG_LEVEL, "info")
	viper.SetDefault(HTTP_HOST, "localhost")
	viper.SetDefault(HTTP_PORT, 8080)

	var cnf Config
	err := viper.Unmarshal(&cnf)

	return &cnf, err

}
