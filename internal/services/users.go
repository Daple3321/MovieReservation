package services

import (
	"context"
	"errors"
	"os"

	"github.com/Daple3321/MovieReservation/internal/entity"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrDuplicateUsername = errors.New("username already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserNotAuthenticated = errors.New("user not authenticated")
var ErrUserNotAuthorized = errors.New("user not authorized")

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Register(ctx context.Context, user entity.UserDTO) (uint, error) {
	if user.Username == "" || user.Password == "" {
		return 0, ErrInvalidCredentials
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	userId, err := s.Create(ctx, user.Username, string(passwordHash))
	if err != nil {
		if errors.Is(err, ErrDuplicateUsername) {
			return 0, ErrDuplicateUsername
		}
		return 0, err
	}

	return userId, nil
}

func (s *UserService) Login(ctx context.Context, userDTO entity.UserDTO) (userId uint, username string, err error) {
	if userDTO.Username == "" || userDTO.Password == "" {
		return 0, "", ErrInvalidCredentials
	}

	user, err := s.GetByUsername(ctx, userDTO.Username)
	if err != nil {
		return 0, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userDTO.Password))
	if err != nil {
		return 0, "", ErrInvalidCredentials
	}

	return user.ID, user.Username, nil
}

func (s *UserService) Create(ctx context.Context, username string, passwordHash string) (uint, error) {
	tx := s.db.Begin()
	defer tx.Rollback()

	if tx.Error != nil {
		return 0, tx.Error
	}

	newUser := entity.User{
		Username:     username,
		PasswordHash: passwordHash,
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if username == adminUsername {
		newUser.Role = "admin"
	} else {
		newUser.Role = "user"
	}

	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return newUser.ID, nil
}

func (s *UserService) GetByID(ctx context.Context, UserID uint) (*entity.User, error) {
	var user entity.User

	result := s.db.
		WithContext(ctx).
		Preload("Watchlist").
		Preload("Tickets").
		First(&user, UserID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	if username == "" {
		return nil, ErrInvalidCredentials
	}

	var user entity.User

	result := s.db.
		WithContext(ctx).
		Preload("Watchlist").
		Preload("Tickets").
		Where("username = ?", username).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}
