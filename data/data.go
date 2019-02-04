// Packag data contains various data interfaces and prototypical
// implementations.
package data

import "math"

// ----------------------------------------------------------------------------
// (X,Y), (U,V)

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

// ----------------------------------------------------------------------------
// Text

// XYText wraps the Len and XYText methods.
type XYTexter interface {
	// Len returns the number of data points.
	Len() int

	// XYText returns a coordinate (x, y) and a string.
	XYText(int) (x, y float64, t string)
}

// XYTexts implements the XYTexter interface.
type XYTexts []struct {
	X, Y float64
	Text string
}

func (d XYTexts) Len() int                                 { return len(d) }
func (d XYTexts) XYText(i int) (x, y float64, text string) { return d[i].X, d[i].Y, d[i].Text }

// ----------------------------------------------------------------------------
// Boxplot

// Boxplotter wraps the Len and Boxplot methods.
type Boxplotter interface {
	// Len returns the number of boxes.
	Len() int

	// Boxplot returns an data for the i'th boxplot.
	Boxplot(i int) (x, min, q1, median, q3, max float64, outlier []float64)
}

// Boxplots implememnts the Boxplotter interface.
type Boxplots []struct {
	X                        float64
	Min, Q1, Median, Q3, Max float64
	Outlier                  []float64
}

func (b Boxplots) Len() int { return len(b) }

func (b Boxplots) Boxplot(i int) (x, min, q1, median, q3, max float64, outlier []float64) {
	return b[i].X, b[i].Min, b[i].Q1, b[i].Median, b[i].Q3, b[i].Max, b[i].Outlier
}
