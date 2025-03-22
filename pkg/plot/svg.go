package plot

import (
	"fmt"
	"strings"
)

type PropertyVG string

const (
	FillPropertyVG    PropertyVG = "fill"
	StrokePropertyVG  PropertyVG = "stroke"
	OpacityPropertyVG PropertyVG = "opacity"
	AnchorPropertyVG  PropertyVG = "text-anchor"
)

type VectorGraphic interface {
	AddProperty(name PropertyVG, value string)
	String() string
}

type SVG struct {
	width    uint32
	height   uint32
	graphics []VectorGraphic
}

func NewSVG(width, height uint32) *SVG {
	svg := &SVG{
		width:    width,
		height:   height,
		graphics: make([]VectorGraphic, 0),
	}
	return svg
}

func (svg *SVG) AddVG(graphic VectorGraphic) {
	svg.graphics = append(svg.graphics, graphic)
}

func (svg *SVG) String() string {
	out := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg"
		width="%d"
		height="%d"
		viewBox="0 0 %d %d"
		fill="none">`, svg.width, svg.height, svg.width, svg.height)
	for _, g := range svg.graphics {
		out += "\n" + g.String()
	}
	out += "</svg>"
	return out
}

type CircleVG struct {
	cx     uint32
	cy     uint32
	radius uint32
	props  []string
}

func NewCircleVG(x, y, radius uint32) *CircleVG {
	return &CircleVG{
		cx:     x,
		cy:     y,
		radius: radius,
		props:  make([]string, 0),
	}
}

func (c *CircleVG) AddProperty(name PropertyVG, value string) {
	c.props = append(c.props, fmt.Sprintf(`%s="%s"`, name, value))
}

func (c *CircleVG) String() string {
	out := fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d"`, c.cx, c.cy, c.radius)
	out += " " + strings.Join(c.props, " ")
	out += "/>"
	return out
}

type PathVG struct {
	path  string
	props []string
}

func NewPathVG() *PathVG {
	return &PathVG{
		props: make([]string, 0),
	}
}

func (p *PathVG) MoveTo(x, y uint32)   { p.path += fmt.Sprintf(" M %d %d", x, y) }
func (p *PathVG) LineTo(x, y uint32)   { p.path += fmt.Sprintf(" L %d %d", x, y) }
func (p *PathVG) Waypoint(x, y uint32) { p.path += fmt.Sprintf(" %d %d", x, y) }
func (p *PathVG) Close()               { p.path += " Z" }

func (p *PathVG) AddProperty(name PropertyVG, value string) {
	p.props = append(p.props, fmt.Sprintf(`%s="%s"`, name, value))
}

func (p *PathVG) String() string {
	out := fmt.Sprintf(`<path d="%s"`, p.path)
	out += " " + strings.Join(p.props, " ")
	out += "/>"
	return out
}

type TextVG struct {
	x     uint32
	y     uint32
	text  string
	props []string
}

func NewTextVG(x, y uint32, text string) *TextVG {
	return &TextVG{
		x:     x,
		y:     y,
		text:  text,
		props: make([]string, 0),
	}
}

func (t *TextVG) AddProperty(name PropertyVG, value string) {
	t.props = append(t.props, fmt.Sprintf(`%s="%s"`, name, value))
}

func (t *TextVG) String() string {
	out := fmt.Sprintf(`<text x="%d" y="%d"`, t.x, t.y)
	out += " " + strings.Join(t.props, " ")
	out += fmt.Sprintf(">%s</text>", t.text)
	return out
}
