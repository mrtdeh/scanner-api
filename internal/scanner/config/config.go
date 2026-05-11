package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	DEBUG           = "debug"
	LOG_LEVEL       = "log_level"
	GRPC_HOST       = "grpc_host"
	GRPC_PORT       = "grpc_port"
	MONGO_USER      = "mongo_user"
	MONGO_PASS      = "mongo_pass"
	MONGO_HOST      = "mongo_host"
	MONGO_PORT      = "mongo_port"
	MONGO_DB        = "mongo_db"
	YARA_RULES_PATH = "yara_rules_path"
	YARA_IMAGE      = "yara_image"
)

type Config struct {
	GRPCHost      string `mapstructure:"grpc_host"`
	GRPCPort      uint   `mapstructure:"grpc_port"`
	MongoUsername string `mapstructure:"mongo_user"`
	MongoPassword string `mapstructure:"mongo_pass"`
	MongoHost     string `mapstructure:"mongo_host"`
	MongoPort     uint   `mapstructure:"mongo_port"`
	MongoDatabase string `mapstructure:"mongo_db"`
	YaraImage     string `mapstructure:"yara_image"`
	YaraRulesPath string `mapstructure:"yara_rules_path"`
}

func LoadConfiguration() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Bind environment variables
	viper.BindEnv(DEBUG, "DEBUG")
	viper.BindEnv(LOG_LEVEL, "LOG_LEVEL")
	viper.BindEnv(GRPC_HOST, "GRPC_HOST")
	viper.BindEnv(GRPC_PORT, "GRPC_PORT")
	viper.BindEnv(YARA_RULES_PATH, "YARA_RULES_PATH")
	viper.BindEnv(MONGO_USER, "MONGO_USER")
	viper.BindEnv(MONGO_PASS, "MONGO_PASS")
	viper.BindEnv(MONGO_HOST, "MONGO_HOST")
	viper.BindEnv(MONGO_PORT, "MONGO_PORT")
	viper.BindEnv(MONGO_DB, "MONGO_DB")
	viper.BindEnv(YARA_IMAGE, "YARA_IMAGE")
	viper.BindEnv(YARA_RULES_PATH, "YARA_RULES_PATH")

	// Set the default configuration
	viper.SetDefault(DEBUG, false)
	viper.SetDefault(LOG_LEVEL, "info")
	viper.SetDefault(GRPC_PORT, 10000)
	viper.SetDefault(MONGO_PORT, 27017)
	viper.SetDefault(YARA_IMAGE, "yara")

	var cnf Config
	err := viper.Unmarshal(&cnf)

	return &cnf, err

}
