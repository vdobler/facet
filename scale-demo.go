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
	p := geom.Point{XY: xy}
	p.Size = func(i int) float64 { return float64(i) }

	f := facet.NewSimplePlot()

	f.Title = "Scaling"

	rainbow := &facet.Rainbow{Saturation: 0.9, Value: 0.9}
	rainbow.SetAlpha(1)
	rainbow.HueGap = 1.0 / 6.0
	rainbow.StartHue = 0.5 / 6.0

	// f.Scales[facet.ColorScale].ScaleType = facet.Linear
	f.ColorMap = rainbow
	f.FillMap = rainbow

	f.Scales[facet.ShapeScale].ScaleType = facet.Linear

	f.Panels[0][0].Geoms = []facet.Geom{p}
	f.XScales[0].Trans = facet.Log10Trans
	f.XScales[0].Expand.Releative = 0.05
	img := vgimg.New(600, 480)
	dc := draw.New(img)
	f.Prepare()
	f.Draw(dc)
	write(img, fmt.Sprintf("testdata/scale-00.png"))
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
