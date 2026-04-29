package ternary

// If returns t if cond is true, otherwise returns f.
func If[T any](cond bool, t T, f T) T { //revive:disable:flag-parameter
	if cond {
		return t
	}
	return f
}
