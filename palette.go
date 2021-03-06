package facet

import (
	"image/color"
	"math"

	"gonum.org/v1/plot/palette"
)

var DefaultColorMap = &Rainbow{
	Value:      0.9,
	Saturation: 0.9,
	HueGap:     1.0 / 6.0,
	min:        0,
	max:        1,
	alpha:      1,
}

// Rainbow is a equaly spaced hue rainbow color map.
type Rainbow struct {
	Value      float64 // Value of the generated colors
	Saturation float64 // Saturation of the generated colors.
	StartHue   float64 // StartHue is the hue used for the Min value.
	HueGap     float64 // HueGap determines the fraction of the hue space which is not used.
	Number     int     // Number of colors generated by the Colors method.

	min, max, alpha float64
}

// At returns the color mapped for x.
func (r *Rainbow) At(x float64) (color.Color, error) {
	h := r.StartHue + (1-r.HueGap)*(x-r.min)/(r.max-r.min)
	if h > 1 {
		h = h - math.Trunc(h)
	}
	c := palette.HSVA{
		H: h,
		S: r.Saturation,
		V: r.Value,
		A: r.alpha,
	}
	return c, nil
}

// Max returns the current maximum value of the ColorMap.
func (r *Rainbow) Max() float64 {
	return r.max
}

// SetMax sets the maximum value of the ColorMap.
func (r *Rainbow) SetMax(max float64) {
	r.max = max
}

// Min returns the current minimum value of the ColorMap.
func (r *Rainbow) Min() float64 {
	return r.min
}

// SetMin sets the minimum value of the ColorMap.
func (r *Rainbow) SetMin(min float64) {
	r.min = min
}

// Alpha returns the opacity value of the ColorMap.
func (r *Rainbow) Alpha() float64 {
	return r.alpha
}

// SetAlpha sets the opacity value of the ColorMap. Zero is transparent
// and one is completely opaque. The default value of alpha should be
// expected to be one. The function should be expected to panic
// if alpha is not between zero and one.
func (r *Rainbow) SetAlpha(alpha float64) {
	if alpha < 0 || alpha > 1 {
		panic(alpha)
	}
	r.alpha = alpha
}

// Palette records the number of colors and retunrs itself as a
// palettte.Palette so that subsequent calls to Color yield a palette
// with colors many colors.
func (r *Rainbow) Palette(colors int) palette.Palette {
	r.Number = colors
	return r
}

// Colors implements palette.Palette.Colors.
func (r *Rainbow) Colors() []color.Color {
	colors := make([]color.Color, r.Number)
	for i := range colors {
		colors[i] = palette.HSVA{
			H: float64(i) / float64(r.Number+1),
			S: r.Saturation,
			V: r.Value,
			A: r.alpha,
		}
	}
	return colors
}
