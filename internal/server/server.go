package server

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

// New создает новый экземпляр HTTP-сервера с заданным адресом и маршрутизатором.
func New(address string, router *chi.Mux) *http.Server {
	return &http.Server{
		Addr:    address,
		Handler: router,
	}
}
