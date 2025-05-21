package util

import (
	"os"
	"strings"
)

// StringInSlice проверяет, содержится ли заданная строка в указанном списке строк
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// IsSystemInTestMode проверяет, работает ли система в режиме тестирования
func IsSystemInTestMode() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}
