// Package geom provides basic geometric objects to display data in a plot.
package geom

import (
	"fmt"
	"image/color"
	"math"

	"github.com/vdobler/facet"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type StyleFunc func(i int) float64

type XYUV struct {
	X, Y, U, V float64
}

// ----------------------------------------------------------------------------
// Rectangle

// Rectangle draws rectangles.
type Rectangle struct {
	XYUV              []XYUV
	Color, Fill, Line StyleFunc
}

// Draw implements facet.Geom.Draw.
func (rp Rectangle) Draw(p *facet.Panel) {
	for _, xyuv := range rp.XYUV {
		r := vg.Rectangle{
			Min: p.Map(xyuv.X, xyuv.X),
			Max: p.Map(xyuv.U, xyuv.V),
		}
		p.Canvas.SetColor(color.RGBA{0xff, 0x77, 0x77, 0xff})
		p.Canvas.Fill(r.Path())
		p.Canvas.SetColor(color.RGBA{0xff, 0x22, 0x22, 0xff})
		p.Canvas.Stroke(r.Path())
	}
}

func (rp Rectangle) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for i, xyuv := range rp.XYUV {
		dr[facet.XScale].Update(xyuv.X)
		dr[facet.YScale].Update(xyuv.Y)
		dr[facet.XScale].Update(xyuv.U)
		dr[facet.YScale].Update(xyuv.V)

		if rp.Fill != nil {
			x := rp.Fill(i)
			dr[facet.FillScale].Update(x)
		}
		// Same for Color and Line...
	}

	return dr
}

// ----------------------------------------------------------------------------
// Point

// Point draws circular points.
type Point struct {
	XY     plotter.XYer
	Color  StyleFunc
	Size   StyleFunc
	Symbol StyleFunc

	Default draw.GlyphStyle
}

func (p Point) Draw(panel *facet.Panel) {
	dye := p.Default.Color
	if dye == nil {
		dye = color.RGBA{0x22, 0x22, 0x22, 0xff}
	}
	colorScale := panel.Scales[facet.ColorScale]

	size := p.Default.Radius
	if size == 0 {
		size = vg.Length(4)
	}

	symbol := p.Default.Shape
	if symbol == nil {
		symbol = draw.GlyphDrawer(draw.CircleGlyph{})
	}

	for i := 0; i < p.XY.Len(); i++ {
		x, y := p.XY.XY(i)
		center := panel.Map(x, y)

		if p.Size != nil {
			val := p.Size(i)
			size = panel.Scales[facet.SizeScale].SizeMap(val)
		}

		if p.Symbol != nil {
			symbol = plotutil.Shape(int(math.Round(p.Symbol(i))))
		}

		if p.Color != nil {
			val := p.Color(i)
			if !colorScale.InRange(val) {
				fmt.Println("==> ", colorScale.Min, val, colorScale.Max)
				dye = color.RGBA{0x22, 0x22, 0x22, 0x44}
			} else {
				u := colorScale.DataToUnit(val)
				var err error
				dye, err = colorScale.ColorMap.At(u)
				if err != nil {
					panic(err)
				}
			}
		}

		sty := draw.GlyphStyle{
			Color:  dye,
			Radius: size,
			Shape:  symbol,
		}
		panel.Canvas.DrawGlyph(sty, center)
	}
}

func (p Point) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for i := 0; i < p.XY.Len(); i++ {
		fmt.Println("### Size", p.Size)

		x, y := p.XY.XY(i)
		dr[facet.XScale].Update(x)
		dr[facet.YScale].Update(y)

		if p.Color != nil {
			x := p.Color(i)
			dr[facet.ColorScale].Update(x)
		}
		if p.Size != nil {
			x := p.Size(i)
			fmt.Println("### Update size", x)
			dr[facet.SizeScale].Update(x)
		}
		if p.Symbol != nil {
			x := p.Symbol(i)
			dr[facet.SymbolScale].Update(x)
		}
	}

	return dr
}

