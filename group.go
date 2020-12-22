package facet

import (
	"fmt"
	"math"
)

// A GroupID identifies a set of values belonging together.
type GroupID struct {
	Row, Col string
}

type Grouper interface {
	Group() GroupID
}

// Aestetic is a function mapping a certain data point to an aestehtic.
type Aesthetic func(i int) float64

// DiscreteAestetic is a function mapping a certain data point to a discrete
// aesthetic like Shape or Stroke.
type DiscreteAesthetic func(i int) int

type GroupBy struct {
	FacetRow DiscreteAesthetic
	FacetCol DiscreteAesthetic
	Alpha    Aesthetic
	Color    Aesthetic
	Fill     Aesthetic
	Shape    DiscreteAesthetic
	Size     Aesthetic
	Stroke   DiscreteAesthetic
}

// Faceting
type Faceting struct {
	Rows   []string
	Cols   []string
	Groups map[GroupID][]int // Groups contains the indices for each group
}

func NewFaceting() *Faceting {
	return &Faceting{
		Groups: make(map[GroupID][]int),
	}
}

func (f1 *Faceting) Add(group GroupID) {

}

// A Partitioner can be used to turn a continuous value into a discrete factor.
type Partitioner struct {
	Partitions int
	Range      Interval
}

func (p *Partitioner) Learn(x ...float64) { p.Range.Update(x...) }
func (p *Partitioner) Partition(x float64) string {
	min, max := p.Range.Min, p.Range.Max

	if x < min {
		return fmt.Sprintf("(-∞, %g)", x)
	}
	if x >= max {
		return fmt.Sprintf("[%g, ∞)", x)
	}

	w := (max - min) / float64(p.Partitions)
	k := math.Floor((x - min) / w)
	return fmt.Sprintf("[%g, %g)", min+k*w, min+(k+1)*w)
}

/*
How to learn groups?

Original Data --- group-by-field-G---> Grouped Data --> Work on each group indiv.
That works fine in R and is "undoable" in Go.

Idea:
 - Geom provides a func (i int) GroupID.
 - Plot owns Geoms.
 - Plot calls GroupID for each element i in the Geom.
 - A panel is selected based on the group and the facting choosen.
Works well if Dataset already contains the group, e.g. as a string or int field.
But hard to use if grouping is done on a continuous field, e.g. a float64.
Need an adaptor which can be trained easily.

G might be:
 - strings (label)
 - integers
 - floats (must be put into intervals)

Basically the same like for a discrete scale.

For grouping continuous data: Type which works as an adoptor.

How to draw a Point Geom in a faceted way?
 - The Plot itself owns the Point Geom
 - The Plot iterates the Points and determines the Group
 - The Plot chooses the appropriate Panel and...
 - ... hands drawing of a single point in that panel to the Point Geom
Same applies to Boxplots or whatnot.

Sizing of e.g. a Boxplot is done on the full dataset before grouping.


Can Stats be added?
Binning for Histograms? Input is Data with X and Group.
Plot iterates Data, determines Group. Records X for that Group.
Can be done as Binning can be done iteratively.
Boxplots? Need full dataset befor Geom can be computed.
Seems doable.
If plot owns Geom: How to compute boxplot? How is faceting done?

Input: Some data with X and Y and optional Color
Step 1: Partition data by group into facets
Step 3: Per Group/facet compute fiveval foraech X



Example Boxplot
---------------
Data:   X, Y, A, B, C, D, E
Faceting by (A,B)
Color mapped from C
Stroke mapped from D
Stat: fiveval/boxplot

Statistics has to be done on Data/(A,B,C,D)
Grouping is done on faceting and optional aesthetics










*/
