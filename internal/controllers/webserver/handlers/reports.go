package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/johnfercher/maroto/v2"
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
	"github.com/open-uem/openuem-console/internal/views/agents_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/reports_views"
)

func (h *Handler) Reports(c echo.Context, successMessage string) error {
	return RenderView(c, reports_views.ReportsIndex("| Reports", reports_views.Reports(c, h.SessionManager, successMessage)))
}

func (h *Handler) GenerateReport(c echo.Context, successMessage string) error {

	dstPath := filepath.Join(h.DownloadDir, "report_test.pdf")

	f, err := h.GetAgentFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not apply filters", false))
	}

	allAgents, err := h.Model.GetAllAgents(*f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not get all agents", false))
	}

	m, err := GetAgentsReport(allAgents)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not initiate report", false))
	}

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

func GetAgentsReport(agents []*ent.Agent) (core.Maroto, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithOrientation(orientation.Horizontal).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	if err := m.RegisterHeader(getPageHeader()); err != nil {
		return nil, err
	}

	m.AddRows(text.NewRow(10, "Agents List", props.Text{
		Top:   3,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))

	m.AddRows(getTransactions(agents)...)

	return m, nil
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
	return row.New(10).Add(
		image.NewFromFileCol(3, "assets/img/openuem.png", props.Rect{
			Percent: 75,
		}),
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

func (h *Handler) GetAgentFilters(c echo.Context) (*filters.AgentFilter, error) {
	f := filters.AgentFilter{}

	f.Hostname = c.FormValue("filterByHostname")

	filteredAgentStatusOptions := []string{}
	for index := range agents_views.AgentStatus {
		value := c.FormValue(fmt.Sprintf("filterByStatusAgent%d", index))
		if value != "" {
			filteredAgentStatusOptions = append(filteredAgentStatusOptions, value)
		}
	}
	f.AgentStatusOptions = filteredAgentStatusOptions

	availableOSes, err := h.Model.GetAgentsUsedOSes()
	if err != nil {
		return nil, err
	}
	filteredAgentOSes := []string{}
	for index := range availableOSes {
		value := c.FormValue(fmt.Sprintf("filterByAgentOS%d", index))
		if value != "" {
			filteredAgentOSes = append(filteredAgentOSes, value)
		}
	}
	f.AgentOSVersions = filteredAgentOSes

	appliedTags, err := h.Model.GetAppliedTags()
	if err != nil {
		return nil, err
	}

	for _, tag := range appliedTags {
		if c.FormValue(fmt.Sprintf("filterByTag%d", tag.ID)) != "" {
			f.Tags = append(f.Tags, tag.ID)
		}
	}

	contactFrom := c.FormValue("filterByContactDateFrom")
	if contactFrom != "" {
		f.ContactFrom = contactFrom
	}
	contactTo := c.FormValue("filterByContactDateTo")
	if contactTo != "" {
		f.ContactTo = contactTo
	}

	return &f, nil
}
