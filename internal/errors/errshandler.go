package errors

import "errors"

var (
	ErrTokenIsEmpty = errors.New("token is empty") // ErrTokenIsEmpty представляет ошибку, возникающую при попытке использования пустого токена.
	ErrNoToken      = errors.New("no token")       // ErrNoToken представляет ошибку, возникающую при отсутствии токена.
)
