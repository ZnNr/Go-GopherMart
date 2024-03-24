package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	errors2 "github.com/ZnNr/Go-GopherMart.git/internal/errors"
	"github.com/ZnNr/Go-GopherMart.git/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var jwtKey = []byte("my_secret_key")

// GetBalanceHandler обрабатывает запрос на получение баланса пользователя.
func (h *Handler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	// Получаем логин пользователя из токена и проверяем статус
	login, status := h.getUsernameFromTokenAndExtractClaims(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	// Получаем информацию о балансе пользователя из базы данных.
	userBalance, err := h.db.GetBalanceInfo(login)
	if err != nil {
		h.log.Errorf("error while getting user balance from db: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(userBalance)
}

// GetWithdrawalsHandler обрабатывает запрос на получение информации о выводах средств пользователя.
func (h *Handler) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	login, status := h.getUsernameFromTokenAndExtractClaims(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	userWithrdawals, err := h.db.GetWithdrawals(login)
	if err != nil {
		if errors.Is(err, errors2.ErrNoData) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.log.Errorf("error while getting withdrawals from db: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(userWithrdawals)
}

// WithdrawHandler принимает и обрабатывает запрос на вывод средств пользователя.
func (h *Handler) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var withdrawInfo *models.WithdrawInfo
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		h.log.Errorf("eeor while reading request body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Декодирование JSON-данных в структуру WithdrawInfo
	if err := json.Unmarshal(buf.Bytes(), &withdrawInfo); err != nil {
		h.log.Errorf("error while unmarshalling request body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Проверка формата заказа
	if !h.checkOrderLuhnValidation(withdrawInfo.OrderID) {
		h.log.Error("invalid order format")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	login, status := h.getUsernameFromTokenAndExtractClaims(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	// Выполнение операции вывода средств
	if err := h.db.Withdraw(login, withdrawInfo.OrderID, withdrawInfo.Amount); err != nil {
		if errors.Is(err, errors2.ErrInsufficientBalance) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		h.log.Errorf("error while trying to withdraw %f from user %q: %s", withdrawInfo.Amount, login, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.log.Infof("withdrawn %f from user %q for order %q", withdrawInfo.Amount, login, withdrawInfo.OrderID)
}

// GetOrdersHandler обрабатывает запрос на получение заказов пользователя.
func (h *Handler) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	/// Получаем логин пользователя из токена и проверяем статус
	login, status := h.getUsernameFromTokenAndExtractClaims(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	// Получение заказов пользователя из базы данных
	userOrders, err := h.db.GetUserOrders(login)
	if err != nil {
		if errors.Is(err, errors2.ErrNoData) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.log.Errorf("error while getting orders from db: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(userOrders)
}

// LoadOrderHandler обрабатывает запрос на загрузку заказа.
func (h *Handler) LoadOrderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	// Читаем данные из тела запроса
	var data bytes.Buffer
	if _, err := data.ReadFrom(r.Body); err != nil {
		h.log.Errorf("error while reading request body: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Получаем логин пользователя из токена и проверяем статус
	login, status := h.getUsernameFromTokenAndExtractClaims(r)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	// Получаем заказ из данных запроса
	order := data.String()
	if !h.checkOrderLuhnValidation(order) {
		h.log.Error("invalid order format")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	// Загружаем заказ в базу данных
	if err := h.db.LoadOrder(login, order); err != nil {
		if errors.Is(err, errors2.ErrCreatedBySameUser) {
			h.log.Info(fmt.Sprintf("order %q was alredy created by the same user", order))
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, errors2.ErrCreatedDiffUser) {
			h.log.Info(fmt.Sprintf("order %q was alredy created by the other user", order))
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.log.Errorf("error while loading order to db: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Успешное завершение обработки запроса
	w.WriteHeader(http.StatusAccepted)
}

// LoginHandler обрабатывает запрос на авторизацию пользователя.
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	// Парсинг введенных пользователем данных.
	user, success := h.parseInputUser(r)
	if !success {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Проверка соответствия логина и пароля в базе данных.
	if err := h.db.Login(user.Login, user.Password); err != nil {
		h.log.Errorf("error while login user: %s", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Генерация и установка токена авторизации.
	expirationTime := time.Now().Add(time.Hour)
	token, err := createToken(user.Login, expirationTime)
	if err != nil {
		h.log.Errorf("error while create token for user: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	// Устанавливаем cookie с токеном авторизации.
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: expirationTime,
	})
	h.log.Info(fmt.Sprintf("user %q is successfully authorized", user.Login))
}

// RegisterHandler обрабатывает запрос на регистрацию пользователя.
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	// Парсинг введенных пользователем данных.
	user, success := h.parseInputUser(r)
	if !success {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Регистрация пользователя в базе данных.
	if err := h.db.Register(user.Login, user.Password); err != nil {
		if errors.Is(err, errors2.ErrUserAlreadyExists) {
			h.log.Errorf("login is already taken: %s", err.Error())
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.log.Errorf("error while register user: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Авторизация пользователя после регистрации.
	if err := h.db.Login(user.Login, user.Password); err != nil {
		h.log.Errorf("error while login user: %s", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Создание токена для авторизованного пользователя.
	expirationTime := time.Now().Add(time.Hour)
	token, err := createToken(user.Login, expirationTime)
	if err != nil {
		h.log.Errorf("error while create token for user: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Установка заголовка с авторизационным токеном и установка куки.
	w.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: expirationTime,
	})
	h.log.Info(fmt.Sprintf("user %q is successfully registered and authorized", user.Login))
}

// AuthenticateRequest проверяет наличие и валидность токена авторизации.
func (h *Handler) AuthenticateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")
		if tokenHeader == "" {
			h.log.Errorf("token is empty")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Извлечение токена из заголовка Authorization.
		tkn, err := h.extractJwtToken(r)
		if err != nil {
			// Обработка ошибок связанных с некорректным или истекшим токеном.
			if errors.Is(err, jwt.ErrSignatureInvalid) ||
				errors.Is(err, jwt.ErrTokenExpired) ||
				errors.Is(err, errors2.ErrTokenIsEmpty) ||
				errors.Is(err, errors2.ErrNoToken) {
				h.log.Errorf(err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h.log.Errorf(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Проверка валидности токена.
		if !tkn.Valid {
			h.log.Errorf("invalid token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Передача заголовка Authorization в следующий обработчик.
		w.Header().Add("Authorization", tokenHeader)
		next.ServeHTTP(w, r)
	})
}

// ExtractJwtToken извлекает и разбирает токен JWT из заголовка авторизации.
func (h *Handler) extractJwtToken(r *http.Request) (*jwt.Token, error) {
	tokenHeader := r.Header.Get("Authorization")
	if tokenHeader == "" {
		h.log.Errorf("token is empty")
		return nil, errors2.ErrTokenIsEmpty
	}
	splitted := strings.Split(tokenHeader, " ")
	if len(splitted) != 2 {
		h.log.Errorf("no token")
		return nil, errors2.ErrNoToken
	}

	tknStr := splitted[1]
	claims := &models.Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	return tkn, err
}

// ParseInputUser парсит и читает информацию о пользователе из HTTP-запроса.
func (h *Handler) parseInputUser(r *http.Request) (*models.User, bool) {
	var userFromRequest *models.User
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		h.log.Errorf("error while reading request body: %s", err.Error())
		return nil, false
	}
	if err := json.Unmarshal(buf.Bytes(), &userFromRequest); err != nil {
		h.log.Errorf("error while unmarshalling request body: %s", err.Error())
		return nil, false
	}
	if userFromRequest.Login == "" || userFromRequest.Password == "" {
		h.log.Errorf("login or password is empty")
		return nil, false
	}
	return userFromRequest, true
}

// checkOrderLuhnValidation проверяет ID заказа на соответствие алгоритму Luhn.
func (h *Handler) checkOrderLuhnValidation(orderID string) bool {
	// Преобразование ID заказа в целое число.
	orderAsInteger, err := strconv.Atoi(orderID)
	if err != nil {
		return false
	}
	// Инициализация переменных для выполнения алгоритма Luhn.
	number := orderAsInteger / 10
	luhn := 0
	// Выполнение алгоритма Luhn на числе number.
	for i := 0; number > 0; i++ {
		c := number % 10
		if i%2 == 0 {
			c *= 2
			if c > 9 {
				c = c%10 + c/10
			}
		}
		luhn += c
		number /= 10
	}
	// Проверка, прошел ли номер заказа валидацию алгоритмом Luhn.
	return (orderAsInteger%10+luhn)%10 == 0
}

// getUsernameFromTokenAndExtractClaims извлекает имя пользователя из токена и возвращает его вместе с HTTP-статусом.
func (h *Handler) getUsernameFromTokenAndExtractClaims(r *http.Request) (string, int) {
	// Чтение тела запроса в буфер.
	var data bytes.Buffer
	if _, err := data.ReadFrom(r.Body); err != nil {
		h.log.Errorf("error while reading request body: %s", err.Error())
		return "", http.StatusBadRequest
	}
	// Извлечение токена из запроса.
	tkn, err := h.extractJwtToken(r)
	if err != nil {
		h.log.Errorf("error while extracting token: %s", err.Error())
		return "", http.StatusInternalServerError
	}
	// Проверка типа утверждений и извлечение имени пользователя.
	claims, ok := tkn.Claims.(*models.Claims)
	if !ok {
		h.log.Errorf("error while getting claims")
		return "", http.StatusInternalServerError
	}
	// Возвращение имени пользователя и HTTP-статуса.
	return claims.Username, http.StatusOK
}

// New создает новый экземпляр структуры Handler и возвращает его.
func New(db DBManager, log *zap.SugaredLogger) *Handler {
	return &Handler{
		db:  db,
		log: log,
	}
}

type Handler struct {
	db  DBManager
	log *zap.SugaredLogger
}

// DBManager представляет интерфейс для взаимодействия с базой данных.
//
//go:generate mockery --disable-version-string --filename db_mock.go --inpackage --name dbManager
type DBManager interface {
	GetBalanceInfo(login string) ([]byte, error)              // GetBalanceInfo возвращает информацию о балансе пользователя по его логину.
	GetWithdrawals(login string) ([]byte, error)              // GetWithdrawals возвращает список выводов пользователя по его логину.
	Withdraw(login string, orderID string, sum float64) error // Withdraw осуществляет вывод средств для заданного пользователя, заказа и суммы.
	GetUserOrders(login string) ([]byte, error)               // GetUserOrders возвращает список заказов пользователя по его логину.
	LoadOrder(login string, orderID string) error             // LoadOrder загружает информацию о заданном заказе пользователя по его логину и идентификатору заказа.
	Register(login string, password string) error             // Register регистрирует нового пользователя с заданным логином и паролем.
	Login(login string, password string) error                // Login выполняет вход пользователя с заданным логином и паролем.
}

// createToken создает токен аутентификации для заданного пользователя и времени истечения срока действия.
func createToken(userName string, expirationTime time.Time) (string, error) {
	// Создаем структуру Claims с информацией о пользователе и времени истечения срока действия токена.
	claims := &models.Claims{
		Username: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	// Создаем новый токен с заданными Claims и методом подписи.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
