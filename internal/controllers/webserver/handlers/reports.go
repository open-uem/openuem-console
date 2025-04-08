package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/invopop/ctxi18n/i18n"
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
	"github.com/open-uem/openuem-console/internal/models"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views/agents_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/reports_views"
	"github.com/open-uem/utils"
)

func (h *Handler) Reports(c echo.Context) error {
	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	return RenderView(c, reports_views.ReportsIndex("| Reports", reports_views.Reports(c, h.SessionManager, h.Version, latestServerRelease.Version, "")))
}

func (h *Handler) GenerateAgentsReport(c echo.Context) error {

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, err := h.GetAgentFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not apply filters", false))
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "")

	allAgents, err := h.Model.GetAgentsByPage(p, *f, true)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not get all agents", false))
	}

	m, err := GetAgentsReport(c, allAgents)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not initiate report", false))
	}

	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not generate report", false))
	}

	err = document.Save(dstPath)
	if err != nil {
		return err
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func GetAgentsReport(c echo.Context, agents []*ent.Agent) (core.Maroto, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithOrientation(orientation.Horizontal).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	tableHeader := []core.Row{
		getPageHeader(i18n.T(c.Request().Context(), "Agents")),
		row.New(5).Add(
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.hostname"), props.Text{Size: 9, Align: align.Left, Left: 3, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "Status"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.os"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.version"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "IP Address"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.last_contact"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
		).WithStyle(&props.Cell{BackgroundColor: getDarkGreenColor()}),
	}

	if err := m.RegisterHeader(tableHeader...); err != nil {
		return nil, err
	}

	m.AddRows(getAgentsTransactions(agents)...)

	return m, nil
}

func getAgentsTransactions(agents []*ent.Agent) []core.Row {
	rows := []core.Row{}

	var contentsRow []core.Row

	for i, agent := range agents {
		osImage := ""
		switch agent.Os {
		case "windows":
			osImage = getWindowsPNG()
		case "debian":
			osImage = getDebianPNG()
		case "ubuntu":
			osImage = getUbuntuPNG()
		}

		r := row.New(4).Add(
			text.NewCol(2, agent.Hostname, props.Text{Size: 8, Left: 3, Align: align.Left}),
			text.NewCol(2, string(agent.AgentStatus), props.Text{Size: 8, Align: align.Center}),
			image.NewFromFileCol(2, osImage, props.Rect{
				Center:  true,
				Percent: 75,
			}),
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

	tags, err := h.Model.GetAllTags()
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		if c.FormValue(fmt.Sprintf("filterByTag%d", tag.ID)) != "" {
			f.Tags = append(f.Tags, tag.ID)
		}
	}

	return &f, nil
}

func (h *Handler) GenerateComputersReport(c echo.Context) error {

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, err := h.GetComputerFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not apply filters", false))
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "")

	allComputers, err := h.Model.GetComputersByPage(p, *f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not get all computers", false))
	}

	m, err := GetComputersReport(c, allComputers)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not initiate report", false))
	}

	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not generate report", false))
	}

	err = document.Save(dstPath)
	if err != nil {
		return err
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func GetComputersReport(c echo.Context, computers []models.Computer) (core.Maroto, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithOrientation(orientation.Horizontal).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	tableHeader := []core.Row{
		getPageHeader(i18n.T(c.Request().Context(), "Computers")),
		row.New(5).Add(
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.hostname"), props.Text{Size: 9, Left: 3, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(1, "OS", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.version"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.username"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(1, i18n.T(c.Request().Context(), "agents.manufacturer"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.model"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, "S/N", props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
		).WithStyle(&props.Cell{BackgroundColor: getDarkGreenColor()}),
	}

	if err := m.RegisterHeader(tableHeader...); err != nil {
		return nil, err
	}

	m.AddRows(getComputersTransactions(computers)...)

	return m, nil
}

func getComputersTransactions(computers []models.Computer) []core.Row {
	rows := []core.Row{}

	var contentsRow []core.Row

	for i, computer := range computers {
		osImage := ""
		switch computer.OS {
		case "windows":
			osImage = getWindowsPNG()
		case "debian":
			osImage = getDebianPNG()
		case "ubuntu":
			osImage = getUbuntuPNG()
		}

		r := row.New(4).Add(
			text.NewCol(2, computer.Hostname, props.Text{Size: 8, Left: 3, Align: align.Left}),
			image.NewFromFileCol(1, osImage, props.Rect{
				Center:  true,
				Percent: 75,
			}),
			text.NewCol(2, computer.Version, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(2, computer.Username, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(1, computer.Manufacturer, props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, computer.Model, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(2, computer.Serial, props.Text{Size: 7, Align: align.Left}),
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

func (h *Handler) GetComputerFilters(c echo.Context) (*filters.AgentFilter, error) {
	f := filters.AgentFilter{}

	f.Hostname = c.FormValue("filterByHostname")
	f.Username = c.FormValue("filterByUsername")

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

	versions, err := h.Model.GetOSVersions(f)
	if err != nil {
		return nil, err
	}
	filteredVersions := []string{}
	for index := range versions {
		value := c.FormValue(fmt.Sprintf("filterByOSVersion%d", index))
		if value != "" {
			filteredVersions = append(filteredVersions, value)
		}
	}
	f.OSVersions = filteredVersions

	filteredComputerManufacturers := []string{}
	vendors, err := h.Model.GetComputerManufacturers()
	if err != nil {
		return nil, err
	}
	for index := range vendors {
		value := c.FormValue(fmt.Sprintf("filterByComputerManufacturer%d", index))
		if value != "" {
			filteredComputerManufacturers = append(filteredComputerManufacturers, value)
		}
	}
	f.ComputerManufacturers = filteredComputerManufacturers

	filteredComputerModels := []string{}
	models, err := h.Model.GetComputerModels(f)
	if err != nil {
		return nil, err
	}
	for index := range models {
		value := c.FormValue(fmt.Sprintf("filterByComputerModel%d", index))
		if value != "" {
			filteredComputerModels = append(filteredComputerModels, value)
		}
	}
	f.ComputerModels = filteredComputerModels

	tags, err := h.Model.GetAllTags()
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		if c.FormValue(fmt.Sprintf("filterByTag%d", tag.ID)) != "" {
			f.Tags = append(f.Tags, tag.ID)
		}
	}

	return &f, nil
}

func (h *Handler) GenerateAntivirusReport(c echo.Context) error {

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, _, _, err := h.GetAntiviriFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not apply filters", false))
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "")

	allAntiviri, err := h.Model.GetAntiviriByPage(p, *f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not get all antivirus systems", false))
	}

	m, err := GetAntiviriReport(c, allAntiviri)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not initiate report", false))
	}

	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not generate report", false))
	}

	err = document.Save(dstPath)
	if err != nil {
		return err
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func GetAntiviriReport(c echo.Context, antiviri []models.Antivirus) (core.Maroto, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithOrientation(orientation.Horizontal).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	tableHeader := []core.Row{
		getPageHeader(i18n.T(c.Request().Context(), "Antivirus")),
		row.New(5).Add(
			text.NewCol(3, i18n.T(c.Request().Context(), "agents.hostname"), props.Text{Size: 9, Left: 3, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, "OS", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(3, i18n.T(c.Request().Context(), "Antivirus"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "antivirus.enabled"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "antivirus.updated"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
		).WithStyle(&props.Cell{BackgroundColor: getDarkGreenColor()}),
	}

	if err := m.RegisterHeader(tableHeader...); err != nil {
		return nil, err
	}

	m.AddRows(getAntiviriTransactions(antiviri)...)

	return m, nil
}

func getAntiviriTransactions(antiviri []models.Antivirus) []core.Row {
	rows := []core.Row{}

	var contentsRow []core.Row

	for i, antivirus := range antiviri {
		osImage := ""
		switch antivirus.OS {
		case "windows":
			osImage = getWindowsPNG()
		case "debian":
			osImage = getDebianPNG()
		case "ubuntu":
			osImage = getUbuntuPNG()
		}

		r := row.New(4).Add(
			text.NewCol(3, antivirus.Hostname, props.Text{Size: 8, Left: 3, Align: align.Left}),
			image.NewFromFileCol(2, osImage, props.Rect{
				Center:  true,
				Percent: 75,
			}),
			text.NewCol(3, antivirus.Name, props.Text{Size: 8, Align: align.Left}),
			image.NewFromFileCol(2, getCheckEmoji(antivirus.IsActive), props.Rect{
				Center:  true,
				Percent: 75,
			}),
			image.NewFromFileCol(2, getCheckEmoji(antivirus.IsUpdated), props.Rect{
				Center:  true,
				Percent: 75,
			}),
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

func (h *Handler) GenerateUpdatesReport(c echo.Context) error {

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, _, _, err := h.GetSystemUpdatesFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not apply filters", false))
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "")

	allSystemUpdates, err := h.Model.GetSystemUpdatesByPage(p, *f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not get all systems", false))
	}

	m, err := GetSystemUpdatesReport(c, allSystemUpdates)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not initiate report", false))
	}

	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not generate report", false))
	}

	err = document.Save(dstPath)
	if err != nil {
		return err
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func GetSystemUpdatesReport(c echo.Context, updates []models.SystemUpdate) (core.Maroto, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithOrientation(orientation.Horizontal).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	tableHeader := []core.Row{
		getPageHeader(i18n.T(c.Request().Context(), "updates.title")),
		row.New(5).Add(
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.hostname"), props.Text{Size: 9, Left: 3, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(1, "OS", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(3, i18n.T(c.Request().Context(), "updates.status"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "updates.last_search"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "updates.last_install"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(2, i18n.T(c.Request().Context(), "updates.pending_updates"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: getWhiteColor()}),
		).WithStyle(&props.Cell{BackgroundColor: getDarkGreenColor()}),
	}

	if err := m.RegisterHeader(tableHeader...); err != nil {
		return nil, err
	}

	m.AddRows(getSystemUpdatesTransactions(c, updates)...)

	return m, nil
}

func getSystemUpdatesTransactions(c echo.Context, updates []models.SystemUpdate) []core.Row {
	rows := []core.Row{}

	var contentsRow []core.Row

	for i, update := range updates {
		osImage := ""
		switch update.OS {
		case "windows":
			osImage = getWindowsPNG()
		case "debian":
			osImage = getDebianPNG()
		case "ubuntu":
			osImage = getUbuntuPNG()
		}

		r := row.New(4).Add(
			text.NewCol(2, update.Hostname, props.Text{Size: 8, Left: 3, Align: align.Left}),
			image.NewFromFileCol(1, osImage, props.Rect{
				Center:  true,
				Percent: 75,
			}),
			text.NewCol(3, i18n.T(c.Request().Context(), update.SystemUpdateStatus), props.Text{Size: 8, Align: align.Left}),
			text.NewCol(2, update.LastSearch.Format("2006-01-02 15:03"), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, update.LastInstall.Format("2006-01-02 15:03"), props.Text{Size: 8, Align: align.Center}),
			image.NewFromFileCol(2, getWarningEmoji(update.PendingUpdates), props.Rect{
				Center:  true,
				Percent: 75,
			}),
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

func (h *Handler) GenerateSoftwareReport(c echo.Context) error {

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, err := h.GetSoftwareFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not apply filters", false))
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "")

	allSoftware, err := h.Model.GetAppsByPage(p, *f)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not get all software", false))
	}

	m, err := GetSoftwareReport(c, allSoftware)
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not initiate report", false))
	}

	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage("could not generate report", false))
	}

	err = document.Save(dstPath)
	if err != nil {
		return err
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func GetSoftwareReport(c echo.Context, software []models.App) (core.Maroto, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithOrientation(orientation.Horizontal).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	tableHeader := []core.Row{
		getPageHeader(i18n.T(c.Request().Context(), "Software")),
		row.New(5).Add(
			text.NewCol(4, i18n.T(c.Request().Context(), "apps.name"), props.Text{Size: 9, Left: 3, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(4, i18n.T(c.Request().Context(), "apps.publisher"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
			text.NewCol(4, i18n.T(c.Request().Context(), "apps.num_installations"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: getWhiteColor()}),
		).WithStyle(&props.Cell{BackgroundColor: getDarkGreenColor()}),
	}
	if err := m.RegisterHeader(tableHeader...); err != nil {
		return nil, err
	}

	m.AddRows(getSoftwareTransactions(software)...)

	return m, nil
}

func getSoftwareTransactions(apps []models.App) []core.Row {
	var contentsRow []core.Row

	rows := []core.Row{}

	for i, app := range apps {
		r := row.New(4).Add(
			text.NewCol(4, app.Name, props.Text{Size: 8, Left: 3, Align: align.Left}),
			text.NewCol(4, app.Publisher, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(4, strconv.Itoa(app.Count), props.Text{Size: 8, Align: align.Left}),
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

func getPageHeader(title string) core.Row {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return nil
	}

	return row.New(10).Add(
		image.NewFromFileCol(3, filepath.Join(cwd, "assets", "img", "openuem.png"), props.Rect{
			Percent: 75,
		}),
		text.NewCol(6, title, props.Text{
			Top:   3,
			Style: fontstyle.Bold,
			Align: align.Center,
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

func getCheckEmoji(value bool) string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}

	if value {
		return filepath.Join(cwd, "assets", "img", "reports", "check.png")
	} else {
		return filepath.Join(cwd, "assets", "img", "reports", "x.png")
	}
}

func getWarningEmoji(value bool) string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	if value {
		return filepath.Join(cwd, "assets", "img", "reports", "warning.png")
	} else {
		return filepath.Join(cwd, "assets", "img", "reports", "check.png")
	}
}

func getWindowsPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "windows.png")
}

func getDebianPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "debian.png")
}

func getUbuntuPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "ubuntu.png")
}
