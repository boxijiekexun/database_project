package config

import "os"

type Config struct {
	Address       string
	MySQLDSN      string
	JWTSecret     string
	TokenTTLHours int
}

func Load() Config {
	address := os.Getenv("ADDR")
	if address == "" {
		address = ":8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-key"
	}

	return Config{
		Address:       address,
		MySQLDSN:      os.Getenv("MYSQL_DSN"),
		JWTSecret:     secret,
		TokenTTLHours: 24,
	}
}
