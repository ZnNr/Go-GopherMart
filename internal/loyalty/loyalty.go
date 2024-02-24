package loyalty

import (
	"encoding/json"
	"fmt"
	"github.com/ZnNr/Go-GopherMart.git/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// UpdateOrdersInfo обновляет информацию о заказах в системе лояльности.
func (ls *LoyaltySystemManager) UpdateOrdersInfo() error {
	// Получаем все заказы из базы данных
	allOrders, err := ls.db.GetAllOrders()
	if err != nil {
		return fmt.Errorf("error while getting all orders from db for updating info: %w", err)
	}
	// Обновляем информацию по каждому заказу
	for _, o := range allOrders {
		actualInfo, err := ls.getActualInfo(o)
		if err != nil {
			return fmt.Errorf("error while getting actual info for order %q: %w", o, err)
		}
		// Обновляем информацию о заказе в базе данных
		if err = ls.db.UpdateOrderInfo(actualInfo); err != nil {
			return fmt.Errorf("error while updating order info: %w", err)
		}
		ls.log.Infof("order %q updated with accrual: %f", *actualInfo.Order, actualInfo.Accrual)
	}
	return nil
}

// getActualInfo получает актуальную информацию о заказе.
func (ls *LoyaltySystemManager) getActualInfo(orderID string) (*models.OrderInfo, error) {
	// Выполняем запрос к системе для получения информации о заказе
	orderFromSystem, err := resty.New().R().Get(fmt.Sprintf("%s/api/orders/%s", ls.addr, orderID))
	if err != nil {
		return nil, fmt.Errorf("error while requesting for order %q: %w", orderID, err)
	}
	var info models.OrderInfo
	// Декодируем тело ответа в структуру OrderInfo
	if err = json.Unmarshal(orderFromSystem.Body(), &info); err != nil {
		return nil, fmt.Errorf("error while unmarshalling order body: %w", err)
	}
	return &info, nil
}

func New(addr string, db DBManager, logger *zap.SugaredLogger) *LoyaltySystemManager {
	return &LoyaltySystemManager{
		addr: addr,
		db:   db,
		log:  logger,
	}
}

type LoyaltySystemManager struct {
	addr string
	db   DBManager
	log  *zap.SugaredLogger
}

type DBManager interface {
	GetAllOrders() ([]string, error)
	UpdateOrderInfo(orderInfo *models.OrderInfo) error
}
