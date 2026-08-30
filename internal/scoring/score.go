package scoring

import "math"

func Calculate(story, acting, direction, technical, impact float64) float64 {
	score := story*0.3 + acting*0.2 + direction*0.2 + technical*0.15 + impact*0.15
	return math.Round(score*10) / 10
}