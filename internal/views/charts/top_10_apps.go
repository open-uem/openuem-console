package charts

import (
	"github.com/doncicuto/openuem-console/internal/models"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
)

func Top10Apps(topApps []models.App) render.ChartSnippet {

	pie := charts.NewPie()

	pieData := []opts.PieData{}

	for _, app := range topApps {
		pieData = append(pieData, opts.PieData{
			Name:  app.Name,
			Value: app.Count,
		})
	}

	pie.AddSeries("Top 10 apps", pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Formatter: "{b}: {c}"}),
		charts.WithPieChartOpts(opts.PieChart{
			Radius: []string{"40%", "75%"},
		}),
	)

	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Show: opts.Bool(false)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		charts.WithColorsOpts(opts.Colors{"#dd6b66", "#759aa0", "#e69d87"}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "600px",
			Height: "300px",
		}))

	return pie.RenderSnippet()
}
