// +build ignore

package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/vdobler/facet"
	"github.com/vdobler/facet/geom"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

var xy = plotter.XYs{}

func init() {
	xy = make(plotter.XYs, 50)
	x := 10.0
	for i := range xy {
		t := x + rand.NormFloat64() + x/10
		u := float64(i) + 2*rand.NormFloat64()
		xy[i].X, xy[i].Y = t, u
		x *= 1.2
	}
}

func main() {
	xy = make(plotter.XYs, 11)
	for i := range xy {
		xy[i].X = float64(3 * i)
		xy[i].Y = 5 + float64(i)
	}

	f := facet.NewSimplePlot()
	f.Panels[0][0].Geoms = []facet.Geom{
		geom.Point{
			XY:   xy,
			Size: func(i int) float64 { return xy[i].Y },
		},
	}
	img := vgimg.New(900, 600)
	dc := draw.New(img)

	// Atoscaleing
	f.Title = "Default, SqrtTrans"
	f.Prepare()
	f.Scales[facet.SizeScale].Trans = facet.SqrtTrans
	dc.Rectangle.Min.X, dc.Rectangle.Max.X = 0, 300
	dc.Rectangle.Min.Y, dc.Rectangle.Max.Y = 400, 600
	f.Draw(dc)

	f.Title = "Default, SqrtTransFix0"
	f.Prepare()
	f.Scales[facet.SizeScale].Trans = facet.SqrtTransFix0
	dc.Rectangle.Min.X, dc.Rectangle.Max.X = 300, 600
	dc.Rectangle.Min.Y, dc.Rectangle.Max.Y = 400, 600
	f.Draw(dc)

	f.Title = "Default, LinearTrans"
	f.Prepare()
	f.Scales[facet.SizeScale].Trans = facet.LinearTrans
	dc.Rectangle.Min.X, dc.Rectangle.Max.X = 600, 900
	dc.Rectangle.Min.Y, dc.Rectangle.Max.Y = 400, 600
	f.Draw(dc)

	// Expanded Limits
	f.Title = "Limit [0, 20], SqrtTrans"
	f.Prepare()
	f.Scales[facet.SizeScale].Trans = facet.SqrtTrans
	f.Scales[facet.SizeScale].Limit.Min = 0
	f.Scales[facet.SizeScale].Limit.Max = 20
	dc.Rectangle.Min.X, dc.Rectangle.Max.X = 0, 300
	dc.Rectangle.Min.Y, dc.Rectangle.Max.Y = 200, 400
	f.Draw(dc)

	f.Title = "Limit [0, 20], SqrtTransFix0"
	f.Prepare()
	f.Scales[facet.SizeScale].Trans = facet.SqrtTransFix0
	f.Scales[facet.SizeScale].Limit.Min = 0
	f.Scales[facet.SizeScale].Limit.Max = 20
	dc.Rectangle.Min.X, dc.Rectangle.Max.X = 300, 600
	dc.Rectangle.Min.Y, dc.Rectangle.Max.Y = 200, 400
	f.Draw(dc)

	f.Title = "Limit [0, 20], LinearTrans"
	f.Prepare()
	f.Scales[facet.SizeScale].Limit.Min = 0
	f.Scales[facet.SizeScale].Limit.Max = 20
	f.Scales[facet.SizeScale].Trans = facet.LinearTrans
	dc.Rectangle.Min.X, dc.Rectangle.Max.X = 600, 900
	dc.Rectangle.Min.Y, dc.Rectangle.Max.Y = 200, 400
	f.Draw(dc)

	// Expanded Range
	f.Title = "Range [0, 20], SqrtTrans"
	f.Prepare()
	f.Scales[facet.SizeScale].Trans = facet.SqrtTrans
	f.Scales[facet.SizeScale].Range.Min = 0
	f.Scales[facet.SizeScale].Range.Max = 20
	dc.Rectangle.Min.X, dc.Rectangle.Max.X = 0, 300
	dc.Rectangle.Min.Y, dc.Rectangle.Max.Y = 0, 200
	f.Draw(dc)

	f.Title = "Range [0, 20], SqrtTransFix0"
	f.Prepare()
	f.Scales[facet.SizeScale].Trans = facet.SqrtTransFix0
	f.Scales[facet.SizeScale].Range.Min = 0
	f.Scales[facet.SizeScale].Range.Max = 20
	dc.Rectangle.Min.X, dc.Rectangle.Max.X = 300, 600
	dc.Rectangle.Min.Y, dc.Rectangle.Max.Y = 0, 200
	f.Draw(dc)

	f.Title = "Range [0, 20], LinearTrans"
	f.Prepare()
	f.Scales[facet.SizeScale].Range.Min = 0
	f.Scales[facet.SizeScale].Range.Max = 20
	f.Scales[facet.SizeScale].Trans = facet.LinearTrans
	dc.Rectangle.Min.X, dc.Rectangle.Max.X = 600, 900
	dc.Rectangle.Min.Y, dc.Rectangle.Max.Y = 0, 200
	f.Draw(dc)

	write(img, fmt.Sprintf("testdata/size.png"))
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
