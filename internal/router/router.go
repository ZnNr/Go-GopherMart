package router

import (
	"github.com/ZnNr/Go-GopherMart.git/internal/database"
	"github.com/ZnNr/Go-GopherMart.git/internal/handlers"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// POST /api/user/register — регистрация пользователя;
// POST /api/user/login — аутентификация пользователя;
// POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
// GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
// GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
// POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
// GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
// SetupRouter настраивает маршрутизатор для обработки запросов API.
func SetupRouter(dbManager *database.Manager, log *zap.SugaredLogger) *chi.Mux {
	handler := handlers.New(dbManager, log)
	r := chi.NewRouter()
	// Группа маршрутов для регистрации и входа пользователей.
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", handler.RegisterHandler)
		r.Post("/api/user/login", handler.LoginHandler)
	})
	// Группа маршрутов для работы с заказами, балансом и выводами.
	r.Group(func(r chi.Router) {
		r.Use(handler.AuthenticateRequest)
		r.Post("/api/user/orders", handler.LoadOrderHandler)
		r.Post("/api/user/balance/withdraw", handler.WithdrawHandler)
		r.Get("/api/user/orders", handler.GetOrdersHandler)
		r.Get("/api/user/withdrawals", handler.GetWithdrawalsHandler)
		r.Get("/api/user/balance", handler.GetBalanceHandler)
	})

	return r
}
