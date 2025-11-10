package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	// ? Load .env if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env found using system environment variables")
	}

	// * Set defaults
	setDefaults()

	// * Read environment variables (including those from .env file)
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	// ? Server Config
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.base_url", "http://localhost:8080")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.gin_mode", "debug")

	// ? App Config (Default TTL: 720 hours = 30 days = 1 month)
	viper.SetDefault("app.default_ttl", "720h")
	viper.SetDefault("app.max_ttl", "8760h") // 1 year
	viper.SetDefault("app.short_code_length", 7)
	viper.SetDefault("app.custom_code_length", 50)
	viper.SetDefault("app.max_custom_code_length", 50)
	viper.SetDefault("app.enable_stats", true)
	viper.SetDefault("app.enable_geoip", true)

	// ? Database defaults
	viper.SetDefault("database.uri", "mongodb://localhost:27017")
	viper.SetDefault("database.name", "tether")
	viper.SetDefault("database.max_pool_size", 100)
	viper.SetDefault("database.min_pool_size", 10)

	// ? Cache defaults (Redis)
	viper.SetDefault("cache.addr", "localhost:6379")
	viper.SetDefault("cache.password", "")
	viper.SetDefault("cache.db", 0)
	viper.SetDefault("cache.ttl", "24h")

	// ? Security defaults
	viper.SetDefault("security.rate_limit", 100)
	viper.SetDefault("security.rate_limit_window", "1h")
	viper.SetDefault("security.allowed_origins", []string{"*"})
	viper.SetDefault("security.trusted_proxies", []string{})
	viper.SetDefault("security.enable_cors", true)
	viper.SetDefault("security.blocked_domains", []string{})
	viper.SetDefault("security.required_headers", []string{})

	// ? Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// * Environment variable bindings
	// ! Note: viper.BindEnv is designed to not return errors for valid inputs
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.host", "SERVER_HOST")
	_ = viper.BindEnv("server.base_url", "BASE_URL")
	_ = viper.BindEnv("server.gin_mode", "GIN_MODE")

	_ = viper.BindEnv("database.uri", "MONGO_URI")
	_ = viper.BindEnv("database.name", "MONGO_DATABASE")

	_ = viper.BindEnv("cache.db", "REDIS_DB")
	_ = viper.BindEnv("cache.addr", "REDIS_ADDR")
	_ = viper.BindEnv("cache.password", "REDIS_PASSWORD")

	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
}
