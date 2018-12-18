package facet

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type debugging int

func (d debugging) V(a ...interface{}) {
	if d > 0 {
		log.Println(a...)
	}
}
func (d debugging) VV(a ...interface{}) {
	if d > 1 {
		a = append([]interface{}{"..."}, a...)
		log.Println(a...)
	}
}
func (d debugging) VVV(a ...interface{}) {
	if d > 2 {
		a = append([]interface{}{"... ..."}, a...)
		log.Println(a...)
	}
}

var debug = debugging(3)

type AllDataRanger interface {
	// AllDataRanges returns the ranges covered by the data for all
	// used scales.
	AllDataRanges() DataRanges
}

// DataRanges contains all the ranges covered by some data.
type DataRanges [numScales]Interval

// NewDataRange returns a DataRange with all intervals unset, i.e. [NaN,NaN].
func NewDataRanges() DataRanges {
	dr := DataRanges{}
	for i := range dr {
		dr[i].Min, dr[i].Max = math.NaN(), math.NaN()
	}
	return dr
}

const (
	XScale int = iota
	YScale
	FillScale
	SizeScale
	ColorScale
	StyleScale
	SymbolScale
	numScales
)

var scaleName = []string{"X-Scale", "Y-Scale", "Fill-Scale", "Size-Scale", "Color-Scale", "Style-Scale", "Symbol-Scale"}

type Geom interface {
	Draw(p *Panel)
}

// ----------------------------------------------------------------------------
// Panel

// A Panel represents one panel in a faceted plot.
type Panel struct {
	Geoms  []Geom
	Canvas draw.Canvas
	Scales [numScales]*Scale
}

// Map the data coordinate (x,y) into a canvas point.
func (p *Panel) Map(x, y float64) vg.Point {
	// fmt.Println(p)
	size := p.Canvas.Size()
	xu, yu := p.Scales[XScale].DataToUnit(x), p.Scales[YScale].DataToUnit(y)
	return vg.Point{
		X: p.Canvas.Min.X + vg.Length(xu)*size.X,
		Y: p.Canvas.Min.Y + vg.Length(yu)*size.Y,
	}
}

// ----------------------------------------------------------------------------
// Facet

// Facet describes a facetted plot.
type Facet struct {
	Title                string
	Rows, Cols           int
	Panels               [][]*Panel
	RowLabels, ColLabels []string
	XScales, YScales     []*Scale
	Scales               [numScales]*Scale // Except X and Y
	Style                FacetStyle
}

// NewFacet creates a new faceted plot with row x col many panels.
// All columns share the same X-sclae and all rows share the same Y-scale
// unless freeX or respectively freeY is specified.
func NewFacet(rows, cols int, freeX, freeY bool) *Facet {
	f := Facet{
		Rows:      rows,
		Cols:      cols,
		Panels:    make([][]*Panel, rows),
		RowLabels: make([]string, rows),
		ColLabels: make([]string, cols),
		XScales:   make([]*Scale, cols),
		YScales:   make([]*Scale, rows),
		Style:     DefaultFacetStyle(12),
	}

	for r := 0; r < f.Rows; r++ {
		f.Panels[r] = make([]*Panel, cols)
		for c := 0; c < f.Cols; c++ {
			f.Panels[r][c] = new(Panel)
		}
	}

	// The different X-scales.
	if freeX {
		for c := range f.XScales {
			f.XScales[c] = NewScale()
		}
	} else {
		common := NewScale()
		for c := range f.XScales {
			f.XScales[c] = common
		}
	}

	// The different X-scales.
	if freeY {
		for r := range f.YScales {
			f.YScales[r] = NewScale()
		}
	} else {
		common := NewScale()
		for r := range f.YScales {
			f.YScales[r] = common
		}
	}

	// The other scales.
	for i := range f.Scales {
		f.Scales[i] = NewScale()
	}

	return &f
}

