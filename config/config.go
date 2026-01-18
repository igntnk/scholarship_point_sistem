package config

import (
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strings"
)

const (
	EnvPrefix = "SPS"
)

type Config struct {
	Database struct {
		URI string `yaml:"uri" mapstructure:"uri"`
	} `yaml:"database" mapstructure:"database"`
	Server struct {
		RESTPort int `mapstructure:"rest_port"`
	} `yaml:"server" mapstructure:"server"`
	Secure struct {
		PasswordBcryptCost   int    `mapstructure:"password_bcrypt_cost"`
		AccessTokenDuration  int    `mapstructure:"token_duration"`
		RefreshTokenDuration int    `mapstructure:"refresh_token_duration"`
		AdminGroupName       string `mapstructure:"admin_group_name"`
		AdminRoleName        string `mapstructure:"admin_role_name"`
		JWTPrivateKeyPath    string `mapstructure:"jwt_private_key_path"`
		AdminPassword        string `mapstructure:"admin_password"`
		AdminEmail           string `mapstructure:"admin_email"`
	} `yaml:"secure" mapstructure:"secure"`
}

func Get(logger zerolog.Logger) *Config {
	v := viper.New()
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AddConfigPath("./config/")
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	err := v.ReadInConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read config")
	}

	for _, key := range v.AllKeys() {
		val := v.Get(key)
		if val == nil {
			continue
		}

		if reflect.TypeOf(val).Kind() == reflect.String {
			v.Set(key, os.ExpandEnv(val.(string)))
		}
	}

	var cfg *Config
	err = v.Unmarshal(&cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to unmarshal config")
	}

	if cfg.Secure.AdminGroupName == "" {
		cfg.Secure.AdminGroupName = "Админитраторы"
	}

	if cfg.Secure.AdminRoleName == "" {
		cfg.Secure.AdminRoleName = "Админитраторы"
	}

	if cfg.Server.RESTPort == 0 {
		cfg.Server.RESTPort = 10000
	}

	if cfg.Secure.AccessTokenDuration == 0 {
		cfg.Secure.AccessTokenDuration = 86400
	}

	if cfg.Secure.RefreshTokenDuration == 0 {
		cfg.Secure.RefreshTokenDuration = 172800
	}

	if cfg.Secure.JWTPrivateKeyPath == "" {
		cfg.Secure.JWTPrivateKeyPath = "./cert/jwtRS256.key"
	}

	return cfg
}
