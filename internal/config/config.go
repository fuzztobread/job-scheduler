// internal/config/config.go
package config

import (
	"strings"
	
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	URLs                []string
	ScrapeInterval      string
	NotifierType        string
	DiscordWebhookURL   string
	SlackToken          string
	SlackChannel        string
	EmailSMTP           string
	EmailFrom           string
	EmailTo             string
	LogLevel            string
	LogFormat           string
}

// LoadConfig loads the configuration from environment variables or config file
func LoadConfig() (*Config, error) {
	viper.SetDefault("ScrapeInterval", "*/5 * * * *")
	viper.SetDefault("NotifierType", "discord")
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("LogFormat", "json")
	
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	
	// Read from environment variables
	viper.SetEnvPrefix("CAREERSCRAPER")
	viper.AutomaticEnv()
	
	// Read from config file
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	
	config := &Config{
		ScrapeInterval:    viper.GetString("ScrapeInterval"),
		NotifierType:      viper.GetString("NotifierType"),
		DiscordWebhookURL: viper.GetString("DiscordWebhookURL"),
		SlackToken:        viper.GetString("SlackToken"),
		SlackChannel:      viper.GetString("SlackChannel"),
		EmailSMTP:         viper.GetString("EmailSMTP"),
		EmailFrom:         viper.GetString("EmailFrom"),
		EmailTo:           viper.GetString("EmailTo"),
		LogLevel:          viper.GetString("LogLevel"),
		LogFormat:         viper.GetString("LogFormat"),
	}
	
	// Parse URLs
	urlsStr := viper.GetString("URLs")
	if urlsStr != "" {
		config.URLs = strings.Split(urlsStr, ",")
	}
	
	return config, nil
}