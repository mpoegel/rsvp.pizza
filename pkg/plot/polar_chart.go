package plot

import (
	"math"
	"strconv"
)

type KeyValDatum struct {
	Key string
	Val float64
}

type PolarChart struct {
	width  uint32
	height uint32
	data   []KeyValDatum
}

func NewPolarChart(width, height uint32) *PolarChart {
	pc := &PolarChart{
		width:  width,
		height: height,
		data:   make([]KeyValDatum, 0),
	}
	return pc
}

func (pc *PolarChart) SetData(newData []KeyValDatum) {
	pc.data = newData
}

func (pc *PolarChart) Render() string {
	g := NewSVG(pc.width, pc.height)

	centerX := pc.width / 2
	centerY := pc.height / 2
	circleJump := int(centerX-25) / 5
	theta := 2.0 * math.Pi / float64(len(pc.data))
	maxRadius := uint32(4 + circleJump*4)
	dataMax := 0.0
	for _, d := range pc.data {
		dataMax = max(d.Val, dataMax)
	}
	scalingFactor := float64(maxRadius) / float64(dataMax)

	for i := range 5 {
		r := uint32(4 + circleJump*i)
		c := NewCircleVG(centerX, centerY, r)
		c.AddProperty(FillPropertyVG, "none")
		c.AddProperty(StrokePropertyVG, "black")
		g.AddVG(c)

		txt := NewTextVG(centerX, centerY-r, strconv.FormatInt(int64(r/uint32(scalingFactor)), 10))
		txt.AddProperty(FillPropertyVG, "grey")
		g.AddVG(txt)
	}

	path := NewPathVG()
	path.AddProperty(OpacityPropertyVG, "0.7")
	path.AddProperty(FillPropertyVG, "green")

	for i, d := range pc.data {
		th := theta * float64(i)
		r := d.Val * scalingFactor
		x := r * (math.Cos(th))
		y := r * (math.Sin(th))
		x += float64(centerX)
		y += float64(centerY)
		c := NewCircleVG(uint32(x), uint32(y), 4)
		c.AddProperty(FillPropertyVG, "green")
		g.AddVG(c)

		if i == 0 {
			path.MoveTo(uint32(x), uint32(y))
		} else if i == 1 {
			path.LineTo(uint32(x), uint32(y))
		} else {
			path.Waypoint(uint32(x), uint32(y))
		}

		x = float64(maxRadius+15)*(math.Cos(th)) + float64(centerX)
		y = float64(maxRadius+15)*(math.Sin(th)) + float64(centerY)
		txt := NewTextVG(uint32(x), uint32(y), d.Key)
		txt.AddProperty(FillPropertyVG, "grey")
		if th > math.Pi/2 && th < 3*math.Pi/2 {
			txt.AddProperty(AnchorPropertyVG, "end")
		} else {
			txt.AddProperty(AnchorPropertyVG, "start")
		}
		g.AddVG(txt)
	}
	path.Close()
	g.AddVG(path)

	return g.String()
}
