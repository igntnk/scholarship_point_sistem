package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"golang.org/x/crypto/bcrypt"
	"unicode"
	"unicode/utf8"
)

type PasswordManager interface {
	ValidatePassword(password string) error
	HashPassword(password string) (string, string, error)
	CompareHashAndPassword(hash, salt, password string) error
}

type passwordManager struct {
	bcryptCost int
}

func NewPasswordManager(bcryptCost int) PasswordManager {
	return &passwordManager{bcryptCost: bcryptCost}
}

func (m *passwordManager) ValidatePassword(password string) error {
	if password == "" {
		return errors.New("Пароль обязателен для создания пользователя")
	}
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("Пароль должен быть не менее 8 символов")
	}
	if hasNoUppercase(password) {
		return errors.New("Пароль должен иметь заглавные буквы")
	}
	if hasNoDigits(password) {
		return errors.New("Пароль должен иметь цифры")
	}

	return nil
}

func (m *passwordManager) HashPassword(password string) (hash string, salt string, err error) {
	generatedUUID, err := uuid.NewUUID()
	if err != nil {
		return hash, salt, errors.Join(err, unexpected.InternalErr)
	}
	salt = generatedUUID.String()[:m.bcryptCost]
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password+salt), m.bcryptCost)
	if err != nil {
		return hash, salt, errors.Join(err, unexpected.InternalErr)
	}

	hash = string(hashBytes)
	return hash, salt, nil
}

func (m *passwordManager) CompareHashAndPassword(hash, salt, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+salt))
}

func hasNoUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func hasNoDigits(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
