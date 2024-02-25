package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	errors2 "github.com/ZnNr/Go-GopherMart.git/internal/errors"
	"github.com/ZnNr/Go-GopherMart.git/internal/models"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// GetBalanceInfo возвращает информацию о балансе пользователя и сумме снятых средств.
func (m *Manager) GetBalanceInfo(login string) ([]byte, error) {
	// Получение текущего баланса пользователя
	userBalance, err := m.getUserBalance(login)
	if err != nil {
		return nil, fmt.Errorf("error while getting curent user balance: %w", err)
	}
	// Получение суммы снятых средств у пользователя
	getUserWithdrawn := "select sum(amount) as withdrawn from withdraw where login = $1"
	row := m.db.QueryRow(getUserWithdrawn, login)
	var userWithdrawn sql.NullFloat64
	if err = row.Scan(&userWithdrawn); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error while getting user withdrawn info: %w", err)
		}
		userWithdrawn = sql.NullFloat64{
			Float64: 0,
			Valid:   true,
		}
	}
	// Формирование структуры с информацией о балансе пользователя и сумме снятых средств
	info := models.BalanceInfo{
		Withdrawn: userWithdrawn.Float64,
		Current:   userBalance,
	}
	// Преобразование структуры в JSON
	result, err := json.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling user balance info : %w", err)
	}
	return result, nil
}

