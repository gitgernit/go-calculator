package auth

import (
	"errors"
	"time"

	db "github.com/gitgernit/go-calculator/internal/infra/gorm"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserInteractor struct {
	JWTSecretKey string
}

func (i *UserInteractor) Create(login, password string) error {
	var existing db.User
	if err := db.Db.Where("login = ?", login).First(&existing).Error; err == nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := db.User{
		Login:    login,
		Password: string(hashedPassword),
	}

	if err := db.Db.Create(&user).Error; err != nil {
		return err
	}

	return nil
}

func (i *UserInteractor) Authorize(login, password string) (string, error) {
	var user db.User
	if err := db.Db.Where("login = ?", login).First(&user).Error; err != nil {
		return "", errors.New("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(i.JWTSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (i *UserInteractor) CheckToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(i.JWTSecretKey), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	login, ok := claims["login"].(string)
	if !ok {
		return "", errors.New("invalid login claim")
	}

	return login, nil
}
