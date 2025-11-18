package utils

import (
	"image/color"
	"math"
)

// GenerateDistinctColors returns n visually distinct colors
func GenerateDistinctColors(n int) []color.RGBA {
	colors := make([]color.RGBA, n)
	for i := 0; i < n; i++ {
		h := float64(i) / float64(n) // Hue: 0.0 - 1.0
		s := 0.65                    // Saturation
		v := 0.95                    // Value (brightness)
		r, g, b := HSVtoRGB(h, s, v)
		colors[i] = color.RGBA{R: r, G: g, B: b, A: 255}
	}
	return colors
}

// HSVtoRGB converts HSV to RGB (values 0-255)
func HSVtoRGB(h, s, v float64) (r, g, b uint8) {
	var rf, gf, bf float64

	i := math.Floor(h * 6)
	f := h*6 - i
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)

	switch int(i) % 6 {
	case 0:
		rf, gf, bf = v, t, p
	case 1:
		rf, gf, bf = q, v, p
	case 2:
		rf, gf, bf = p, v, t
	case 3:
		rf, gf, bf = p, q, v
	case 4:
		rf, gf, bf = t, p, v
	case 5:
		rf, gf, bf = v, p, q
	}

	r = uint8(rf * 255)
	g = uint8(gf * 255)
	b = uint8(bf * 255)
	return
}
