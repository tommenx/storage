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
