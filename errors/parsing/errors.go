package parsing

import (
	"errors"
)

var (
	InputDataErr  = errors.New("Возникла ошибка в результате обработки входящих данных")
	OutputDataErr = errors.New("Возникла ошибка при попытке подготовить ответ")
)
