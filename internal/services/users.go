package services

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/Daple3321/MovieReservation/internal/entity"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrDuplicateUsername = errors.New("username already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserNotAuthenticated = errors.New("user not authenticated")
var ErrUserNotAuthorized = errors.New("user not authorized")
var ErrInternalServer = errors.New("internal server error")

// postgresUniqueViolation is SQLSTATE 23505 (unique_violation).
func postgresUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

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
		slog.Error("register: bcrypt hash", "err", err)
		return 0, ErrInternalServer
	}

	userId, err := s.Create(ctx, user.Username, string(passwordHash))
	if err != nil {
		if errors.Is(err, ErrDuplicateUsername) {
			return 0, ErrDuplicateUsername
		}
		if errors.Is(err, ErrInternalServer) {
			return 0, ErrInternalServer
		}
		slog.Error("register: unexpected error from create", "err", err)
		return 0, ErrInternalServer
	}

	return userId, nil
}

func (s *UserService) Login(ctx context.Context, userDTO entity.UserDTO) (userId uint, username string, isAdmin bool, err error) {
	if userDTO.Username == "" || userDTO.Password == "" {
		return 0, "", false, ErrInvalidCredentials
	}

	user, err := s.GetByUsername(ctx, userDTO.Username)
	if err != nil {
		return 0, "", false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userDTO.Password))
	if err != nil {
		return 0, "", false, ErrInvalidCredentials
	}

	isUserAdmin := false
	switch user.Role {
	case "admin":
		isUserAdmin = true
	default:
		isUserAdmin = false
	}

	return user.ID, user.Username, isUserAdmin, nil
}

func (s *UserService) Create(ctx context.Context, username string, passwordHash string) (uint, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("user create: begin tx", "err", tx.Error)
		return 0, ErrInternalServer
	}
	defer tx.Rollback()

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
		if postgresUniqueViolation(err) {
			return 0, ErrDuplicateUsername
		}
		slog.Error("user create: insert", "err", err)
		return 0, ErrInternalServer
	}

	if err := tx.Commit().Error; err != nil {
		if postgresUniqueViolation(err) {
			return 0, ErrDuplicateUsername
		}
		slog.Error("user create: commit", "err", err)
		return 0, ErrInternalServer
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
