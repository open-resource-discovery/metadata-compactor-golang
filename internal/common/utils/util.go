package utils

func First[S any](s S, _ ...any) S {
	return s
}

func Ternary[V any](condition bool, first V, second V) V {
	if condition {
		return first
	}

	return second
}
