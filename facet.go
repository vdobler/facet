package facet

import (
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/palette/moreland"
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

// DataRanges contains all the ranges covered by some data.
type DataRanges [numScales]Interval

// NewDataRange returns a DataRange with all intervals unset, i.e. [NaN,NaN].
func NewDataRanges() DataRanges {
	dr := DataRanges{}
	for i := range dr {
		dr[i] = UnsetInterval
	}
	return dr
}

const (
	XScale int = iota
	YScale
	AlphaScale
	ColorScale
	FillScale
	ShapeScale
	SizeScale
	StrokeScale
	numScales
)

var scaleName = []string{
	"X-Scale",
	"Y-Scale",
	"Alpha-Scale",
	"Color-Scale",
	"Fill-Scale",
	"Shape-Scale",
	"Size-Scale",
	"Stroke-Scale"}

// A Geom is the geometrical representation of some data.
type Geom interface {
	// DataRange returns what ranges on which scales are covered by
	// covered by the geoms indexed by subset
	DataRange() DataRanges

	// Draw is called to draw the geoms indexed by subset onto p.
	Draw(p *Panel)
}

// A FGeom is the geometrical representation of some faceted data.
type FGeom interface {
	// N returns the number of geoms in this data set.
	N() int

	// Group of the i'th geom in this data set.
	Group(i int) GroupID

	// DataRange returns what ranges on which scales are covered by
	// covered by the geoms indexed by subset
	DataRange(subset []int) DataRanges

	// Draw is called to draw the geoms indexed by subset onto p.
	Draw(p *Panel, subset []int)
}

// ----------------------------------------------------------------------------
// FacetPlot

type Layer struct {
}

// FacetPlot describes a automatically facetted plot.
type FacetPlot struct {
	// Title is the optional plot title.
	Title string

	Geoms []FGeom

	// Rows and Cols are number of rows and columns in the faceted plot.
	Rows, Cols int
}

func GeneratePlot(fp FacetPlot) *Plot {
	for _, g := range fp.Geoms {
		g.N()
	}
	return nil // TODO
}

// ----------------------------------------------------------------------------
// Plot

// Plot describes a facetted plot.
type Plot struct {
	// Title is the optional plot title.
	Title string

	// Rows and Cols are number of rows and columns in the faceted plot.
	Rows, Cols int

	// Panels is the matirx of plot panels: Panels[r][c] is the
	// panel at row r and column c.
	Panels [][]*Panel

	// RowLabels and ColLabels contain the optional strip titles
	// for the row and column strips of a grid layout.
	RowLabels, ColLabels []string

	// XScales are the scales for the Col many x-axes. If the x scales
	// are not free all x-axes will share the same scale.
	XScales []*Scale

	// YScales are the scales for the Row many y-axes. If the y scales
	// are not free all y-axes will share the same scale.
	YScales []*Scale

	// Scales contains the rest of the scales like Color, Fill, Shape, etc.
	Scales [numScales]*Scale // Except X and Y

	// ColorMap and FillMap are used to map the ColorScale and FillScale
	// to a color
	ColorMap, FillMap palette.ColorMap

	// Style used during plotting. TODO: Keep here?
	Style Style

	// Messages is used to report warnings and errors during creation
	// of the plot.
	Messages io.Writer
}

// NewSimple creates a new un-faceted plot, that is a plot with just one panel.
func NewSimplePlot() *Plot {
	return NewPlot(1, 1, false, false)
}

// NewPlot creates a new faceted plot in a grid layout with row * col many
// panels.
// All columns share the same X-scale and all rows share the same Y-scale
// unless freeX or respectively freeY is specified.
func NewPlot(rows, cols int, freeX, freeY bool) *Plot {
	plot := &Plot{
		Rows:      rows,
		Cols:      cols,
		Panels:    make([][]*Panel, rows),
		RowLabels: make([]string, rows),
		ColLabels: make([]string, cols),
		XScales:   make([]*Scale, cols),
		YScales:   make([]*Scale, rows),
		Style:     DefaultFacetStyle(12),
		Messages:  ioutil.Discard,
	}

	for r := 0; r < plot.Rows; r++ {
		plot.Panels[r] = make([]*Panel, cols)
		for c := 0; c < plot.Cols; c++ {
			plot.Panels[r][c] = new(Panel)
			plot.Panels[r][c].Plot = plot
		}
	}

	// The different X-scales.
	if freeX {
		for c := range plot.XScales {
			plot.XScales[c] = NewScale()

		}
	} else {
		common := NewScale()
		for c := range plot.XScales {
			plot.XScales[c] = common
		}
	}

	// The different Y-scales.
	if freeY {
		for r := range plot.YScales {
			plot.YScales[r] = NewScale()
		}

	} else {
		common := NewScale()
		for r := range plot.YScales {
			plot.YScales[r] = common
		}
	}

	// The other scales.
	for i := range plot.Scales {
		plot.Scales[i] = NewScale()
	}

	plot.setScaleDefaults()

	// The two color maps.
	rainbow := &Rainbow{
		Saturation: 0.9,
		Value:      0.9,
		StartHue:   0,
		HueGap:     1.0 / 6.0,
		alpha:      1,
	}
	plot.ColorMap = rainbow
	plot.FillMap = moreland.ExtendedBlackBody()

	return plot
}

func (p *Plot) setScaleDefaults() {
	// The positional scales look good if the scale is 5% longer than
	// the actual data range on each side.
	for _, s := range p.XScales {
		s.Autoscaling.Expand.Releative = p.Style.XAxis.Expand.Releative
		s.Autoscaling.Expand.Absolute = p.Style.XAxis.Expand.Absolute
		s.Trans = LinearTrans
	}
	for _, s := range p.YScales {
		s.Autoscaling.Expand.Releative = p.Style.YAxis.Expand.Releative
		s.Autoscaling.Expand.Absolute = p.Style.YAxis.Expand.Absolute
		s.Trans = LinearTrans
	}

	// TODO: Trans for other scales.

	// The size scale normaly maps the size aestethics to area
	// so use an sqrt transform and do not map 0 to visually nothing.
	p.Scales[SizeScale].Trans = SqrtTrans
}

// LearnDataRange determines the the range the data covers in all scales.
func (p *Plot) LearnDataRange() {
	for _, s := range p.XScales {
		s.Data = UnsetInterval
	}
	for _, s := range p.YScales {
		s.Data = UnsetInterval
	}
	for _, s := range p.Scales {
		s.Data = UnsetInterval
	}

	for row := 0; row < p.Rows; row++ {
		p.Scales[YScale] = p.YScales[row]
		for col := 0; col < p.Cols; col++ {
			p.Scales[XScale] = p.XScales[col]
			for _, geom := range p.Panels[row][col].Geoms {
				for s, r := range geom.DataRange() {
					p.Scales[s].UpdateData(r)
				}
			}
		}
	}
	p.debugScales("After learning data ranges")

}

// Autoscale all scales based on the current Data range.
func (p *Plot) Autoscale() {
	p.applyToScales((*Scale).Autoscale)
	p.debugScales("After autoscaling")
}

func (p *Plot) fillRange() {
	p.applyToScales((*Scale).fillRange)
	p.debugScales("After filling Range")
}

func (f *Plot) applyToScales(m func(*Scale)) {
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
func (p *Plot) Warnf(f string, args ...interface{}) {
	if !strings.HasSuffix(f, "\n") {
		f += "\n"
	}
	fmt.Fprintf(os.Stderr, f, args...)
}

func (p *Plot) debugScales(info string) {
	debug.V(info)
	for i, s := range p.XScales {
		debug.VV("X-Axis", i, s)
	}
	for i, s := range p.YScales {
		debug.VV("Y-Axis", i, s)
	}
	for i, s := range p.Scales {
		if i == XScale || i == YScale {
			continue
		}
		debug.VV("Scale", scaleName[i], s)
	}
}

// DeDegenerateXandY makes sure the Limit intervall for all X and Y scales in p
// are not degenerated: NaN and Inf are turned into -1 (Min) or +1 (Max)
// degenerate intervalls of the form [a, a] are exapnded around a.
func (p *Plot) DeDegenerateXandY() {
	// X- and Y-scales must not be unset or degenerate
	for i, s := range p.XScales {
		if s.Limit.Degenerate() {
			p.Warnf("Corrected degeneration of %dth X scale", i)
		}
	}
	for i, s := range p.YScales {
		if s.Limit.Degenerate() {
			p.Warnf("Corrected degeneration of %dth Y scale", i)
		}
	}
	p.debugScales("After de-degenerating X and Y")
}

// Prepare learns the Data range of each scale, autoscales each scale's limit,
// clears each scales's range and degenrated the X and Y scales.
func (p *Plot) Prepare() {
	p.LearnDataRange()
	p.Autoscale()
	p.DeDegenerateXandY()
	p.fillRange()

	p.setupColorAndSizeMaps() // TODO: this should go somewhere else
}

func (p *Plot) setupColorAndSizeMaps() {
	p.ColorMap.SetMin(0)
	p.ColorMap.SetMax(1)
	p.FillMap.SetMin(0)
	p.FillMap.SetMax(1)
}

func (p *Plot) needGuides() bool {
	for s := AlphaScale; s < numScales; s++ {
		if p.Scales[s].HasData() {
			return true
		}
	}
	return false
}

// Draw renders f to c.
func (f *Plot) Draw(c draw.Canvas) error {
	debug.V("Drawing to canvas from ", c.Min.X, ",", c.Min.Y, " to ", c.Max.X, ",", c.Max.Y)
	if f.Title != "" {
		c.FillText(f.Style.Title, vg.Point{X: c.Center().X, Y: c.Max.Y}, f.Title)
		c.Max.Y -= f.Style.TitleHeight
	}

	if f.needGuides() {
		// TODO: guides should be vertically centered.
		guideWidth := f.Style.Legend.Discrete.Size * 3 // TODO: this 3 should be calculated or settable
		gc := c
		gc.Min.X = gc.Max.X - guideWidth

		for _, combo := range f.combineGuides() {
			gc.Max.Y = f.drawGuides(gc, combo)
		}

		c.Max.X -= guideWidth + f.Style.Legend.Discrete.Pad
	}

	var h1, h2, h3, h4 vg.Length
	var w1, w2, w3, w4 vg.Length

	// Determine various widths in main plot area.
	if f.YScales[0].Title != "" {
		w1 = f.Style.YAxis.TitleWidth
	}
	w2 = 30 // TODO: Dynamic
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
	for c, s := range f.XScales {
		xticks[c] = s.Trans.Ticker.Ticks(s.Limit.Min, s.Limit.Max)
	}
	for r, s := range f.YScales {
		yticks[r] = s.Trans.Ticker.Ticks(s.Limit.Min, s.Limit.Max)
	}

	// Setup the panel canvases, draw their background and draw the facet
	// column and row labels.
	padx, pady := f.Style.Panel.PadX, f.Style.Panel.PadY
	numCols, numRows := vg.Length(f.Cols), vg.Length(f.Rows)
	width := (w3 - padx*(numCols-1)) / numCols
	height := (h3 - pady*(numRows-1)) / numRows
	havePanelTitle := f.havePanelTitle()

	// Point (x0,y0) is the top-left corner of each panel
	y0 := c.Max.Y - h4
	for row, panels := range f.Panels {
		x0 := c.Min.X + w1 + w2

		for col, panel := range panels {
			if panel == nil {
				continue
			}
			f.setupPanel(panel, row, col, c, havePanelTitle,
				x0, y0, width, height,
				xticks[col], yticks[row])

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
			r := panel.MapXY(tick.Value, 0)
			sty := f.Style.XAxis.MajorTick.LineStyle
			length := f.Style.XAxis.MajorTick.Length
			align := vg.Length(f.Style.XAxis.MajorTick.Align)
			if tick.IsMinor() {
				sty = f.Style.XAxis.MinorTick.LineStyle
				length = f.Style.XAxis.MinorTick.Length
				align = vg.Length(f.Style.XAxis.MinorTick.Align)
			}
			canvas := panel.Canvas
			y0 := canvas.Min.Y
			canvas.StrokeLine2(sty, r.X, y0+align*length, r.X, y0+(align-1)*length)
			if tick.IsMinor() {
				continue
			}
			canvas.FillText(f.Style.XAxis.MajorTick.Label,
				vg.Point{r.X, y0 - length}, tick.Label)
		}
	}
	for r, ytick := range yticks {
		for _, tick := range ytick {
			panel := f.Panels[r][0]
			r := panel.MapXY(0, tick.Value)
			sty := f.Style.YAxis.MajorTick.LineStyle
			length := f.Style.YAxis.MajorTick.Length
			align := vg.Length(f.Style.YAxis.MajorTick.Align)
			if tick.IsMinor() {
				sty = f.Style.YAxis.MinorTick.LineStyle
				length = f.Style.YAxis.MinorTick.Length
				align = vg.Length(f.Style.YAxis.MinorTick.Align)
			}
			canvas := panel.Canvas
			x0 := canvas.Min.X
			canvas.StrokeLine2(sty, x0+(align-1)*length, r.Y, x0+align*length, r.Y)
			if tick.IsMinor() {
				continue
			}
			canvas.FillText(f.Style.YAxis.MajorTick.Label,
				vg.Point{x0 - length, r.Y}, tick.Label)
		}
	}

	return nil
}

func (p *Plot) havePanelTitle() bool {
	for _, panels := range p.Panels {
		for _, panel := range panels {
			if panel != nil && panel.Title != "" {
				return true
			}
		}

	}
	return false
}

func (p *Plot) setupPanel(panel *Panel, row, col int, canvas draw.Canvas,
	havePanelTitle bool,
	x0, y0, width, height vg.Length,
	xticks, yticks []plot.Tick) {
	debug.V("setupPanel", row, ",", col, " at ", x0, ",", y0, " size ", width, "x", height)

	panel.Canvas.Canvas = canvas.Canvas
	panel.Canvas.Min.X = x0
	panel.Canvas.Min.Y = y0 - height
	panel.Canvas.Max.X = x0 + width
	panel.Canvas.Max.Y = y0

	if havePanelTitle {
		min := vg.Point{x0, y0}
		max := vg.Point{x0 + width, y0 + p.Style.HStrip.Height}
		p.drawStrip(canvas, panel.Title, min, max, p.Style.HStrip.TextStyle)
	}

	panel.Scales = p.Scales
	panel.Scales[XScale] = p.XScales[col]
	panel.Scales[YScale] = p.YScales[row]
	panel.Canvas.SetColor(p.Style.Panel.Background)
	panel.Canvas.Fill(panel.Canvas.Rectangle.Path())
	if p.Style.Grid.Major.Color != nil {
		for _, xtic := range xticks {
			r := panel.MapXY(xtic.Value, 0)
			sty := p.Style.Grid.Major
			if xtic.IsMinor() {
				sty = p.Style.Grid.Minor
			}
			panel.Canvas.StrokeLine2(sty,
				r.X, y0, r.X, y0-height)
		}
		for _, ytic := range yticks {
			r := panel.MapXY(0, ytic.Value)
			sty := p.Style.Grid.Major
			if ytic.IsMinor() {
				sty = p.Style.Grid.Minor
			}
			panel.Canvas.StrokeLine2(sty,
				x0, r.Y, x0+width, r.Y)
		}
	}

}

func (p *Plot) drawStrip(c draw.Canvas, text string, min, max vg.Point, style draw.TextStyle) {
	cb := c
	cb.Min = min
	cb.Max = max
	cb.SetColor(p.Style.VStrip.Background)
	cb.Fill(cb.Rectangle.Path())
	cb.FillText(style, cb.Center(), text)
}

// MapSize maps the data value s to a display length via f's size scale.
// Values outside of of the range of the size scale are mapped to 0.
func (p *Plot) MapSize(v float64) vg.Length {
	min := 2.0
	max := float64(0.5 * p.Style.Legend.Discrete.Size)
	s := p.Scales[SizeScale]
	t := s.Trans.Trans(s.Range, Interval{min, max}, v)

	if !p.Scales[SizeScale].InRange(v) || math.IsNaN(t) {
		return 0
	}
	return vg.Length(t)
}

// MapColor maps the data value v to a color via p's ColorMap or
// FillMap if fill is true.
// Values outside if the relevant scale's intervall are mapped to
// Gray50 (which is what ggplot2 does).
func (p *Plot) MapColor(v float64, fill bool) color.Color {
	scale, cm := p.Scales[ColorScale], p.ColorMap
	if fill {
		scale, cm = p.Scales[FillScale], p.FillMap
	}
	if !scale.InRange(v) {
		return color.Gray{0x7f}
	}

	t := scale.Map(v)
	if math.IsNaN(t) {
		return color.Gray{0x7f}
	}

	if t < 0 || t > 1 {
		panic(fmt.Sprintf("MapColor(%g,%t), t=%f", v, fill, t))
	}
	cm.SetMin(0)
	cm.SetMax(1)
	col, err := cm.At(t)
	if err != nil {
		panic(fmt.Sprintf("MapColor(%g,%t), t=%f: %v", v, fill, t, err))
	}

	return col
}

// combineGuides returns which combinations of guides need to be drawn and
// how they should be combined.
func (f *Plot) combineGuides() [][]int {
	debug.V("Combining scales")
	combinations := [][]int{}
	for j := AlphaScale; j < numScales; j++ {
		debug.VV(scaleName[j], "data range", f.Scales[j].Data.Min, f.Scales[j].Data.Max, f.Scales[j].HasData())
		if !f.Scales[j].HasData() {
			debug.VV(scaleName[j], "has no data")
			continue // This scale has no data, so no need to combine it.
		}

		combinable := false
		for i, combi := range combinations {
			combinable = true
			for _, k := range combi {
				if !f.canCombineScales(j, k) {
					combinable = false
					break
				}
			}
			if combinable {
				combinations[i] = append(combinations[i], j)
				debug.VV(scaleName[j], "combined into", combinations[i])
				break
			}
		}
		if !combinable {
			combinations = append(combinations, []int{j})
			debug.VV(scaleName[j], "uncombinable")
		}
	}
	debug.V("Combined scales", combinations)
	return combinations
}

// Guides for different scales are combined iff:
func (p *Plot) canCombineScales(j, k int) bool {
	s1, s2 := p.Scales[j], p.Scales[k]

	// 1. The two scales are of the same kind (linear, discrete, time, ...)
	if s1.ScaleType != s2.ScaleType {
		debug.VVV("different type for", j, k)
		return false
	}

	// 2. The two scales have the same range.
	if s1.Limit.Min != s2.Limit.Min || s1.Limit.Max != s2.Limit.Max {
		debug.VVV("different range for", j, s1.Limit.Min, s1.Limit.Max,
			"and", k, s2.Limit.Min, s2.Limit.Max)
		return false
	}

	// 3. The two scales have the same Title or the Title is empty.
	if s1.Title != s2.Title && s1.Title != "" && s2.Title != "" {
		return false
	}

	// 4. The scales must use the same Ticker.
	if s1.Ticker != nil && s2.Ticker != nil && s1.Ticker != s2.Ticker {
		t1, t2 := s1.Ticker.Ticks(s1.Limit.Min, s1.Limit.Max), s2.Ticker.Ticks(s2.Limit.Min, s2.Limit.Max)
		if len(t1) != len(t2) {
			return false
		}
		for i := range t1 {
			if t1[i].Value != t2[i].Value || t1[i].Label != t2[i].Label {
				return false
			}
		}
	}

	// 5. Fill and Color can be combined if they use the same ColorMap or one is empty.
	if (j == FillScale && k == ColorScale) ||
		(k == FillScale && j == ColorScale) {
		if p.ColorMap != p.FillMap && p.ColorMap != nil && p.FillMap != nil {
			return false
		}
	}

	return true

}

// There are two major types of guides:
//   A. Color guides for continuous scales drawn as a continuous "rainbow".
//   B. Discrete guides where each label is shown as a small rectangle
//      containing lines, symbols, etc.
func (p *Plot) drawGuides(c draw.Canvas, scales []int) vg.Length {
	if title := p.titleFor(scales); title != "" {
		pos := vg.Point{
			X: c.Min.X,
			Y: c.Max.Y,
		}
		c.FillText(p.Style.Legend.Title, pos, title)
		c.Max.Y -= 2 * p.Style.Legend.Title.Font.Size
	}

	if p.isContinuousColorGuide(scales) {
		s := p.Scales[scales[0]]
		m := p.colorMapFor(scales)
		return p.drawContinuousColorGuide(c, s, m)
	}
	return p.drawDiscreteGuides(c, scales)
}

func (f *Plot) titleFor(scales []int) string {
	for _, s := range scales {
		if title := f.Scales[s].Title; title != "" {
			return title
		}
	}
	return ""
}

// Finding a suitable ticker is complicated: If one of the scales
// is a Symbol or Style scale only integer values are allowed and
// all used values should be ticked.
// TODO: Maybe Style and Symbol must be different kind of scales
// as these cannot be anything than discrete as anything else cannont
// be mapped to an aesthetics.
func (f *Plot) tickerFor(scales []int) plot.Ticker {
	for _, s := range scales {
		if f.Scales[s].Ticker != nil {
			return f.Scales[s].Ticker
		}
	}
	if containsInt(scales, StrokeScale) || containsInt(scales, ShapeScale) {
		return DiscreteTicks{}

	}

	return DefaultTicks(6)
}

type DiscreteTicks struct{}

var _ plot.Ticker = DiscreteTicks{}

// Ticks makes DiscreteTicks implements plot.Ticker.
func (DiscreteTicks) Ticks(min, max float64) []plot.Tick {
	min, max = math.Ceil(min), math.Floor(max)

	ticks := []plot.Tick{}
	for ; min <= max; min++ {
		ticks = append(ticks, plot.Tick{
			Value: min,
			Label: fmt.Sprintf("%d", int(min)),
		})
	}
	return ticks
}

// colorMapFor looks for a color map defined on one of the given scales.
// Only Fill- and ColorScales are inspected.
func (p *Plot) colorMapFor(scales []int) palette.ColorMap {
	for _, s := range scales {
		if s != FillScale && s != ColorScale {
			continue
		}
		if s == ColorScale {
			return p.ColorMap
		}
		return p.FillMap
	}
	return p.FillMap // TODO: panic ??
}

func (f *Plot) SizeMap() func(x float64) vg.Length {
	if !f.Scales[SizeScale].HasData() {
		return func(x float64) vg.Length { return 5 }
	}

	min, max := vg.Length(3), f.Style.Legend.Discrete.Size

	return func(x float64) vg.Length {
		return min + vg.Length(x)*(max-min)
	}
}

func (f *Plot) isContinuousColorGuide(scales []int) bool {
	if f.Scales[scales[0]].ScaleType == Discrete {
		return false
	}
	for _, s := range scales {
		if s != FillScale && s != ColorScale {
			return false
		}
	}
	return true
}

func (plot *Plot) drawDiscreteGuides(c draw.Canvas, scales []int) vg.Length {
	debug.V("Drawing descrete scales", scales)
	showAlpha := containsInt(scales, AlphaScale)
	showColor := containsInt(scales, ColorScale)
	showFill := containsInt(scales, FillScale)
	showShape := containsInt(scales, ShapeScale)
	showSize := containsInt(scales, SizeScale)
	showStroke := containsInt(scales, StrokeScale)
	scale := plot.Scales[scales[0]] // all have same range (otherwise they would not have been combined), so take the first
	ticker := plot.tickerFor(scales)
	ticks := ticker.Ticks(scale.Limit.Min, scale.Limit.Max)

	boxSize, pad := plot.Style.Legend.Discrete.Size, vg.Length(3)
	r := vg.Rectangle{
		Min: vg.Point{c.Min.X, c.Max.Y - boxSize},
		Max: vg.Point{c.Min.X + boxSize, c.Max.Y},
	}

	labelSty := plot.Style.Legend.Label
	labelSty.XAlign = draw.XLeft

	var pal []color.Color
	if showColor || showFill {
		cm := plot.colorMapFor(scales)
		pal = cm.Palette(len(ticks)).Colors()
	}

	shape := draw.GlyphDrawer(draw.CircleGlyph{})
	basecolor := plot.Style.GeomDefault.Color
	size := boxSize / 5

	for i, tick := range ticks {
		if tick.Label == "" {
			debug.VV("skiping tick at", tick.Value)
			continue
		}
		debug.VV("tick", tick.Label, "@", tick.Value)
		// The background box.
		c.SetColor(color.Gray{0xee})
		c.Fill(r.Path())

		center := vg.Point{X: (r.Min.X + r.Max.X) / 2, Y: (r.Min.Y + r.Max.Y) / 2}

		col := basecolor
		// The actual indicators.
		if pal != nil {
			col = pal[i]
		}
		if showAlpha {
			r, g, b, a := col.RGBA()
			alpha := plot.Scales[AlphaScale].Map(tick.Value)
			col = color.NRGBA64{uint16(r), uint16(g), uint16(b), uint16(float64(a) * alpha)}
		}

		if showSize {
			size = plot.MapSize(tick.Value)
		}
		if showShape {
			shape = plotutil.Shape(i)
		}

		if showStroke {
			lsty := draw.LineStyle{
				Color:  col,
				Width:  1,
				Dashes: plotutil.Dashes(i),
			}
			c.StrokeLine2(lsty, r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)
		}

		// Do not draw the shape if not needed.
		if showShape || showFill || showSize || showColor || (showAlpha && !showStroke) {
			gsty := draw.GlyphStyle{
				Color:  col,
				Radius: size,
				Shape:  shape,
			}
			c.DrawGlyph(gsty, center)
		}
		// The label.
		c.FillText(labelSty, vg.Point{r.Max.X + pad, (r.Min.Y + r.Max.Y) / 2}, tick.Label)

		// The box border
		c.SetColor(color.Black)
		c.SetLineDash(nil, 0)
		c.SetLineWidth(vg.Length(0.3))
		c.Stroke(r.Path())

		r.Min.Y -= boxSize + pad
		r.Max.Y -= boxSize + pad
	}

	return r.Min.Y + boxSize - 2*pad
}

func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func (p *Plot) drawDiscreteColorGuide(c draw.Canvas, fill bool) vg.Length {
	scale := p.Scales[ColorScale]
	cm := p.ColorMap
	if fill {
		scale = p.Scales[FillScale]
		cm = p.FillMap
	}
	a, e := int(scale.Data.Min), int(scale.Data.Max)
	size, pad := p.Style.Legend.Discrete.Size, vg.Length(3)
	r := vg.Rectangle{
		Min: vg.Point{c.Min.X, c.Max.Y - size},
		Max: vg.Point{c.Min.X + size, c.Max.Y},
	}

	labelSty := p.Style.Legend.Label
	labelSty.XAlign = draw.XLeft

	for level := e; level >= a; level-- {
		col, _ := cm.At(scale.Map(float64(level)))
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

func (p *Plot) drawContinuousColorGuide(c draw.Canvas, scale *Scale, colMap palette.ColorMap) vg.Length {
	width := p.Style.Legend.Continuous.Size
	height := p.Style.Legend.Continuous.Length
	scale2Canvas := func(x float64) vg.Length {
		t := scale.Map(x)
		return c.Max.Y - height + height*vg.Length(t)
	}
	rect := vg.Rectangle{
		Min: vg.Point{c.Min.X, scale2Canvas(scale.Limit.Min)},
		Max: vg.Point{c.Min.X + width, scale2Canvas(scale.Limit.Max)},
	}
	step := height / 101
	r := rect
	for i := 0; i <= 100; i++ {
		col, err := colMap.At(float64(i) / 100)
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
	ticks := plot.DefaultTicks{}.Ticks(scale.Limit.Min, scale.Limit.Max)
	for _, tick := range ticks {
		if tick.IsMinor() {
			continue
		}
		sty := p.Style.Legend.Continuous.Tick.LineStyle
		length := p.Style.Legend.Continuous.Tick.Length
		align := vg.Length(p.Style.Legend.Continuous.Tick.Align)
		y := scale2Canvas(tick.Value)
		x := rect.Max.X
		c.StrokeLine2(sty, x-align*length, y, x+(1-align)*length, y)

		if p.Style.Legend.Continuous.Tick.Mirror {
			x := rect.Min.X
			c.StrokeLine2(sty, x+(align-1)*length, y, x+align*length, y)
		}
		tsty := p.Style.Legend.Label
		tsty.XAlign = draw.XLeft
		c.FillText(tsty,
			vg.Point{x + (1-align)*length, y}, " "+tick.Label)
	}

	return rect.Min.Y
}