// ----------------------------------------------------------------------------
// Lines

// Line draws
type Lines struct {
	XY    []plotter.XYer
	Color StyleFunc
	Style StyleFunc
	Size  StyleFunc

	Default draw.LineStyle
}

func (l Lines) Draw(panel *facet.Panel) {
	dye := l.Default.Color
	if dye == nil {
		dye = color.RGBA{0, 0, 0x22, 0xff}
	}
	colorScale := panel.Scales[facet.ColorScale]

	width := l.Default.Width
	if width == 0 {
		width = vg.Length(1)
	}

	dashes := l.Default.Dashes

	canvas := panel.Canvas
	for g, xy := range l.XY {
		ps := make([]vg.Point, xy.Len())
		for i := 0; i < xy.Len(); i++ {
			x, y := xy.XY(i)
			ps[i] = panel.Map(x, y)
		}

		if l.Color != nil {
			val := l.Color(g)
			dye = colorScale.MapColor(val)
		}
		if l.Style != nil {
			val := l.Style(g)
			dashes = plotutil.Dashes(int(math.Round(val)))
		}
		if l.Size != nil {
			val := l.Size(g)
			width = vg.Length(val)
		}

		sty := draw.LineStyle{
			Color:  dye,
			Width:  width,
			Dashes: dashes,
		}

		canvas.StrokeLines(sty, canvas.ClipLinesXY(ps)...)
	}
}

func (l Lines) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for g, xy := range l.XY {
		for i := 0; i < xy.Len(); i++ {
			x, y := xy.XY(i)
			dr[facet.XScale].Update(x)
			dr[facet.YScale].Update(y)

			if l.Color != nil {
				x := l.Color(g)
				dr[facet.ColorScale].Update(x)
			}
			if l.Style != nil {
				x := l.Style(g)
				dr[facet.StyleScale].Update(x)
			}
		}
	}

	return dr
}

// ----------------------------------------------------------------------------
// LinesPoints

// LinesPoints draws Points connected by lines
type LinesPoints struct {
	XY    []plotter.XYer
	Color StyleFunc
	Style StyleFunc

	Size   StyleFunc // of Points
	Symbol StyleFunc // of Points

	LineDefault  draw.LineStyle
	PointDefault draw.GlyphStyle
}

func (lp LinesPoints) Draw(panel *facet.Panel) {
	lines := Lines{
		XY:      lp.XY,
		Default: lp.LineDefault,
	}
	if lp.Color != nil {
		lines.Color = func(i int) float64 { return lp.Color(i) }
	}
	if lp.Style != nil {
		lines.Style = func(i int) float64 { return lp.Style(i) }
	}
	lines.Draw(panel)

	for g, xy := range lp.XY {
		points := Point{
			XY:      xy,
			Default: lp.PointDefault,
		}
		if lp.Color != nil {
			points.Color = func(i int) float64 { return lp.Color(g) }
		}
		if lp.Symbol != nil {
			points.Symbol = func(i int) float64 { return lp.Symbol(g) }
		}
		if lp.Size != nil {
			points.Size = func(i int) float64 { return lp.Size(g) }
		}
		points.Draw(panel)
	}
}

func (lp LinesPoints) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for g, xy := range lp.XY {
		for i := 0; i < xy.Len(); i++ {
			x, y := xy.XY(i)
			dr[facet.XScale].Update(x)
			dr[facet.YScale].Update(y)

			if lp.Color != nil {
				dr[facet.ColorScale].Update(lp.Color(g))
			}
			if lp.Style != nil {
				dr[facet.StyleScale].Update(lp.Style(g))
			}
			if lp.Size != nil {
				dr[facet.SizeScale].Update(lp.Size(g))
			}
			if lp.Symbol != nil {
				dr[facet.SymbolScale].Update(lp.Symbol(g))
			}
		}
	}

	return dr
}
