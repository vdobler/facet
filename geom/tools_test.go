package geom

import (
	"fmt"
	"testing"
)

func TestCopyAesthetics(t *testing.T) {
	h := HLine{
		Alpha: func(i int) float64 { return float64(2 * i) },
	}
	v := VLine{}

	double := func(n int) int { return 2 * n }
	addone := func(n int) int { return n + 1 }
	CopyAesthetics(&v, h, nil)
	fmt.Println(v.Alpha(3))
	CopyAesthetics(&v, h, double)
	fmt.Println(v.Alpha(3))
	CopyAesthetics(&v, &h, addone)
	fmt.Println(v.Alpha(3))
}
