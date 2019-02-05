package geom

import (
	"image/color"
	"math"
	"reflect"

	"github.com/vdobler/facet"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// Aestetic is a function mapping a certain data point to an aestehtic.
type Aesthetic func(i int) float64

// DiscreteAestetic is a function mapping a certain data point to a discrete
// aesthetic like Shape or Stroke.
type DiscreteAesthetic func(i int) int

// UpdateAestheticsRanges is a helper to update the data ranges dr based on
// the non-nil aesthetics functions evaluated for all n data points.
func UpdateAestheticsRanges(dr *facet.DataRanges, n int,
	alpha Aesthetic,
	color Aesthetic,
	fill Aesthetic,
	shape DiscreteAesthetic,
	size Aesthetic,
	stroke DiscreteAesthetic) {

	for i := 0; i < n; i++ {
		if alpha != nil {
			dr[facet.AlphaScale].Update(alpha(i))
		}
		if color != nil {
			dr[facet.ColorScale].Update(color(i))
		}
		if fill != nil {
			dr[facet.FillScale].Update(fill(i))
		}
		if shape != nil {
			dr[facet.ShapeScale].Update(float64(shape(i)))
		}
		if size != nil {
			dr[facet.SizeScale].Update(size(i))
		}
		if stroke != nil {
			dr[facet.StrokeScale].Update(float64(stroke(i))) // TODO: StrokeScale should be discrete from the start
		}
	}
}

// CopyAesthetics copies the non-nil aesthetics from src to dst.
// The destination must be a pointer to a struct, the source may be a struct
// or a pointer to one.
// The index function can be used to reindex the aestetics functions between
// src and dst.
func CopyAesthetics(dst, src interface{}, index func(int) int) {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}
	dstVal := reflect.ValueOf(dst).Elem()

	for _, aes := range []string{"Alpha", "Color", "Fill", "Shape", "Size", "Stroke"} {
		srcAes := srcVal.FieldByName(aes)
		if !srcAes.IsValid() {
			continue
		}
		dstAes := dstVal.FieldByName(aes)
		if !dstAes.IsValid() {
			continue
		}

		if index == nil || srcAes.IsNil() {
			dstAes.Set(srcAes)
			continue
		}

		f := reflect.MakeFunc(srcAes.Type(), func(in []reflect.Value) []reflect.Value {
			n := int(in[0].Int())
			m := index(n)
			val := srcAes.Call([]reflect.Value{reflect.ValueOf(m)})
			return val
		})
		dstAes.Set(f)
	}
}

// BoxStyle combines a line style (for the border with a fill color for
// the interior of a geom.  TODO: add the rest like size, symbol too?
type BoxStyle struct {
	Fill   color.Color
	Border draw.LineStyle
}

// CanonicRectangle returns the canonical form of r, i.e. its Min points
// having smaller coordinates than its Max point.
func CanonicRectangle(r vg.Rectangle) vg.Rectangle {
	if r.Min.X > r.Max.X {
		r.Min.X, r.Max.X = r.Max.X, r.Min.X
	}
	if r.Min.Y > r.Max.Y {
		r.Min.Y, r.Max.Y = r.Max.Y, r.Min.Y
	}
	return r
}

func determineColor(col color.Color, panel *facet.Panel, i int, colorF, alphaF Aesthetic) (color.Color, bool) {
	if colorF != nil {
		col = panel.MapColor(colorF(i))
	}

	if col == nil {
		return col, false
	}

	if alphaF != nil {
		alpha := panel.Scales[facet.AlphaScale].Map(alphaF(i))
		if alpha < 0 || alpha > 1 || math.IsNaN(alpha) {
			return col, false
		}
		r, g, b, a := col.RGBA()
		col = color.NRGBA64{
			uint16(r),
			uint16(g),
			uint16(b),
			uint16(float64(a) * alpha),
		}
	}

	return col, true
}
