// +build ignore

package main

import (
	"image/color"
	"os"

	"github.com/vdobler/facet"
	"github.com/vdobler/facet/data"
	"github.com/vdobler/facet/geom"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func main() {
	f := facet.NewPlot(2, 3, true, true)

	f.Title = "Facet"
	f.XScales[0].Title = "X-Axis"
	f.YScales[0].Title = "Y-Axis"
	f.RowLabels[0] = "Row 0"
	f.RowLabels[1] = "Row 1"
	f.ColLabels[0] = "Col 0"
	f.ColLabels[1] = "Col 1"
	f.ColLabels[2] = "Col 2"

	rainbow := &facet.Rainbow{Saturation: 0.9, Value: 0.9}
	rainbow.SetAlpha(1)
	rainbow.HueGap = 1.0 / 6.0
	rainbow.StartHue = 2.0 / 6.0
	f.Scales[facet.ColorScale].Title = "Color"
	f.Scales[facet.ColorScale].ScaleType = facet.Linear
	// f.Scales[facet.ColorScale].ColorMap = rainbow

	f.Scales[facet.FillScale].Title = "Fill"
	f.Scales[facet.FillScale].ScaleType = facet.Linear
	// f.Scales[facet.FillScale].ColorMap = moreland.Kindlmann()

	f.Scales[facet.SizeScale].Title = "Size"
	f.Scales[facet.StrokeScale].Title = "Stroke"

	f.Scales[facet.ShapeScale].Title = "Shape"
	f.Scales[facet.ShapeScale].ScaleType = facet.Discrete

	// Rectangles
	xyuv := data.XYUVs{
		{10, 10, 20, 15},
		{5, 0, 15, 8},
		{14, 7, 18, 20},
	}
	rectGeom := geom.Rectangle{XYUV: xyuv}
	rectGeom.Default.Fill = color.RGBA{0x77, 0xff, 0x77, 0xff}
	rectGeom.Default.Border.Color = color.RGBA{0, 0xcc, 0, 0xff}
	rectGeom.Default.Border.Width = 2
	f.Panels[0][0].Geoms = []facet.Geom{rectGeom}

	// Bubble plot
	xyz := plotter.XYZs{
		{3.0, 2.0, -4},
		{3.5, 2.5, -3},
		{4.0, 1.0, -2},
		{4.8, 3.0, -3},
		{5.2, 4.0, 0},
		{6.5, 3.5, 2},
		{7.0, 4.0, 1},
		{7.2, 3.3, 1.5},
		{7.5, 5.0, 2},
		{8.0, 4.5, 3},
	}
	f.Panels[1][1].Geoms = []facet.Geom{
		geom.Point{
			XY:   plotter.XYValues{xyz},
			Size: func(i int) float64 { return xyz[i].Z },
			Color: func(i int) float64 {
				return xyz[i].X + xyz[i].Y
			},
		},
		geom.HLine{
			Y: plotter.Values{4.2, 8.4, 10.5},
		},
		geom.VLine{
			X:       plotter.Values{2.5, 7.5},
			Default: draw.LineStyle{Color: color.RGBA{0xff, 0, 0, 0xff}},
		},
		geom.VLine{
			X:       plotter.Values{3.75, 5.25},
			Default: draw.LineStyle{Color: color.RGBA{0, 0xff, 0, 0xff}},
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
		geom.Line{XY: exp1,
			Color:  func(i int) float64 { return 1.0 },
			Stroke: func(i int) int { return 1 },
			Size:   func(i int) float64 { return 2.0 },
		},
		geom.Line{XY: exp2,
			Color:  func(i int) float64 { return 2.0 },
			Stroke: func(i int) int { return 2 },
			Size:   func(i int) float64 { return 4.0 },
		},
		geom.Line{XY: exp3,
			Color:  func(i int) float64 { return 3.0 },
			Stroke: func(i int) int { return 3 },
			Size:   func(i int) float64 { return 6.0 },
		},
		geom.Text{
			XYText: data.XYTexts{
				{5, 5, "Foo"},
				{7, 15, "Bar"},
				{3, 17, "Wok"},
				{8, 2, "Qux"},
			},
			Color: func(i int) float64 { return float64(i) * 4 },
			Size:  func(i int) float64 { return float64(i)*5 + 5 },
			Default: draw.TextStyle{
				XAlign: draw.XCenter,
				YAlign: draw.YCenter,
			},
		},
	}

	// Bar plot
	spending := plotter.XYZs{
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
	f.Panels[0][2].Geoms = []facet.Geom{
		geom.Bar{
			XY:       plotter.XYValues{spending},
			Fill:     func(i int) float64 { return spending[i].Z },
			Position: "stack",
		},
	}

	// Boxplot
	box := data.Boxplots{
		{6, 0, 2, 3, 5, 7, nil},
		{10, 3, 5, 7, 8, 10, []float64{2, 11, 11.5}},
		{14, 4, 5, 6, 7, 8, nil},
		{18, 3, 7, 9, 10, 13, []float64{1.5, 2, 14.5}},
		{10, 2, 3, 4, 6, 6, []float64{1, 7}},
		{14, 2.2, 3.5, 3.8, 4.1, 4.5, []float64{1, 2, 5, 5.5}},
	}
	f.Panels[1][0].Geoms = []facet.Geom{
		geom.Boxplot{
			Boxplot:  box,
			Position: "dodge",
			Color:    func(i int) float64 { return box[i].Median },
			Default: geom.BoxStyle{
				Fill: color.Transparent,
				Border: draw.LineStyle{
					Color: color.RGBA{0xcc, 0, 0, 0xff},
					Width: 2,
				},
			},
			DefaultPoint: draw.GlyphStyle{
				Color:  color.RGBA{0xcc, 0, 0, 0xff},
				Radius: 2,
			},
		},
	}

	img := vgimg.New(800, 600)
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
