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
	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart/v2"
)

type barChart struct {
	p   *Painter
	opt *BarChartOption
}

// NewBarChart returns a bar chart renderer
func NewBarChart(p *Painter, opt BarChartOption) *barChart {
	if opt.Theme == nil {
		opt.Theme = defaultTheme
	}
	return &barChart{
		p:   p,
		opt: &opt,
	}
}

type BarChartOption struct {
	// The theme
	Theme ColorPalette
	// The font size
	Font *truetype.Font
	// The data series list
	SeriesList SeriesList
	// The x axis option
	XAxis XAxisOption
	// The padding of line chart
	Padding Box
	// The y axis option
	YAxisOptions []YAxisOption
	// The option of title
	Title TitleOption
	// The legend option
	Legend LegendOption
}

type barChartLabelRenderOption struct {
	Text  string
	Style Style
	X     int
	Y     int
}

func (b *barChart) render(result *defaultRenderResult, seriesList SeriesList) (Box, error) {
	p := b.p
	opt := b.opt
	seriesPainter := result.seriesPainter

	xRange := NewRange(AxisRangeOption{
		DivideCount: len(opt.XAxis.Data),
		Size:        seriesPainter.Width(),
	})
	x0, x1 := xRange.GetRange(0)
	width := int(x1 - x0)
	// 每一块之间的margin
	margin := 10
	// 每一个bar之间的margin
	barMargin := 5
	if width < 20 {
		margin = 2
		barMargin = 2
	} else if width < 50 {
		margin = 5
		barMargin = 3
	}
	seriesCount := len(seriesList)
	// 总的宽度-两个margin-(总数-1)的barMargin
	barWidth := (width - 2*margin - barMargin*(seriesCount-1)) / len(seriesList)
	barMaxHeight := seriesPainter.Height()
	theme := opt.Theme
	seriesNames := seriesList.Names()

	markPointPainter := NewMarkPointPainter(seriesPainter)
	markLinePainter := NewMarkLinePainter(seriesPainter)
	rendererList := []Renderer{
		markPointPainter,
		markLinePainter,
	}
	labelRenderOptions := make([]barChartLabelRenderOption, 0)
	for index := range seriesList {
		series := seriesList[index]
		yRange := result.axisRanges[series.AxisIndex]
		seriesColor := theme.GetSeriesColor(series.index)

		divideValues := xRange.AutoDivide()
		points := make([]Point, len(series.Data))
		for j, item := range series.Data {
			if j >= xRange.divideCount {
				continue
			}
			x := divideValues[j]
			x += margin
			if index != 0 {
				x += index * (barWidth + barMargin)
			}

			h := int(yRange.getHeight(item.Value))
			fillColor := seriesColor
			if !item.Style.FillColor.IsZero() {
				fillColor = item.Style.FillColor
			}
			top := barMaxHeight - h

			seriesPainter.OverrideDrawingStyle(Style{
				FillColor: fillColor,
			}).Rect(chart.Box{
				Top:    top,
				Left:   x,
				Right:  x + barWidth,
				Bottom: barMaxHeight - 1,
			})
			// 用于生成marker point
			points[j] = Point{
				// 居中的位置
				X: x + barWidth>>1,
				Y: top,
			}
			// 用于生成marker point
			points[j] = Point{
				// 居中的位置
				X: x + barWidth>>1,
				Y: top,
			}
			// 如果label不需要展示，则返回
			if !series.Label.Show {
				continue
			}
			distance := series.Label.Distance
			if distance == 0 {
				distance = 5
			}
			text := NewValueLabelFormatter(seriesNames, series.Label.Formatter)(index, item.Value, -1)
			labelStyle := Style{
				FontColor: theme.GetTextColor(),
				FontSize:  labelFontSize,
				Font:      opt.Font,
			}
			if !series.Label.Color.IsZero() {
				labelStyle.FontColor = series.Label.Color
			}

			textBox := seriesPainter.MeasureText(text)

			labelRenderOptions = append(labelRenderOptions, barChartLabelRenderOption{
				Text:  text,
				Style: labelStyle,
				X:     x + (barWidth-textBox.Width())>>1,
				Y:     barMaxHeight - h - distance,
			})
		}

		markPointPainter.Add(markPointRenderOption{
			FillColor: seriesColor,
			Font:      opt.Font,
			Series:    series,
			Points:    points,
		})
		markLinePainter.Add(markLineRenderOption{
			FillColor:   seriesColor,
			FontColor:   opt.Theme.GetTextColor(),
			StrokeColor: seriesColor,
			Font:        opt.Font,
			Series:      series,
			Range:       yRange,
		})
	}
	for _, labelOpt := range labelRenderOptions {
		seriesPainter.OverrideTextStyle(labelOpt.Style)
		seriesPainter.Text(labelOpt.Text, labelOpt.X, labelOpt.Y)
	}
	// 最大、最小的mark point
	err := doRender(rendererList...)
	if err != nil {
		return BoxZero, err
	}

	return p.box, nil
}

func (b *barChart) Render() (Box, error) {
	p := b.p
	opt := b.opt
	renderResult, err := defaultRender(p, defaultRenderOption{
		Theme:        opt.Theme,
		Padding:      opt.Padding,
		SeriesList:   opt.SeriesList,
		XAxis:        opt.XAxis,
		YAxisOptions: opt.YAxisOptions,
		TitleOption:  opt.Title,
		LegendOption: opt.Legend,
	})
	if err != nil {
		return BoxZero, err
	}
	seriesList := opt.SeriesList.Filter(ChartTypeLine)
	return b.render(renderResult, seriesList)
}
