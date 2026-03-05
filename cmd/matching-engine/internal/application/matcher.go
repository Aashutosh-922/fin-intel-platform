package application

import (
	"container/heap"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/domain"
)

type Broadcaster interface {
	Broadcast(symbol string, data []byte)
	BroadcastTrade(symbol string, data []byte)
}

type Matcher struct {
	sync.RWMutex
	orderBooks       map[string]*domain.OrderBook
	orderIndex       map[string]*domain.Order
	stopOrders       map[string][]*domain.Order
	takeProfitOrders map[string][]*domain.Order
	trailingOrders   map[string][]*domain.Order
	hub              Broadcaster
}

func NewMatcher(hub Broadcaster) *Matcher {
	return &Matcher{
		orderBooks:       make(map[string]*domain.OrderBook),
		orderIndex:       make(map[string]*domain.Order),
		stopOrders:       make(map[string][]*domain.Order),
		takeProfitOrders: make(map[string][]*domain.Order),
		trailingOrders:   make(map[string][]*domain.Order),
		hub:              hub,
	}
}

func (m *Matcher) Process(order domain.Order) []domain.Trade {
	m.Lock()
	defer m.Unlock()

	// if order.Type == domain.StopLoss {
	// 	m.stopOrders[order.Symbol] = append(
	// 		m.stopOrders[order.Symbol],
	// 		&order,
	// 	)
	// 	return nil
	// }

	switch order.Type {

	case domain.StopLoss:
		m.stopOrders[order.Symbol] = append(m.stopOrders[order.Symbol], &order)
		return nil

	case domain.TakeProfit:
		m.takeProfitOrders[order.Symbol] = append(m.takeProfitOrders[order.Symbol], &order)
		return nil

	case domain.TrailingStop:
		order.HighestSeen = order.Price
		order.LowestSeen = order.Price
		m.trailingOrders[order.Symbol] = append(m.trailingOrders[order.Symbol], &order)
		return nil
	}

	book, exists := m.orderBooks[order.Symbol]
	if !exists {
		book = domain.NewOrderBook()
		m.orderBooks[order.Symbol] = book
	}

	orderCopy := order
	m.orderIndex[order.OrderID] = &orderCopy

	var trades []domain.Trade

	if order.Type == domain.Market {
		trades = m.processMarketOrder(book, order)
	} else if order.Side == "BUY" {
		trades = m.matchBuy(book, order)
	} else {
		trades = m.matchSell(book, order)
	}

	m.evaluateTriggers(order.Symbol, trades)

	return trades
}

// func (m *Matcher) matchBuy(book *domain.OrderBook, buy domain.Order) []domain.Trade {
// 	var trades []domain.Trade

// 	for buy.Quantity > 0 && book.SellOrders.Len() > 0 {

// 		bestSell := book.SellOrders[0]

// 		if buy.Price < bestSell.Price {
// 			break
// 		}

// 		heap.Pop(&book.SellOrders)

// 		qty := min(buy.Quantity, bestSell.Quantity)

// 		trade := domain.Trade{
// 			TradeID:   generateID(),
// 			BuyOrder:  buy.OrderID,
// 			SellOrder: bestSell.OrderID,
// 			Symbol:    buy.Symbol,
// 			Price:     bestSell.Price,
// 			Quantity:  qty,
// 			Timestamp: time.Now().Unix(),
// 		}

// 		trades = append(trades, trade)

// 		buy.Quantity -= qty
// 		bestSell.Quantity -= qty

// 		if bestSell.Quantity > 0 {
// 			heap.Push(&book.SellOrders, bestSell)
// 		}
// 	}

// 	if buy.Quantity > 0 {
// 		heap.Push(&book.BuyOrders, &buy)
// 	}

