package charts

import (
	"context"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/open-uem/openuem-console/internal/models"
)

func AgentsByOs(ctx context.Context, agents []models.Agent, countAllAgents int) render.ChartSnippet {
	pie := charts.NewPie()

	// preformat data
	pieData := []opts.PieData{}

	for _, a := range agents {
		pieData = append(pieData, opts.PieData{Name: a.OS, Value: a.Count})
	}

	// put data into chart
	pie.AddSeries(i18n.T(ctx, "charts.os_version"), pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(false), Formatter: "{b}: {c}"}),
		charts.WithPieChartOpts(opts.PieChart{
			Radius: []string{"40%", "75%"},
			Center: []string{"25%", "50%"},
		}),
	)

	leftTitle := "19.5%"
	if countAllAgents < 10 {
		leftTitle = "22%"
	}

	textStyle := opts.TextStyle{FontSize: 36, Color: "#777"}

	labelStyle := opts.TextStyle{Color: "#777"}

	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: strconv.Itoa(countAllAgents), Left: leftTitle, Top: "43%", TitleStyle: &textStyle}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), TextStyle: &labelStyle, Type: "scroll", Orient: "vertical", X: "left", Y: "center", Left: "55%"}),
		charts.WithColorsOpts(opts.Colors{"#9e0142", "#f46d43", "#fdae61", "#fee08b", "#e6f598", "#abdda4", "#66c2a5", "#3288bd", "#5e4fa2"}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "480px",
			Height: "300px",
		}),
	)

	return pie.RenderSnippet()
}
