package application

import (
	"context"
	"time"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/domain"
)

type Producer interface {
	Publish(ctx context.Context, event domain.OrderCreatedEvent) error
}

type IdempotencyRepo interface {
	Exists(orderID string) bool
	Save(orderID string)
}

type Service struct {
	producer Producer
	idRepo   IdempotencyRepo
}

type OrderCancelled struct {
	OrderID string
	Symbol  string
}

func NewService(p Producer, r IdempotencyRepo) *Service {
	return &Service{producer: p, idRepo: r}
}

func (s *Service) CreateOrder(ctx context.Context, order domain.Order) error {

	if s.idRepo.Exists(order.OrderID) {
		return domain.ErrDuplicateOrder
	}

	if err := ValidateOrder(order); err != nil {
		return err
	}

	if order.Type == domain.Limit && order.Price <= 0 {
		return domain.ErrInvalidPrice
	}

	event := domain.OrderCreatedEvent{
		EventID:   generateID(),
		OrderID:   order.OrderID,
		UserID:    order.UserID,
		Symbol:    order.Symbol,
		Side:      string(order.Side),
		Price:     order.Price,
		Quantity:  order.Quantity,
		Type:      string(order.Type),
		CreatedAt: time.Now().Unix(),
	}

	if err := s.producer.Publish(ctx, event); err != nil {
		return err
	}

	s.idRepo.Save(order.OrderID)
	return nil
}