//		return trades
//	}
func (m *Matcher) matchBuy(book *domain.OrderBook, buy domain.Order) []domain.Trade {
	var trades []domain.Trade

	for buy.Quantity > 0 {

		// 🔥 Clean top of SELL heap (lazy removal)
		for book.SellOrders.Len() > 0 {
			top := book.SellOrders[0]

			if top.Cancelled || top.Quantity == 0 {
				heap.Pop(&book.SellOrders)
				continue
			}
			break
		}

		// No valid sell orders left
		if book.SellOrders.Len() == 0 {
			break
		}

		bestSell := book.SellOrders[0]

		// Price check
		if buy.Price < bestSell.Price {
			break
		}

		heap.Pop(&book.SellOrders)

		qty := min(buy.Quantity, bestSell.Quantity)

		trade := domain.Trade{
			TradeID:    generateID(),
			BuyOrder:   buy.OrderID,
			SellOrder:  bestSell.OrderID,
			BuyUserID:  buy.UserID,
			SellUserID: bestSell.UserID,
			Symbol:     buy.Symbol,
			Price:      bestSell.Price,
			Quantity:   qty,
			Timestamp:  time.Now().Unix(),
		}

		trades = append(trades, trade)

		buy.Quantity -= qty
		bestSell.Quantity -= qty

		if bestSell.Quantity > 0 {
			heap.Push(&book.SellOrders, bestSell)
		}
	}

	// Remaining quantity goes back to BUY heap
	if buy.Quantity > 0 && !buy.Cancelled {
		heap.Push(&book.BuyOrders, &buy)
	}

	return trades
}

func (m *Matcher) matchSell(book *domain.OrderBook, sell domain.Order) []domain.Trade {
	var trades []domain.Trade

	for sell.Quantity > 0 {

		// 🔥 Lazy cleanup of BUY heap
		for book.BuyOrders.Len() > 0 {
			top := book.BuyOrders[0]

			if top.Cancelled || top.Quantity == 0 {
				heap.Pop(&book.BuyOrders)
				continue
			}
			break
		}

		// No valid buy orders left
		if book.BuyOrders.Len() == 0 {
			break
		}

		bestBuy := book.BuyOrders[0]

		// Price check
		if sell.Price > bestBuy.Price {
			break
		}

		heap.Pop(&book.BuyOrders)

		qty := min(sell.Quantity, bestBuy.Quantity)

		trade := domain.Trade{
			TradeID:    generateID(),
			BuyOrder:   bestBuy.OrderID,
			SellOrder:  sell.OrderID,
			BuyUserID:  bestBuy.UserID,
			SellUserID: sell.UserID,
			Symbol:     sell.Symbol,
			Price:      bestBuy.Price,
			Quantity:   qty,
			Timestamp:  time.Now().Unix(),
		}

		trades = append(trades, trade)

		sell.Quantity -= qty
		bestBuy.Quantity -= qty

		if bestBuy.Quantity > 0 {
			heap.Push(&book.BuyOrders, bestBuy)
		}
	}

	// Push remaining SELL order into heap
	if sell.Quantity > 0 && !sell.Cancelled {
		heap.Push(&book.SellOrders, &sell)
	}

	return trades
}

// func (m *Matcher) CancelOrder(cancel domain.OrderCancel) bool {

// 	order, exists := m.orderIndex[cancel.OrderID]
// 	if !exists {
// 		return false
// 	}

// 	order.Cancelled = true
// 	return true
// }

func (m *Matcher) CancelOrder(cancel domain.OrderCancel) bool {

	m.Lock()
	defer m.Unlock()

	order, exists := m.orderIndex[cancel.OrderID]
	if !exists {
		return false
	}

	order.Cancelled = true

	return true
}

func (m *Matcher) processMarketOrder(book *domain.OrderBook, order domain.Order) []domain.Trade {

	var trades []domain.Trade

	if order.Side == "BUY" {

		for order.Quantity > 0 {

			// Clean SELL heap
			for book.SellOrders.Len() > 0 {
				top := book.SellOrders[0]
				if top.Cancelled || top.Quantity == 0 {
					heap.Pop(&book.SellOrders)
					continue
				}
				break
			}

			if book.SellOrders.Len() == 0 {
				break
			}

			bestSell := heap.Pop(&book.SellOrders).(*domain.Order)

			qty := min(order.Quantity, bestSell.Quantity)

			trade := domain.Trade{
				TradeID:    generateID(),
				BuyOrder:   order.OrderID,
				SellOrder:  bestSell.OrderID,
				BuyUserID:  order.UserID,
				SellUserID: bestSell.UserID,
				Symbol:     order.Symbol,
				Price:      bestSell.Price, // market order takes best price
				Quantity:   qty,
				Timestamp:  time.Now().Unix(),
			}

			trades = append(trades, trade)

			order.Quantity -= qty
			bestSell.Quantity -= qty

			if bestSell.Quantity > 0 {
				heap.Push(&book.SellOrders, bestSell)
			}
		}

	} else { // MARKET SELL

		for order.Quantity > 0 {

			// Clean BUY heap
			for book.BuyOrders.Len() > 0 {
				top := book.BuyOrders[0]
				if top.Cancelled || top.Quantity == 0 {
					heap.Pop(&book.BuyOrders)
					continue
				}
				break
			}

			if book.BuyOrders.Len() == 0 {
				break
			}

			bestBuy := heap.Pop(&book.BuyOrders).(*domain.Order)

			qty := min(order.Quantity, bestBuy.Quantity)

			trade := domain.Trade{
				TradeID:    generateID(),
				BuyOrder:   bestBuy.OrderID,
				SellOrder:  order.OrderID,
				BuyUserID:  bestBuy.UserID,
				SellUserID: order.UserID,
				Symbol:     order.Symbol,
				Price:      bestBuy.Price,
				Quantity:   qty,
				Timestamp:  time.Now().Unix(),
			}

			trades = append(trades, trade)

			order.Quantity -= qty
			bestBuy.Quantity -= qty

			if bestBuy.Quantity > 0 {
				heap.Push(&book.BuyOrders, bestBuy)
			}
		}
	}

	// 🚨 IMPORTANT:
	// If market order not fully filled → discard remaining
	// Market orders NEVER go to heap.

	return trades
}

