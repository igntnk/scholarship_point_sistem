package validation

import (
	"errors"
)

var (
	RecordAlreadyExistsErr = errors.New("Запись с такими параметрами уже существует и не может дублироваться")
	NoDataFoundErr         = errors.New("Запись с такими данными не найдена")
	WrongInputErr          = errors.New("Данные не прошли валидация")
)
