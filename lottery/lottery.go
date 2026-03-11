package lottery

import (
	"math/rand"
)

// NewLottery create a new lottery pool
func NewLottery[T comparable]() *Lottery[T] {
	var l Lottery[T]
	return &l
}

type Lottery[T comparable] struct {
	total  int
	weight int
	pool   []*item[T]
}

type Roller func(int) int

// AddItem add an item for Lottery draw
func (l *Lottery[T]) AddItem(ID T, weight int) {
	l.total++
	l.weight += weight
	l.pool = append(l.pool, newItem(ID, weight))
}

// Draw draw a lottery
func (l *Lottery[T]) Draw(random ...Roller) T {
	var roller Roller
	if len(random) == 0 {
		roller = rand.Intn
	} else {
		roller = random[0]
		if roller == nil {
			roller = rand.Intn
		}
	}
	r := roller(l.weight)
	return l.result(r).ID
}

// Len get amount of items
func (l *Lottery[T]) Len() int {
	return l.total
}

// Weight get total weight of items
func (l *Lottery[T]) Weight() int {
	return l.weight
}

func (l *Lottery[T]) result(r int) *item[T] {
	_min := 0
	for _, v := range l.pool {
		_min += v.Weight
		if r < _min {
			return v
		}
	}
	return l.pool[0]
}

func (l *Lottery[T]) Items() map[T]int {
	items := make(map[T]int, l.total)
	for _, v := range l.pool {
		items[v.ID] = v.Weight
	}
	return items
}
