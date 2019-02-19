package facet

import (
	"image/color"

	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// ----------------------------------------------------------------------------
// Panel

// A Panel represents one panel in a faceted plot.
type Panel struct {
	Title  string
	Plot   *Plot
	Geoms  []Geom
	Canvas draw.Canvas
	Scales [numScales]*Scale
}

func (p *Panel) InRangeXY(x, y float64) bool {
	return p.Scales[XScale].InRange(x) && p.Scales[YScale].InRange(y)
}

// MapXY maps the data coordinate (x,y) to a canvas point.
func (p *Panel) MapXY(x, y float64) vg.Point {
	xs, ys := p.Scales[XScale], p.Scales[YScale]
	cx := Interval{float64(p.Canvas.Min.X), float64(p.Canvas.Max.X)}
	cy := Interval{float64(p.Canvas.Min.Y), float64(p.Canvas.Max.Y)}
	xu := xs.Trans.Trans(xs.Range, cx, x)
	yu := ys.Trans.Trans(ys.Range, cy, y)
	return vg.Point{X: vg.Length(xu), Y: vg.Length(yu)}
}

// MapSize maps a data value v to a display size by calling p.Plot.MapSize.
func (p *Panel) MapSize(v float64) vg.Length {
	return p.Plot.MapSize(v)
}

// MapColor maps a data value v to a color by calling p.Plot.MapColor(v,false).
func (p *Panel) MapColor(v float64) color.Color {
	return p.Plot.MapColor(v, false)
}

// MapFill maps a data value v to a color by calling p.Plot.MapColor(v,true).
func (p *Panel) MapFill(v float64) color.Color {
	return p.Plot.MapColor(v, true)
}
