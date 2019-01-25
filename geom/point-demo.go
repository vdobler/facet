// +build ignore

package main

import (
	"os"

	"github.com/vdobler/facet"
	"github.com/vdobler/facet/geom"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func main() {
	plot := facet.NewPlot(5, 5, true, true)

	plot.Title = "Geom Points"
	plot.XScales[0].Title = "X-Axis"
	plot.YScales[0].Title = "Y-Axis"

	plot.Scales[facet.AlphaScale].Title = "Alpha"
	plot.Scales[facet.ColorScale].Title = "Color"
	plot.Scales[facet.ShapeScale].Title = "Shape"
	plot.Scales[facet.ShapeScale].ScaleType = facet.Discrete
	plot.Scales[facet.SizeScale].Title = "Size"

	xyz := plotter.XYZs{
		{1, 1, 1},
		{2, 2, 2},
		{3, 3, 3},
		{4, 4, 4},
		{5, 3, 6},
		{6, 2, 8},
		{7, 1, 10},
	}

	plot.Panels[0][0].Geoms = []facet.Geom{
		geom.Point{
			XY: plotter.XYValues{xyz},
		},
	}
	plot.Panels[1][0].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Alpha: func(i int) float64 { return 10 - xyz[i].Z },
		},
	}
	plot.Panels[2][0].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Color: func(i int) float64 { return xyz[i].Z / 2 },
		},
	}
	plot.Panels[3][0].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Shape: func(i int) int { return i },
		},
	}
	plot.Panels[4][0].Geoms = []facet.Geom{
		geom.Point{
			XY:   plotter.XYValues{xyz},
			Size: func(i int) float64 { return xyz[i].Z * 2 },
		},
	}

	plot.Panels[0][1].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Alpha: func(i int) float64 { return xyz[i].Z },
			Color: func(i int) float64 { return xyz[i].Z / 2 },
		},
	}
	plot.Panels[1][1].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Alpha: func(i int) float64 { return 10 - xyz[i].Z },
			Shape: func(i int) int { return i },
		},
	}
	plot.Panels[2][1].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Alpha: func(i int) float64 { return 10 - xyz[i].Z },
			Size:  func(i int) float64 { return xyz[i].Z * 2 },
		},
	}
	plot.Panels[3][1].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Color: func(i int) float64 { return xyz[i].Z / 2 },
			Size:  func(i int) float64 { return xyz[i].Z * 2 },
		},
	}
	plot.Panels[4][1].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Color: func(i int) float64 { return xyz[i].Z / 2 },
			Shape: func(i int) int { return i },
		},
	}

	plot.Panels[0][2].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Alpha: func(i int) float64 { return 10 - xyz[i].Z },
			Color: func(i int) float64 { return xyz[i].Z / 2 },
			Shape: func(i int) int { return i },
		},
	}
	plot.Panels[1][2].Geoms = []facet.Geom{
		geom.Point{
			XY:    plotter.XYValues{xyz},
			Alpha: func(i int) float64 { return 10 - xyz[i].Z },
			Color: func(i int) float64 { return xyz[i].Z / 2 },
			Size:  func(i int) float64 { return xyz[i].Z * 2 },
		},
	}

	img := vgimg.New(1000, 800)
	dc := draw.New(img)
	plot.Range()
	plot.Draw(dc)

	w, err := os.Create("testdata/point.png")
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
