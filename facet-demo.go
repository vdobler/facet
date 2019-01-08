// +build ignore

package main

import (
	"os"

	"github.com/vdobler/facet"
	"github.com/vdobler/facet/geom"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func main() {
	f := facet.NewFacet(2, 3, true, true)

	f.Title = "Facet"
	f.XScales[0].Title = "X-Axis"
	f.YScales[0].Title = "Y-Axis"
	f.RowLabels[0] = "(R) 0 j g"
	f.RowLabels[1] = "R 1"
	f.ColLabels[0] = "C 0"
	f.ColLabels[1] = "(C) 1 j g"
	f.ColLabels[2] = "C 2"

	rainbow := &facet.Rainbow{Saturation: 0.9, Value: 0.9}
	rainbow.SetAlpha(1)
	rainbow.HueGap = 1.0 / 6.0
	rainbow.StartHue = 2.0 / 6.0
	f.Scales[facet.ColorScale].Title = "Rnbw"
	f.Scales[facet.ColorScale].ScaleType = facet.Linear
	f.Scales[facet.ColorScale].ColorMap = rainbow

	f.Scales[facet.FillScale].Title = "Heat"
	f.Scales[facet.FillScale].ScaleType = facet.Linear
	f.Scales[facet.FillScale].ColorMap = moreland.Kindlmann()

	f.Scales[facet.SymbolScale].Title = "User"
	f.Scales[facet.SymbolScale].ScaleType = facet.Discrete

	// Rectangles
	xyuv := []geom.XYUV{
		{10, 10, 20, 15},
		{5, 0, 15, 8},
		{14, 7, 18, 20},
	}
	f.Panels[0][0].Geoms = []facet.Geom{
		geom.Rectangle{XYUV: xyuv},
	}

	// Bubble plot
	xyz := plotter.XYZs{
		{3.0, 2.0, -4},
		{4.0, 1.0, -2},
		{4.8, 3.0, -3},
		{5.2, 4.0, 0},
		{6.5, 3.5, 2},
		{7.0, 4.0, 1},
		{7.5, 5.0, 2},
		{8.0, 4.5, 3},
	}
	f.Panels[1][1].Geoms = []facet.Geom{
		geom.Point{
			XY:   plotter.XYValues{xyz},
			Size: func(i int) float64 { return xyz[i].Z },
			Color: func(i int) float64 {
				k := (i + 4) % len(xyz)
				return 7 + xyz[k].Z
			},
		},
	}

	// Lines plot
	exp1 := plotter.XYs{
		{2, 10},
		{3, 12},
		{4, 13},
		{5, 18},
		{7, 17},
	}
	exp2 := plotter.XYs{
		{3, -2},
		{4, 0},
		{5, 3},
		{6, 6},
		{7, 6.5},
		{7.5, 9},
	}
	exp3 := plotter.XYs{
		{2, 15},
		{4, 10},
		{6, 7},
		{8, 0},
	}
	f.Panels[0][1].Geoms = []facet.Geom{
		geom.Lines{
			XY: []plotter.XYer{exp1, exp2, exp3},
			Color: func(i int) float64 {
				return float64(i)
			},
			Style: func(i int) float64 {
				return float64(i)
			},
			Size: func(i int) float64 {
				return float64(i + 2)
			},
		},
	}

	img := vgimg.New(600, 600)
	dc := draw.New(img)
	f.Range()
	f.Draw(dc)

	w, err := os.Create("testdata/facet.png")
	defer w.Close()
	if err != nil {
		panic(err)
	}
	png := vgimg.PngCanvas{Canvas: img}
	if _, err = png.WriteTo(w); err != nil {
		panic(err)
	}
	if err = w.Close(); err != nil {
		panic(err)
	}
}
