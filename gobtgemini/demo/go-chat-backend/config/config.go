package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 全局配置结构
type Config struct {
	Port             string
	GinMode          string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	JWTSecret        string
	JWTExpires       int
	ChromaHost       string
	ChromaPort       string
	ChromaCollection string
	LLMAPIURL        string
	LLMAPIKey        string
	LLMModel         string
	LogLevel         string
	LogFile          string
}

var cfg *Config

// LoadConfig 加载配置
func LoadConfig() {
	// 加载.env文件
	err := godotenv.Load()
	if err != nil {
		log.Println("警告: 未找到.env文件，将使用系统环境变量")
	}

	cfg = &Config{
		Port:             GetString("PORT", "8080"),
		GinMode:          GetString("GIN_MODE", "debug"),
		DBHost:           GetString("DB_HOST", "localhost"),
		DBPort:           GetString("DB_PORT", "5432"),
		DBUser:           GetString("DB_USER", "postgres"),
		DBPassword:       GetString("DB_PASSWORD", "password123"),
		DBName:           GetString("DB_NAME", "go_chat_db"),
		DBSSLMode:        GetString("DB_SSLMODE", "disable"),
		JWTSecret:        GetString("JWT_SECRET", "your_super_secret_jwt_key_change_this_in_production"),
		JWTExpires:       GetInt("JWT_EXPIRES_HOURS", 24),
		ChromaHost:       GetString("CHROMA_HOST", "chroma"),
		ChromaPort:       GetString("CHROMA_PORT", "8000"),
		ChromaCollection: GetString("CHROMA_COLLECTION_NAME", "chat_memory"),
		LLMAPIURL:        GetString("LLM_API_URL", "https://pictureai.pp.ua/v1beta/models/gemini-2.0-flash:generateContent"),
		LLMAPIKey:        GetString("LLM_API_KEY", "AIzaSyB0BToz8pj_TS-ZkqsccG4gnfy7tJm7oUE"),
		LLMModel:         GetString("LLM_MODEL", "gemini-2.0-flash"),
		LogLevel:         GetString("LOG_LEVEL", "info"),
		LogFile:          GetString("LOG_FILE", "logs/app.log"),
	}
}

// GetString 获取字符串配置
func GetString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetInt 获取整数配置
func GetInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetBool 获取布尔配置
func GetBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// Get 获取配置实例
func Get() *Config {
	return cfg
}
