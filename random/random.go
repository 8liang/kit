package random

type Random struct {
    seed int
}

func NewRandom(seed int) *Random {
    r := Random{seed: seed}
    return &r
}
func (c *Random) Random(weight int) int {
    c.seed = (c.seed*9301 + 49297) % 233280
    r := float64(c.seed) / 233280.0
    return int(r * float64(weight))
}

func (c *Random) SetSeed(i int) {
    c.seed = i
}
