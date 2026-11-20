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

// UpdateRunsByStatus shows how many agents of the organization are classified by the
// result of their last software update run, with the number of queried agents in the
// center of the pie, like the Deploy update status page does
func UpdateRunsByStatus(ctx context.Context, counts models.UpdateRunStatusCounts, countAllAgents int) render.ChartSnippet {
	pie := charts.NewPie()

	pieData := []opts.PieData{}
	if countAllAgents > 0 {
		pieData = []opts.PieData{
			{Name: i18n.T(ctx, "charts.updates_run_finished"), Value: counts.Finished},
			{Name: i18n.T(ctx, "charts.updates_run_failed"), Value: counts.Failed},
			{Name: i18n.T(ctx, "charts.updates_run_running"), Value: counts.Running},
			{Name: i18n.T(ctx, "charts.updates_run_never"), Value: counts.Never},
			{Name: i18n.T(ctx, "charts.updates_run_unreachable"), Value: counts.Unreachable},
		}
	}

	pie.AddSeries(i18n.T(ctx, "charts.updates_run_status"), pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(false), Formatter: "{b}: {c}"}),
		charts.WithPieChartOpts(opts.PieChart{
			Radius: []string{"40%", "75%"},
			Center: []string{"25%", "50%"},
		}),
	)

	leftTitle := getLeftTitlePercentage(countAllAgents)

	// green for finished, red for failures, blue for running an update and greys for
	// the agents that didn't update yet or didn't answer
	colors := opts.Colors{"#48C639", "#C63948", "#3288bd", "#999999", "#555555"}

	textStyle := opts.TextStyle{FontSize: 36, Color: "#777"}
	labelStyle := opts.TextStyle{Color: "#777"}

	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: strconv.Itoa(countAllAgents), Left: leftTitle, Top: "43%", TitleStyle: &textStyle}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), TextStyle: &labelStyle, Type: "scroll", Orient: "vertical", X: "left", Y: "center", Left: "55%"}),
		charts.WithColorsOpts(colors),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "480px",
			Height: "300px",
		}),
	)

	return pie.RenderSnippet()
}
