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

// DeploymentsByStatus shows how many software deployments are installed, failed or
// still waiting for the agent's result, with the total number of deployments in
// the center of the pie
func DeploymentsByStatus(ctx context.Context, counts models.DeploymentStatusCounts, countAllDeployments int) render.ChartSnippet {
	pie := charts.NewPie()

	pieData := []opts.PieData{}
	if countAllDeployments > 0 {
		pieData = []opts.PieData{
			{Name: i18n.T(ctx, "charts.deployment_installed"), Value: counts.Installed},
			{Name: i18n.T(ctx, "charts.deployment_failed"), Value: counts.Failed},
			{Name: i18n.T(ctx, "charts.deployment_pending"), Value: counts.Pending},
		}
	}

	pie.AddSeries(i18n.T(ctx, "charts.deployments_status"), pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(false), Formatter: "{b}: {c}"}),
		charts.WithPieChartOpts(opts.PieChart{
			Radius: []string{"40%", "75%"},
			Center: []string{"25%", "50%"},
		}),
	)

	leftTitle := getLeftTitlePercentage(countAllDeployments)

	colors := opts.Colors{"#48C639", "#C63948", "#f0ad4e"}

	textStyle := opts.TextStyle{FontSize: 36, Color: "#777"}
	labelStyle := opts.TextStyle{Color: "#777"}

	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: strconv.Itoa(countAllDeployments), Left: leftTitle, Top: "43%", TitleStyle: &textStyle}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), TextStyle: &labelStyle, Type: "scroll", Orient: "vertical", X: "left", Y: "center", Left: "50%"}),
		charts.WithColorsOpts(colors),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "480px",
			Height: "300px",
		}),
	)

	return pie.RenderSnippet()
}
