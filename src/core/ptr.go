package core

// Ptr returns a pointer to the given value; avoids the need for one-off variables.
func Ptr[T any](x T) *T {
	return &x
}
