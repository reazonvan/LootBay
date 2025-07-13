package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config основная конфигурация приложения
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	MongoDB    MongoDBConfig    `mapstructure:"mongodb"`
	RabbitMQ   RabbitMQConfig   `mapstructure:"rabbitmq"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Payment    PaymentConfig    `mapstructure:"payment"`
	Telegram   TelegramConfig   `mapstructure:"telegram"`
	Email      EmailConfig      `mapstructure:"email"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	MaxBodySize  int64         `mapstructure:"max_body_size"`
}

// DatabaseConfig конфигурация PostgreSQL
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"`
}

// RedisConfig конфигурация Redis
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// MongoDBConfig конфигурация MongoDB
type MongoDBConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
}

// RabbitMQConfig конфигурация RabbitMQ
type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	VHost    string `mapstructure:"vhost"`
}

// JWTConfig конфигурация JWT
type JWTConfig struct {
	SecretKey        string        `mapstructure:"secret_key"`
	AccessTokenTTL   time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `mapstructure:"refresh_token_ttl"`
	PasswordResetTTL time.Duration `mapstructure:"password_reset_ttl"`
}

// PaymentConfig конфигурация платежных систем
type PaymentConfig struct {
	Stripe StripeConfig `mapstructure:"stripe"`
	PayPal PayPalConfig `mapstructure:"paypal"`
}

// StripeConfig конфигурация Stripe
type StripeConfig struct {
	PublicKey  string `mapstructure:"public_key"`
	SecretKey  string `mapstructure:"secret_key"`
	WebhookKey string `mapstructure:"webhook_key"`
}

// PayPalConfig конфигурация PayPal
type PayPalConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Mode         string `mapstructure:"mode"` // sandbox, live
}

// TelegramConfig конфигурация Telegram
type TelegramConfig struct {
	BotToken   string `mapstructure:"bot_token"`
	WebhookURL string `mapstructure:"webhook_url"`
	Debug      bool   `mapstructure:"debug"`
}

// EmailConfig конфигурация email
type EmailConfig struct {
	Provider  string `mapstructure:"provider"` // sendgrid, smtp
	APIKey    string `mapstructure:"api_key"`
	FromEmail string `mapstructure:"from_email"`
	FromName  string `mapstructure:"from_name"`
	SMTPHost  string `mapstructure:"smtp_host"`
	SMTPPort  int    `mapstructure:"smtp_port"`
	SMTPUser  string `mapstructure:"smtp_user"`
	SMTPPass  string `mapstructure:"smtp_pass"`
}

// StorageConfig конфигурация хранилища
type StorageConfig struct {
	Provider        string `mapstructure:"provider"` // s3, minio, local
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	BucketName      string `mapstructure:"bucket_name"`
	Region          string `mapstructure:"region"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	LocalPath       string `mapstructure:"local_path"`
}

// MonitoringConfig конфигурация мониторинга
type MonitoringConfig struct {
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Jaeger     JaegerConfig     `mapstructure:"jaeger"`
}

// PrometheusConfig конфигурация Prometheus
type PrometheusConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// JaegerConfig конфигурация Jaeger
type JaegerConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service_name"`
}

// bindEnvSafe безопасная обертка для BindEnv
func bindEnvSafe(key string, envVar string) {
	_ = viper.BindEnv(key, envVar)
}

