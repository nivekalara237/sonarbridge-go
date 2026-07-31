package utils

func Ternary[T any](cond bool, True, False T) T {
	if cond {
		return True
	}
	return False
}
