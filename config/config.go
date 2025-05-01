package config

import (
	"fmt"

	"github.com/atnomoverflow/auth-server/pkg/logger"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Http     HTTP     `yaml:"http"`
		Log      Log      `yaml:"logger"`
		Database Database `yaml:"db"`
		Hashers  Hashers  `yaml:"hasher"`
		Auth     Auth     `yaml:"auth"`
		App      App      `yaml:"app"`
		Mail     Mail     `yaml:"mail"`
	}

	App struct {
		Name     string `yaml:"name" env:"APP_NAME"`
		Issuer   string `yaml:"issuer" env:"APP_ISSUER"`
		Audience string `yaml:"audience" env:"APP_AUDIENCE"`
	}
	Log struct {
		Level logger.LogLevel `env-required:"true" yaml:"level" env:"LOG_LEVEL"`
	}

	HTTP struct {
		Port   string   `env-required:"true" yaml:"port" env:"HTTP_PORT"`
		Origin []string `yaml:"origin" env:"HTTP_ORIGIN"`
	}

	Database struct {
		ConnectionString string `yaml:"connectionString" env:"CONNECTION_STRING"`
	}
	Hashers struct {
		Password Hasher `yaml:"password"`
	}
	Hasher struct {
		Algo        string `env-required:"true" yaml:"algo" env:"PASSWORD_HASH_ALGO"`
		Memory      uint32 `yaml:"memory" env:"PASSWORD_HASH_ALGO"`
		Iterations  uint32 `yaml:"iterations" env:"PASSWORD_HASH_ITERATION"`
		Parallelism uint8  `yaml:"parallelism" env:"PASSWORD_HASH_PARALLELISIM"`
		SaltLength  uint32 `yaml:"saltLength" env:"PASSWORD_HASH_SALTL_ENGTH"`
		KeyLength   uint32 `yaml:"keyLength" env:"PASSWORD_HASH_KEY_LENGTH"`
	}
	Auth struct {
		VerificationCodeLength uint16 `yaml:"verificationCodeLength" env:"VERIFICATION_CODE_LENGTH"`
		JwtSecret              string `yaml:"jwtSecret" env:"JWT_SECRET"`
	}
	Mail struct {
		Smtp Smtp `yaml:"smtp"`
	}
	Smtp struct {
		From     string `yaml:"from" env:"MAIL_SMTP_FROM"`
		Host     string `yaml:"host" env:"MAIL_SMTP_HOST"`
		Password string `yaml:"password" env:"MAIL_SMTP_PASSWORD"`
		Port     uint16 `yaml:"port" env:"MAIL_SMTP_PORT"`
	}
)

func New(configPath string) (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading config from file %s: %w", configPath, err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading config from env file: %w", err)
	}

	return cfg, nil
}
