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
    var ok bool
    if len(random) == 0 {
        roller = rand.Intn
    } else {
        roller = random[0]
        if !ok {
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

func (l *Lottery[T]) result(r int) *item[T] {
    min := 0
    for _, v := range l.pool {
        min += v.Weight
        if r < min {
            return v
        }
    }
    return l.pool[0]
}
