package utils

import "testing"

func TestGetInt64(t *testing.T) {
	str := "12346"
	num := GetInt64(str)
	t.Log(num)
}