// Learn all data ranges for all scales for all plotters in all panels in f.
func (f *Facet) learnDataRange() {
	for row := 0; row < f.Rows; row++ {
		f.Scales[YScale] = f.YScales[row]
		for col := 0; col < f.Cols; col++ {
			f.Scales[XScale] = f.XScales[col]
			for _, plt := range f.Panels[row][col].Geoms {
				if adr, ok := plt.(AllDataRanger); ok {
					dr := adr.AllDataRanges()
					for i := range dr {
						f.Scales[i].UpdateData(dr[i])
					}
				} else if sdr, ok := plt.(plot.DataRanger); ok {
					xmin, xmax, ymin, ymax := sdr.DataRange()
					f.Scales[XScale].Update(xmin)
					f.Scales[XScale].Update(xmax)
					f.Scales[YScale].Update(ymin)
					f.Scales[YScale].Update(ymax)
				}
			}
		}
	}
}

func (f *Facet) applyToScales(m func(*Scale)) {
	done := make(map[*Scale]bool)
	for _, s := range f.XScales {
		if done[s] {
			continue
		}
		m(s)
		done[s] = true
	}
	for _, s := range f.YScales {
		if done[s] {
			continue
		}
		m(s)
		done[s] = true
	}
	for _, s := range f.Scales {
		if done[s] {
			continue
		}
		m(s)
		done[s] = true
	}
}

func (f *Facet) debugScales(info string) {
	debug.V(info)
	for i, s := range f.XScales {
		debug.VV("X-Axis", i, s)
	}
	for i, s := range f.YScales {
		debug.VV("Y-Axis", i, s)
	}
	for i, s := range f.Scales {
		if i == XScale || i == YScale {
			continue
		}
		debug.VV("Scale", scaleName[i], s)
	}
}

func (f *Facet) deDegenerateXandY() error {
	// X- and Y-scales must not be unset or degenerate
	for _, s := range f.XScales {
		if math.IsNaN(s.Min) {
			s.Min = -1
		}
		if math.IsNaN(s.Max) {
			s.Max = 1
		}
	}
	for _, s := range f.YScales {
		if math.IsNaN(s.Min) {
			s.Min = -1
		}
		if math.IsNaN(s.Max) {
			s.Max = 1
		}
	}
	return nil
}

// Range prepares all panels and scales of f.
func (f *Facet) Range( /* Todo */ ) error {
	// We start by finding the all the actual, sharp data ranges.
	// Then we apply autoscaling constraints and expand the data ranges.
	f.learnDataRange()
	f.debugScales("After learning data ranges")

	f.applyToScales((*Scale).autoscale)
	f.debugScales("After autoscaling")

	f.deDegenerateXandY()
	f.debugScales("After de-degenerating X and Y")

	f.applyToScales((*Scale).buildConversionFuncs)
	f.debugScales("After building CF")

	f.Scales[FillScale].setupColor()
	f.Scales[ColorScale].setupColor()

	return nil // TODO: fail for illegal log scales, etc.
}

func (f *Facet) needGuides() bool {
	for s := FillScale; s < numScales; s++ {
		if f.Scales[ColorScale].HasData() {
			return true
		}
	}
	return false
}

