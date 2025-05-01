package repository

import (
	"context"

	"github.com/atnomoverflow/auth-server/internal/entity"
	"github.com/uptrace/bun"
)

// depends on the database used
type Users interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uint64) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User, fields []string) error
	GetUser2Fauth(ctx context.Context, userID uint64) ([]entity.TwoFAMethod, error)
	RevokeToken(ctx context.Context, refrechToken string) error
	// Verify(ctx context.Context) error
	IsRefreshTokenBlackListed(ctx context.Context, refreshToken string) error
	Delete(ctx context.Context, id uint64) error
}

type Repositories struct {
	User Users
}

func NewRepositories(db *bun.DB) *Repositories {
	return &Repositories{
		User: NewUserRepository(db),
	}
}
