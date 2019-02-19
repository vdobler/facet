package facet

import (
	"fmt"
	"math"
	"testing"
)

var transformationTests = []struct {
	trans   Transformation
	a, b    float64 // from
	u, v    float64 // to
	x, want float64
}{
	{IdentityTrans, 10, 20, 0, 1, 7, 7},

	{LinearTrans, 10, 20, 10, 20, 12, 12},
	{LinearTrans, 10, 20, 100, 200, 12, 120},
	{LinearTrans, 3, 5, 0, 1, 3, 0},
	{LinearTrans, 3, 5, 0, 1, 4, 0.5},
	{LinearTrans, 3, 5, 0, 1, 5, 1},

	{SqrtTrans, -10, 30, 2, 20, -10, 2.0},
	{SqrtTrans, -10, 30, 2, 20, 0, 10.15},
	{SqrtTrans, -10, 30, 2, 20, 10, 14.21},
	{SqrtTrans, -10, 30, 2, 20, 20, 17.35},
	{SqrtTrans, -10, 30, 2, 20, 30, 20.00},

	{SqrtTransFix0, 10, 20, 3, 4, 0, 0},
	{SqrtTransFix0, 10, 20, 3, 4, 10, 2 * math.Sqrt2},
	{SqrtTransFix0, 10, 20, 3, 4, 20, 4},
}

func equal64(a, b float64) bool {
	ai, af := math.Modf(a)
	bi, bf := math.Modf(b)
	if af == 0 && bf == 0 {
		return ai == bi
	}
	return math.Abs(a-b) < 0.006
}

func TestTransform(t *testing.T) {
	for i, tc := range transformationTests {
		t.Run(fmt.Sprintf("%s/%d", tc.trans.Name, i), func(t *testing.T) {
			from, to := Interval{tc.a, tc.b}, Interval{tc.u, tc.v}
			if got := tc.trans.Trans(from, to, tc.x); !equal64(got, tc.want) {
				t.Errorf("%s.Trans(%v,%v,%f) = %f, want %f",
					tc.trans.Name, from, to, tc.x, got, tc.want)
			}
		})
	}
}
