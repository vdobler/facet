// +build ignore

package main

import (
	"image/color"
	"os"

	"github.com/vdobler/facet"
	"github.com/vdobler/facet/geom"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func main() {
	f := facet.NewPlot(3, 2, true, true)

	f.Title = "Facet"
	f.XScales[0].Title = "X-Axis"
	f.YScales[0].Title = "Y-Axis"
	f.RowLabels[0] = "Stack"
	f.RowLabels[1] = "Fill"
	f.RowLabels[2] = "Dodge"
	f.ColLabels[0] = "A"
	f.ColLabels[1] = "B"

	rainbow := &facet.Rainbow{Saturation: 0.9, Value: 0.9}
	rainbow.SetAlpha(1)
	rainbow.HueGap = 1.0 / 6.0
	rainbow.StartHue = 0 // 2.0 / 6.0
	f.Scales[facet.ColorScale].Title = "Color"
	f.Scales[facet.ColorScale].ScaleType = facet.Linear
	f.Scales[facet.ColorScale].ColorMap = moreland.Kindlmann()

	f.Scales[facet.FillScale].Title = "Fill"
	f.Scales[facet.FillScale].ScaleType = facet.Discrete
	f.Scales[facet.FillScale].ColorMap = rainbow

	f.Scales[facet.SizeScale].Title = "Size"
	f.Scales[facet.StyleScale].Title = "Style"

	f.Scales[facet.SymbolScale].Title = "Symbol"
	f.Scales[facet.SymbolScale].ScaleType = facet.Discrete

	xyz := plotter.XYZs{
		{10, 5, 1},
		{20, 3, 1},
		{30, 7, 1},
		{40, 2, 1},
		{50, 6, 1},
		{10, 2, 2},
		{20, 4, 2},
		{30, 1, 2},
		{40, 3, 2},
		{50, 5, 2},
		{20, 4, 3},
		{40, 2, 3},
		{50, 1, 3},
	}

	// First Column
	f.Panels[0][0].Geoms = []facet.Geom{
		geom.Bar{
			XY:       plotter.XYValues{xyz},
			Fill:     func(i int) float64 { return xyz[i].Z },
			Position: "stack",
		},
	}
	f.Panels[1][0].Geoms = []facet.Geom{
		geom.Bar{
			XY:       plotter.XYValues{xyz},
			Fill:     func(i int) float64 { return xyz[i].Z },
			Position: "fill",
		},
	}
	f.Panels[2][0].Geoms = []facet.Geom{
		geom.Bar{
			XY:       plotter.XYValues{xyz},
			Fill:     func(i int) float64 { return xyz[i].Z },
			Position: "dodge",
		},
	}

	// Second Column
	f.Panels[0][1].Geoms = []facet.Geom{
		geom.Bar{
			XY:       plotter.XYValues{xyz},
			Color:    func(i int) float64 { return xyz[i].Z },
			Position: "stack",
			Default: geom.BoxStyle{
				Fill: color.Transparent,
				Border: draw.LineStyle{
					Width: 6,
				},
			},
		},
	}
	f.Panels[1][1].Geoms = []facet.Geom{
		geom.Bar{
			XY:       plotter.XYValues{xyz},
			Size:     func(i int) float64 { return xyz[i].Z },
			Position: "fill",
		},
	}

	f.Panels[2][1].Geoms = []facet.Geom{
		geom.Bar{
			XY:       plotter.XYValues{xyz},
			Style:    func(i int) int { return int(xyz[i].Z) },
			Position: "dodge",
		},
	}

	img := vgimg.New(800, 600)
	dc := draw.New(img)
	f.Range()
	f.Draw(dc)

	w, err := os.Create("testdata/bar.png")
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
