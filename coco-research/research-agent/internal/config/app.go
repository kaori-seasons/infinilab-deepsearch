package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	ES       ESConfig       `mapstructure:"elasticsearch"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// ESConfig Elasticsearch配置
type ESConfig struct {
	Hosts    []string `mapstructure:"hosts"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
}

// LLMConfig LLM配置
type LLMConfig struct {
	Provider  string         `mapstructure:"provider"`  // 默认LLM提供商：openai, claude, deepseek
	APIKey    string         `mapstructure:"api_key"`   // 默认API密钥
	OpenAI    OpenAIConfig   `mapstructure:"openai"`
	Claude    ClaudeConfig   `mapstructure:"claude"`
	DeepSeek  DeepSeekConfig `mapstructure:"deepseek"`
	Embedding EmbeddingConfig `mapstructure:"embedding"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey     string `mapstructure:"api_key"`
	BaseURL    string `mapstructure:"base_url"`
	Model      string `mapstructure:"model"`
	MaxTokens  int    `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// ClaudeConfig Claude配置
type ClaudeConfig struct {
	APIKey     string `mapstructure:"api_key"`
	BaseURL    string `mapstructure:"base_url"`
	Model      string `mapstructure:"model"`
	MaxTokens  int    `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// DeepSeekConfig DeepSeek配置
type DeepSeekConfig struct {
	APIKey     string `mapstructure:"api_key"`
	BaseURL    string `mapstructure:"base_url"`
	Model      string `mapstructure:"model"`
	MaxTokens  int    `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// EmbeddingConfig 嵌入模型配置
type EmbeddingConfig struct {
	Model     string `mapstructure:"model"`      // 嵌入模型名称
	Provider  string `mapstructure:"provider"`   // 提供商：huggingface, openai, claude等
	APIKey    string `mapstructure:"api_key"`    // API密钥
	BaseURL   string `mapstructure:"base_url"`   // 基础URL
	MaxLength int    `mapstructure:"max_length"` // 最大文本长度
	Dimension int    `mapstructure:"dimension"`  // 向量维度
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("app")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	// 数据库默认配置
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "coco")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "coco_research")
	viper.SetDefault("database.sslmode", "disable")

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Elasticsearch默认配置
	viper.SetDefault("elasticsearch.hosts", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")

	// LLM默认配置
	viper.SetDefault("llm.provider", "openai")
	viper.SetDefault("llm.openai.model", "gpt-4")
	viper.SetDefault("llm.openai.max_tokens", 4096)
	viper.SetDefault("llm.openai.temperature", 0.7)
	viper.SetDefault("llm.claude.model", "claude-3-sonnet-20240229")
	viper.SetDefault("llm.claude.max_tokens", 4096)
	viper.SetDefault("llm.claude.temperature", 0.7)
	viper.SetDefault("llm.deepseek.model", "deepseek-chat")
	viper.SetDefault("llm.deepseek.max_tokens", 4096)
	viper.SetDefault("llm.deepseek.temperature", 0.7)
	
	// 嵌入模型默认配置
	viper.SetDefault("llm.embedding.model", "iic/nlp_corom_sentence-embedding_chinese-base")
	viper.SetDefault("llm.embedding.provider", "huggingface")
	viper.SetDefault("llm.embedding.max_length", 512)
	viper.SetDefault("llm.embedding.dimension", 768)

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
} 