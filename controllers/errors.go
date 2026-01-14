package controllers

import "errors"

var (
	ErrParsingBody = errors.New("Ошибка при попытке обработать тело запроса")
)
