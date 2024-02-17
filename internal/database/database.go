package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/Go-GopherMart.git/internal/models"
)

func (m *Manager) GetBalanceInfo(login string) ([]byte, error) {

}

func (m *Manager) GetWithdrawals(login string) ([]byte, error) {

}

func (m *Manager) Withdraw(login string, orderID string, sum float64) error {

}

func (m *Manager) GetUserOrders(login string) ([]byte, error) {

}

func (m *Manager) GetAllOrders() ([]string, error) {

}

func (m *Manager) UpdateOrderInfo(orderInfo *models.OrderInfo) error {

}

func (m *Manager) LoadOrder(login string, orderID string) error {

}

func (m *Manager) Register(login string, password string) error {

}

func (m *Manager) Login(login string, password string) error {

}

func (m *Manager) getUserBalance(login string) (float64, error) {

}

func (m *Manager) init(ctx context.Context) error {
	createRegisteredQuery := `CREATE TABLE IF NOT EXISTS registered_users (login TEXT PRIMARY KEY, password TEXT)`
	if _, err := m.db.ExecContext(ctx, createRegisteredQuery); err != nil {
		return fmt.Errorf("error while trying to create table with registered users: %w", err)
	}
	createOrdersQuery := `CREATE TABLE IF NOT EXISTS orders (order_id TEXT UNIQUE, login TEXT, uploaded_at TIMESTAMP WITH TIME ZONE, status TEXT, accrual DOUBLE PRECISION, PRIMARY KEY(order_id))`
	if _, err := m.db.ExecContext(ctx, createOrdersQuery); err != nil {
		return fmt.Errorf("error while trying to create table with orders: %w", err)
	}
	createWithdrawQuery := `CREATE TABLE IF NOT EXISTS withdraw (login TEXT, order_id TEXT UNIQUE, processed_at TIMESTAMP WITH TIME ZONE, amount DOUBLE PRECISION, PRIMARY KEY(login, order_id))`
	if _, err := m.db.ExecContext(ctx, createWithdrawQuery); err != nil {
		return fmt.Errorf("error while trying to create table with orders: %w", err)
	}
	return nil
}

func New(ctx context.Context, db *sql.DB) (*Manager, error) {
	m := Manager{
		db: db,
	}
	if err := m.init(ctx); err != nil {
		return nil, err
	}
	return &m, nil
}

type Manager struct {
	db *sql.DB
}