// GetWithdrawals возвращает информацию о снятых средствах пользователя.
func (m *Manager) GetWithdrawals(login string) ([]byte, error) {
	getUserWithdrawals := `select order_id, amount, processed_at from withdraw where login = $1 order by processed_at`
	rows, err := m.db.Query(getUserWithdrawals, login)
	if err != nil {
		return nil, fmt.Errorf("error while searching for user withdrawals: %w", err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	userWithdrawals := make([]models.WithdrawInfo, 0)
	for rows.Next() {
		var (
			orderID     string
			amount      float64
			processedAt time.Time
		)
		if err = rows.Scan(&orderID, &amount, &processedAt); err != nil {
			return nil, fmt.Errorf("error while scanning rows from userWithdrawals: %w", err)
		}
		userWithdrawals = append(userWithdrawals, models.WithdrawInfo{
			OrderID:     orderID,
			ProcessedAt: &processedAt,
			Amount:      amount,
		})
	}
	if len(userWithdrawals) == 0 {
		return nil, errors2.ErrNoData
	}
	result, err := json.Marshal(userWithdrawals)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling user withdrawals info: %w", err)
	}
	return result, nil
}

// Withdraw осуществляет снятие средств со счета пользователя.
func (m *Manager) Withdraw(login string, orderID string, sum float64) error {
	userBalance, err := m.getUserBalance(login)
	if err != nil {
		return fmt.Errorf("error while checking user balance: %w", err)
	}
	if userBalance < sum {
		return errors2.ErrInsufficientBalance
	}
	withdraw := "insert into withdraw values ($1, $2, now(), $3)"
	if _, err = m.db.Exec(withdraw, login, orderID, sum); err != nil {
		return fmt.Errorf("error while trying to withdraw: %w", err)
	}
	return nil
}

// GetUserOrders получает информацию о заказах пользователя.
func (m *Manager) GetUserOrders(login string) ([]byte, error) {
	// Запрос на получение заказов пользователя из базы данных.
	getUserOrdersQuery := `select order_id, status, accrual, uploaded_at from orders where login = $1`
	rows, err := m.db.Query(getUserOrdersQuery, login)
	if err != nil {
		return nil, fmt.Errorf("error while getting orders from db for user %q: %w", login, err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	// Слайс для хранения информации о заказах пользователя.
	userOrders := make([]models.OrderInfo, 0)
	// Перебираем строки, полученные из базы данных.
	for rows.Next() {
		var (
			orderID    string
			status     models.OrderStatus
			accrual    sql.NullFloat64
			uploadedAt time.Time
		)
		// Сканируем строки и извлекаем значения в переменные.
		if err = rows.Scan(&orderID, &status, &accrual, &uploadedAt); err != nil {
			return nil, fmt.Errorf("error while scanning rows: %w", err)
		}
		// Добавляем информацию о заказе в слайс.
		userOrders = append(userOrders, models.OrderInfo{
			OrderID:   orderID,
			Accrual:   accrual.Float64,
			CreatedAt: &uploadedAt,
			Status:    status,
		})
	}
	if len(userOrders) == 0 {
		return nil, errors2.ErrNoData
	}
	result, err := json.Marshal(userOrders)
	if err != nil {
		return nil, fmt.Errorf("error while marshaling user orders info: %w", err)
	}
	return result, nil
}

// GetAllOrders получает все заказы.
func (m *Manager) GetAllOrders() ([]string, error) {
	getAllOrdersQuery := `select order_id from orders`
	rows, err := m.db.Query(getAllOrdersQuery)
	if err != nil {
		return nil, fmt.Errorf("error while getting all orders from db: %w", err)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	//Слайс для хранения заказов.
	orders := make([]string, 0)
	for rows.Next() {
		var orderID string
		if err = rows.Scan(&orderID); err != nil {
			return nil, fmt.Errorf("error while scanning rows: %w", err)
		}
		orders = append(orders, orderID)
	}
	return orders, nil
}

// UpdateOrderInfo обновляет информацию о заказе.
func (m *Manager) UpdateOrderInfo(orderInfo *models.OrderInfo) error {
	// Запрос на обновление информации о заказе в базе данных.
	updateOrderInfoQuery := `update orders set status=$1, accrual=$2 where order_id=$3`
	if _, err := m.db.Exec(updateOrderInfoQuery, string(orderInfo.Status), orderInfo.Accrual, orderInfo.Order); err != nil {
		return fmt.Errorf("error while updating order info: %w", err)
	}
	return nil
}

// LoadOrder загружает заказ для указанного логина и идентификатора заказа.
func (m *Manager) LoadOrder(login string, orderID string) error {
	// Проверяем, существует ли заказ с указанным идентификатором.
	getOrderByIDQuery := `select login from orders where order_id = $1`
	row := m.db.QueryRow(getOrderByIDQuery, orderID)

	var userName string
	err := row.Scan(&userName)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// Если заказ не существует, создаем новый заказ.
		loadOrderQuery := `insert into orders values ($1, $2, now(), $3, $4)`
		if _, err = m.db.Exec(loadOrderQuery, orderID, login, models.OrderStatus("NEW"), 0); err != nil {
			return fmt.Errorf("error while loading order %s: %w", orderID, err)
		}
		return nil
	case err == nil:
		// Если заказ уже существует, проверяем, создан ли он тем же пользователем.
		if userName == login {
			return errors2.ErrCreatedBySameUser
		}
		return errors2.ErrCreatedDiffUser
	default:
		return fmt.Errorf("error while scanning rows: %w", err)
	}
}

// Register регистрирует нового пользователя с указанным логином и паролем.
func (m *Manager) Register(login string, password string) error {
	// Хэшируем пароль с использованием bcrypt.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error, this password is not allowed: %w", err)
	}
	// Запрос для добавления нового зарегистрированного пользователя.
	registerUserQuery := `insert into registered_users values ($1, $2)`
	if _, err = m.db.Exec(registerUserQuery, login, hash); err != nil {
		// Обрабатываем возможные ошибки при выполнении запроса.
		duplicateKeyErr := errors2.ErrDuplicateKey{Key: "registered_users_pkey"}
		if err.Error() == duplicateKeyErr.Error() {
			return errors2.ErrUserAlreadyExists
		}
		return fmt.Errorf("error while executing register user query: %w", err)
	}
	return nil

}

// Login выполняет аутентификацию пользователя с указанным логином и паролем.
func (m *Manager) Login(login string, password string) error {
	// Запрос для получения зарегистрированных пользователей.
	getRegisteredUserQuery := "select login, password from registered_users"
	rows, err := m.db.Query(getRegisteredUserQuery)
	if err != nil {
		return fmt.Errorf("error while executing search query: %w", err)

	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	// Проверяем каждого пользователя в результатах запроса.
	for rows.Next() {
		// Проверяем, соответствует ли логин и пароль.
		var loginFromDB, passwordFromDB string
		if err = rows.Scan(&loginFromDB, &passwordFromDB); err != nil {
			return fmt.Errorf("error while scanning rows: %w", err)
		}
		if loginFromDB == login {
			if err = bcrypt.CompareHashAndPassword([]byte(passwordFromDB), []byte(password)); err != nil {
				return errors2.ErrInvalidCredentials
			}
			return nil
		}
	}
	// Если пользователь не найден, возвращаем ошибку.
	return errors2.ErrNoSuchUser
}

// GetUserBalance возвращает баланс пользователя с указанным логином.
func (m *Manager) getUserBalance(login string) (float64, error) {
	// Запрос для получения баланса пользователя.
	getUserBalanceQuery := "select coalesce(sum(accrual), 0) - coalesce(sum(amount), 0) as balance from orders o left join withdraw w on o.login = w.login where o.login = $1 group by o.login;"
	row := m.db.QueryRow(getUserBalanceQuery, login)
	var balance sql.NullFloat64
	if err := row.Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("error while getting user balance: %w", err)

	}
	return balance.Float64, nil

}

// init создает необходимые таблицы, если они еще не существуют.
func (m *Manager) init(ctx context.Context) error {
	// Создание таблицы зарегистрированных пользователей, если она не существует.
	createRegisteredQuery := `create table if not exists registered_users (login text primary key, password text)`
	if _, err := m.db.ExecContext(ctx, createRegisteredQuery); err != nil {
		return fmt.Errorf("error while trying to create table with registered users: %w", err)
	}
	// Создание таблицы заказов, если она не существует.
	createOrdersQuery := `create table if not exists orders (order_id text unique, login text, uploaded_at timestamp with time zone, status text, accrual double precision, primary key(order_id))`
	if _, err := m.db.ExecContext(ctx, createOrdersQuery); err != nil {
		return fmt.Errorf("error while trying to create table with orders: %w", err)
	}
	// Создание таблицы выводов, если она не существует.
	createWithdrawQuery := `create table if not exists withdraw (login text, order_id text unique, processed_at timestamp with time zone, amount double precision, primary key(login, order_id))`
	if _, err := m.db.ExecContext(ctx, createWithdrawQuery); err != nil {
		return fmt.Errorf("error while trying to create table with orders: %w", err)
	}
	return nil
}

// New создает новый экземпляр Manager с переданным db и инициализирует необходимые таблицы.
func New(ctx context.Context, db *sql.DB) (*Manager, error) {
	m := Manager{
		db: db,
	}
	// Инициализация таблиц.
	if err := m.init(ctx); err != nil {
		return nil, err
	}
	return &m, nil
}

// Manager представляет менеджер базы данных.
type Manager struct {
	db *sql.DB
}
