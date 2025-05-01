package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/atnomoverflow/auth-server/internal/entity"
	"github.com/uptrace/bun"
)

type userRepository struct {
	DB *bun.DB
}

var (
	ErrorNotFound     = errors.New("user not Found")
	ErrorTokenBlocked = errors.New("token is blocked")
)

func NewUserRepository(db *bun.DB) *userRepository {
	return &userRepository{
		DB: db,
	}
}

func (rep *userRepository) Create(ctx context.Context, user *entity.User) error {
	_, err := rep.DB.NewInsert().Model(user).Exec(ctx)
	return err
}

func (rep *userRepository) FindByID(ctx context.Context, id uint64) (*entity.User, error) {
	user := &entity.User{}
	_, err := rep.DB.NewSelect().Model(user).Where("id = ?", id).Exec(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorNotFound
		}
		return nil, err
	}
	return user, nil
}

func (rep *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}
	_, err := rep.DB.NewSelect().Model(user).Where("email = ?", email).Exec(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorNotFound
		}
		return nil, err
	}
	return user, nil
}

// TODO: The fields should be valid since we will have a dto in the controller string is fine
// to get the fields we can unmarchel the update request into map[string]interface{} then
// loop over the fields
func (rep *userRepository) Update(ctx context.Context, user *entity.User, fields []string) error {
	user.UpdatedAt = time.Now()
	query := rep.DB.NewUpdate().Model(user).Where("id = ?", user.ID)
	for _, field := range fields {
		query.Column(field)
	}
	query.Column("updated_at")
	_, err := query.Exec(ctx)
	return err
}

func (rep *userRepository) Delete(ctx context.Context, id uint64) error {
	_, err := rep.DB.NewDelete().Model(&entity.User{}).Where("id = ?", id).Exec(ctx)
	return err
}

func (rep *userRepository) GetUser2Fauth(ctx context.Context, userID uint64) ([]entity.TwoFAMethod, error) {

	var twoFauthMethod []entity.TwoFAMethod
	err := rep.DB.NewSelect().Model(&twoFauthMethod).Where("user_id = ?", userID).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorNotFound
		}
		return nil, err
	}

	return twoFauthMethod, nil
}

func (rep *userRepository) IsRefreshTokenBlackListed(ctx context.Context, refreshToken string) error {

	var token []entity.BlackListToken
	ok, err := rep.DB.NewSelect().Model(&token).Where("token = ?", refreshToken).Exists(ctx)

	if err != nil {
		// thechnicly it should not happen
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if ok {
		return ErrorTokenBlocked
	}
	return nil
}

func (rep *userRepository) RevokeToken(ctx context.Context, refrechToken string) error {
	token := &entity.BlackListToken{
		Token: refrechToken,
	}
	_, err := rep.DB.NewInsert().Model(token).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
