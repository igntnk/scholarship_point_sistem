package unexpected

import "errors"

var (
	RequestErr = errors.New("Возникла неожиданная ошибка при запросе в базу данных")
)
