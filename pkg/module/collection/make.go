package collection

// MakeSlice creates a slice with the given capacity. If capacity is 0, returns nil.
func MakeSlice[T any](capacity int) []T {
	if capacity == 0 {
		return nil
	}
	return make([]T, 0, capacity)
}
