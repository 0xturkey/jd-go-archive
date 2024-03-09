package utils

import (
	"math/rand"
	"time"
)

func GetRandomInt(min int, max int) int {
	// Create a new random number generator and seed it.
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)
	return min + rnd.Intn(max-min)
}
