// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/spf13/cobra"
)

// lineCmd represents the line command
var lineCmd = &cobra.Command{
	Use:   "line",
	Short: "line plot and scatter plot",
	Long: `line plot and scatter plot

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		plotConfig := getPlotConfigs(cmd)

		files := getFileList(args)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		lineWidth := vg.Points(getFlagPositiveFloat64(cmd, "line-width"))
		pointSize := vg.Length(getFlagPositiveFloat64(cmd, "point-size"))
		scatter := getFlagBool(cmd, "scatter")
		dataFieldXStr := getFlagString(cmd, "data-field-x")
		if dataFieldXStr == "" {
			checkError(fmt.Errorf("flag -x (--data-field-x) needed"))
		}
		if strings.Index(dataFieldXStr, ",") >= 0 {
			checkError(fmt.Errorf("only one field allowed for flag -x (--data-field-x)"))
		}
		if dataFieldXStr[0] == '-' {
			checkError(fmt.Errorf("unselect not allowed for flag -x (--data-field-x)"))
		}

		dataFieldYStr := getFlagString(cmd, "data-field-y")
		if dataFieldYStr == "" {
			checkError(fmt.Errorf("flag -x (--data-field-y) needed"))
		}
		if strings.Index(dataFieldYStr, ",") >= 0 {
			checkError(fmt.Errorf("only one field allowed for flag -y (--data-field-y)"))
		}
		if dataFieldXStr[0] == '-' {
			checkError(fmt.Errorf("unselect not allowed for flag -y (--data-field-y)"))
		}

		groupFieldStr := getFlagString(cmd, "group-field")
		if len(groupFieldStr) > 0 {
			if strings.Index(groupFieldStr, ",") >= 0 {
				checkError(fmt.Errorf("only one field allowed for flag --group-field"))
			}
			if groupFieldStr[0] == '-' {
				checkError(fmt.Errorf("unselect not allowed for flag --group-field"))
			}
			plotConfig.fieldStr = dataFieldXStr + "," + dataFieldYStr + "," + groupFieldStr
		} else {
			plotConfig.fieldStr = dataFieldXStr + "," + dataFieldYStr
		}

		file := files[0]
		headerRow, fields, data, _, _ := parseCSVfile(cmd, config, file, plotConfig.fieldStr, false)

		// =======================================

		if config.OutFile == "-" {
			if scatter {
				config.OutFile = "scater.png"
			} else {
				config.OutFile = "lineplot.png"
			}
		}
		groups := make(map[string]plotter.XYs)
		var x, y float64
		var err error
		var ok bool
		var groupName string
		for _, d := range data {
			x, err = strconv.ParseFloat(d[0], 64)
			if err != nil {
				if len(headerRow) > 0 {
					checkError(fmt.Errorf("fail to parse X value: %s at column: %s. please choose the right column by flag --data-field-x", d[0], headerRow[0]))
				} else {
					checkError(fmt.Errorf("fail to parse X value: %s at column: %d. please choose the right column by flag --data-field-x", d[0], fields[0]))
				}
			}
			y, err = strconv.ParseFloat(d[1], 64)
			if err != nil {
				if len(headerRow) > 0 {
					checkError(fmt.Errorf("fail to parse Y value: %s at column: %s. please choose the right column by flag --data-field-y", d[1], headerRow[1]))
				} else {
					checkError(fmt.Errorf("fail to parse Y value: %s at column: %d. please choose the right column by flag --data-field-y", d[1], fields[1]))
				}
			}

			if len(d) > 2 {
				groupName = d[2]
			} else {
				groupName = ""
			}
			if _, ok = groups[groupName]; !ok {
				groups[groupName] = make(plotter.XYs, 0)
			}
			groups[groupName] = append(groups[groupName], struct{ X, Y float64 }{X: x, Y: y})
		}

		p, err := plot.New()
		if err != nil {
			checkError(err)
		}

		i := 0
		for g, v := range groups {
			if !scatter {
				lines, points, err := plotter.NewLinePoints(v)
				checkError(err)
				lines.Color = plotutil.Color(i)
				lines.LineStyle.Dashes = plotutil.Dashes(i)
				lines.LineStyle.Width = lineWidth
				points.Shape = plotutil.Shape(i)
				points.Color = plotutil.Color(i)
				points.Radius = pointSize
				p.Add(lines, points)
				p.Legend.Add(g, lines, points)
			} else {
				points, err := plotter.NewScatter(v)
				checkError(err)
				points.Shape = plotutil.Shape(i)
				points.Color = plotutil.Color(i)
				points.Radius = pointSize
				p.Add(points)
				p.Legend.Add(g, points)
			}

			i++
		}
		if lineWidth > pointSize {
			p.Legend.Padding = vg.Length(lineWidth)
		} else {
			p.Legend.Padding = vg.Length(pointSize)
		}
		p.Legend.Top = getFlagBool(cmd, "legend-top")
		p.Legend.Left = getFlagBool(cmd, "legend-left")

		if plotConfig.ylab == "" {
			if len(headerRow) > 0 {
				plotConfig.ylab = headerRow[1]
			} else {
				plotConfig.ylab = "Y Values"
			}
		}
		if plotConfig.xlab == "" {
			if len(headerRow) > 0 {
				plotConfig.xlab = headerRow[0]
			} else {
				plotConfig.xlab = "X Values"
			}
		}

		p.Title.Text = plotConfig.title
		p.Title.TextStyle.Font.Size = plotConfig.titleSize
		p.X.Label.Text = plotConfig.xlab
		p.Y.Label.Text = plotConfig.ylab
		p.X.Label.TextStyle.Font.Size = plotConfig.labelSize
		p.Y.Label.TextStyle.Font.Size = plotConfig.labelSize
		p.X.Width = plotConfig.axisWidth
		p.Y.Width = plotConfig.axisWidth
		p.X.Tick.Width = plotConfig.tickWidth
		p.Y.Tick.Width = plotConfig.tickWidth

		if plotConfig.xminStr != "" {
			p.X.Min = plotConfig.xmin
		}
		if plotConfig.xmaxStr != "" {
			p.X.Max = plotConfig.xmax
		}
		if plotConfig.yminStr != "" {
			p.Y.Min = plotConfig.ymin
		}
		if plotConfig.ymaxStr != "" {
			p.Y.Max = plotConfig.ymax
		}

		// Save the plot to a PNG file.
		if err := p.Save(plotConfig.width*vg.Inch,
			plotConfig.height*vg.Inch, config.OutFile); err != nil {
			checkError(err)
		}
	},
}

func init() {
	plotCmd.AddCommand(lineCmd)
	lineCmd.Flags().StringP("data-field-x", "x", "", `column index or column name of X for command line`)
	lineCmd.Flags().StringP("data-field-y", "y", "", `column index or column name of Y for command line`)

	lineCmd.Flags().BoolP("legend-top", "", false, "locate legend along the top edge of the plot")
	lineCmd.Flags().BoolP("legend-left", "", false, "locate legend along the left edge of the plot")

	lineCmd.Flags().BoolP("scatter", "", false, "only plot points")
	lineCmd.Flags().Float64P("line-width", "", 1.5, "line width")
	lineCmd.Flags().Float64P("point-size", "", 3, "point size")
}