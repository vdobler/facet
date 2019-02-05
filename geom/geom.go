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

// ----------------------------------------------------------------------------
// Point

// Point draws points / symbols.
type Point struct {
	XY plotter.XYer

	Alpha Aesthetic
	Color Aesthetic
	Shape DiscreteAesthetic
	Size  Aesthetic

	Default draw.GlyphStyle
}

func (p Point) Draw(panel *facet.Panel) {
	baseColor := p.Default.Color
	if baseColor == nil {
		baseColor = panel.Plot.Style.GeomDefault.Color
	}

	size := p.Default.Radius
	if size == 0 {
		size = panel.Plot.Style.GeomDefault.Size
	}

	shape := p.Default.Shape
	if shape == nil {
		shape = draw.GlyphDrawer(draw.CircleGlyph{})
	}

	for i := 0; i < p.XY.Len(); i++ {
		x, y := p.XY.XY(i)
		center, ok := panel.MapXY(x, y)
		if !ok {
			continue // TODO: should notify Plot/Panel about dropped data point.
		}

		col, ok := determineColor(baseColor, panel, i, p.Color, p.Alpha)
		if !ok {
			continue
		}

		if p.Shape != nil {
			shape = plotutil.Shape(p.Shape(i))
		}

		if p.Size != nil {
			size = panel.MapSize(p.Size(i))
			if size == 0 {
				continue
			}
		}

		sty := draw.GlyphStyle{
			Color:  col,
			Radius: size,
			Shape:  shape,
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

	UpdateAestheticsRanges(&dr, p.XY.Len(), p.Alpha, p.Color, nil, p.Shape, p.Size, nil)

	return dr
}

// ----------------------------------------------------------------------------
// Rectangle

// Rectangle draws rectangles.
// The coordinates are the outside coordinates, i.e. if the border is drawn for
// the rectangle then this border is drawn inside the rectangle given by the
// coordinates.
type Rectangle struct {
	XYUV data.XYUVer

	Alpha  Aesthetic
	Color  Aesthetic
	Fill   Aesthetic
	Size   Aesthetic
	Stroke DiscreteAesthetic

	Default BoxStyle
}

// clipRect clips rect to canvas. The returned rectangle is in the canonical form.
func clipRect(rect vg.Rectangle, canvas draw.Canvas) vg.Rectangle {
	rect = CanonicRectangle(rect)
	limit := CanonicRectangle(canvas.Rectangle)

	if rect.Min.X < limit.Min.X {
		rect.Min.X = limit.Min.X
	}
	if rect.Min.Y < limit.Min.Y {
		rect.Min.Y = limit.Min.Y
	}

	if rect.Max.X > limit.Max.X {
		rect.Max.X = limit.Max.X
	}
	if rect.Max.Y > limit.Max.Y {
		rect.Min.Y = limit.Min.Y
	}
	return rect
}

// Draw implements facet.Geom.Draw.
func (r Rectangle) Draw(panel *facet.Panel) {
	fill := r.Default.Fill
	border := r.Default.Border
	if fill == nil && border.Color == nil {
		border.Color = color.RGBA{0, 0, 0x10, 0xff}
		border.Width = 2 // TODO ??
	}
	if fill == nil && border.Color == nil {
		border.Color = color.RGBA{0, 0, 0x10, 0xff}
		border.Width = 2 // TODO ??
	}

	for i := 0; i < r.XYUV.Len(); i++ {
		x, y, u, v := r.XYUV.XYUV(i)
		min, minok := panel.MapXY(x, y)
		max, maxok := panel.MapXY(u, v)
		if !minok && !maxok {
			continue // both corners outside of scale range
		}
		rect := vg.Rectangle{Min: min, Max: max}
		rect = clipRect(rect, panel.Canvas)

		if fillCol, ok := determineColor(fill, panel, i, r.Fill, r.Alpha); ok {
			panel.Canvas.SetColor(fillCol)
			panel.Canvas.Fill(rect.Path())
		}
		if r.Size != nil {
			border.Width = panel.MapSize(r.Size(i))
		}
		if border.Width <= 0 {
			continue
		}

		if borderCol, ok := determineColor(border.Color, panel, i, r.Color, r.Alpha); ok {
			w := 0.499 * border.Width
			rect.Min.X += w
			rect.Min.Y += w
			rect.Max.X -= w
			rect.Max.Y -= w
			panel.Canvas.SetColor(borderCol)
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
	UpdateAestheticsRanges(&dr, r.XYUV.Len(), r.Alpha, r.Color, r.Fill, nil, r.Size, r.Stroke)
	return dr
}

// ----------------------------------------------------------------------------
// Bar

// Bar draws rectangles standing/hanging from y=0.
type Bar struct {
	XY plotter.XYer

	Alpha  Aesthetic
	Color  Aesthetic
	Fill   Aesthetic
	Size   Aesthetic
	Stroke DiscreteAesthetic

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

	rect := Rectangle{XYUV: XYUV}
	CopyAesthetics(&rect, b, nil)
	return rect
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
// Path

// Path connects the given points in data order through straight line segments.
// The aestetics map the individual line segments based on their first point.
//
// (To draw them in order of x values see Line.)
type Path struct {
	XY plotter.XYer

	Alpha  Aesthetic
	Color  Aesthetic
	Size   Aesthetic
	Stroke DiscreteAesthetic

	Default draw.LineStyle
}

func (p Path) Draw(panel *facet.Panel) {
	baseColor := p.Default.Color
	if baseColor == nil {
		baseColor = panel.Plot.Style.GeomDefault.Color
	}

	width := p.Default.Width
	if width == 0 {
		width = panel.Plot.Style.GeomDefault.LineWidth
	}

	dashes := p.Default.Dashes

	canvas := panel.Canvas
	for i := 0; i < p.XY.Len()-1; i++ {
		left, _ := panel.MapXY(p.XY.XY(i))      // Clipping done below.
		right, _ := panel.MapXY(p.XY.XY(i + 1)) // Clipping done below.

		col, ok := determineColor(baseColor, panel, i, p.Color, p.Alpha)
		if !ok {
			continue // TODO: report dropping of data to Plot/Panel
		}
		if p.Stroke != nil {
			dashes = plotutil.Dashes(p.Stroke(i))
		}
		if p.Size != nil {
			width = panel.MapSize(p.Size(i))
		}

		sty := draw.LineStyle{
			Color:  col,
			Width:  width,
			Dashes: dashes,
		}

		// TODO: What if dropped completely? Report?
		canvas.StrokeLines(sty, canvas.ClipLinesXY([]vg.Point{left, right})...)
	}
}

func (p Path) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for i := 0; i < p.XY.Len(); i++ {
		x, y := p.XY.XY(i)
		dr[facet.XScale].Update(x)
		dr[facet.YScale].Update(y)
	}
	UpdateAestheticsRanges(&dr, p.XY.Len(), p.Alpha, p.Color, nil, nil, p.Size, p.Stroke)
	return dr
}

// ----------------------------------------------------------------------------
// Line

// Line connects the given points in order of the x values by straight line segments.
// The aestetics map the individual line segments based on their first point.
//
// (To draw them in data order see Path.)
type Line struct {
	XY plotter.XYer

	Alpha  Aesthetic
	Color  Aesthetic
	Size   Aesthetic
	Stroke DiscreteAesthetic

	Default draw.LineStyle
}

func (l Line) toPath() Path {
	path := Path(l)

	xy := make(plotter.XYs, l.XY.Len())
	for i := range xy {
		xy[i].X, xy[i].Y = l.XY.XY(i)
	}
	sort.Slice(xy, func(i, j int) bool { return xy[i].X < xy[j].X })
	path.XY = xy

	return path
}

func (l Line) Draw(panel *facet.Panel) {
	path := l.toPath()
	path.Draw(panel)
}

func (l Line) AllDataRanges() facet.DataRanges {
	path := Path(l) // no need to sort
	return path.AllDataRanges()
}

// ----------------------------------------------------------------------------
// Step

// Step produces a stairstep plot of the given data.
type Step struct {
	XY plotter.XYer

	Alpha  Aesthetic
	Color  Aesthetic
	Size   Aesthetic
	Stroke DiscreteAesthetic

	// Vertical changes the step to "vertical then horizontal".
	Vertical bool

	Default draw.LineStyle
}

func (s Step) toPath() Path {
	path := Path{Alpha: s.Alpha, Color: s.Color, Size: s.Size,
		Stroke: s.Stroke, Default: s.Default}

	N := s.XY.Len()
	xy := make(plotter.XYs, 2*N-1)
	for i := 0; i < N; i++ {
		xy[i].X, xy[i].Y = s.XY.XY(i)
	}
	sort.Slice(xy[:N], func(i, j int) bool { return xy[i].X < xy[j].X })

	for i := len(xy) - 1; i > 0; i -= 2 {
		xy[i] = xy[i/2]
	}

	for i := 1; i < len(xy); i += 2 {
		if s.Vertical {
			xy[i].X, xy[i].Y = xy[i-1].X, xy[i+1].Y
		} else {
			xy[i].X, xy[i].Y = xy[i+1].X, xy[i-1].Y
		}
	}

	path.XY = xy

	return path
}

func (s Step) Draw(panel *facet.Panel) {
	path := s.toPath()
	path.Draw(panel)
}

func (s Step) AllDataRanges() facet.DataRanges {
	// all additional points lie inside the range spaned by the original data points.
	path := Path{Alpha: s.Alpha, Color: s.Color, Size: s.Size,
		Stroke: s.Stroke, Default: s.Default}
	return path.AllDataRanges()
}

// ----------------------------------------------------------------------------
// Segment

// Segment draws line segments between two points (X,Y) and (U,V).
type Segment struct {
	XYUV data.XYUVer

	Alpha  Aesthetic
	Color  Aesthetic
	Size   Aesthetic
	Stroke DiscreteAesthetic

	Default draw.LineStyle
}

func (s Segment) Draw(panel *facet.Panel) {
	baseColor := s.Default.Color
	if baseColor == nil {
		baseColor = panel.Plot.Style.GeomDefault.Color
	}

	width := s.Default.Width
	if width == 0 {
		width = panel.Plot.Style.GeomDefault.LineWidth
	}

	dashes := s.Default.Dashes

	canvas := panel.Canvas
	for i := 0; i < s.XYUV.Len(); i++ {
		x, y, u, v := s.XYUV.XYUV(i)
		left, _ := panel.MapXY(x, y)  // Clipping done below.
		right, _ := panel.MapXY(u, v) // Clipping done below.

		col, ok := determineColor(baseColor, panel, i, s.Color, s.Alpha)
		if !ok {
			continue // TODO: report dropping of data to Plot/Panel
		}
		if s.Stroke != nil {
			dashes = plotutil.Dashes(s.Stroke(i))
		}
		if s.Size != nil {
			width = panel.MapSize(s.Size(i))
		}

		sty := draw.LineStyle{
			Color:  col,
			Width:  width,
			Dashes: dashes,
		}

		// TODO: What if dropped completely? Report?
		canvas.StrokeLines(sty, canvas.ClipLinesXY([]vg.Point{left, right})...)
	}
}

func (s Segment) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for i := 0; i < s.XYUV.Len(); i++ {
		x, y, u, v := s.XYUV.XYUV(i)
		dr[facet.XScale].Update(x)
		dr[facet.YScale].Update(y)
		dr[facet.XScale].Update(u)
		dr[facet.YScale].Update(v)
	}
	UpdateAestheticsRanges(&dr, s.XYUV.Len(), s.Alpha, s.Color, nil, nil, s.Size, s.Stroke)
	return dr
}

// ----------------------------------------------------------------------------
// HLine

// HLine draws horizontal reference (or rule) lines at the given Y values.
type HLine struct {
	Y plotter.Valuer

	Alpha  Aesthetic
	Color  Aesthetic
	Size   Aesthetic
	Stroke DiscreteAesthetic

	Default draw.LineStyle
}

func (h HLine) Draw(panel *facet.Panel) {
	N := h.Y.Len()
	xyuv := make(data.XYUVs, N)
	xscale := panel.Scales[facet.XScale]
	xmin, xmax := xscale.Min, xscale.Max
	for i := 0; i < N; i++ {
		y := h.Y.Value(i)
		xyuv[i].X, xyuv[i].Y, xyuv[i].U, xyuv[i].V = xmin, y, xmax, y
	}
	segment := Segment{XYUV: xyuv, Default: h.Default}
	CopyAesthetics(&segment, h, nil)
	segment.Draw(panel)
}

func (h HLine) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for i := 0; i < h.Y.Len(); i++ {
		dr[facet.YScale].Update(h.Y.Value(i))
	}
	UpdateAestheticsRanges(&dr, h.Y.Len(), h.Alpha, h.Color, nil, nil, h.Size, h.Stroke)
	return dr
}

// ----------------------------------------------------------------------------
// VLine

// VLine draws vertical reference (or rule) lines at the given X values.
type VLine struct {
	X plotter.Valuer

	Alpha  Aesthetic
	Color  Aesthetic
	Size   Aesthetic
	Stroke DiscreteAesthetic

	Default draw.LineStyle
}

func (v VLine) Draw(panel *facet.Panel) {
	N := v.X.Len()
	xyuv := make(data.XYUVs, N)
	yscale := panel.Scales[facet.YScale]
	ymin, ymax := yscale.Min, yscale.Max
	for i := 0; i < N; i++ {
		x := v.X.Value(i)
		xyuv[i].X, xyuv[i].Y, xyuv[i].U, xyuv[i].V = x, ymin, x, ymax
	}
	segment := Segment{XYUV: xyuv, Default: v.Default}
	CopyAesthetics(&segment, v, nil)
	segment.Draw(panel)
}

func (v VLine) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for i := 0; i < v.X.Len(); i++ {
		dr[facet.XScale].Update(v.X.Value(i))
	}
	UpdateAestheticsRanges(&dr, v.X.Len(), v.Alpha, v.Color, nil, nil, v.Size, v.Stroke)
	return dr
}

// ----------------------------------------------------------------------------
// Boxplot

// Boxplot draws rectangles.
// The coordinates are the outside coordinates, i.e. if the border is drawn for
// the rectangle then this border is drawn inside the rectangle given by the
// coordinates.
type Boxplot struct {
	Boxplot data.Boxplotter

	Alpha  Aesthetic
	Color  Aesthetic
	Fill   Aesthetic
	Shape  DiscreteAesthetic
	Size   Aesthetic
	Stroke DiscreteAesthetic

	Position     string
	Default      BoxStyle
	DefaultPoint draw.GlyphStyle
	GGap, BGap   float64
}

// Draw implements facet.Geom.Draw.
func (b Boxplot) Draw(panel *facet.Panel) {
	// A Boxplot is drawn by:
	//     - Rectangle in XYUV: One per data point.
	//     - Lines in Seg: Three per data point
	//     - Points in XYZ: arbitrary many per data point
	N := b.Boxplot.Len()
	XYUV := make(data.XYUVs, N)
	Seg := make(data.XYUVs, 3*N)
	XYZ := plotter.XYZs{}

	g := NewBarGroups(b.Position, b.GGap, b.BGap, true)
	for i := 0; i < N; i++ {
		x, _, _, _, _, _, _ := b.Boxplot.Boxplot(i)
		g.Record(x, i)
	}

	for i := 0; i < N; i++ {
		x, min, q1, median, q3, max, out := b.Boxplot.Boxplot(i)

		// The box.
		center, halfwidth := g.Width(x, i)
		xmin, xmax := center-halfwidth, center+halfwidth
		XYUV[i].X, XYUV[i].U = xmin, xmax
		XYUV[i].Y, XYUV[i].V = q1, q3

		// The lines
		Seg[3*i].X, Seg[3*i].Y, Seg[3*i].U, Seg[3*i].V = xmin, median, xmax, median
		Seg[3*i+1].X, Seg[3*i+1].Y, Seg[3*i+1].U, Seg[3*i+1].V = center, min, center, q1
		Seg[3*i+2].X, Seg[3*i+2].Y, Seg[3*i+2].U, Seg[3*i+2].V = center, q3, center, max

		// The outliers
		for _, o := range out {
			z := 0.0
			if b.Color != nil {
				z = b.Color(i)
			}
			XYZ = append(XYZ, struct{ X, Y, Z float64 }{center, o, z})
		}
	}
	rect := Rectangle{XYUV: XYUV, Default: b.Default}
	segment := Segment{XYUV: Seg, Default: b.Default.Border}
	point := Point{XY: plotter.XYValues{XYZ}, Default: b.DefaultPoint}
	CopyAesthetics(&rect, b, nil)
	CopyAesthetics(&segment, b, func(n int) int { return n / 3 })
	CopyAesthetics(&point, b, nil)
	if b.Color != nil {
		point.Color = func(i int) float64 { return XYZ[i].Z }
	}

	rect.Draw(panel)
	segment.Draw(panel)
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

	UpdateAestheticsRanges(&dr, b.Boxplot.Len(), b.Alpha, b.Color, b.Fill, nil, b.Size, b.Stroke)
	return dr
}

// ----------------------------------------------------------------------------
// Text

// Text draws points / symbols.
type Text struct {
	XYText data.XYTexter

	Alpha Aesthetic
	Color Aesthetic
	Size  Aesthetic

	Default draw.TextStyle
}

func (t Text) Draw(panel *facet.Panel) {
	baseColor := t.Default.Color
	if baseColor == nil {
		baseColor = panel.Plot.Style.GeomDefault.Color
	}

	font := panel.Plot.Style.XAxis.Title.Font
	if t.Default.Font != (vg.Font{}) {
		font = t.Default.Font
	}
	size := font.Size

	for i := 0; i < t.XYText.Len(); i++ {
		x, y, text := t.XYText.XYText(i)
		center, ok := panel.MapXY(x, y)
		if !ok {
			continue // TODO: should notify Plot/Panel about dropped data point.
		}

		col, ok := determineColor(baseColor, panel, i, t.Color, t.Alpha)
		if !ok {
			continue
		}

		if t.Size != nil {
			size = panel.MapSize(t.Size(i))
			if size == 0 {
				continue
			}
		}

		sty := t.Default
		sty.Color = col
		sty.Font = font
		sty.Font.Size = 2 * size
		panel.Canvas.FillText(sty, center, text)
	}
}

func (t Text) AllDataRanges() facet.DataRanges {
	dr := facet.NewDataRanges()
	for i := 0; i < t.XYText.Len(); i++ {
		x, y, _ := t.XYText.XYText(i)
		dr[facet.XScale].Update(x)
		dr[facet.YScale].Update(y)
	}
	UpdateAestheticsRanges(&dr, t.XYText.Len(), t.Alpha, t.Color, nil, nil, t.Size, nil)
	return dr
}
