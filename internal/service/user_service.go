package service

import (
	"github.com/atnomoverflow/auth-server/internal/repository"
)

type UserService struct {
	userRepository repository.Users
}

func NewUserService(userRepository repository.Users) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}
