package config

import "time"

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type SecurityConfig struct {
	AllowedOrigins  []string `mapstructure:"allowed_origins"`
	TrustedProxies  []string `mapstructure:"trusted_proxies"`
	EnableCors      bool     `mapstructure:"enable_cors"`
	BlockedDomains  []string `mapstructure:"blocked_domains"`
	RequiredHeaders []string `mapstructure:"required_headers"`
}

type RedisConfig struct {
	Addr     string        `mapstructure:"addr"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	TTL      time.Duration `mapstructure:"ttl"`
}

type DatabaseConfig struct {
	URI         string `mapstructure:"uri"`
	Name        string `mapstructure:"name"`
	MaxPoolSize int    `mapstructure:"max_pool_size"`
	MinPoolSize int    `mapstructure:"min_pool_size"`
}

type AppConfig struct {
	DefaultTTL    time.Duration `mapstructure:"default_ttl"`
	MaxTTL        time.Duration `mapstructure:"max_ttl"`
	ShortCodeLen  int           `mapstructure:"short_code_length"`
	CustomCodeLen int           `mapstructure:"custom_code_length"`
	MaxCustomCode int           `mapstructure:"max_custom_code_length"`
	EnableStats   bool          `mapstructure:"enable_stats"`
	EnableGeoIP   bool          `mapstructure:"enable_geoip"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	BaseURL      string        `mapstructure:"base_url"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	GinMode      string        `mapstructure:"gin_mode"`
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    RedisConfig    `mapstructure:"cache"`
	Security SecurityConfig `mapstructure:"security"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}
