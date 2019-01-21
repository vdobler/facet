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

// BoxStyle combines a line style (for the border with a fill color for
// the interior of a geom.  TODO: add the rest like size, symbol too?
type BoxStyle struct {
	Fill   color.Color
	Border draw.LineStyle
}

// CanonicRectangle returns the canonical form of r, i.e. its Min points
// having smaller coordinates than its Max point.
func CanonicRectangle(r vg.Rectangle) vg.Rectangle {
	if r.Min.X > r.Max.X {
		r.Min.X, r.Max.X = r.Max.X, r.Min.X
	}
	if r.Min.Y > r.Max.Y {
		r.Min.Y, r.Max.Y = r.Max.Y, r.Min.Y
	}
	return r
}

// ----------------------------------------------------------------------------
// Rectangle

// Rectangle draws rectangles.
// The coordinates are the outside coordinates, i.e. if the border is drawn for
// the rectangle then this border is drawn inside the rectangle given by the
// coordinates.
type Rectangle struct {
	XYUV        data.XYUVer
	Color, Fill Aesthetic
	Size, Alpha Aesthetic
	Style       DiscreteAesthetic

	Default BoxStyle
}

// Draw implements facet.Geom.Draw.
func (r Rectangle) Draw(panel *facet.Panel) {
	fill := r.Default.Fill
	border := r.Default.Border
	if fill == nil && border.Color == nil {
		border.Color = color.RGBA{0, 0, 0x10, 0xff}
		border.Width = 2 // TODO ??
	}

	for i := 0; i < r.XYUV.Len(); i++ {
		x, y, u, v := r.XYUV.XYUV(i)
		rect := vg.Rectangle{Min: panel.Map(x, y), Max: panel.Map(u, v)}
		rect = CanonicRectangle(rect)
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

		if border.Color != nil && border.Width > 0 {
			w := 0.4999 * border.Width
			rect.Min.X += w
			rect.Min.Y += w
			rect.Max.X -= w
			rect.Max.Y -= w
			panel.Canvas.SetColor(border.Color)
			panel.Canvas.SetLineWidth(border.Width)
			panel.Canvas.SetLineDash(border.Dashes, border.DashOffs)
			panel.Canvas.Stroke(rect.Path())
		}
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
	XY    plotter.XYer
	Fill  Aesthetic
	Color Aesthetic
	Size  Aesthetic
	Style DiscreteAesthetic

	Position string  // "stack" (default), "dogde" or "fill"
	GGap     float64 // Gap between groups as fraction of sample distance.
	BGap     float64 // Gap inside a group as fraction of sample distance.

	Default BoxStyle
}

// Draw implements facet.Geom.Draw.
func (b Bar) Draw(p *facet.Panel) {
	rect := b.rects()
	rect.Default = b.Default
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
	XYUV := make(data.XYUVs, b.XY.Len())

	g := b.groups()

	/*
		minDelta := g.minDelta()
		halfBarWidth := (1 - b.Gap) * minDelta / 2
		if b.Position == "dodge" {
			maxGroupSize = g.maxGroupSize()
			halfBarWidth /= float64(maxGroupSize)
		}
	*/
	for _, x := range g.Xs() {
		is := g.Group[x] // indices of all bars to draw at x
		switch b.Position {
		case "stack", "fill":
			ymin, ymax := 0.0, 0.0
			Y, V := 0.0, 0.0
			for _, i := range is {
				center, halfwidth := g.Width(x, i)
				_, y := b.XY.XY(i)
				if y < 0 {
					Y, V = ymin, ymin+y
					ymin += y

				} else {
					Y, V = ymax, ymax+y
					ymax += y
				}
				XYUV[i].X, XYUV[i].Y = center-halfwidth, Y
				XYUV[i].U, XYUV[i].V = center+halfwidth, V
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
			/*
				n := len(is)
				x -= float64(n) * halfBarWidth
				barWidth := 2 * halfBarWidth
			*/
			for _, i := range is {
				center, halfwidth := g.Width(x, i)
				_, y := b.XY.XY(i)
				XYUV[i].X, XYUV[i].Y = center-halfwidth, 0
				XYUV[i].U, XYUV[i].V = center+halfwidth, y
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

func (b Bar) groups() *BarGroups {
	g := NewBarGroups(b.Position, b.GGap, b.BGap, true)
	for i := 0; i < b.XY.Len(); i++ {
		x, _ := b.XY.XY(i)
		g.Record(x, i)
	}
	return g
}

// ----------------------------------------------------------------------------
// BarGroups helps determing bar sizes for Bar or Boxplots

type BarGroups struct {
	Group    map[float64][]int
	Position string  // "dodge" or something else
	Ggap     float64 // between groups
	Dgap     float64 // between bars inside a group if dodged
	Same     bool    // Same width for all bars?

	xs []float64
	md float64
	lg int
}

// NewBarGroups creates a BarGroups for dodged bar positioning with
// sensible gaps between bars.
func NewBarGroups(position string, groupGap, barGap float64, sameWidth bool) *BarGroups {
	if groupGap == 0 {
		groupGap = 0.2
	}
	return &BarGroups{
		Group:    make(map[float64][]int),
		Position: position,
		Ggap:     groupGap,
		Dgap:     barGap,
		Same:     sameWidth,
	}
}

// Record the point i with the given x coordinate.
func (bg *BarGroups) Record(x float64, i int) {
	bg.Group[x] = append(bg.Group[x], i)
	bg.xs = nil
}

// Bar returns the center and the halfwidth for the bar i at x.
func (bg *BarGroups) Width(x float64, i int) (center float64, halfwidth float64) {
	minDelta := bg.MinDelta()
	nonGapWidth := minDelta * (1 - bg.Ggap)

	if bg.Position != "dodge" {
		return x, nonGapWidth / 2
	}

	n := len(bg.Group[x])
	if bg.Same {
		n = bg.MaxGroupSize()
	}
	if n == 0 {
		panic(fmt.Sprintf("No data at %g", x))
	}
	halfwidth = nonGapWidth / float64(2*n)

	g := -1
	for j, k := range bg.Group[x] {
		if k == i {
			g = j
			break
		}
	}
	if g == -1 {
		panic(fmt.Sprintf("No point %d at %g", i, x))
	}

	center = x
	m := len(bg.Group[x])
	center += float64(2*g-m+1) * halfwidth

	halfwidth -= minDelta * bg.Dgap

	return center, halfwidth
}

// Xs returns the sorted list of recorded x values.
func (bg *BarGroups) Xs() []float64 {
	bg.recalc()
	return bg.xs
}

// MinDelta returns the smallest difference between recorded x-values.
func (bg *BarGroups) MinDelta() float64 {
	bg.recalc()
	return bg.md
}

// MaxGroupSize determines the maximum number of values recorded per x-values.
func (bg *BarGroups) MaxGroupSize() int {
	bg.recalc()
	return bg.lg
}

func (bg *BarGroups) XRange() (xmin float64, xmax float64) {
	bg.recalc()

	left, right := bg.xs[0], bg.xs[len(bg.xs)-1]

	li := bg.Group[left][0]
	c, hw := bg.Width(left, li)
	xmin = c - hw

	rg := bg.Group[right]
	ri := rg[len(rg)-1]
	c, hw = bg.Width(right, ri)
	xmax = c + hw

	return xmin, xmax
}

func (bg *BarGroups) recalc() {
	if bg.xs != nil {
		return
	}

	// xs: all x-valuses in sorted order
	bg.xs = make([]float64, 0, len(bg.Group))
	for x := range bg.Group {
		bg.xs = append(bg.xs, x)
	}
	sort.Float64s(bg.xs)

	// md: minumum distance between two x-valuse
	if len(bg.Group) == 0 {
		bg.md = 0
	}
	if len(bg.Group) == 1 {
		bg.md = 1
	}
	bg.md = bg.xs[1] - bg.xs[0]
	for i := 2; i < len(bg.xs); i++ {
		if m := bg.xs[i] - bg.xs[i-1]; m < bg.md {
			bg.md = m
		}
	}

	// lg: largest groups size
	bg.lg = 0
	for _, is := range bg.Group {
		if len(is) > bg.lg {
			bg.lg = len(is)
		}
	}
}

// ----------------------------------------------------------------------------
// Point

// Point draws points / symbols.
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

// ----------------------------------------------------------------------------
// Boxplot

// Boxplot draws rectangles.
// The coordinates are the outside coordinates, i.e. if the border is drawn for
// the rectangle then this border is drawn inside the rectangle given by the
// coordinates.
type Boxplot struct {
	Boxplot     data.Boxplotter
	Color, Fill Aesthetic
	Size, Alpha Aesthetic
	Stroke      DiscreteAesthetic
	Shape       DiscreteAesthetic

	Position     string
	Default      BoxStyle
	DefaultPoint draw.GlyphStyle
	GGap, BGap   float64
}

// Draw implements facet.Geom.Draw.
func (b Boxplot) Draw(panel *facet.Panel) {
	// A Boxplot is drawn by:
	//     - Rectangle in XYUV: One per data point.
	//     - Lines in XY: Three per data point
	//     - Points in XYZ: arbitrary many per data point
	XYUV := make(data.XYUVs, b.Boxplot.Len())
	XY := make([]plotter.XYer, 3*b.Boxplot.Len())
	for j := range XY {
		XY[j] = make(plotter.XYs, 2)
	}
	XYZ := plotter.XYZs{}

	g := NewBarGroups(b.Position, b.GGap, b.BGap, true)
	for i := 0; i < b.Boxplot.Len(); i++ {
		x, _, _, _, _, _, _ := b.Boxplot.Boxplot(i)
		g.Record(x, i)
	}

	for i := 0; i < b.Boxplot.Len(); i++ {
		x, min, q1, median, q3, max, out := b.Boxplot.Boxplot(i)
		// TODO: box width and dodging

		// The box.
		center, halfwidth := g.Width(x, i)
		xmin, xmax := center-halfwidth, center+halfwidth
		XYUV[i].X, XYUV[i].U = xmin, xmax
		XYUV[i].Y, XYUV[i].V = q1, q3

		// The lines
		hor := make(plotter.XYs, 2)
		hor[0].X, hor[0].Y = xmin, median
		hor[1].X, hor[1].Y = xmax, median
		XY[i*3] = hor
		vert1 := make(plotter.XYs, 2)
		vert1[0].X, vert1[0].Y = center, q3
		vert1[1].X, vert1[1].Y = center, max
		XY[i*3+1] = vert1
		vert2 := make(plotter.XYs, 2)
		vert2[0].X, vert2[0].Y = center, q1
		vert2[1].X, vert2[1].Y = center, min
		XY[i*3+2] = vert2

		// The outliers
		for _, o := range out {
			z := 0.0
			if b.Color != nil {
				z = b.Color(i)
			}
			XYZ = append(XYZ, struct{ X, Y, Z float64 }{center, o, z})
		}
	}
	rect := Rectangle{
		XYUV:  XYUV,
		Color: b.Color,
		Fill:  b.Fill,
		Size:  b.Size,
		Style: b.Stroke,

		Default: b.Default,
	}
	line := Lines{
		XY: XY,

		Default: b.Default.Border,
	}
	point := Point{
		XY:      plotter.XYValues{XYZ},
		Default: b.DefaultPoint,
	}
	if b.Color != nil {
		line.Color = func(i int) float64 { return b.Color(i / 3) }
		point.Color = func(i int) float64 { return XYZ[i].Z }
	} // TODO: same for Size and Stroke

	rect.Draw(panel)
	line.Draw(panel)
	point.Draw(panel)
}

func (b Boxplot) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	g := NewBarGroups(b.Position, b.GGap, b.BGap, true)

	for i := 0; i < b.Boxplot.Len(); i++ {
		x, min, _, _, _, max, out := b.Boxplot.Boxplot(i)
		g.Record(x, i)
		dr[facet.XScale].Update(x)
		dr[facet.YScale].Update(min, max)
		dr[facet.YScale].Update(out...)
	}
	xmin, xmax := g.XRange()
	dr[facet.XScale].Update(xmin, xmax)

	UpdateAestheticsRanges(&dr, b.Boxplot.Len(), b.Fill, b.Color, b.Size, b.Stroke, nil)
	return dr
}
