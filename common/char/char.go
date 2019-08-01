package char

func IsDigit(a byte) bool {
	if a >= 48 && a <= 57 {
		return true
	}
	return false
}
