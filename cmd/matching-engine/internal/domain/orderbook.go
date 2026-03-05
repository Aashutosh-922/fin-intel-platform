package domain

import "container/heap"

type BuyHeap []*Order
type SellHeap []*Order

type OrderBook struct {
	BuyOrders  BuyHeap
	SellOrders SellHeap
}

// Max Heap for BUY
func (h BuyHeap) Len() int { return len(h) }
func (h BuyHeap) Less(i, j int) bool {
	if h[i].Price == h[j].Price {
		return h[i].Timestamp < h[j].Timestamp
	}
	return h[i].Price > h[j].Price
}
func (h BuyHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *BuyHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}
func (h *BuyHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Min Heap for SELL
func (h SellHeap) Len() int { return len(h) }
func (h SellHeap) Less(i, j int) bool {
	if h[i].Price == h[j].Price {
		return h[i].Timestamp < h[j].Timestamp
	}
	return h[i].Price < h[j].Price
}
func (h SellHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *SellHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}
func (h *SellHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func NewOrderBook() *OrderBook {
	ob := &OrderBook{}
	heap.Init(&ob.BuyOrders)
	heap.Init(&ob.SellOrders)
	return ob
}