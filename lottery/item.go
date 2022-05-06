package lottery

func newItem[T comparable](ID T, weight int) *item[T] {
    var i item[T]
    i.ID = ID
    i.Weight = weight
    return &i
}

type item[T comparable] struct {
    ID     T
    Weight int
}
