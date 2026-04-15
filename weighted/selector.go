package weighted

import (
	"math/rand"
)

// New create a new lottery pool
func New[T comparable]() *Selector[T] {
	var l Selector[T]
	return &l
}

type Selector[T comparable] struct {
	total  int
	weight int
	pool   []*item[T]
}

type Roller func(int) int

// Add add an item for Selector draw
func (l *Selector[T]) Add(ID T, weight int) {
	l.total++
	l.weight += weight
	l.pool = append(l.pool, newItem(ID, weight))
}

// Draw draw a lottery
func (l *Selector[T]) Draw(opts ...Option) T {
	options := &drawOptions{
		roller: rand.Intn,
	}
	for _, opt := range opts {
		opt(options)
	}
	r := options.roller(l.weight)
	return l.result(r).ID
}

// Len get amount of items
func (l *Selector[T]) Len() int {
	return l.total
}

// Weight get total weight of items
func (l *Selector[T]) Weight() int {
	return l.weight
}

func (l *Selector[T]) result(r int) *item[T] {
	_min := 0
	for _, v := range l.pool {
		_min += v.Weight
		if r < _min {
			return v
		}
	}
	return l.pool[0]
}

func (l *Selector[T]) Items() map[T]int {
	items := make(map[T]int, l.total)
	for _, v := range l.pool {
		items[v.ID] = v.Weight
	}
	return items
}
