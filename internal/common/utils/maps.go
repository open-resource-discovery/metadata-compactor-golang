package utils

func ContainsKey[K comparable, V any](value map[K]V, key K) bool {
	_, ok := value[key]

	return ok
}
