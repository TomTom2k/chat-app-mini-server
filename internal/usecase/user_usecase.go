package usecase

import (
	"errors"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
	"github.com/TomTom2k/chat-app/server/pkg/jwt"
	"github.com/TomTom2k/chat-app/server/pkg/utils"
)

type UserUseCase struct {
	Repo      domain.UserRepository
	JWTSecret string
}

type RegisterResult struct {
	Token string      `json:"token"`
	User  entity.User `json:"user"`
}

func (u *UserUseCase) Register(user entity.User) (*RegisterResult, error) {
	// Check if user with email already exists
	existingUser, _ := u.Repo.GetByEmail(user.Email)
	if existingUser.Email != "" {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	// Set timestamps
	now := time.Now()
	newUser := entity.User{
		Email:     user.Email,
		Name:      user.Name,
		FullName:  user.Name, // Store in both fields for compatibility
		Password:  hashed,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Create user
	err = u.Repo.CreateUser(newUser)
	if err != nil {
		return nil, err
	}

	// Get created user to get ID
	createdUser, err := u.Repo.GetByEmail(user.Email)
	if err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := jwt.GenerateToken(createdUser.ID, createdUser.Email, u.JWTSecret)
	if err != nil {
		return nil, err
	}

	// Remove password from response and map fields
	createdUser.Password = ""
	if createdUser.FullName != "" && createdUser.Name == "" {
		createdUser.Name = createdUser.FullName
	}

	return &RegisterResult{
		Token: token,
		User:  createdUser,
	}, nil
}

func (u *UserUseCase) Login(email string, password string) (*RegisterResult, error) {
	user, err := u.Repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	if user.Email == "" {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := jwt.GenerateToken(user.ID, user.Email, u.JWTSecret)
	if err != nil {
		return nil, err
	}

	// Remove password from response and map fields
	user.Password = ""
	if user.FullName != "" && user.Name == "" {
		user.Name = user.FullName
	}

	return &RegisterResult{
		Token: token,
		User:  user,
	}, nil
}

func (u *UserUseCase) GetMe(userID string) (entity.User, error) {
	user, err := u.Repo.GetByID(userID)
	if err != nil {
		return entity.User{}, err
	}

	// Remove password from response and map fields
	user.Password = ""
	if user.FullName != "" && user.Name == "" {
		user.Name = user.FullName
	}
	return user, nil
}
