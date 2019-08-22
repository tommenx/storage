package utils

import "strconv"

func GetInt64(val string) int64 {
	v, _ := strconv.ParseInt(val, 0, 64)
	return v
}
