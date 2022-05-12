package lottery

import (
    "github.com/stretchr/testify/suite"
    "math/rand"
    "testing"
    "time"
)

func TestNewLottery(t *testing.T) {
    suite.Run(t, &TestLotterySuite{})
}

type TestLotterySuite struct {
    suite.Suite
}

func (l *TestLotterySuite) TestInt() {
    rand.Seed(time.Now().Unix())
    lt := NewLottery[int]()
    lt.AddItem(1, 1)
    lt.AddItem(2, 3)
    lt.AddItem(3, 7)
    l.Equal(3, lt.Len())
    counter := make(map[int]int)
    for i := 0; i < 100; i++ {
        r := lt.Draw()
        l.NotNil(r)
        counter[r]++
    }
    l.Greater(counter[3], counter[2])
    l.Greater(counter[2], counter[1])
}

func (l *TestLotterySuite) TestString() {
    lt := NewLottery[string]()
    lt.AddItem("A", 1)
    lt.AddItem("B", 1)
    lt.AddItem("C", 1)
    l.Equal(3, lt.total)
    i := lt.Draw()
    l.NotNil(i)
}
