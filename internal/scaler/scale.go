package scaler

import "math"

type Resolution struct {
	W int
	H int
}

func Auto(res Resolution, mode string) Resolution {
	ratio := float64(res.W) / float64(res.H)

	factor := 1.5
	if mode == "down" {
		factor = 2.0 / 3.0
	}

	h := int(math.Round(float64(res.H) * factor))
	w := int(math.Round(float64(h) * ratio))

	w -= w % 2
	h -= h % 2

	if w < 320 {
		w = 320
		h = int(float64(w) / ratio)
		h -= h % 2
	}

	return Resolution{W: w, H: h}
}
