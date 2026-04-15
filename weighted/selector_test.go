package weighted

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestNewLottery(t *testing.T) {
	suite.Run(t, &TestLotterySuite{})
}

type TestLotterySuite struct {
	suite.Suite
}

func (l *TestLotterySuite) TestInt() {
	rand.Seed(time.Now().Unix())
	lt := New[int]()
	lt.Add(1, 1)
	lt.Add(2, 3)
	lt.Add(3, 7)
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
	lt := New[string]()
	lt.Add("A", 1)
	lt.Add("B", 1)
	lt.Add("C", 1)
	l.Equal(3, lt.total)
	i := lt.Draw()
	l.NotNil(i)
}
