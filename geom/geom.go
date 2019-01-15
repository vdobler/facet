// Package geom provides basic geometric objects to display data in a plot.
//
// The overall concept is loosely based in ggplot2's geoms. Each geom has
// some required aesthetics, typically an (x,y) coordinate and may provide
// the ability to optionally map other aestetics like line or fill color
// or size.
//
// The required aestethics are a field like XY in the various geoms while the
// optional aestehtics are mapped through optional (Discrete)Aesthetics
// functions which provide a (discrete) value for a data point.
//
// The different geoms have singular names like Rectangle or Point even if
// they may draw several rectangles or points to match the naming in ggplot2.
package geom

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/vdobler/facet"
	"github.com/vdobler/facet/data"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// Aestetic is a function mappinge a certain data point to an aestehtic.
type Aesthetic func(i int) float64

// DiscreteAestetic is a function mappinge a certain data point to a discrete
// aesthetic like Symbol or Style.
type DiscreteAesthetic func(i int) int

// UpdateAestheticsRanges is a helper to update the data ranges dr based on
// the non-nil aesthetics functions evaluated for all n data points.
func UpdateAestheticsRanges(dr *facet.DataRanges, n int,
	fill, color, size Aesthetic,
	style, symbol DiscreteAesthetic) {

	for i := 0; i < n; i++ {
		if fill != nil {
			dr[facet.FillScale].Update(fill(i))
		}
		if color != nil {
			dr[facet.ColorScale].Update(color(i))
		}
		if size != nil {
			dr[facet.SizeScale].Update(size(i))
		}
		if style != nil {
			dr[facet.StyleScale].Update(float64(style(i))) // TODO: StyleScale should be discrete from the start
		}
		if symbol != nil {
			dr[facet.SymbolScale].Update(float64(symbol(i)))
		}
	}
}

// ----------------------------------------------------------------------------
// Rectangle

// Rectangle draws rectangles.
type Rectangle struct {
	XYUV              data.XYUVer
	Color, Fill, Size Aesthetic
	Style             DiscreteAesthetic

	Default struct {
		Fill   color.Color
		Border draw.LineStyle
	}
}

// Draw implements facet.Geom.Draw.
func (r Rectangle) Draw(panel *facet.Panel) {
	var fill color.Color = color.RGBA{0, 0, 0x10, 0xff}
	if r.Default.Fill != nil {
		fill = r.Default.Fill
	}
	border := r.Default.Border

	for i := 0; i < r.XYUV.Len(); i++ {
		x, y, u, v := r.XYUV.XYUV(i)
		rect := vg.Rectangle{Min: panel.Map(x, y), Max: panel.Map(u, v)}
		if r.Fill != nil {
			fill = panel.Scales[facet.FillScale].MapColor(r.Fill(i))
		}
		panel.Canvas.SetColor(fill)
		panel.Canvas.Fill(rect.Path())

		if r.Color != nil {
			border.Color = panel.Scales[facet.ColorScale].MapColor(r.Color(i))
		}
		if r.Size != nil {
			border.Width = panel.Scales[facet.SizeScale].SizeMap(r.Size(i))
		}
		// TODO: Style

		panel.Canvas.SetColor(border.Color)
		panel.Canvas.SetLineWidth(border.Width)
		panel.Canvas.SetLineDash(border.Dashes, border.DashOffs)
		panel.Canvas.Stroke(rect.Path())
	}
}

func (r Rectangle) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	xmin, xmax, ymin, ymax, umin, umax, vmin, vmax := data.XYUVRange(r.XYUV)
	dr[facet.XScale].Update(xmin, xmax, umin, umax)
	dr[facet.YScale].Update(ymin, ymax, vmin, vmax)
	UpdateAestheticsRanges(&dr, r.XYUV.Len(), r.Fill, r.Color, r.Size, r.Style, nil)
	return dr
}

// ----------------------------------------------------------------------------
// Bar

// Bar draws rectangles standing/hanging from y=0.
type Bar struct {
	XY       plotter.XYer
	Fill     Aesthetic
	Color    Aesthetic
	Size     Aesthetic
	Style    DiscreteAesthetic
	Position string  // "stack" (default), "dogde" or "fill"
	Gap      float64 // Gap between bars as fraction of bar width.
	// TODO: Spacing in a group of dodged bars.
}

// Draw implements facet.Geom.Draw.
func (b Bar) Draw(p *facet.Panel) {
	rect := b.rects()
	rect.Draw(p)
}

func (b Bar) AllDataRanges() facet.DataRanges {
	rect := b.rects()
	return rect.AllDataRanges()
}

