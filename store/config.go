package store

type Config struct {
	DatabasePath string `toml:"database_path"`
}

func NewConfig() *Config {
	return &Config{}
}