func (m *Matcher) GetSnapshot(symbol string, depth int) domain.OrderBookSnapshot {
	m.RLock()
	defer m.RUnlock()

	book, exists := m.orderBooks[symbol]
	if !exists {
		return domain.OrderBookSnapshot{
			Symbol: symbol,
			Bids:   []domain.Level{},
			Asks:   []domain.Level{},
		}
	}

	bids := m.extractLevels(book.BuyOrders, depth)
	asks := m.extractLevels(book.SellOrders, depth)

	var spread float64
	if len(bids) > 0 && len(asks) > 0 {
		spread = asks[0].Price - bids[0].Price
	}

	return domain.OrderBookSnapshot{
		Symbol: symbol,
		Bids:   bids,
		Asks:   asks,
		Spread: spread,
	}
}

func (m *Matcher) extractLevels(orders interface{}, depth int) []domain.Level {

	var heapCopy []*domain.Order
	isBuy := false

	switch h := orders.(type) {
	case domain.BuyHeap:
		heapCopy = append([]*domain.Order{}, h...)
		isBuy = true
	case domain.SellHeap:
		heapCopy = append([]*domain.Order{}, h...)
	}

	agg := make(map[float64]int)

	for _, order := range heapCopy {

		if order.Cancelled || order.Quantity == 0 {
			continue
		}

		agg[order.Price] += order.Quantity
	}

	prices := make([]float64, 0, len(agg))
	for p := range agg {
		prices = append(prices, p)
	}

	if isBuy {
		sort.Slice(prices, func(i, j int) bool { return prices[i] > prices[j] })
	} else {
		sort.Slice(prices, func(i, j int) bool { return prices[i] < prices[j] })
	}

	levels := make([]domain.Level, 0, min(depth, len(prices)))
	for i, p := range prices {
		if i >= depth {
			break
		}
		levels = append(levels, domain.Level{
			Price:    p,
			Quantity: agg[p],
		})
	}

	return levels
}

func (m *Matcher) broadcastSnapshot(symbol string) {

	if m.hub == nil {
		return
	}

	snapshot := m.GetSnapshot(symbol, 10)

	data, err := json.Marshal(snapshot)
	if err != nil {
		return
	}

	m.hub.Broadcast(symbol, data)
}

func (m *Matcher) broadcastTrade(trade domain.Trade) {

	if m.hub == nil {
		return
	}

	data, err := json.Marshal(trade)
	if err != nil {
		return
	}

	m.hub.BroadcastTrade(trade.Symbol, data)
}

// func (m *Matcher) evaluateStops(symbol string, trades []domain.Trade) {

// 	if len(trades) == 0 {
// 		return
// 	}

// 	lastPrice := trades[len(trades)-1].Price

// 	var remaining []*domain.Order
// 	triggered := false

// 	for _, stopOrder := range m.stopOrders[symbol] {

// 		trigger := false

// 		if stopOrder.Side == "SELL" && lastPrice <= stopOrder.StopPrice {
// 			trigger = true
// 		}

// 		if stopOrder.Side == "BUY" && lastPrice >= stopOrder.StopPrice {
// 			trigger = true
// 		}

// 		if trigger {

// 			// Convert to market
// 			stopOrder.Type = domain.Market

// 			// Execute immediately
// 			stopTrades := m.processMarketOrder(
// 				m.orderBooks[symbol],
// 				*stopOrder,
// 			)

