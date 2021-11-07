package util

import (
	"strconv"
	"time"
)

func ParseDateString(dateString string) (time.Time, error) {
	return time.Parse("2006-01-02", dateString)
}

func ParseInteger64String(int64String string) int64 {
	i, err := strconv.ParseInt(int64String, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func ParseFloat64String(float64String string) float64 {
	f, err := strconv.ParseFloat(float64String, 64)
	if err != nil {
		return -1
	}
	return f
}
