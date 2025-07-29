package utils

import (
	"math/rand"
	"time"
)

func IsHitGrey(rate int) bool {
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(101)
	return randomNum < rate
}
