package utils

import (
	"slices"
)

func None[E any](elements []E, predicate func(E) bool) bool {
	for _, element := range elements {
		if predicate(element) {
			return false
		}
	}

	return true
}

func Last[E any](elements []E) E {
	return elements[len(elements)-1]
}

func Map[I any, O any](elements []I, mapper func(int, I) O) []O {
	result := make([]O, 0, len(elements))

	for idx, element := range elements {
		result = append(result, mapper(idx, element))
	}

	return result
}

func OneOf[T comparable](value T, expected ...T) bool {
	return slices.Index(expected, value) >= 0
}
