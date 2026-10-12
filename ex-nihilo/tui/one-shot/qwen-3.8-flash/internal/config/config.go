package config

import "os"

type Config struct {
	DBDir string
}

func Load() Config {
	dir := os.Getenv("DB_DIR")
	if dir == "" {
		dir = "./db"
	}

	return Config{
		DBDir: dir,
	}
}