// Draw renders f to c.
func (f *Facet) Draw(c draw.Canvas) error {
	if f.Title != "" {
		c.FillText(f.Style.Title, vg.Point{X: c.Center().X, Y: c.Max.Y}, f.Title)
		c.Max.Y -= f.Style.TitleHeight
	}

	if f.needGuides() {
		guideWidth := f.Style.Guide.Size * 3 // TODO: this 3 should be calculated or settable
		gc := c
		gc.Min.X = gc.Max.X - guideWidth

		for _, combo := range f.combineGuides() {

		}

		gc.Max.Y = f.drawDiscreteScales(gc)

		// TODO: Combine and Layout the relevant guides
		if f.Scales[FillScale].HasData() {
			gc.Max.Y = f.drawColorScale(gc, f.Scales[FillScale])
		}
		if f.Scales[ColorScale].HasData() {
			gc.Max.Y = f.drawColorScale(gc, f.Scales[ColorScale])
		}

		c.Max.X -= guideWidth + f.Style.Guide.Pad
	}

	var h1, h2, h3, h4 vg.Length
	var w1, w2, w3, w4 vg.Length

	// Determine various widths in main plot area.
	if f.YScales[0].Title != "" {
		w1 = f.Style.YAxis.TitleWidth
	}
	w2 = 20
	for _, rl := range f.RowLabels {
		if rl != "" {
			w4 = f.Style.VStrip.Width
			break
		}
	}
	w3 = c.Max.X - c.Min.X - w1 - w2 - w4

	// Determine various heights in main plot area.
	if f.XScales[0].Title != "" {
		h1 = f.Style.XAxis.TitleHeight
	}
	h2 = 20 // Tics and tic labels. TODO: calculate from style
	for _, cl := range f.ColLabels {
		if cl != "" {
			h4 = f.Style.HStrip.Height
			break
		}
	}
	h3 = c.Max.Y - c.Min.Y - h1 - h2 - h4

	// Draw the X and Y axis titles
	c.FillText(f.Style.XAxis.Title, vg.Point{X: c.Min.X + w1 + w2 + w3/2, Y: c.Min.Y}, f.XScales[0].Title)
	c.FillText(f.Style.YAxis.Title, vg.Point{X: c.Min.X, Y: c.Min.Y + h1 + h2 + h3/2}, f.YScales[0].Title)

	xticks := make([][]plot.Tick, f.Cols)
	yticks := make([][]plot.Tick, f.Rows)
	marker := plot.DefaultTicks{}
	for c, s := range f.XScales {
		xticks[c] = marker.Ticks(s.Min, s.Max)
	}
	for r, s := range f.YScales {
		yticks[r] = marker.Ticks(s.Min, s.Max)
	}

	// Setup the panel canvases, draw their background and draw the facet
	// column and row labels.
	padx, pady := f.Style.Panel.PadX, f.Style.Panel.PadY
	numCols, numRows := vg.Length(f.Cols), vg.Length(f.Rows)
	width := (w3 - padx*(numCols-1)) / numCols
	height := (h3 - pady*(numRows-1)) / numRows
	// Point (x0,y0) is the top-left corner of each panel
	y0 := c.Max.Y - h4
	for row, panels := range f.Panels {
		x0 := w1 + w2
		for col, panel := range panels {
			panel.Canvas.Canvas = c.Canvas
			panel.Canvas.Min.X = x0
			panel.Canvas.Min.Y = y0 - height
			panel.Canvas.Max.X = x0 + width
			panel.Canvas.Max.Y = y0
			panel.Scales = f.Scales
			panel.Scales[XScale] = f.XScales[col]
			panel.Scales[YScale] = f.YScales[row]
			panel.Canvas.SetColor(f.Style.Panel.Background)
			panel.Canvas.Fill(panel.Canvas.Rectangle.Path())
			if f.Style.Grid.Major.Color != nil {
				for _, xtic := range xticks[col] {
					r := panel.Map(xtic.Value, 0)
					sty := f.Style.Grid.Major
					if xtic.IsMinor() {
						sty = f.Style.Grid.Minor
					}
					panel.Canvas.StrokeLine2(sty,
						r.X, y0, r.X, y0-height)
				}
				for _, ytic := range yticks[row] {
					r := panel.Map(0, ytic.Value)
					sty := f.Style.Grid.Major
					if ytic.IsMinor() {
						sty = f.Style.Grid.Minor
					}
					panel.Canvas.StrokeLine2(sty,
						x0, r.Y, x0+width, r.Y)
				}
			}

			if row == 0 {
				cb := c
				cb.Min.X = panel.Canvas.Min.X
				cb.Min.Y = panel.Canvas.Max.Y
				cb.Max.X = panel.Canvas.Max.X
				cb.Max.Y = cb.Min.Y + w4
				cb.SetColor(f.Style.HStrip.Background)
				cb.Fill(cb.Rectangle.Path())
				cb.FillText(f.Style.HStrip.TextStyle, cb.Center(), f.ColLabels[col])
			}
			x0 += width + padx
		}
		cb := c
		panel := f.Panels[row][f.Cols-1]
		cb.Min = panel.Canvas.Rectangle.Max
		cb.Max.X = cb.Min.X + w4
		cb.Max.Y = panel.Canvas.Rectangle.Min.Y
		cb.SetColor(f.Style.VStrip.Background)
		cb.Fill(cb.Rectangle.Path())
		cb.FillText(f.Style.VStrip.TextStyle, cb.Center(), f.RowLabels[row])

		y0 -= height + pady
	}

	// Draw the actual data.
	for _, panels := range f.Panels {
		for _, panel := range panels {
			for _, geom := range panel.Geoms {
				geom.Draw(panel)
			}
		}
	}

	// Draw the tics
	for c, xtick := range xticks {
		for _, tick := range xtick {
			panel := f.Panels[f.Rows-1][c]
			r := panel.Map(tick.Value, 0)
			sty := f.Style.XAxis.Tick.Major
			length := f.Style.XAxis.Tick.Length
			if tick.IsMinor() {
				sty = f.Style.XAxis.Tick.Minor
				length /= 2
			}
			canvas := panel.Canvas
			y0 := canvas.Min.Y
			canvas.StrokeLine2(sty, r.X, y0, r.X, y0-length)
			if tick.IsMinor() {
				continue
			}
			canvas.FillText(f.Style.XAxis.Tick.Label,
				vg.Point{r.X, y0 - length}, tick.Label)
		}
	}
	for r, ytick := range yticks {
		for _, tick := range ytick {
			panel := f.Panels[r][0]
			r := panel.Map(0, tick.Value)
			sty := f.Style.YAxis.Tick.Major
			length := f.Style.YAxis.Tick.Length
			if tick.IsMinor() {
				sty = f.Style.YAxis.Tick.Minor
				length /= 2
			}
			canvas := panel.Canvas
			x0 := canvas.Min.X
			canvas.StrokeLine2(sty, x0-length, r.Y, x0, r.Y)
			if tick.IsMinor() {
				continue
			}
			canvas.FillText(f.Style.YAxis.Tick.Label,
				vg.Point{x0 - length, r.Y}, tick.Label)
		}
	}

	return nil
}

