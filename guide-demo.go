// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/vdobler/facet"
	"github.com/vdobler/facet/data"
	"github.com/vdobler/facet/geom"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

var xy = plotter.XYs{
	{1, 1},
	{2, 2},
	{3, 3},
	{4, 4},
	{5, 5},
	{6, 6},
	{7, 7},
	{8, 8},
	{9, 9},
}

func points(alpha, color, fill, shape, size, stroke bool) facet.Geom {
	p := geom.Point{XY: xy}

	if alpha {
		p.Alpha = func(i int) float64 { return float64(i) }
	}
	if color {
		p.Color = func(i int) float64 { return float64(i) }
	}
	if shape {
		p.Shape = func(i int) int { return i }
	}
	if size {
		p.Size = func(i int) float64 { return float64(i) }
	}

	return p
}

func segments(alpha, color, fill, shape, size, stroke bool) facet.Geom {
	xyuv := make(data.XYUVs, len(xy))
	for i := range xyuv {
		xyuv[i].X = xy[i].X - 0.5
		xyuv[i].U = xy[i].X + 0.5

		if i%2 == 0 {
			xyuv[i].Y = xy[i].X - 0.5
			xyuv[i].V = xy[i].X + 0.5
		} else {
			xyuv[i].Y = xy[i].X + 0.5
			xyuv[i].V = xy[i].X - 0.5

		}
	}
	seg := geom.Segment{
		XYUV: xyuv,
	}

	if alpha {
		seg.Alpha = func(i int) float64 { return float64(i) }
	}
	if color {
		seg.Color = func(i int) float64 { return float64(i) }
	}
	if size {
		seg.Size = func(i int) float64 { return float64(i) }
	}
	if stroke {
		seg.Stroke = func(i int) int { return i }
	}

	return seg
}

func rectangles(alpha, color, fill, shape, size, stroke bool) facet.Geom {
	xyuv := make(data.XYUVs, len(xy))
	for i := range xyuv {
		xyuv[i].X = xy[i].X - 0.4
		xyuv[i].U = xy[i].X + 0.4
		xyuv[i].Y = 10 - xy[i].X
		xyuv[i].V = 9 - xy[i].X
	}
	rect := geom.Rectangle{
		XYUV: xyuv,
	}

	if alpha {
		rect.Alpha = func(i int) float64 { return float64(i) }
	}
	if color {
		rect.Color = func(i int) float64 { return float64(i) }
	}
	if fill {
		rect.Fill = func(i int) float64 { return float64(i) }
	}
	if size {
		rect.Size = func(i int) float64 { return float64(i) }
	}
	if stroke {
		rect.Stroke = func(i int) int { return i }
	}

	return rect
}

func sample(alpha, color, fill, shape, size, stroke bool) *facet.Plot {
	f := facet.NewSimplePlot()

	features := []string{}
	if alpha {
		features = append(features, "Alpha")
	}
	if color {
		features = append(features, "Color")
	}
	if fill {
		features = append(features, "Fill")
	}
	if shape {
		features = append(features, "Shape")
	}
	if size {
		features = append(features, "Size")
	}
	if stroke {
		features = append(features, "Stroke")
	}
	if len(features) == 0 {
		features = append(features, "-none-")
	}
	f.Title = strings.Join(features, ", ")

	rainbow := &facet.Rainbow{Saturation: 0.9, Value: 0.9}
	rainbow.SetAlpha(1)
	rainbow.HueGap = 1.0 / 6.0
	rainbow.StartHue = 0.5 / 6.0

	// f.Scales[facet.ColorScale].ScaleType = facet.Linear
	f.ColorMap = rainbow
	f.FillMap = rainbow

	f.Scales[facet.ShapeScale].ScaleType = facet.Linear

	f.Panels[0][0].Geoms = []facet.Geom{
		rectangles(alpha, color, fill, shape, size, stroke),
		segments(alpha, color, fill, shape, size, stroke),
		points(alpha, color, fill, shape, size, stroke),
	}

	return f
}

func main() {
	for m := uint(0); m <= 64; m++ {
		fmt.Println()
		alpha, color, fill, shape, size, stroke := m&0x01 != 0, m&0x02 != 0, m&0x04 != 0, m&0x08 != 0, m&0x10 != 0, m&0x20 != 0
		fmt.Println("====== ", m, " ======")
		img := vgimg.New(600, 480)
		dc := draw.New(img)
		c := dc
		f := sample(alpha, color, fill, shape, size, stroke)
		f.Prepare()
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
