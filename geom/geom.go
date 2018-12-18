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
	XY   plotter.XYer
	Fill StyleFunc
	Size StyleFunc

	Default draw.GlyphStyle
	Color   color.Color
	Radius  vg.Length
}

func (p Point) Draw(panel *facet.Panel) {
	fill := p.Color
	if fill == nil {
		fill = color.RGBA{0x22, 0x22, 0x22, 0xff}
	}
	fillScale := panel.Scales[facet.FillScale]
	size := p.Radius
	if size == 0 {
		size = vg.Length(5)
	}

	for i := 0; i < p.XY.Len(); i++ {
		x, y := p.XY.XY(i)
		center := panel.Map(x, y)

		if p.Size != nil {
			val := p.Size(i)
			u := panel.Scales[facet.SizeScale].DataToUnit(val)
			size = vg.Length(u*10 + 1)
		}

		if p.Fill != nil {
			val := p.Fill(i)
			if !fillScale.InRange(val) {
				// TODO
				fill = color.RGBA{0x22, 0x22, 0x22, 0x44}
			} else {
				u := fillScale.DataToUnit(val)
				var err error
				fill, err = fillScale.ColorMap.At(u)
				if err != nil {
					panic(err)
				}
			}
		}

		sty := draw.GlyphStyle{
			Color:  fill,
			Radius: size,
			Shape:  draw.CircleGlyph{},
		}
		panel.Canvas.DrawGlyph(sty, center)
	}
}

func (p Point) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for i := 0; i < p.XY.Len(); i++ {
		x, y := p.XY.XY(i)
		dr[facet.XScale].Update(x)
		dr[facet.YScale].Update(y)

		if p.Fill != nil {
			x := p.Fill(i)
			dr[facet.FillScale].Update(x)
		}
		if p.Size != nil {
			x := p.Size(i)
			dr[facet.SizeScale].Update(x)
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
		width = vg.Length(2)
	}

	dashes := l.Default.Dashes

	canvas := panel.Canvas
	for g, xy := range l.XY {
		fmt.Println("Lines: drawing group", g)
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
