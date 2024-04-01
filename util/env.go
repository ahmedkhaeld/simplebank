package util

import (
	"time"

	"github.com/spf13/viper"
)

type Env struct {
	Environment          string        `mapstructure:"ENV"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

func LoadEnv(path string) (env Env, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("dev")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&env)
	return
}
