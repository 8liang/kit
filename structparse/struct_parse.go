package structparse

func Parse[T any](r T) *Recipe[T] {
    var recipe Recipe[T]
    recipe.Parse(r)
    return &recipe
}