// MapSize maps the data value s to a display length via f's size scale.
func (f *Facet) MapSize(s float64) vg.Length {
	if f.Scales[SizeScale].DataToUnit == nil {
		return 0
	}

	max := 0.5 * f.Style.Guide.Size
	t := f.Scales[SizeScale].DataToUnit(s)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	return max * vg.Length(t)
}

// MapColor maps the data value s to a color via one of f's color scales.
func (f *Facet) MapColor(s float64, scale int) color.Color {
	if scale != ColorScale && scale != FillScale {
		panic(scale)
	}
	return f.Scales[scale].MapColor(s)
}

// combineGuides returns which combinations of guides need to be drawn and
// how they should be combined.
func (f *Facet) combineGuides() [][]int {
	combinations := [][]int{}
	for j := FillScale; j < numScales; j++ {
		if !f.Scales[j].HasData() {
			continue // This scale has no data, so no need to combine it.
		}
		s := f.Scales[j]
		combined := false
		for i, combi := range combinations {
			for k := range combi {
				r := f.Scales[combi[k]]
				if canCombineScales(s, r) {
					combinations[i] = append(combinations[i], j)
					combined = true
					break
				}
			}
		}
		if !combined {
			combinations = append(combinations, []int{j})
		}
	}
	return combinations
}

// Guides for different scales are combined iff:
func canCombineScales(s1, s2 *Scale) bool {
	// 1. The two scales are of the same kind (discrete, continuous, ...)
	if s1.ScaleType != s2.ScaleType {
		return false
	}

	// 2. The two scales have the same range.
	if s1.Min != s2.Min || s1.Max != s2.Max {
		return false
	}

	// 3. The two scales have the same Title or the Title is empty.
	if s1.Title != s2.Title && s1.Title != "" && s2.Title != "" {
		return false
	}

	// 4. The scales must use the same Ticker.
	// TODO

	// 5. Fill and Color can be combined if they use the same ColorMap or one is empty.
	if s1.ColorMap != s2.ColorMap && s1.ColorMap != nil && s2.ColorMap != nil {
		return false
	}

	return true

}