// LoadConfig загружает конфигурацию из файла и переменных окружения
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Настройка переменных окружения с префиксами
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Маппинг переменных окружения на поля конфигурации
	bindEnvSafe("server.host", "SERVER_HOST")
	bindEnvSafe("server.port", "SERVER_PORT")
	bindEnvSafe("server.mode", "SERVER_MODE")
	bindEnvSafe("server.read_timeout", "SERVER_READ_TIMEOUT")
	bindEnvSafe("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	bindEnvSafe("server.max_body_size", "SERVER_MAX_BODY_SIZE")

	bindEnvSafe("database.host", "DATABASE_HOST")
	bindEnvSafe("database.port", "DATABASE_PORT")
	bindEnvSafe("database.user", "DATABASE_USER")
	bindEnvSafe("database.password", "DATABASE_PASSWORD")
	bindEnvSafe("database.db_name", "DATABASE_NAME")
	bindEnvSafe("database.ssl_mode", "DATABASE_SSL_MODE")
	bindEnvSafe("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	bindEnvSafe("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	bindEnvSafe("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")
	bindEnvSafe("database.auto_migrate", "DATABASE_AUTO_MIGRATE")

	bindEnvSafe("redis.host", "REDIS_HOST")
	bindEnvSafe("redis.port", "REDIS_PORT")
	bindEnvSafe("redis.password", "REDIS_PASSWORD")
	bindEnvSafe("redis.db", "REDIS_DB")
	bindEnvSafe("redis.pool_size", "REDIS_POOL_SIZE")
	bindEnvSafe("redis.min_idle_conns", "REDIS_MIN_IDLE_CONNS")

	bindEnvSafe("mongodb.host", "MONGODB_HOST")
	bindEnvSafe("mongodb.port", "MONGODB_PORT")
	bindEnvSafe("mongodb.user", "MONGODB_USER")
	bindEnvSafe("mongodb.password", "MONGODB_PASSWORD")
	bindEnvSafe("mongodb.db_name", "MONGODB_DB_NAME")

	bindEnvSafe("rabbitmq.host", "RABBITMQ_HOST")
	bindEnvSafe("rabbitmq.port", "RABBITMQ_PORT")
	bindEnvSafe("rabbitmq.user", "RABBITMQ_USER")
	bindEnvSafe("rabbitmq.password", "RABBITMQ_PASSWORD")
	bindEnvSafe("rabbitmq.vhost", "RABBITMQ_VHOST")

	bindEnvSafe("jwt.secret_key", "JWT_SECRET_KEY")
	bindEnvSafe("jwt.access_token_ttl", "JWT_ACCESS_TOKEN_TTL")
	bindEnvSafe("jwt.refresh_token_ttl", "JWT_REFRESH_TOKEN_TTL")
	bindEnvSafe("jwt.password_reset_ttl", "JWT_PASSWORD_RESET_TTL")

	bindEnvSafe("payment.stripe.public_key", "STRIPE_PUBLIC_KEY")
	bindEnvSafe("payment.stripe.secret_key", "STRIPE_SECRET_KEY")
	bindEnvSafe("payment.stripe.webhook_key", "STRIPE_WEBHOOK_KEY")
	bindEnvSafe("payment.paypal.client_id", "PAYPAL_CLIENT_ID")
	bindEnvSafe("payment.paypal.client_secret", "PAYPAL_CLIENT_SECRET")
	bindEnvSafe("payment.paypal.mode", "PAYPAL_MODE")

	bindEnvSafe("telegram.bot_token", "TELEGRAM_BOT_TOKEN")
	bindEnvSafe("telegram.webhook_url", "TELEGRAM_WEBHOOK_URL")
	bindEnvSafe("telegram.debug", "TELEGRAM_DEBUG")

	bindEnvSafe("email.provider", "EMAIL_PROVIDER")
	bindEnvSafe("email.api_key", "SENDGRID_API_KEY")
	bindEnvSafe("email.from_email", "EMAIL_FROM_EMAIL")
	bindEnvSafe("email.from_name", "EMAIL_FROM_NAME")
	bindEnvSafe("email.smtp_host", "SMTP_HOST")
	bindEnvSafe("email.smtp_port", "SMTP_PORT")
	bindEnvSafe("email.smtp_user", "SMTP_USER")
	bindEnvSafe("email.smtp_pass", "SMTP_PASS")

	bindEnvSafe("storage.provider", "STORAGE_PROVIDER")
	bindEnvSafe("storage.endpoint", "S3_ENDPOINT")
	bindEnvSafe("storage.access_key_id", "S3_ACCESS_KEY_ID")
	bindEnvSafe("storage.secret_access_key", "S3_SECRET_ACCESS_KEY")
	bindEnvSafe("storage.bucket_name", "S3_BUCKET_NAME")
	bindEnvSafe("storage.region", "S3_REGION")
	bindEnvSafe("storage.use_ssl", "S3_USE_SSL")
	bindEnvSafe("storage.local_path", "STORAGE_LOCAL_PATH")

	bindEnvSafe("monitoring.prometheus.enabled", "PROMETHEUS_ENABLED")
	bindEnvSafe("monitoring.prometheus.port", "PROMETHEUS_PORT")
	bindEnvSafe("monitoring.prometheus.path", "PROMETHEUS_PATH")
	bindEnvSafe("monitoring.jaeger.enabled", "JAEGER_ENABLED")
	bindEnvSafe("monitoring.jaeger.endpoint", "JAEGER_ENDPOINT")
	bindEnvSafe("monitoring.jaeger.service_name", "JAEGER_SERVICE_NAME")

	// Читаем конфигурацию из файла (если существует)
	if err := viper.ReadInConfig(); err != nil {
		// Если файл не найден, используем только переменные окружения
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfigForService загружает конфигурацию для конкретного сервиса
func LoadConfigForService(path string, serviceName string) (*Config, error) {
	config, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}

	// Переопределяем порт для конкретного сервиса
	switch serviceName {
	case "user-service":
		config.Server.Port = "8081"
	case "product-service":
		config.Server.Port = "8082"
	case "order-service":
		config.Server.Port = "8083"
	case "payment-service":
		config.Server.Port = "8084"
	case "chat-service":
		config.Server.Port = "8085"
	case "notification-service":
		config.Server.Port = "8086"
	}

	return config, nil
}
