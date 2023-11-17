package structparse

import (
    "reflect"
)

type Recipe[T any] struct {
    rType reflect.Type
    isPtr bool
    name  string
    path  string
}

func (r *Recipe[T]) Parse(instance T) {
    r.rType = reflect.TypeOf(instance)
    if r.rType.Kind() == reflect.Ptr {
        r.isPtr = true
        r.rType = r.rType.Elem()
    }
    r.name = r.rType.Name()
    r.path = r.rType.PkgPath()
}

func (r *Recipe[T]) Spawn() T {
    var instance reflect.Value
    instance = reflect.New(r.rType).Elem()
    if r.isPtr {
        instance = instance.Addr()
    }
    return instance.Interface().(T)
}

func (r *Recipe[T]) Name() string {
    return r.name
}

func (r *Recipe[T]) Path() string {
    return r.path
}
