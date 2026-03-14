package service

import (
	"errors"
	"time"

	"job-tracker/internal/dto"
	"job-tracker/internal/entity"
	"job-tracker/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken      = errors.New("email already in use")
	ErrInvalidCreds    = errors.New("invalid email or password")
	ErrUserNotFound    = errors.New("user not found")
)

type AuthService interface {
	Register(req *dto.RegisterRequest) (*entity.User, error)
	Login(req *dto.LoginRequest) (string, *entity.User, error)
	GetProfile(userID int) (*entity.User, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{userRepo: userRepo, jwtSecret: []byte(jwtSecret)}
}

func (s *authService) Register(req *dto.RegisterRequest) (*entity.User, error) {
	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hash),
		Phone:    req.Phone,
		UUID:     uuid.New(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(req *dto.LoginRequest) (string, *entity.User, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return "", nil, ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", nil, ErrInvalidCreds
	}

	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *authService) GetProfile(userID int) (*entity.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *authService) generateToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
