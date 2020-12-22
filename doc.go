// Package facet is an experimental package to produce faceted plots.
//
// It tries to use or enhance  gonum.org/v1/plot.
//
// Scales
//
// The concept of a scale is taken from ggplot2. Package facet knows about
// the following scales:
//   - X-Scale        The x- and y scale are mandatory for all faceted
//   - Y-Scale        plots. The scales are drawn as axis, not as guides.
//   - Alpha-Scale    The alpha (opacity) value
//   - Color-Scale    The line and symbol color
//   - Fill-Scale     The fill color
//   - Shape-Scale    The symbol used to draw points
//   - Size-Scale     The size of points and width of lines
//   - Stroke-Scale   The line style (dash pattern)
//
// The Shape and Stroke scales cannot be continouos but must be discrete
// because only a discrete set of symbol types and line styles are available.
// The other scales can be discrete or continouos.
//
// If a scale is used in a faceted plot a scale Guide is drawn to show how
// the scales range maps to aesthetics. Guides for different scales are
// combined iff:
//   1. The two scales are of the same kind (discrete, continuous, ...)
//   2. The two scales have the same range.
//   3. The two scales have the same Title or the Title is empty.
//   4. The scales must use the same Ticker.
//   5. Fill and Color can be combined if they use the same ColorMap or one is empty.
//
//
// Faceted Plots and Grouping
//
// There are one dimensional groupings and two-dimensional groupings
// corresponding to facet_row, facet_column and facet_grid. Grouping
// is done on discrete values (which may be an intervall).
package facet
