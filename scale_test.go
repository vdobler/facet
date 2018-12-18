package facet

import (
	"math"
	"strconv"
	"testing"
)

var nan = math.NaN()

var intervallUpdateTests = []struct {
	old  Interval
	x    float64
	want Interval
}{
	{Interval{3, 6}, 4, Interval{3, 6}},
	{Interval{3, 6}, 2, Interval{2, 6}},
	{Interval{3, 6}, 7, Interval{3, 7}},
	{Interval{nan, nan}, nan, Interval{nan, nan}},
	{Interval{nan, nan}, 5, Interval{5, 5}},
	{Interval{5, 5}, nan, Interval{5, 5}},
}

func TestIntervalUpdate(t *testing.T) {
	for i, tc := range intervallUpdateTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tc.old
			got.Update(tc.x)
			if !got.Equal(tc.want) {
				t.Errorf("%v update %v = %v, want %v",
					tc.old, tc.x, got, tc.want)
			}
		})
	}
}
