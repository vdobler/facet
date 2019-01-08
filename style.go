package facet

import (
	"image/color"
	"math"

	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// A Style controls how a Plot is drawn.
type Style struct {
	Background color.Color

	Title       draw.TextStyle
	SubTitle    draw.TextStyle
	TitleHeight vg.Length

	Panel struct {
		Background color.Color
		PadX       vg.Length
		PadY       vg.Length
	}
	HStrip struct {
		Background color.Color
		Height     vg.Length
		draw.TextStyle
	}
	VStrip struct {
		Background color.Color
		Width      vg.Length
		draw.TextStyle
	}

	Grid struct {
		Major draw.LineStyle
		Minor draw.LineStyle
	}

	XAxis struct {
		Title       draw.TextStyle
		TitleHeight vg.Length
		Line        draw.LineStyle
		MajorTick   struct {
			draw.LineStyle
			Length vg.Length
			Align  draw.YAlignment
			Label  draw.TextStyle
		}
		MinorTick struct {
			draw.LineStyle
			Length vg.Length
			Align  draw.YAlignment
		}
	}

	YAxis struct {
		Title      draw.TextStyle
		TitleWidth vg.Length
		Line       draw.LineStyle
		MajorTick  struct {
			draw.LineStyle
			Length vg.Length
			Align  draw.XAlignment
			Label  draw.TextStyle
		}
		MinorTick struct {
			draw.LineStyle
			Length vg.Length
			Align  draw.XAlignment
		}
	}

	Legend struct {
		Position string // left
		Title    draw.TextStyle
		Label    draw.TextStyle

		Discrete struct {
			Size vg.Length
			Pad  vg.Length
		}
		Continuous struct {
			Size   vg.Length
			Length vg.Length
			Tick   struct {
				draw.LineStyle
				Length vg.Length
				Align  draw.XAlignment
				Mirror bool
			}
		}
	}
}

// DefaultFacetStyle returns a FacetStyle which mimics the appearance of ggplot2.
// The baseFontSize is the font size for axis titles and strip labels, the title
// is a bit bigger, tick labels a bit smaller.
func DefaultFacetStyle(baseFontSize vg.Length) Style {
	scale := func(x vg.Length, f float64) vg.Length {
		return vg.Length(math.Round(f * float64(x)))
	}

	titleFont, err := vg.MakeFont("Helvetica-Bold", scale(baseFontSize, 1.2))
	if err != nil {
		panic(err)
	}
	baseFont, err := vg.MakeFont("Helvetica-Bold", baseFontSize)
	if err != nil {
		panic(err)
	}
	tickFont, err := vg.MakeFont("Helvetica-Bold", scale(baseFontSize, 1/1.2))
	if err != nil {
		panic(err)
	}

	fs := Style{}
	fs.Background = color.Transparent

	fs.TitleHeight = scale(baseFontSize, 3)
	fs.Title.Color = color.Black
	fs.Title.Font = titleFont
	fs.Title.XAlign = draw.XCenter
	fs.Title.YAlign = draw.YTop

	fs.Panel.Background = color.Gray16{0xeeee}
	fs.Panel.PadX = scale(baseFontSize, 0.5)
	fs.Panel.PadY = fs.Panel.PadX

	fs.HStrip.Background = color.Gray16{0xcccc}
	fs.HStrip.Font = baseFont
	fs.HStrip.Height = scale(baseFontSize, 2)
	fs.HStrip.XAlign = draw.XCenter
	fs.HStrip.YAlign = -0.3 // draw.YCenter

	fs.VStrip.Background = color.Gray16{0xcccc}
	fs.VStrip.Font = baseFont
	fs.VStrip.Width = scale(baseFontSize, 2.5)
	fs.VStrip.XAlign = draw.XCenter
	fs.VStrip.YAlign = -0.3 // draw.YCenter
	fs.VStrip.Rotation = -math.Pi / 2

	fs.Grid.Major.Color = color.White
	fs.Grid.Major.Width = vg.Length(1)
	fs.Grid.Minor.Color = color.White
	fs.Grid.Minor.Width = vg.Length(0.5)

	fs.XAxis.Title.Color = color.Black
	fs.XAxis.Title.Font = baseFont
	fs.XAxis.Title.Rotation = 0
	fs.XAxis.Title.XAlign = draw.XCenter
	fs.XAxis.Title.YAlign = draw.YAlignment(0.3)
	fs.XAxis.TitleHeight = scale(baseFontSize, 2)

	fs.XAxis.Line.Width = 0

	fs.XAxis.MajorTick.Color = color.Gray16{0x1111}
	fs.XAxis.MajorTick.Width = vg.Length(1)
	fs.XAxis.MajorTick.Length = vg.Length(5)
	fs.XAxis.MajorTick.Align = 0

	fs.XAxis.MinorTick.Color = nil
	fs.XAxis.MinorTick.Width = vg.Length(0)
	fs.XAxis.MinorTick.Length = vg.Length(0)
	fs.XAxis.MinorTick.Align = 0

	fs.XAxis.MajorTick.Label.Color = color.Black
	fs.XAxis.MajorTick.Label.Font = tickFont
	fs.XAxis.MajorTick.Label.XAlign = draw.XCenter
	fs.XAxis.MajorTick.Label.YAlign = draw.YTop

	fs.YAxis.Title.Color = color.Black
	fs.YAxis.Title.Font = baseFont
	fs.YAxis.Title.Rotation = math.Pi / 2
	fs.YAxis.Title.XAlign = draw.XCenter
	fs.YAxis.Title.YAlign = draw.YTop
	fs.YAxis.TitleWidth = scale(baseFontSize, 2)

	fs.YAxis.Line.Width = 0

	// Major Ticks and Labels
	fs.YAxis.MajorTick.Color = color.Gray16{0x1111}
	fs.YAxis.MajorTick.Width = vg.Length(1)
	fs.YAxis.MajorTick.Length = vg.Length(5)
	fs.YAxis.MajorTick.Align = 0
	fs.YAxis.MajorTick.Label.Color = color.Black
	fs.YAxis.MajorTick.Label.Font = tickFont
	fs.YAxis.MajorTick.Label.XAlign = draw.XRight
	fs.YAxis.MajorTick.Label.YAlign = -0.3 // draw.YCenter

	// No minor ticks
	fs.YAxis.MinorTick.Color = nil
	fs.YAxis.MinorTick.Width = 0
	fs.YAxis.MinorTick.Length = 0
	fs.YAxis.MinorTick.Align = 0

	fs.Legend.Position = "right"

	fs.Legend.Label.Color = color.Black
	fs.Legend.Label.Font = tickFont
	fs.Legend.Label.YAlign = -0.3 // draw.YCenter

	fs.Legend.Title.Color = color.Black
	fs.Legend.Title.Font = baseFont
	fs.Legend.Title.XAlign = draw.XLeft
	fs.Legend.Title.YAlign = draw.YTop

	fs.Legend.Discrete.Size = vg.Length(20)
	fs.Legend.Discrete.Pad = vg.Length(4)

	fs.Legend.Continuous.Size = vg.Length(20)
	fs.Legend.Continuous.Length = vg.Length(150)
	fs.Legend.Continuous.Tick.Color = color.Black
	fs.Legend.Continuous.Tick.Width = vg.Length(1)
	fs.Legend.Continuous.Tick.Length = vg.Length(3)
	fs.Legend.Continuous.Tick.Align = 1
	fs.Legend.Continuous.Tick.Mirror = true

	return fs
}
