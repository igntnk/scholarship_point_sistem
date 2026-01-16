package authorization

import "errors"

var (
	UnauthorizedErr    = errors.New("Токен не найден")
	TokenExpiredErr    = errors.New("Срок жизни токена закончился")
	TokenDeniedErr     = errors.New("Токен отклонен")
	HasNoPermissionErr = errors.New("Нет прав на данный ресурс")
)