func (f *Facet) drawDiscreteScales(c draw.Canvas) vg.Length {
	if f.Scales[SymbolScale].HasData() {
		return f.drawSymbolGuide(c, f.Scales[SymbolScale], true, false, false, true)
	}
	return c.Min.Y
}

func (f *Facet) drawSymbolGuide(c draw.Canvas, scale *Scale, showSymbol, showColor, showSize, showStyle bool) vg.Length {
	if scale.Title != "" {
		p := vg.Point{
			X: c.Min.X + 0.5*f.Style.Guide.Size,
			Y: c.Max.Y,
		}
		c.FillText(f.Style.Guide.Title, p, scale.Title)
		c.Max.Y -= 2 * f.Style.Guide.Title.Font.Size
	}

	a, e := int(scale.Data.Min), int(scale.Data.Max)
	boxSize, pad := f.Style.Guide.Size, vg.Length(3)
	r := vg.Rectangle{
		Min: vg.Point{c.Min.X, c.Max.Y - boxSize},
		Max: vg.Point{c.Min.X + boxSize, c.Max.Y},
	}

	labelSty := f.Style.YAxis.Tick.Label
	labelSty.XAlign = draw.XLeft

	var pal []color.Color
	if showColor {
		pal = scale.ColorMap.Palette(e - 1 + 1).Colors()
	}

	shape := draw.GlyphDrawer(draw.CircleGlyph{})
	col := color.Color(color.Black)
	radius := boxSize / 2

	for level := a; level <= e; level++ {
		// The background box.
		c.SetColor(color.Gray{0xee})
		c.Fill(r.Path())

		center := vg.Point{X: (r.Min.X + r.Max.X) / 2, Y: (r.Min.Y + r.Max.Y) / 2}

		// The actual indicators.
		if showColor {
			col = pal[level-a]
		}
		if showSize {

		}
		if showSymbol {
			shape = plotutil.Shape(level)
		}

		if showStyle {
			lsty := draw.LineStyle{
				Color:  col,
				Width:  1,
				Dashes: plotutil.Dashes(level),
			}
			c.StrokeLine2(lsty, c.Min.X, center.Y, c.Min.X+boxSize, center.Y)
		}

		gsty := draw.GlyphStyle{
			Color:  col,
			Radius: radius,
			Shape:  shape,
		}
		c.DrawGlyph(gsty, center)

		// The label.
		c.FillText(labelSty, vg.Point{r.Max.X + pad, (r.Min.Y + r.Max.Y) / 2}, strconv.Itoa(level))

		// The box border
		c.SetColor(color.Black)
		c.SetLineWidth(vg.Length(0.3))
		c.Stroke(r.Path())

		r.Min.Y -= boxSize + pad
		r.Max.Y -= boxSize + pad
	}

	return r.Min.Y + boxSize - 2*pad
}

// Draw continuous color scales.
func (f *Facet) drawColorScale(c draw.Canvas, scale *Scale) vg.Length {
	fmt.Println("drawColorSclae", scale.Title)
	if scale.Title != "" {
		p := vg.Point{
			X: c.Min.X + 0.5*f.Style.Guide.Size,
			Y: c.Max.Y,
		}
		c.FillText(f.Style.Guide.Title, p, scale.Title)
		c.Max.Y -= 2 * f.Style.Guide.Title.Font.Size
	}

	if scale.ScaleType == Discrete {
		return f.drawDiscreteColorGuide(c, scale)
	}
	return f.drawContinuousColorGuide(c, scale)
}

