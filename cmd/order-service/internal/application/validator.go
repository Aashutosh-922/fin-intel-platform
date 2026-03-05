package application

import "github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/domain"

func ValidateOrder(order domain.Order) error {
	if order.Quantity <= 0 {
		return domain.ErrInvalidQuantity
	}
	if order.Type == domain.Limit && order.Price <= 0 {
		return domain.ErrInvalidPrice
	}
	return nil
}