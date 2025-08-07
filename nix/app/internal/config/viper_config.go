package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// LoadConfig loads configuration using Viper
// Priority order: Environment variables > Config file > Defaults
func LoadConfig(configPath string) (*ServerConfig, error) {
	v := viper.New()

	// Set config file details
	v.SetConfigName("server")
	v.SetConfigType("yaml")

	// Add config paths
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/treacherest")
	}

	// Enable environment variable binding
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables
	// These allow both TREACHEREST_SERVER_PORT and PORT to work
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.host", "HOST")
	v.BindEnv("server.loglevel", "LOG_LEVEL")
	v.BindEnv("server.logformat", "LOG_FORMAT")
	v.BindEnv("server.ratelimit", "RATE_LIMIT")
	v.BindEnv("server.ratelimitburst", "RATE_LIMIT_BURST")
	v.BindEnv("server.maxrequestsize", "MAX_REQUEST_SIZE")
	v.BindEnv("server.maxsseconnections", "MAX_SSE_CONNECTIONS")
	v.BindEnv("server.enablemetrics", "ENABLE_METRICS")
	v.BindEnv("server.metricsport", "METRICS_PORT")

	// Set defaults for safe settings
	v.SetDefault("server.maxplayersperroom", 20)
	v.SetDefault("server.minplayersperroom", 1)
	v.SetDefault("server.defaultgamesize", 5)
	v.SetDefault("server.roomcodelength", 5)
	v.SetDefault("server.roomtimeout", "24h")

	// Timeout defaults
	v.SetDefault("server.readtimeout", "0s")
	v.SetDefault("server.writetimeout", "0s")
	v.SetDefault("server.idletimeout", "0s") // 0 for SSE support
	v.SetDefault("server.shutdowntimeout", "0s")
	
	// Request timeout for middleware (separate from server timeouts)
	v.SetDefault("server.requesttimeout", "60s")     // Default 60s for regular requests
	v.SetDefault("server.ssetimeout", "24h")         // 24 hours for SSE connections (or 0 to disable)

	// Rate limiting defaults
	v.SetDefault("server.ratelimit", 10.0)
	v.SetDefault("server.ratelimitburst", 20)

	// Request limits
	v.SetDefault("server.maxrequestsize", 1048576) // 1MB
	v.SetDefault("server.maxsseconnections", 1000)

	// Monitoring defaults
	v.SetDefault("server.enablemetrics", false)
	v.SetDefault("server.loglevel", "info")
	v.SetDefault("server.logformat", "text")

	// Try to read config file (it's optional)
	if err := v.ReadInConfig(); err != nil {
		// If a specific config file was requested and not found, that's OK
		// We'll continue with env vars and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// For other errors (like permission issues), check if it's just file not found
			if strings.Contains(err.Error(), "no such file or directory") {
				// File doesn't exist, continue with defaults
			} else {
				// Config file was found but another error occurred
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
		// Config file not found; continue with env vars and defaults
	}

	// Create config struct
	cfg := &ServerConfig{}

	// Unmarshal into the struct
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Validate required fields
	if v.GetString("server.port") == "" {
		return nil, fmt.Errorf("PORT environment variable must be set")
	}
	if v.GetString("server.host") == "" {
		return nil, fmt.Errorf("HOST environment variable must be set")
	}

	// Load default roles if not in config file
	if len(cfg.Roles.Available) == 0 {
		cfg.Roles = DefaultConfig().Roles
	}

	// Additional validation
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// InitViper sets up Viper for use throughout the application
// This can be called once at startup to access config values directly
func InitViper() *viper.Viper {
	v := viper.New()

	// Set config details
	v.SetConfigName("server")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Enable environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config if available
	v.ReadInConfig()

	return v
}
