package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Google   GoogleConfig
	AI       AIConfig
	Storage  StorageConfig
}

type ServerConfig struct {
	Port string
	Mode string // "debug" or "release"
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
	RefreshSecret   string
}

type GoogleConfig struct {
	ClientID         string
	ClientSecret     string
	RedirectURL      string
	ServiceAccountKey string
	DriveFolderID    string
}

type AIConfig struct {
	Provider string // "gemini", "deepseek", "ollama"
	APIKey   string
	BaseURL  string
	Model    string
}

type StorageConfig struct {
	MaxFileSize    int64
	AllowedTypes   []string
	CompressImages bool
	CompressPDFs   bool
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "archiving"),
			Password: getEnv("DB_PASSWORD", "archiving_secret"),
			DBName:   getEnv("DB_NAME", "archiving_qa"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "change-me-in-production"),
			ExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
			RefreshSecret:   getEnv("JWT_REFRESH_SECRET", "change-me-refresh-secret"),
		},
		Google: GoogleConfig{
			ClientID:         getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret:     getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:      getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
			ServiceAccountKey: getEnv("GOOGLE_SERVICE_ACCOUNT_KEY", ""),
			DriveFolderID:    getEnv("GOOGLE_DRIVE_FOLDER_ID", ""),
		},
		AI: AIConfig{
			Provider: getEnv("AI_PROVIDER", "ollama"),
			APIKey:   getEnv("AI_API_KEY", ""),
			BaseURL:  getEnv("AI_BASE_URL", "http://localhost:11434"),
			Model:    getEnv("AI_MODEL", "llama3"),
		},
		Storage: StorageConfig{
			MaxFileSize:    getEnvInt64("MAX_FILE_SIZE", 50*1024*1024), // 50MB
			AllowedTypes:   []string{".pdf", ".doc", ".docx", ".jpg", ".jpeg", ".png", ".tiff"},
			CompressImages: getEnvBool("COMPRESS_IMAGES", true),
			CompressPDFs:   getEnvBool("COMPRESS_PDFS", true),
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode
}

func (c *RedisConfig) Addr() string {
	return c.Host + ":" + c.Port
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}
