package service

import "errors"

var (
	ParsingDataErr         = errors.New("Возникла ошибка в результате обработки входящих данных")
	RecordAlreadyExistsErr = errors.New("Запись с такими параметрами уже существует и не может дублироваться")
)
