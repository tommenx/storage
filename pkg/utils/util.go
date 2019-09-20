package utils

import (
	"math"
	"strconv"
)

func GetInt64(val string) int64 {
	v, _ := strconv.ParseFloat(val, 64)
	num := math.Ceil(v)
	return int64(num)
}
func Int64(val string) int64 {
	v, _ := strconv.ParseInt(val, 0, 64)
	return v
}
