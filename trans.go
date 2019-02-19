// Scale Transformations
//
// Scale transformations should work like the ones in ggplot2.
package facet

import (
	"math"

	"gonum.org/v1/plot"
)

// A Transformation bundles two functions Trans and Inverse together with
// an appropiate Ticker. The two functions map two intervals.
type Transformation struct {
	Name    string
	Trans   func(from, to Interval, x float64) float64
	Inverse func(from, to Interval, y float64) float64
	Ticker  plot.Ticker
}

// IdentityTrans does not transform at all.
var IdentityTrans = Transformation{
	Name:    "Identity",
	Trans:   func(from, to Interval, x float64) float64 { return x },
	Inverse: func(from, to Interval, y float64) float64 { return y },
	Ticker:  DefaultTicks(4),
}

// LinearTrans implements a linear mapping of from to to.
var LinearTrans = Transformation{
	Name: "Linear",
	Trans: func(from, to Interval, x float64) float64 {
		return to.Min + (to.Max-to.Min)*(x-from.Min)/(from.Max-from.Min)
	},
	Inverse: func(from, to Interval, y float64) float64 {
		return to.Min + (to.Max-to.Min)*(y-from.Min)/(from.Max-from.Min)
	},
	Ticker: DefaultTicks(4),
}

// SqrtTrans implements a square root transformation suitable to map
// the Size aesthetic to the size of a point. (Ggplots scale_size)
var SqrtTrans = Transformation{
	Name: "SquareRoot",
	Trans: func(from, to Interval, x float64) float64 {
		area := Interval{to.Min * to.Min, to.Max * to.Max}
		return math.Sqrt(LinearTrans.Trans(from, area, x))
	},
	Inverse: func(from, to Interval, y float64) float64 {
		area := Interval{from.Min * from.Min, from.Max * from.Max}
		return LinearTrans.Trans(area, to, y*y)
	},
	Ticker: DefaultTicks(5),
}

// SqrtTransFix0 implements a square root transformation suitable to map
// the Size aesthetic to the area of a point. It maps 0 to 0.
// (Ggplot's scale_size_are)
var SqrtTransFix0 = Transformation{
	Name: "SquareRoot",
	Trans: func(from, to Interval, x float64) float64 {
		from.Min, to.Min = 0, 0
		area := Interval{to.Min * to.Min, to.Max * to.Max}
		return math.Sqrt(LinearTrans.Trans(from, area, x))
	},
	Inverse: func(from, to Interval, y float64) float64 {
		from.Min, to.Min = 0, 0
		area := Interval{from.Min * from.Min, from.Max * from.Max}
		return LinearTrans.Trans(area, to, y*y)
	},
	Ticker: DefaultTicks(5),
}

var Log10Trans = Transformation{
	Name: "Log10",
	Trans: func(from, to Interval, x float64) float64 {
		t := math.Log10(x/from.Min) / math.Log10(from.Max/from.Min)
		y := to.Min + t*(to.Max-to.Min)
		return y
	},
	Inverse: func(from, to Interval, y float64) float64 {
		return to.Min * math.Pow(10, math.Log10(to.Max/to.Min)*(y-from.Min)/(from.Max-from.Min))
	},
	Ticker: plot.LogTicks{},
}
