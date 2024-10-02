package charts

import (
	"context"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/invopop/ctxi18n/i18n"
)

func AgentsByLastReportDate(ctx context.Context, countReportedLast24, countAllAgents int) render.ChartSnippet {
	pie := charts.NewPie()

	pieData := []opts.PieData{}

	// preformat data
	if countAllAgents > 0 {
		pieData = []opts.PieData{
			{Name: i18n.T(ctx, "charts.last_contact_less_24"), Value: countReportedLast24},
			{Name: i18n.T(ctx, "charts.last_contact_more_24"), Value: countAllAgents - countReportedLast24},
		}
	}

	// put data into chart
	pie.AddSeries(i18n.T(ctx, "charts.last_report"), pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(false), Formatter: "{b}: {c}"}),
		charts.WithPieChartOpts(opts.PieChart{
			Radius: []string{"40%", "75%"},
		}),
	)

	textStyle := opts.TextStyle{FontSize: 36}
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: strconv.Itoa(countAllAgents), Left: "center", Top: "center", TitleStyle: &textStyle}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Type: "scroll"}),
		charts.WithColorsOpts(opts.Colors{"#48C639", "#C63948"}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "300px",
			Height: "300px",
		}),
	)

	return pie.RenderSnippet()
}
