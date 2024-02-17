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
func New(dbManager *database.Manager, log *zap.SugaredLogger) *chi.Mux {
	handler := handlers.New(dbManager, log)
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", handler.Register)
		r.Post("/api/user/login", handler.Login)
	})
	r.Group(func(r chi.Router) {
		r.Use(handler.BasicAuth)
		r.Post("/api/user/orders", handler.LoadOrders)
		r.Post("/api/user/balance/withdraw", handler.Withdraw)
		r.Get("/api/user/orders", handler.GetOrders)
		r.Get("/api/user/withdrawals", handler.GetWithdrawals)
		r.Get("/api/user/balance", handler.GetBalance)
	})
	return r
}
