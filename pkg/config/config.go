package config

import (
	"github.com/spf13/viper"

	"testapp/pkg/http"
	"testapp/pkg/pgsql"
)

type Config struct {
	HTTP http.Config
	PgSQL pgsql.Config
}

func LoadConfig(filename, ext, path string) (Config, error) {
	viper.SetConfigName(filename) 
	viper.SetConfigType(ext)
	viper.AddConfigPath(path)     

	if err := viper.ReadInConfig(); err != nil { 
		return Config{}, err
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}