// Packag data contains various data interfaces and prototypical
// implementations.
package data

import "math"

// XYUVer wraps the Len and XYUV methods.
type XYUVer interface {
	// Len returns the number of x, y, u, v quadruples.
	Len() int

	// XYUV returns an x, y, u, v quadruple.
	XYUV(int) (x, y, u, v float64)
}

// XYUVRange returns the minimum and maximum x, y, u and v values.
func XYUVRange(xyuvs XYUVer) (xmin, xmax, ymin, ymax, umin, umax, vmin, vmax float64) {
	xmin, xmax = math.Inf(1), math.Inf(-1)
	ymin, ymax = math.Inf(1), math.Inf(-1)
	umin, umax = math.Inf(1), math.Inf(-1)
	vmin, vmax = math.Inf(1), math.Inf(-1)
	for i := 0; i < xyuvs.Len(); i++ {
		x, y, u, v := xyuvs.XYUV(i)
		xmin, xmax = math.Min(xmin, x), math.Max(xmax, x)
		ymin, ymax = math.Min(ymin, y), math.Max(ymax, y)
		umin, umax = math.Min(umin, u), math.Max(umax, u)
		vmin, vmax = math.Min(vmin, v), math.Max(vmax, v)
	}
	return xmin, xmax, ymin, ymax, umin, umax, vmin, vmax
}

// XYUVs implements the XYUVer interface.
type XYUVs []struct{ X, Y, U, V float64 }

func (d XYUVs) Len() int                        { return len(d) }
func (d XYUVs) XYUV(i int) (x, y, u, v float64) { return d[i].X, d[i].Y, d[i].U, d[i].V }
