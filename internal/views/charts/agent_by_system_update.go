package charts

import (
	"context"
	"strconv"

	"github.com/open-uem/openuem-console/internal/models"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/invopop/ctxi18n/i18n"
)

func AgentsBySystemUpdate(ctx context.Context, agents []models.Agent, countAllAgents int) render.ChartSnippet {
	pie := charts.NewPie()

	// preformat data
	pieData := []opts.PieData{}

	for _, a := range agents {
		pieData = append(pieData, opts.PieData{Name: a.Status, Value: a.Count})
	}

	// put data into chart
	pie.AddSeries(i18n.T(ctx, "charts.update_status"), pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(false), Formatter: "{b}: {c}"}),
		charts.WithPieChartOpts(opts.PieChart{
			Radius: []string{"40%", "75%"},
		}),
	)

	textStyle := opts.TextStyle{FontSize: 36}
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: strconv.Itoa(countAllAgents), Left: "center", Top: "center", TitleStyle: &textStyle}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Type: "scroll"}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "300px",
			Height: "300px",
		}),
	)

	return pie.RenderSnippet()
}
