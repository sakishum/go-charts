// MIT License

// Copyright (c) 2022 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package charts

import (
	"bytes"
	"errors"

	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

const (
	PositionLeft   = "left"
	PositionRight  = "right"
	PositionTop    = "top"
	PositionBottom = "bottom"
)

type Draw struct {
	Render chart.Renderer
	Box    chart.Box
	Font   *truetype.Font
	parent *Draw
}

type DrawOption struct {
	Type   string
	Parent *Draw
	Width  int
	Height int
}

type Option func(*Draw) error

func PaddingOption(padding chart.Box) Option {
	return func(d *Draw) error {
		d.Box.Left += padding.Left
		d.Box.Top += padding.Top
		d.Box.Right -= padding.Right
		d.Box.Bottom -= padding.Bottom
		return nil
	}
}

func NewDraw(opt DrawOption, opts ...Option) (*Draw, error) {
	if opt.Parent == nil && (opt.Width <= 0 || opt.Height <= 0) {
		return nil, errors.New("parent and width/height can not be nil")
	}
	font, _ := chart.GetDefaultFont()
	d := &Draw{
		Font: font,
	}
	width := opt.Width
	height := opt.Height
	if opt.Parent != nil {
		d.parent = opt.Parent
		d.Render = d.parent.Render
		d.Box = opt.Parent.Box.Clone()
	}
	if width != 0 && height != 0 {
		d.Box.Right = width + d.Box.Left
		d.Box.Bottom = height + d.Box.Top
	}
	// 创建render
	if d.parent == nil {
		fn := chart.SVG
		if opt.Type == "png" {
			fn = chart.PNG
		}
		r, err := fn(d.Box.Right, d.Box.Bottom)
		if err != nil {
			return nil, err
		}
		d.Render = r
	}

	for _, o := range opts {
		err := o(d)
		if err != nil {
			return nil, err
		}
	}
	return d, nil
}

func (d *Draw) Parent() *Draw {
	return d.parent
}

func (d *Draw) Top() *Draw {
	if d.parent == nil {
		return nil
	}
	t := d.parent
	// 限制最多查询次数，避免嵌套引用
	for i := 50; i > 0; i-- {
		if t.parent == nil {
			break
		}
		t = t.parent
	}
	return t
}

func (d *Draw) Bytes() ([]byte, error) {
	buffer := bytes.Buffer{}
	err := d.Render.Save(&buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), err
}

func (d *Draw) moveTo(x, y int) {
	d.Render.MoveTo(x+d.Box.Left, y+d.Box.Top)
}

func (d *Draw) lineTo(x, y int) {
	d.Render.LineTo(x+d.Box.Left, y+d.Box.Top)
}

func (d *Draw) circle(radius float64, x, y int) {
	d.Render.Circle(radius, x+d.Box.Left, y+d.Box.Top)
}

func (d *Draw) text(body string, x, y int) {
	d.Render.Text(body, x+d.Box.Left, y+d.Box.Top)
}

func (d *Draw) lineStroke(points []Point, style LineStyle) {
	s := style.Style()
	if !s.ShouldDrawStroke() {
		return
	}
	r := d.Render
	s.GetStrokeOptions().WriteDrawingOptionsToRenderer(r)
	for index, point := range points {
		x := point.X
		y := point.Y
		if index == 0 {
			d.moveTo(x, y)
		} else {
			d.lineTo(x, y)
		}
	}
	r.Stroke()
}

func (d *Draw) setBackground(width, height int, color drawing.Color) {
	r := d.Render
	s := chart.Style{
		FillColor: color,
	}
	s.WriteToRenderer(r)
	r.MoveTo(0, 0)
	r.LineTo(width, 0)
	r.LineTo(width, height)
	r.LineTo(0, height)
	r.LineTo(0, 0)
	r.FillStroke()
}
