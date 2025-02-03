package handlers

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/reports_views"
)

func (h *Handler) Reports(c echo.Context, successMessage string) error {
	return RenderView(c, reports_views.ReportsIndex("| Reports", reports_views.Reports(c, h.SessionManager, successMessage)))
}

func (h *Handler) GenerateReport(c echo.Context, successMessage string) error {

	dstPath := filepath.Join(h.DownloadDir, "report_test.pdf")

	allAgents, err := h.Model.GetAllAgents(filters.AgentFilter{})
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not get all agents", false))
	}

	m := GetMaroto(allAgents)
	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not generate report", false))
	}

	err = document.Save(dstPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Redirect to file
	url := "/download/report_test.pdf"
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func GetMaroto(agents []*ent.Agent) core.Maroto {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithOrientation(orientation.Horizontal).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	err := m.RegisterHeader(getPageHeader())
	if err != nil {
		log.Fatal(err.Error())
	}

	m.AddRows(text.NewRow(10, "Agents List", props.Text{
		Top:   3,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))

	m.AddRows(getTransactions(agents)...)

	return m
}

func getTransactions(agents []*ent.Agent) []core.Row {
	rows := []core.Row{
		row.New(5).Add(
			text.NewCol(2, "Hostname", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, "Status", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, "OS", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, "Version", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, "IP Address", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, "Last Contact", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
		).WithStyle(&props.Cell{BackgroundColor: getDarkGreenColor()}),
	}

	var contentsRow []core.Row

	for i, agent := range agents {
		r := row.New(4).Add(
			text.NewCol(2, agent.Hostname, props.Text{Size: 8, Left: 3, Align: align.Left}),
			text.NewCol(2, string(agent.AgentStatus), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, agent.Os, props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, agent.Edges.Release.Version, props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, agent.IP, props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, agent.LastContact.Format("2006-01-02 15:03"), props.Text{Size: 8, Align: align.Center}),
		)
		if i%2 == 0 {
			gray := getLightGreenColor()
			r.WithStyle(&props.Cell{BackgroundColor: gray})
		}

		contentsRow = append(contentsRow, r)
	}

	rows = append(rows, contentsRow...)

	return rows
}

func getPageHeader() core.Row {
	return row.New(20).Add(
		image.NewFromFileCol(3, "assets/img/openuem.png", props.Rect{
			Percent: 50,
		}),
		col.New(6),
	)
}

func getDarkGreenColor() *props.Color {
	return &props.Color{
		Red:   0,
		Green: 117,
		Blue:  0,
	}
}

func getLightGreenColor() *props.Color {
	return &props.Color{
		Red:   143,
		Green: 204,
		Blue:  143,
	}
}

func getWhiteColor() *props.Color {
	return &props.Color{
		Red:   255,
		Green: 255,
		Blue:  255,
	}
}
