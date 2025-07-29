package utils

import (
	"fmt"
	"strconv"
)

func FormatFloatToFloat64(f float64) (float64, error) {
	formattedStr := fmt.Sprintf("%.2f", f)
	result, err := strconv.ParseFloat(formattedStr, 64)
	return result, err
}

func Str2Float64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
