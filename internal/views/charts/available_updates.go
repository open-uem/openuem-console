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

// AvailableUpdates shows how many software updates the computers need, in four buckets:
// none, under five, five to twenty, and more than twenty. The number in the center of
// the pie is the number of computers that answered the console's last count, computers
// that didn't answer aren't counted
func AvailableUpdates(ctx context.Context, counts models.AvailableUpdatesCounts) render.ChartSnippet {
	pie := charts.NewPie()

	total := counts.Total()

	pieData := []opts.PieData{}
	if total > 0 {
		pieData = []opts.PieData{
			{Name: i18n.T(ctx, "charts.updates_needed_zero"), Value: counts.Zero},
			{Name: i18n.T(ctx, "charts.updates_needed_under_five"), Value: counts.UnderFive},
			{Name: i18n.T(ctx, "charts.updates_needed_five_to_twenty"), Value: counts.FiveToTwenty},
			{Name: i18n.T(ctx, "charts.updates_needed_over_twenty"), Value: counts.OverTwenty},
		}
	}

	pie.AddSeries(i18n.T(ctx, "charts.updates_needed_status"), pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(false), Formatter: "{b}: {c}"}),
		charts.WithPieChartOpts(opts.PieChart{
			Radius: []string{"40%", "75%"},
			Center: []string{"25%", "50%"},
		}),
	)

	leftTitle := getLeftTitlePercentage(total)

	// from green, no updates needed, to red, a lot of them
	colors := opts.Colors{"#abdda4", "#fee08b", "#fdae61", "#d53e4f"}

	textStyle := opts.TextStyle{FontSize: 36, Color: "#777"}
	labelStyle := opts.TextStyle{Color: "#777"}

	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: strconv.Itoa(total), Left: leftTitle, Top: "43%", TitleStyle: &textStyle}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), TextStyle: &labelStyle, Type: "scroll", Orient: "vertical", X: "left", Y: "center", Left: "50%"}),
		charts.WithColorsOpts(colors),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "480px",
			Height: "300px",
		}),
	)

	return pie.RenderSnippet()
}