// 			// 🔥 Broadcast trade tape for triggered trades
// 			for _, t := range stopTrades {
// 				m.broadcastTrade(t)
// 			}

// 			triggered = true
// 			continue
// 		}

// 		remaining = append(remaining, stopOrder)
// 	}

// 	m.stopOrders[symbol] = remaining

//		// 🔥 EXACT PLACE TO ADD SNAPSHOT BROADCAST
//		if triggered {
//			m.broadcastSnapshot(symbol)
//		}
//	}
func (m *Matcher) evaluateTriggers(symbol string, trades []domain.Trade) {

	if len(trades) == 0 {
		return
	}

	lastPrice := trades[len(trades)-1].Price

	triggered := false

	// -----------------------
	// STOP LOSS
	// -----------------------
	var remainingStop []*domain.Order

	for _, o := range m.stopOrders[symbol] {

		shouldTrigger := false

		if o.Side == "SELL" && lastPrice <= o.StopPrice {
			shouldTrigger = true
		}

		if o.Side == "BUY" && lastPrice >= o.StopPrice {
			shouldTrigger = true
		}

		if shouldTrigger {
			m.executeTriggered(symbol, o)
			triggered = true
			continue
		}

		remainingStop = append(remainingStop, o)
	}

	m.stopOrders[symbol] = remainingStop

	// -----------------------
	// TAKE PROFIT
	// -----------------------
	var remainingTP []*domain.Order

	for _, o := range m.takeProfitOrders[symbol] {

		shouldTrigger := false

		if o.Side == "SELL" && lastPrice >= o.TriggerPrice {
			shouldTrigger = true
		}

		if o.Side == "BUY" && lastPrice <= o.TriggerPrice {
			shouldTrigger = true
		}

		if shouldTrigger {
			m.executeTriggered(symbol, o)
			triggered = true
			continue
		}

		remainingTP = append(remainingTP, o)
	}

	m.takeProfitOrders[symbol] = remainingTP

	// -----------------------
	// TRAILING STOP
	// -----------------------
	var remainingTrail []*domain.Order

	for _, o := range m.trailingOrders[symbol] {

		if o.Side == "SELL" {

			if lastPrice > o.HighestSeen {
				o.HighestSeen = lastPrice
			}

			stopLevel := o.HighestSeen - o.TrailAmount

			if lastPrice <= stopLevel {
				m.executeTriggered(symbol, o)
				triggered = true
				continue
			}
		}

		if o.Side == "BUY" {

			if lastPrice < o.LowestSeen {
				o.LowestSeen = lastPrice
			}

			stopLevel := o.LowestSeen + o.TrailAmount

			if lastPrice >= stopLevel {
				m.executeTriggered(symbol, o)
				triggered = true
				continue
			}
		}

		remainingTrail = append(remainingTrail, o)
	}

	m.trailingOrders[symbol] = remainingTrail

	if triggered {
		m.broadcastSnapshot(symbol)
	}
}

func (m *Matcher) executeTriggered(symbol string, order *domain.Order) {

	order.Type = domain.Market

	trades := m.processMarketOrder(
		m.orderBooks[symbol],
		*order,
	)

	for _, t := range trades {
		m.broadcastTrade(t)
	}
}

func (m *Matcher) RestoreFromSnapshots(snapshots map[string]domain.OrderBookSnapshot) {
	m.Lock()
	defer m.Unlock()

	now := time.Now().Unix()
	for symbol, snap := range snapshots {
		book := domain.NewOrderBook()
		ts := now

		for _, lvl := range snap.Bids {
			if lvl.Quantity <= 0 {
				continue
			}
			book.BuyOrders = append(book.BuyOrders, &domain.Order{
				OrderID:   generateID(),
				UserID:    "restored",
				Symbol:    symbol,
				Side:      "BUY",
				Type:      domain.Limit,
				Price:     lvl.Price,
				Quantity:  lvl.Quantity,
				Timestamp: ts,
			})
			ts++
		}

		for _, lvl := range snap.Asks {
			if lvl.Quantity <= 0 {
				continue
			}
			book.SellOrders = append(book.SellOrders, &domain.Order{
				OrderID:   generateID(),
				UserID:    "restored",
				Symbol:    symbol,
				Side:      "SELL",
				Type:      domain.Limit,
				Price:     lvl.Price,
				Quantity:  lvl.Quantity,
				Timestamp: ts,
			})
			ts++
		}

		heap.Init(&book.BuyOrders)
		heap.Init(&book.SellOrders)
		m.orderBooks[symbol] = book
	}
}
