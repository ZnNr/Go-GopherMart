package errors

import (
	"errors"
	"fmt"
)

// ErrDuplicateKey представляет ошибку, возникающую при попытке вставить дублирующийся ключ.
type ErrDuplicateKey struct {
	Key string
}

// Error возвращает текстовое представление ошибки.
func (m ErrDuplicateKey) Error() string {
	return fmt.Sprintf("ERROR: duplicate key value violates unique constraint %q (SQLSTATE 23505)", m.Key)
}

var (
	ErrUserAlreadyExists   = errors.New("user already exists")                         // ErrUserAlreadyExists представляет ошибку, возникающую при попытке создать пользователя, который уже существует.
	ErrCreatedBySameUser   = errors.New("order was already created by the same user")  // ErrCreatedBySameUser представляет ошибку, возникающую при попытке создать заказ, который уже создан тем же пользователем.
	ErrCreatedDiffUser     = errors.New("order was already created by the other user") // ErrCreatedDiffUser представляет ошибку, возникающую при попытке создать заказ, который уже создан другим пользователем.
	ErrNoData              = errors.New("no data")                                     // ErrNoData представляет ошибку, возникающую при отсутствии данных.
	ErrInsufficientBalance = errors.New("insufficient balance")                        // ErrInsufficientBalance представляет ошибку, возникающую при недостаточном балансе.
	ErrNoSuchUser          = errors.New("no such user")                                // ErrNoSuchUser представляет ошибку, возникающую при отсутствии пользователя.
	ErrInvalidCredentials  = errors.New("incorrect password")                          // ErrInvalidCredentials представляет ошибку, возникающую при неверных учетных данных.
)