func (f *Facet) drawDiscreteColorGuide(c draw.Canvas, scale *Scale) vg.Length {
	a, e := int(scale.Data.Min), int(scale.Data.Max)
	n := e - a + 1
	size, pad := f.Style.Guide.Size, vg.Length(3)
	r := vg.Rectangle{
		Min: vg.Point{c.Min.X, c.Max.Y - size},
		Max: vg.Point{c.Min.X + size, c.Max.Y},
	}

	labelSty := f.Style.YAxis.Tick.Label
	labelSty.XAlign = draw.XLeft

	pal := scale.ColorMap.Palette(n).Colors()
	for level := e; level >= a; level-- {
		col := pal[level-a]
		c.SetColor(col)
		c.Fill(r.Path())
		c.SetColor(color.Black)
		c.SetLineWidth(vg.Length(0.3))
		c.Stroke(r.Path())
		c.FillText(labelSty, vg.Point{r.Max.X + pad, (r.Min.Y + r.Max.Y) / 2}, strconv.Itoa(level))

		r.Min.Y -= size + pad
		r.Max.Y -= size + pad
	}

	return r.Min.Y + size - 2*pad
}

func (f *Facet) drawContinuousColorGuide(c draw.Canvas, scale *Scale) vg.Length {
	width := f.Style.Guide.Size
	height := 5 * width
	scale2Canvas := func(x float64) vg.Length {
		t := scale.DataToUnit(x)
		return c.Max.Y - height + height*vg.Length(t)
	}
	rect := vg.Rectangle{
		Min: vg.Point{c.Min.X, scale2Canvas(scale.Min)},
		Max: vg.Point{c.Min.X + width, scale2Canvas(scale.Max)},
	}
	step := height / 101
	r := rect
	for i := 0; i <= 100; i++ {
		col, err := scale.ColorMap.At(float64(i) / 100)
		if err != nil {
			panic(fmt.Sprintf("%d %s", i, err))
		}
		c.SetColor(col)
		c.Fill(r.Path())
		r.Min.Y += step
	}
	c.SetColor(color.Black)
	c.SetLineWidth(vg.Length(0.3))
	c.Stroke(rect.Path())
	ticks := plot.DefaultTicks{}.Ticks(scale.Min, scale.Max)
	// TODO: Do not use YAxis style.
	for _, tick := range ticks {
		sty := f.Style.YAxis.Tick.Major
		length := f.Style.YAxis.Tick.Length
		if tick.IsMinor() {
			sty = f.Style.YAxis.Tick.Minor
			length /= 2
		}
		x := rect.Max.X
		y := scale2Canvas(tick.Value)
		fmt.Println(length, x, y)
		c.StrokeLine2(sty, x, y, x+length, y)
		if tick.IsMinor() {
			continue
		}
		tsty := f.Style.YAxis.Tick.Label
		tsty.XAlign = draw.XLeft
		c.FillText(tsty,
			vg.Point{x + length, y}, tick.Label)
	}

	return rect.Min.Y
}

