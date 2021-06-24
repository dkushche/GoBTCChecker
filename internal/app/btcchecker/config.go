package btcchecker

type Config struct {
	BindAddr    string `toml:"bind_addr"`
	LogLevel    string `toml:"log_level"`
	StoragePath string `toml:"database_path"`
}

func NewConfig() *Config {
	return &Config{
		BindAddr:    ":8080",
		LogLevel:    "debug",
		StoragePath: "storage/db.csv",
	}
}
