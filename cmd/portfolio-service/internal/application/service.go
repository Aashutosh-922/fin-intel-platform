package application

import (
	"context"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/domain"
)

// type Repository interface {
// 	GetPosition(ctx context.Context, userID, symbol string) (*domain.Position, error)
// 	SavePosition(ctx context.Context, pos *domain.Position) error
// }

// type Service struct {
// 	repo Repository
// }

//	func NewService(r Repository) *Service {
//		return &Service{repo: r}
//	}
type Repository interface {
	GetPosition(ctx context.Context, userID, symbol string) (*domain.Position, error)
	GetAllBySymbol(ctx context.Context, symbol string) ([]domain.Position, error)
	SavePosition(ctx context.Context, pos *domain.Position) error
}

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) ProcessTrade(ctx context.Context, trade domain.Trade) error {
	// Preferred event shape has explicit buy/sell users.
	if trade.BuyUserID != "" || trade.SellUserID != "" {
		if trade.BuyUserID != "" {
			if err := s.applyTrade(ctx, trade.BuyUserID, trade.Symbol, trade.Price, trade.Quantity); err != nil {
				return err
			}
		}
		if trade.SellUserID != "" {
			if err := s.applyTrade(ctx, trade.SellUserID, trade.Symbol, trade.Price, -trade.Quantity); err != nil {
				return err
			}
		}
		return nil
	}

	// Backward compatibility for legacy single-user trade events.
	return s.applyTrade(ctx, trade.UserID, trade.Symbol, trade.Price, trade.Quantity)
}

func (s *Service) applyTrade(ctx context.Context, userID, symbol string, price float64, quantity int) error {
	pos, _ := s.repo.GetPosition(ctx, userID, symbol)
	if pos == nil {
		pos = &domain.Position{UserID: userID, Symbol: symbol}
	}

	if quantity > 0 {
		totalCost := pos.AvgPrice*float64(pos.Quantity) + price*float64(quantity)
		pos.Quantity += quantity
		if pos.Quantity > 0 {
			pos.AvgPrice = totalCost / float64(pos.Quantity)
		}
	} else {
		sellQty := -quantity
		pnl := (price - pos.AvgPrice) * float64(sellQty)
		pos.RealizedPnL += pnl
		pos.Quantity -= sellQty
	}

	return s.repo.SavePosition(ctx, pos)
}

func (s *Service) ProcessMarketTick(ctx context.Context, tick domain.MarketTick) error {

	positions, err := s.repo.GetAllBySymbol(ctx, tick.Symbol)
	if err != nil {
		return err
	}

	for _, pos := range positions {

		pos.UnrealizedPnL =
			(tick.Price - pos.AvgPrice) *
				float64(pos.Quantity)

		if err := s.repo.SavePosition(ctx, &pos); err != nil {
			return err
		}
	}

	return nil
}