// DefaultFacetStyle returns a FacetStyle which mimics the appearance of ggplot2.
// The baseFontSize is the font size for axis titles and strip labels, the title
// is a bit bigger, tick labels a bit smaller.
func DefaultFacetStyle(baseFontSize vg.Length) FacetStyle {
	scale := func(x vg.Length, f float64) vg.Length {
		return vg.Length(math.Round(f * float64(x)))
	}

	titleFont, err := vg.MakeFont("Helvetica-Bold", scale(baseFontSize, 1.2))
	if err != nil {
		panic(err)
	}
	baseFont, err := vg.MakeFont("Helvetica-Bold", baseFontSize)
	if err != nil {
		panic(err)
	}
	tickFont, err := vg.MakeFont("Helvetica-Bold", scale(baseFontSize, 1/1.2))
	if err != nil {
		panic(err)
	}

	fs := FacetStyle{}
	fs.TitleHeight = scale(baseFontSize, 3)
	fs.Title.Color = color.Black
	fs.Title.Font = titleFont
	fs.Title.XAlign = draw.XCenter
	fs.Title.YAlign = draw.YTop

	fs.Panel.Background = color.Gray16{0xeeee}
	fs.Panel.PadX = scale(baseFontSize, 0.5)
	fs.Panel.PadY = fs.Panel.PadX

	fs.HStrip.Background = color.Gray16{0xcccc}
	fs.HStrip.Font = baseFont
	fs.HStrip.Height = scale(baseFontSize, 2)
	fs.HStrip.XAlign = draw.XCenter
	fs.HStrip.YAlign = -0.3 // draw.YCenter

	fs.VStrip.Background = color.Gray16{0xcccc}
	fs.VStrip.Font = baseFont
	fs.VStrip.Width = scale(baseFontSize, 2.5)
	fs.VStrip.XAlign = draw.XCenter
	fs.VStrip.YAlign = -0.3 // draw.YCenter
	fs.VStrip.Rotation = -math.Pi / 2

	fs.Grid.Major.Color = color.White
	fs.Grid.Major.Width = vg.Length(1)
	fs.Grid.Minor.Color = color.White
	fs.Grid.Minor.Width = vg.Length(0.5)

	fs.XAxis.Title.Color = color.Black
	fs.XAxis.Title.Font = baseFont
	fs.XAxis.Title.Rotation = 0
	fs.XAxis.Title.XAlign = draw.XCenter
	fs.XAxis.Title.YAlign = draw.YAlignment(0.3)
	fs.XAxis.TitleHeight = scale(baseFontSize, 2)

	fs.XAxis.Line.Width = 0
	fs.XAxis.Tick.Label.Color = color.Black
	fs.XAxis.Tick.Label.Font = tickFont
	fs.XAxis.Tick.Label.XAlign = draw.XCenter
	fs.XAxis.Tick.Label.YAlign = draw.YTop
	fs.XAxis.Tick.Major.Color = color.Gray16{0x1111}
	fs.XAxis.Tick.Major.Width = vg.Length(1)
	fs.XAxis.Tick.Length = vg.Length(5)

	fs.YAxis.Title.Color = color.Black
	fs.YAxis.Title.Font = baseFont
	fs.YAxis.Title.Rotation = math.Pi / 2
	fs.YAxis.Title.XAlign = draw.XCenter
	fs.YAxis.Title.YAlign = draw.YTop
	fs.YAxis.TitleWidth = scale(baseFontSize, 2)

	fs.YAxis.Line.Width = 0
	fs.YAxis.Tick.Label.Color = color.Black
	fs.YAxis.Tick.Label.Font = tickFont
	fs.YAxis.Tick.Label.XAlign = draw.XRight
	fs.YAxis.Tick.Label.YAlign = -0.3 // draw.YCenter
	fs.YAxis.Tick.Major.Color = color.Gray16{0x1111}
	fs.YAxis.Tick.Major.Width = vg.Length(1)
	fs.YAxis.Tick.Length = vg.Length(5)

	fs.Guide.Title.Color = color.Black
	fs.Guide.Title.Font = baseFont
	fs.Guide.Title.XAlign = draw.XCenter
	fs.Guide.Title.YAlign = draw.YTop
	fs.Guide.Size = scale(baseFontSize, 2)
	fs.Guide.Pad = scale(baseFontSize, 1)

	return fs
}

type FacetStyle struct {
	Title       draw.TextStyle
	TitleHeight vg.Length

	Panel struct {
		Background color.Color
		PadX       vg.Length
		PadY       vg.Length
	}
	HStrip struct {
		Background color.Color
		Height     vg.Length
		draw.TextStyle
	}
	VStrip struct {
		Background color.Color
		Width      vg.Length
		draw.TextStyle
	}

	Grid struct {
		Major draw.LineStyle
		Minor draw.LineStyle
	}

	XAxis struct {
		Title       draw.TextStyle
		TitleHeight vg.Length
		Line        draw.LineStyle
		Tick        struct {
			Label  draw.TextStyle
			Major  draw.LineStyle
			Minor  draw.LineStyle
			Length vg.Length
		}
	}

	YAxis struct {
		Title      draw.TextStyle
		TitleWidth vg.Length
		Line       draw.LineStyle
		Tick       struct {
			Label  draw.TextStyle
			Major  draw.LineStyle
			Minor  draw.LineStyle
			Length vg.Length
		}
	}

	Guide struct {
		Title draw.TextStyle
		Size  vg.Length
		Pad   vg.Length
	}
}