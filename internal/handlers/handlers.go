package handlers

import (
	"github.com/ZnNr/Go-GopherMart.git/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var jwtKey = []byte("my_secret_key")

func (h *handler) GetBalance(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) Withdraw(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) GetOrders(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) LoadOrders(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) BasicAuth(w http.Handler) http.Handler {

}

func (h *handler) extractJwtToken(r *http.Request) (*jwt.Token, error) {

}

func (h *handler) parseInputUser(r *http.Request) (*models.User, bool) {

}

func (h *handler) checkOrder(orderID string) bool {

}

func (h *handler) getUsernameFromToken(r *http.Request) (string, int) {

}

func New(db dbManager, log *zap.SugaredLogger) *handler {
	return &handler{
		db:  db,
		log: log,
	}
}

type handler struct {
	db  dbManager
	log *zap.SugaredLogger
}

type dbManager interface {
	GetBalanceInfo(login string) ([]byte, error)
	GetWithdrawals(login string) ([]byte, error)
	Withdraw(login string, orderID string, sum float64) error
	GetUserOrders(login string) ([]byte, error)
	LoadOrder(login string, orderID string) error
	Register(login string, password string) error
	Login(login string, password string) error
}

func createToken(userName string, expirationTime time.Time) (string, error) {

}
