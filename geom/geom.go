// Package geom provides basic geometric objects to display data in a plot.
package geom

import (
	"image/color"

	"github.com/vdobler/facet"
	"gonum.org/v1/plot/plotter"
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

	Color  color.Color
	Radius vg.Length
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
