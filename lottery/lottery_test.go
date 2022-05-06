package lottery

import (
    "github.com/stretchr/testify/suite"
    "testing"
)

func TestNewLottery(t *testing.T) {
    suite.Run(t, &TestLotterySuite{})
}

type TestLotterySuite struct {
    suite.Suite
}

func (l *TestLotterySuite) TestInt() {
    lt := NewLottery[int]()
    lt.AddItem(10, 1)
    lt.AddItem(100, 10)
    lt.AddItem(1000, 100)
    l.Equal(3, lt.Total())
    item := lt.Draw()
    l.NotNil(item)
}

func (l *TestLotterySuite) TestString() {
    lt := NewLottery[string]()
    lt.AddItem("A", 1)
    lt.AddItem("B", 1)
    lt.AddItem("C", 1)
    l.Equal(3, lt.total)
    item := lt.Draw()
    l.NotNil(item)
}
