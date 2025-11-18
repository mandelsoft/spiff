package dynaml

func Optional[T comparable](def T, values ...T) T {
	var _nil T

	for _, v := range values {
		if v != _nil {
			def = v
		}
	}
	return def
}
