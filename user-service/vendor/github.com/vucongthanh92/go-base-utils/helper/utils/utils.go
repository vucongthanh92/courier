package utils

func IterateSlice[T any](params []T, f func(i int, item T)) {
	if len(params) == 0 {
		return
	}
	for i, item := range params {
		f(i, item)
	}
}

func IterateMap[T comparable, K any](params map[T]K, f func(i T, item K)) {
	if params == nil {
		return
	}
	for i, item := range params {
		f(i, item)
	}
}
