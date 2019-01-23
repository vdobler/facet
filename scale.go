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

	// Data is the range covered by actual data.
	Data Interval

	// Interval captures the range of this scale. It may be larger or
	// smaller than the actual Data range.
	Interval

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

// NewScale returns a new linear scale which autoscales to the actual data.
func NewScale() *Scale {
	s := &Scale{
		Data:      unsetInterval(),
		Interval:  unsetInterval(),
		ScaleType: Linear,
		Autoscaling: Autoscaling{
			MinRange: unsetInterval(),
			MaxRange: unsetInterval(),
		},
	}
	s.Autoscaling.Expand.Releative = 0.05

	return s
}

// Map maps the intervall [s.Min, s.Max] to [0, 1].
// Values outside of [s.Min, s.Max] are mapped to values < 0 or > 1.
// If s's Intervall is degenerate or unset Map returns NaN.
func (s *Scale) Map(x float64) float64 {
	if math.IsNaN(s.Min) || math.IsNaN(s.Max) || s.Min == s.Max {
		return math.NaN()
	}

	switch s.ScaleType {
	case Linear, Time, Discrete:
		return (x - s.Min) / (s.Max - s.Min)
	case Logarithmic:
		min, max := math.Log10(s.Min), math.Log10(s.Max)
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
	return x >= s.Min && x <= s.Max
}

func (s *Scale) String() string {
	if s == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Range=[%.2f:%.2f] Data=[%.2f:%.2f] %s %q",
		s.Min, s.Max, s.Data.Min, s.Data.Max, s.ScaleType, s.Title)
}

func have(x float64) bool {
	return !math.IsNaN(x)
}

// Autoscale turns the data range into an actual scale range.
func (s *Scale) autoscale() {
	if !s.HasData() {
		return
	}

	ext := s.Expand.Releative*(s.Data.Max-s.Data.Min) + s.Expand.Absolut

	// Determine the left edge of s.
	if s.MinRange.Min == s.MinRange.Max {
		// Degenerate MinRangeIntervall and non NaN:
		// The user has set a fixed Min.
		s.Min = s.MinRange.Min
	} else {
		s.Min = s.Data.Min

		// Apply expansion.
		switch s.ScaleType {
		case Linear, Time:
			s.Min -= ext
		case Discrete:
			s.Min -= 0.5 + ext
		case Logarithmic:
			s.Min /= s.Expand.Absolut
			// TODO: relative
		default:
			panic(s.ScaleType)
		}

		// Clip autoscaling
		if s.MinRange.Min > s.Min {
			s.Min = s.MinRange.Min
		}
		if s.MinRange.Max < s.Min {
			s.Min = s.MinRange.Max
		}
	}

	// Determine the right edge of s.
	if s.MaxRange.Min == s.MaxRange.Max {
		// Degenerate MaxRangeIntervall and non NaN:
		// The user has set a fixed Max.
		s.Max = s.MaxRange.Min
	} else {
		s.Max = s.Data.Max

		// Apply expansion.
		switch s.ScaleType {
		case Linear, Time:
			s.Max += ext
		case Discrete:
			s.Max += 0.5 + ext
		case Logarithmic:
			s.Max *= s.Expand.Absolut
			// TODO: what if Absolut < 1? ANd handle relative.
		default:
			panic(s.ScaleType)
		}

		// Clip autoscaling
		if s.MaxRange.Min > s.Max {
			s.Max = s.MaxRange.Min
		}
		if s.MaxRange.Max < s.Max {
			s.Max = s.MaxRange.Max
		}
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

func unsetInterval() Interval {
	return Interval{math.NaN(), math.NaN()}
}

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

func (i *Interval) Equal(j Interval) bool {
	if math.IsNaN(i.Min) {
		return math.IsNaN(j.Min)
	}
	if math.IsNaN(i.Max) {
		return math.IsNaN(j.Max)
	}
	return i.Min == j.Min && i.Max == j.Max
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
		Absolut   float64
		Releative float64
	}

	MinRange Interval // MinRange determines the allowed range of the Min of a scale.
	MaxRange Interval // MaxRange determines the allowed range of the Max of a scale.
}