func (b Bar) rects() Rectangle {
	if b.Position == "" {
		b.Position = "stack"
	}
	if b.Gap == 0 {
		b.Gap = 0.2
	}

	XYUV := make(data.XYUVs, b.XY.Len())

	g := b.groups()
	minDelta := g.minDelta()
	halfBarWidth := (1 - b.Gap) * minDelta / 2
	if b.Position == "dodge" {
		maxGroupSize := 0
		for _, is := range g {
			if len(is) > maxGroupSize {
				maxGroupSize = len(is)
			}
		}
		halfBarWidth /= float64(maxGroupSize)
	}

	for _, x := range g.xs() {
		is := g[x] // indices of all bars to draw at x
		switch b.Position {
		case "stack", "fill":
			X, Y := x-halfBarWidth, 0.0
			U, V := x+halfBarWidth, 0.0
			ymin, ymax := 0.0, 0.0
			for _, i := range is {
				_, y := b.XY.XY(i)
				if y < 0 {
					Y, V = ymin, ymin+y
					ymin += y

				} else {
					Y, V = ymax, ymax+y
					ymax += y
				}
				XYUV[i].X, XYUV[i].Y = X, Y
				XYUV[i].U, XYUV[i].V = U, V
			}
			if b.Position == "fill" {
				ymin *= -1
				for _, i := range is {
					if XYUV[i].V < 0 {
						XYUV[i].Y /= ymin
						XYUV[i].V /= ymin
					} else {
						XYUV[i].Y /= ymax
						XYUV[i].V /= ymax
					}
				}
			}
		case "dodge":
			n := len(is)
			x -= float64(n) * halfBarWidth
			barWidth := 2 * halfBarWidth
			for _, i := range is {
				_, y := b.XY.XY(i)
				XYUV[i].X, XYUV[i].Y = x, 0
				XYUV[i].U, XYUV[i].V = x+barWidth, y
				x += 2 * halfBarWidth
			}
		default:
			panic("geom.Bar: unknown value for Position: " + b.Position)
		}
	}

	return Rectangle{
		XYUV:  XYUV,
		Fill:  b.Fill,
		Color: b.Color,
		Size:  b.Size,
		Style: b.Style,
	}

}

func (b Bar) groups() barGroups {
	g := make(barGroups, b.XY.Len())
	for i := 0; i < b.XY.Len(); i++ {
		x, _ := b.XY.XY(i)
		g.add(x, i)
	}
	return g
}

type barGroups map[float64][]int

func (bg barGroups) add(x float64, i int) {
	bg[x] = append(bg[x], i)
}

func (bg barGroups) xs() []float64 {
	xs := make([]float64, 0, len(bg))
	for x := range bg {
		xs = append(xs, x)
	}
	sort.Float64s(xs)
	return xs
}

func (bg barGroups) minDelta() float64 {
	if len(bg) == 0 {
		return 0
	}
	if len(bg) == 1 {
		return 1
	}
	xs := bg.xs()
	min := xs[1] - xs[0]
	for i := 2; i < len(xs); i++ {
		if m := xs[i] - xs[i-1]; m < min {
			min = m
		}
	}
	return min
}

// ----------------------------------------------------------------------------
// Point

// Point draws circular points.
type Point struct {
	XY     plotter.XYer
	Color  Aesthetic
	Size   Aesthetic
	Symbol DiscreteAesthetic

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
			symbol = plotutil.Shape(p.Symbol(i))
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
	xmin, xmax, ymin, ymax := plotter.XYRange(p.XY)
	dr[facet.XScale].Update(xmin)
	dr[facet.XScale].Update(xmax)
	dr[facet.YScale].Update(ymin)
	dr[facet.YScale].Update(ymax)

	UpdateAestheticsRanges(&dr, p.XY.Len(), nil, p.Color, p.Size, nil, p.Symbol)

	return dr
}

// ----------------------------------------------------------------------------
// Lines

// Line draws
type Lines struct {
	XY    []plotter.XYer
	Color Aesthetic
	Size  Aesthetic
	Style DiscreteAesthetic

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
			dye = colorScale.MapColor(l.Color(g))
		}
		if l.Style != nil {
			dashes = plotutil.Dashes(l.Style(g))
		}
		if l.Size != nil {
			width = vg.Length(l.Size(g)) // TODO: Proper mapping!!
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
	for _, xy := range l.XY {
		xmin, xmax, ymin, ymax := plotter.XYRange(xy)
		dr[facet.XScale].Update(xmin)
		dr[facet.XScale].Update(xmax)
		dr[facet.YScale].Update(ymin)
		dr[facet.YScale].Update(ymax)
		UpdateAestheticsRanges(&dr, xy.Len(), nil, l.Color, l.Size, l.Style, nil)
	}
	return dr
}

// ----------------------------------------------------------------------------
// LinesPoints

// LinesPoints draws Points connected by lines
type LinesPoints struct {
	XY    []plotter.XYer
	Color Aesthetic
	Style DiscreteAesthetic

	Size   Aesthetic         // of Points
	Symbol DiscreteAesthetic // of Points

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
		lines.Style = func(i int) int { return lp.Style(i) }
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
		if lp.Size != nil {
			points.Size = func(i int) float64 { return lp.Size(g) }
		}
		if lp.Symbol != nil {
			points.Symbol = func(i int) int { return lp.Symbol(g) }
		}
		points.Draw(panel)
	}
}

func (lp LinesPoints) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for _, xy := range lp.XY {
		for i := 0; i < xy.Len(); i++ {
			x, y := xy.XY(i)
			dr[facet.XScale].Update(x)
			dr[facet.YScale].Update(y)
		}
		UpdateAestheticsRanges(&dr, xy.Len(), nil, lp.Color, lp.Size, lp.Style, lp.Symbol)
	}

	return dr
}
