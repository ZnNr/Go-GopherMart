package flags

import (
	"flag"
	"github.com/ZnNr/Go-GopherMart.git/internal/models"
	"os"
)

const (
	defaultAddr string = "localhost:8080"
)

// WithDatabase добавляет опцию для конфигурации строки подключения к базе данных.
func WithDatabase() models.Option {
	return func(p *models.Config) {
		flag.StringVar(&p.Database.ConnectionString, "d", "postgres://practicum:yandex@localhost:5432/postgres?sslmode=disable", "connection string for db")
		if envDBAddr := os.Getenv("DATABASE_URI"); envDBAddr != "" {
			p.Database.ConnectionString = envDBAddr
		}
	}
}

// WithAddr добавляет опцию для конфигурации адреса и порта сервера.
func WithAddr() models.Option {
	return func(p *models.Config) {
		flag.StringVar(&p.Server.Address, "a", defaultAddr, "address and port to run server")
		if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
			p.Server.Address = envRunAddr
		}
	}
}

// WithAccrual добавляет опцию для конфигурации адреса и порта системы начисления.
func WithAccrual() models.Option {
	return func(p *models.Config) {
		flag.StringVar(&p.AccrualSystem.Address, "r", "", "address and port to run server")
		if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
			p.AccrualSystem.Address = envAccrualAddr
		}
	}
}

// Init инициализирует конфигурацию с заданными опциями.
func Init(opts ...models.Option) *models.Config {
	p := &models.Config{}
	for _, opt := range opts {
		opt(p)
	}
	flag.Parse()
	return p
}
