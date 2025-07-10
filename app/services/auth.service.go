package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/rizkyizh/go-fiber-boilerplate/app/models"
	"github.com/rizkyizh/go-fiber-boilerplate/config"
	"github.com/rizkyizh/go-fiber-boilerplate/database"
)

type AuthService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewAuthService() *AuthService {
	return &AuthService{
		db:    database.DB,
		redis: database.Redis,
	}
}

func (s *AuthService) Register(email, password, firstName, lastName string) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Role:      "user",
		Active:    true,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
	var user models.User
	if err := s.db.Where("email = ? AND active = ?", email, true).First(&user).Error; err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := s.GenerateJWT(user)
	if err != nil {
		return nil, "", err
	}

	// Store token in Redis for session management
	ctx := context.Background()
	s.redis.Set(ctx, "session:"+token, user.ID, 24*time.Hour)

	return &user, token, nil
}

func (s *AuthService) Logout(token string) error {
	ctx := context.Background()
	s.redis.Del(ctx, "session:"+token)
	return nil
}

func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) UpdateProfile(userID uint, firstName, lastName, phone, address, city, country, postalCode string) error {
	// Update user basic info
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"first_name": firstName,
		"last_name":  lastName,
	}).Error; err != nil {
		return err
	}

	// Update or create profile
	var profile models.UserProfile
	result := s.db.Where("user_id = ?", userID).First(&profile)
	if result.Error != nil {
		// Create new profile
		profile = models.UserProfile{
			UserID:     userID,
			Phone:      phone,
			Address:    address,
			City:       city,
			Country:    country,
			PostalCode: postalCode,
		}
		return s.db.Create(&profile).Error
	} else {
		// Update existing profile
		return s.db.Model(&profile).Updates(map[string]interface{}{
			"phone":       phone,
			"address":     address,
			"city":        city,
			"country":     country,
			"postal_code": postalCode,
		}).Error
	}
}

func (s *AuthService) GenerateJWT(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT_SECRET))
}
