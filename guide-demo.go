// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/vdobler/facet"
	"github.com/vdobler/facet/geom"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

var xyz = plotter.XYZs{
	{1, 1, 1},
	{2, 2, 2},
	{3, 3, 3},
	{4, 4, 4},
	{5, 5, 5},
	{6, 6, 6},
	{7, 7, 7},
	{8, 8, 8},
	{9, 9, 9},
}

func points(size, color, symbol bool) facet.Geom {
	p := geom.Point{XY: plotter.XYValues{xyz}}

	if size {
		p.Size = func(i int) float64 { return xyz[i].Z }
	}
	if color {
		p.Color = func(i int) float64 { return xyz[i].Z }
	}
	if symbol {
		p.Symbol = func(i int) float64 { return xyz[i].Z }
	}

	return p
}

func linespoints(size, color, symbol, style bool) facet.Geom {
	lp := geom.LinesPoints{
		XY: make([]plotter.XYer, len(xyz)),
	}

	for i, v := range xyz {
		lp.XY[i] = plotter.XYs{
			{v.X, v.Y},
			{v.X + 3, v.Y + 1},
			{v.X + 5, v.Y - 2},
		}
	}

	if size {
		lp.Size = func(i int) float64 { return float64(i) }
	}
	if color {
		lp.Color = func(i int) float64 { return float64(i) }
	}
	if symbol {
		lp.Symbol = func(i int) float64 { return float64(i) }
	}
	if style {
		lp.Style = func(i int) float64 { return float64(i) }
	}

	return lp
}

func sample(size, color, symbol, style bool) *facet.Plot {
	f := facet.NewSimplePlot()

	features := []string{}
	if size {
		features = append(features, "Size")
	}
	if color {
		features = append(features, "Color")
	}
	if symbol {
		features = append(features, "Symbol")
	}
	if style {
		features = append(features, "Style")
	}
	if len(features) == 0 {
		features = append(features, "-none-")
	}
	f.Title = strings.Join(features, ", ")

	rainbow := &facet.Rainbow{Saturation: 0.9, Value: 0.9}
	rainbow.SetAlpha(1)
	rainbow.HueGap = 1.0 / 6.0
	rainbow.StartHue = 2.0 / 6.0

	f.Scales[facet.ColorScale].ScaleType = facet.Linear
	f.Scales[facet.ColorScale].ColorMap = rainbow

	f.Scales[facet.SymbolScale].ScaleType = facet.Linear

	f.Panels[0][0].Geoms = []facet.Geom{linespoints(size, color, symbol, style)}

	return f
}

func main() {
	for m := uint(0); m < 16; m++ {
		fmt.Println()
		fmt.Println("====== ", m, " ======")
		img := vgimg.New(400, 300)
		dc := draw.New(img)
		c := dc
		// c.Max.X, c.Max.Y = 300, 200
		fmt.Println("Min", c.Min, "  Max", c.Max)
		f := sample(m&0x01 != 0, m&0x02 != 0, m&0x04 != 0, m&0x08 != 0)
		f.Range()
		f.Draw(c)
		if c.Max.X < 900 {
			c.Min.X += 300
			c.Max.X += 300
		} else {
			c.Min.Y += 200
			c.Max.Y += 200
			c.Min.X = 0
			c.Max.X = 300
		}
		write(img, fmt.Sprintf("testdata/guide-%02d.png", m))
	}

}

func write(canvas *vgimg.Canvas, name string) {
	w, err := os.Create(name)
	defer w.Close()
	if err != nil {
		panic(err)
	}
	png := vgimg.PngCanvas{Canvas: canvas}
	if _, err = png.WriteTo(w); err != nil {
		panic(err)
	}
	if err = w.Close(); err != nil {
		panic(err)
	}
}
