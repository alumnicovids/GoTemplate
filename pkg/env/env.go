package env

import (
	"cmp"
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	return cmp.Or(os.Getenv(key), fallback)
}

func GetInt(key string, fallback int) int {
	val, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return val
}
