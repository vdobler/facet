// Scales
//
// Scales have a wide varity of ranges.
//   - Data: The range covered by actual data points
//       This is learend by scanning during Range()
//   - Interval (Limits): Data in this range is processed
//       This range is determined by autoscaling from the learend
//       data range or set manually.
//   - Range: The range of the guides (axis and legends)
//       This normaly is the same as the Interval/Limits. It is the
//       range actually drawn which might clip processed data points
//       or add margin.
//   - Output: This are the pixel ranges on screen for position scales
//     and the Size scale,
// Ticks are generated for the Interval (but ticks outside of Range
// are not drawn).
//
// The distinction of Interval/Limits and range for a position scale
// works like this.
//    Data == [10, 30]
//      |
//      | Autoscaling, relative expansion 5% = 1
//      V
//    Interval == [9, 31]  --> Ticks generated: 10, 20, 30
//      |
//      | Range changed manualy to
//      V
//    Range == [15, 45]
//      The plot area is clipped below 15.
//      Ticks stay the same, only 20 and 30 are drawn.
//
// Being able to limit the Range is useful to "zoom in" into a plot.
// Expanding the range beyond the Limit is useful for the Size scale:
// If your Limits are [20, 40] and you map that directly to point radius
// from [2px, 20px] a point representing a data value of 20 will be
// 10 times smaller in radius and 100 times smaller in area than a data
// point with value 40. Here you might set the range to [0, 40] but still
// draw breaks/ticks and thus legend entries only for 20, 25, 30, 35 and 40.
//
// A Transformation is responsible for two things:
//   - Generate breaks for the given Interval/Limit
//   - Take values from Interval/Limits (or Range of different) and
//     map them to output coordinates
// Output coordinates differ for different scales:
//   - Position Scales: screen coordinates (x,y)
//   - Color Scales: [0, 1] used to index a ColorMap
//   - Size Scale: pixel size for radius or linewidth
//   - Shape and Stroke Scale: ???
package facet

import (
	"fmt"
	"math"
	"time"

	"gonum.org/v1/plot"
)

// ----------------------------------------------------------------------------
// Scale

// Scale is a generalizes axis: While a plot has exactly two axes (the x-axis
// and the y-axis) it can have more scales, e.g. a color scale, a linetype
// scale, a symbol scale or even a size scale.
type Scale struct {
	// Title is the scale's title.
	Title string

	// Data is the range covered by actual data (in data units). It can be
	// populated from the actual data via LearnDataRange.
	Data Interval

	// Limit captures the range (in data units) of this scale. It may be
	// larger or smaller than the actual Data range.
	Limit Interval

	// Range is the output range of this scale and will be in different
	// units depending on the scale type:
	//   - Screen coordinates for X and Y scales
	//   - Screen length for Size scale
	//   - Opacity (between 0 and 1) for the Alpha scale.
	//   - Color (between 0 and 1) for Color and Fill scale
	//   - Integer between 0 and N for discrete scales
	Range Interval

	// Trans implements the mapping from Interval (Limits) to Range for
	// this scale.
	Trans Transformation

	// ScaleType determines the fundamental nature of the scale.
	ScaleType ScaleType

	// Autoscaling can be used to control autoscaling of this scale.
	Autoscaling

	// Ticker is responsible for generating the ticks.
	Ticker plot.Ticker

	// Values contains the nominal values. TODO: replace by Ticker
	Values []string

	// TimeFmt is used to format date/time tics.
	TimeFmt string
	// T0 is the reference time and timezone
	T0 time.Time
}

// NewScale returns a new scale with all intervalls unset, an identitiy
// transformation and full, unexpanded autoscaling
func NewScale() *Scale {
	s := &Scale{
		Limit: UnsetInterval,
		Data:  UnsetInterval,
		Range: UnsetInterval,
		Trans: IdentityTrans,
	}
	s.Autoscaling.MinRange = UnsetInterval
	s.Autoscaling.MaxRange = UnsetInterval
	s.Autoscaling.Expand.Absolute = 0
	s.Autoscaling.Expand.Releative = 0

	s.Trans = IdentityTrans // make sure Trans is not nil

	return s
}

// Map maps the intervall [s.Min, s.Max] to [0, 1].
// Values outside of [s.Min, s.Max] are mapped to values < 0 or > 1.
// If s's Intervall is degenerate or unset Map returns NaN.
func (s *Scale) Map(x float64) float64 {
	U := Interval{0, 1}
	return s.Trans.Trans(s.Limit, U, x)

	// ======  OLD CODE =======
	if math.IsNaN(s.Limit.Min) || math.IsNaN(s.Limit.Max) || s.Limit.Min == s.Limit.Max {
		return math.NaN()
	}

	switch s.ScaleType {
	case Linear, Time, Discrete:
		return (x - s.Limit.Min) / (s.Limit.Max - s.Limit.Min)
	case Logarithmic:
		min, max := math.Log10(s.Limit.Min), math.Log10(s.Limit.Max)
		math.Log10(x)
		return (x - min) / (max - min)
	default:
		panic(s.ScaleType)
	}

}

// UpdateData updates s to cover i.
func (s *Scale) UpdateData(i Interval) {
	s.Data.Update(i.Min)
	s.Data.Update(i.Max)
}

// FixMin fixes the min of s to x. If x is NaN the min is determined by
// autoscaling to the actual data.
func (s *Scale) FixMin(x float64) {
	s.MinRange.Min = x
	s.MinRange.Max = x
}

// FixMax fixes the max of s to x. . If x is NaN the max is determined by
// autoscaling to the actual data.
func (s *Scale) FixMax(x float64) {
	s.MaxRange.Min = x
	s.MaxRange.Max = x
}

