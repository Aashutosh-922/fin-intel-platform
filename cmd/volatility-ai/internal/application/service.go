package application

import (
	"sync"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/volatility-ai/internal/domain"
)

type Service struct {
	detector *Detector
	mu       sync.RWMutex
	alerts   []domain.Alert
}

func NewService(detector *Detector) *Service {
	return &Service{
		detector: detector,
		alerts:   make([]domain.Alert, 0, 100),
	}
}

func (s *Service) ProcessTrade(trade domain.Trade) *domain.Alert {
	alert := s.detector.AddTrade(trade.Symbol, trade.Price)
	if alert != nil {
		s.addAlert(*alert)
	}
	return alert
}

func (s *Service) GetAlerts() []domain.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Alert, len(s.alerts))
	copy(out, s.alerts)
	return out
}

func (s *Service) addAlert(alert domain.Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.alerts = append(s.alerts, alert)
	if len(s.alerts) > 100 {
		s.alerts = s.alerts[len(s.alerts)-100:]
	}
}
