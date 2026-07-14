package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	JWTSecret  string

	NotionClientID     string
	NotionClientSecret string
	NotionRedirectURI  string

	OpenRouterAPIKey string
	OpenRouterModel  string

	DiscordToken         string
	DiscordPrefix        string
	DiscordDefaultUserID int
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", ""),
		DBPort:     getEnv("DB_PORT", ""),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", ""),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", ""),

		NotionClientID:     getEnv("NOTION_CLIENT_ID", ""),
		NotionClientSecret: getEnv("NOTION_CLIENT_SECRET", ""),
		NotionRedirectURI:  getEnv("NOTION_REDIRECT_URI", ""),

		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:  getEnv("OPENROUTER_MODEL", "openai/gpt-oss-120b:free"),

		DiscordToken:         getEnv("DISCORD_TOKEN", ""),
		DiscordPrefix:        getEnv("DISCORD_PREFIX", "!"),
		DiscordDefaultUserID: getEnvInt("DISCORD_DEFAULT_USER_ID", 0),
	}
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" port=" + c.DBPort +
		" sslmode=disable TimeZone=UTC"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if n, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return n
	}
	return fallback
}