// HasData reports whether the Data intervall of s is valid.
func (s *Scale) HasData() bool {
	return !math.IsNaN(s.Data.Min) && !math.IsNaN(s.Data.Max)
}

// InRange reports whether x lies in the the range of s..
func (s *Scale) InRange(x float64) bool {
	return x >= s.Limit.Min && x <= s.Limit.Max
}

func (s *Scale) String() string {
	if s == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Data=[%.2f:%.2f] Limit=[%.2f:%.2f] Range=[%.2f:%.2f]  %s %q",
		s.Data.Min, s.Data.Max, s.Limit.Min, s.Limit.Max, s.Range.Min, s.Range.Max, s.ScaleType, s.Title)
}

func have(x float64) bool {
	return !math.IsNaN(x)
}

// Autoscale turns the Data range into an actual scale range.
func (s *Scale) Autoscale() {
	if !s.HasData() {
		return
	}

	// Determine the left edge of s.
	if s.MinRange.Min == s.MinRange.Max {
		// Degenerate MinRangeIntervall and non NaN:
		// The user has set a fixed Min.
		s.Limit.Min = s.MinRange.Min
	} else {
		s.Limit.Min = s.Data.Min
		s.applyExpansion(true)

		// Clip autoscaling
		if s.MinRange.Min > s.Limit.Min {
			s.Limit.Min = s.MinRange.Min
		}
		if s.MinRange.Max < s.Limit.Min {
			s.Limit.Min = s.MinRange.Max
		}
	}

	// Determine the right edge of s.
	if s.MaxRange.Min == s.MaxRange.Max {
		// Degenerate MaxRangeIntervall and non NaN:
		// The user has set a fixed Max.
		s.Limit.Max = s.MaxRange.Min
	} else {
		s.Limit.Max = s.Data.Max
		s.applyExpansion(false)

		// Clip autoscaling
		if s.MaxRange.Min > s.Limit.Max {
			s.Limit.Max = s.MaxRange.Min
		}
		if s.MaxRange.Max < s.Limit.Max {
			s.Limit.Max = s.MaxRange.Max
		}
	}

}

func (s *Scale) applyExpansion(min bool) {
	U := Interval{0, 1}
	if min {
		s.Limit.Min = s.Trans.Inverse(U, s.Data, -s.Expand.Releative)
		s.Limit.Min -= s.Expand.Absolute
	} else {
		s.Limit.Max = s.Trans.Inverse(U, s.Data, 1+s.Expand.Releative)
		s.Limit.Max += s.Expand.Absolute
	}
}

func (s *Scale) fillRange() {
	if math.IsNaN(s.Range.Min) {
		s.Range.Min = s.Limit.Min
	}
	if math.IsNaN(s.Range.Max) {
		s.Range.Max = s.Limit.Max
	}
}

// ----------------------------------------------------------------------------
// Intervall

// Interval represents a (potentially degenerate) real interval.
// Both edges of the interval may be NaN indicating this edge is not
// set determined.
type Interval struct {
	Min, Max float64
}

// UnsetInterval is the interval with unsepcified (NaN) endpoints.
var UnsetInterval = Interval{math.NaN(), math.NaN()}

// InfinitInterval is the interval [-Inf, +Inf] endpoints.
var InfinitInterval = Interval{math.Inf(-1), math.Inf(+1)}

// Update expands i to include x.
func (i *Interval) Update(x ...float64) {
	for _, v := range x {
		if math.IsNaN(v) {
			continue
		}
		// TODO: replace with math.Min/Max
		if !(i.Min < v) {
			i.Min = v
		}
		if !(i.Max > v) {
			i.Max = v
		}
	}
}

func (i Interval) Equal(j Interval) bool {
	if math.IsNaN(i.Min) {
		return math.IsNaN(j.Min)
	}
	if math.IsNaN(i.Max) {
		return math.IsNaN(j.Max)
	}
	return i.Min == j.Min && i.Max == j.Max
}

// Degenerate the intervall i by replacing NaN and Inf with -1 (for Min)
// and +1 (for Max) and by exapnding collapsed intervals of the form [a, a].
func (i *Interval) Degenerate() (modified bool) {
	mod := false
	if math.IsNaN(i.Min) || math.IsInf(i.Min, 0) {
		i.Min = -1
		mod = true
	}
	if math.IsNaN(i.Max) || math.IsInf(i.Max, 0) {
		i.Max = 1
		mod = true
	}

	if i.Min == i.Max {
		if i.Min == 0 {
			i.Min, i.Max = -1, +1
		} else {
			d := i.Min / 10
			i.Min -= d
			i.Max += d
		}
		mod = true
	}

	if i.Min > i.Max {
		i.Min, i.Max = i.Max, i.Min
		mod = true
	}

	return mod
}

// ----------------------------------------------------------------------------
// ScaleType

// ScaleType selects one of the handful know scale types.
type ScaleType int

// String returns the type of st.
func (st ScaleType) String() string {
	return []string{"linear", "discrete", "time", "log"}[int(st)]
}

const (
	Linear ScaleType = iota
	Discrete
	Time
	Logarithmic
)

// ----------------------------------------------------------------------------
// Autoscaling

// Autoscaling controls how the min and max value of a scale are scaled.
// Setting a range to a degenerate interval [f:f] will turn of autoscaling
// and fix the value to f. A non-degenerate range [u:v] will allow autoscaling
// between u and v. A NaN value works like -Inf for u and +Inf for v.
type Autoscaling struct {
	// Expand determines how much the actual data range is expandend.
	Expand struct {
		Absolute  float64
		Releative float64
	}

	MinRange Interval // MinRange determines the allowed range of the Min of a scale.
	MaxRange Interval // MaxRange determines the allowed range of the Max of a scale.
}
