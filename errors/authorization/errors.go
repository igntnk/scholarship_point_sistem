package authorization

import "errors"

var (
	UnauthorizedErr    = errors.New("Токен не найден")
	TokenExpiredErr    = errors.New("Срок жизни токена закончился")
	TokenDeniedErr     = errors.New("Токен отклонен")
	HasNoPermissionErr = errors.New("Нет прав на данный ресурс")
	HasNoEmailErr      = errors.New("Не предоставлен email")
	HasNoPasswordErr   = errors.New("Не предоставлен пароль")
	WrongPasswordErr   = errors.New("Неправильный пароль или email")
)
