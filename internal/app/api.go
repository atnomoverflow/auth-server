package app

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/atnomoverflow/auth-server/config"
	"github.com/atnomoverflow/auth-server/internal/handler"
	"github.com/atnomoverflow/auth-server/internal/repository"
	"github.com/atnomoverflow/auth-server/internal/service"
	"github.com/atnomoverflow/auth-server/pkg/cache"
	"github.com/atnomoverflow/auth-server/pkg/database"
	"github.com/atnomoverflow/auth-server/pkg/hash"
	"github.com/atnomoverflow/auth-server/pkg/logger"
	"github.com/atnomoverflow/auth-server/pkg/otp"
	"github.com/atnomoverflow/auth-server/pkg/server"
	"github.com/atnomoverflow/auth-server/pkg/token"
	"github.com/rs/cors"
)

func Run(cfg *config.Config) {

	l := logger.New(cfg.Log.Level)
	l.Info(cfg.Database.ConnectionString)
	db := database.New(cfg.Database.ConnectionString)
	defer db.Close()
	mux := http.NewServeMux()

	// init dependencies
	passwordHasher, err := hash.New(hash.Config{
		Algo: cfg.Hashers.Password.Algo,
		Argon2Config: hash.Argon2Config{
			Memory:      cfg.Hashers.Password.Memory,
			Iterations:  cfg.Hashers.Password.Iterations,
			Parallelism: cfg.Hashers.Password.Parallelism,
			SaltLength:  cfg.Hashers.Password.SaltLength,
			KeyLength:   cfg.Hashers.Password.KeyLength,
		},
	})
	if err != nil {
		l.Error("app - Run - Failed to init password hasher service due to: %w", err)
		return
	}
	totpManager := otp.NewTotp(cfg.App.Issuer)
	tokenManager, err := token.NewTokenManager(cfg.Auth.JwtSecret, token.WithIssuer(cfg.App.Issuer), token.WithAudience(cfg.App.Audience))
	if err != nil {
		l.Error("app - Run - Failed to init token manger service due to: %w", err)
		return
	}
	cacheManager := cache.NewMemoryCache()
	// inti repositories
	repositories := repository.NewRepositories(db.DB)
	// inti services
	services := service.NewServices(service.Dependencies{
		Repos:                  repositories,
		PasswordHasher:         passwordHasher,
		TotpManager:            totpManager,
		VerificationCodeLength: cfg.Auth.VerificationCodeLength,
		JWTTokenManger:         tokenManager,
		Logger:                 l,
		Cache:                  cacheManager,
	})
	handlers := handler.New(services, l)
	handlers.Init(*cfg, mux)

	handler := cors.New(cors.Options{
		AllowedOrigins: cfg.Http.Origin,
	}).Handler(mux)

	server := server.New(handler, server.Port(cfg.Http.Port))

	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interupt:
		l.Info("app - Run - signal %s", s.String())
	case err := <-server.Notify():
		l.Error(fmt.Errorf("app - Run - server shutting down %w", err))
	}

	err = server.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - server shutting down %w", err))
	}
}
