package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                  string
	DBDSN                 string
	RedisAddr             string
	RedisPassword         string
	RateLimit             int
	RateWindow            int    // seconds
	SMTPServerIP          string // 仅从 SMTP_SERVER_IP 环境变量读取
	SMTPHostname          string // 邮件服务器场指向的 hostname，不硬编码
	OutboundEmailEnabled  bool
	OutboundMaxRecipients int
	ResendAPIKey          string
	ResendAPIURL          string
	ResendFromAddress     string
	ResendFromName        string
}

func Load() *Config {
	rl, _ := strconv.Atoi(getEnv("RATE_LIMIT", "500"))
	rw, _ := strconv.Atoi(getEnv("RATE_WINDOW", "60"))
	outboundEnabled, _ := strconv.ParseBool(getEnv("OUTBOUND_EMAIL_ENABLED", "false"))
	maxRecipients, _ := strconv.Atoi(getEnv("OUTBOUND_MAX_RECIPIENTS", "50"))

	return &Config{
		// ★ PORT：API 容器内监听端口，默认 8967。
		// 由 .env 中的 API_PORT 注入。修改此端口后需同步：
		//   1. .env / .env.example 的 API_PORT
		//   2. docker-compose.yml api.ports 右边数字
		//   3. nginx/default.conf 所有 proxy_pass http://api:8967
		//   4. postfix/entrypoint.sh curl http://api:8967
		//   5. postfix/mail-receiver.py API_URL 默认值
		Port:  getEnv("PORT", "8967"),
		DBDSN: getEnv("DB_DSN", ""),
		// ★ RedisAddr：Redis 容器内部地址，格式 "host:port"。
		// 默认 "redis:6379"，"redis" 是 Docker 内部服务名，不需要修改。
		RedisAddr:             getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword:         getEnv("REDIS_PASSWORD", ""),
		RateLimit:             rl,
		RateWindow:            rw,
		SMTPServerIP:          os.Getenv("SMTP_SERVER_IP"),
		SMTPHostname:          os.Getenv("SMTP_HOSTNAME"),
		OutboundEmailEnabled:  outboundEnabled,
		OutboundMaxRecipients: maxRecipients,
		ResendAPIKey:          os.Getenv("RESEND_API_KEY"),
		ResendAPIURL:          os.Getenv("RESEND_API_URL"),
		ResendFromAddress:     os.Getenv("RESEND_FROM_ADDRESS"),
		ResendFromName:        getEnv("RESEND_FROM_NAME", "TempMail"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
