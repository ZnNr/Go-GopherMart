package models

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// User представляет собой пользователя с логином и паролем.
type User struct {
	Login    string `json:"login"`    // Login это имя пользователя.
	Password string `json:"password"` // Password это пароль пользователя.
}

// OrderStatus представляет состояние заказа.
type OrderStatus string

// OrderInfo содержит информацию о заказе.
type OrderInfo struct {
	UserName  *string     `json:"user,omitempty"`        // UserName это имя пользователя, который разместил заказ.
	OrderID   string      `json:"number"`                // OrderID это уникальный идентификатор заказа.
	Order     *string     `json:"order,omitempty"`       // Order это детали заказа.
	CreatedAt *time.Time  `json:"uploaded_at,omitempty"` // CreatedAt это временная метка создания заказа.
	Status    OrderStatus `json:"status"`                // Status это состояние заказа.
	Accrual   float64     `json:"accrual"`               // Accrual это сумма начисления заказа.
}

// WithdrawInfo содержит информацию о списании средств(баллов).
type WithdrawInfo struct {
	UserName    *string    `json:"user,omitempty"`         // UserName это имя пользователя.
	OrderID     string     `json:"order"`                  // OrderID это идентификатор заказа.
	ProcessedAt *time.Time `json:"processed_at,omitempty"` // ProcessedAt это временная метка обработки заказа.
	Amount      float64    `json:"sum"`                    // Amount это сумма вывода средств.
}

// BalanceInfo содержит информацию о балансе.
type BalanceInfo struct {
	Current   float64 `json:"current"`   // Current это текущий баланс.
	Withdrawn float64 `json:"withdrawn"` // Withdrawn это сумма вывода средств.
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Credentials содержит информацию о учетных данных.
type Credentials struct {
	Password string `json:"password"` // Password это пароль.
	Username string `json:"login"`    // Username это имя пользователя.
}

// Option определяет функцию для настройки конфигурации.
type Option func(params *Config)

// Config содержит конфигурацию.
type Config struct {
	Server struct {
		Address string
	}
	Database struct {
		ConnectionString string
	}
	AccrualSystem struct {
		Address string
	}
}
