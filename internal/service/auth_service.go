package service

import (
	"errors"
	"officeworker/internal/pkg/jwt"
	"officeworker/internal/pkg/redis"
	"officeworker/internal/repository"
	"officeworker/models"
	"time"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtMgr    *jwt.Manager
	blacklist *redis.Blacklist
	jwtExpiry time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	jwtMgr *jwt.Manager,
	blacklist *redis.Blacklist,
	jwtExpiry time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtMgr:    jwtMgr,
		blacklist: blacklist,
		jwtExpiry: jwtExpiry,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	exists, err = s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     "user",
		Status:   "active",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.jwtMgr.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	if user.Status != "active" {
		return nil, errors.New("account is inactive")
	}

	if !s.userRepo.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid username or password")
	}

	accessToken, refreshToken, err := s.jwtMgr.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (s *AuthService) Refresh(refreshToken string) (*AuthResponse, error) {
	claims, err := s.jwtMgr.ParseToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	accessToken, newRefreshToken, err := s.jwtMgr.RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (s *AuthService) Logout(accessToken string) error {
	return s.blacklist.Add(accessToken, s.jwtExpiry)
}

func (s *AuthService) GetUserInfo(userID uint) (*UserInfo, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}
