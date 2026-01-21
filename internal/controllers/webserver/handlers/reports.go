package handlers

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/ent"
	"github.com/open-uem/openuem-console/internal/models"
	"github.com/open-uem/openuem-console/internal/views/agents_views"
	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem-console/internal/views/reports_views"
	"github.com/open-uem/utils"
	"github.com/xuri/excelize/v2"
)

func (h *Handler) Reports(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	return RenderView(c, reports_views.ReportsIndex("| Reports", reports_views.Reports(c, "", commonInfo), commonInfo))
}

func (h *Handler) GenerateCSVReports(c echo.Context) error {

	fileName := uuid.NewString() + ".csv"
	dstPath := filepath.Join(h.DownloadDir, fileName)
	csvFile, err := os.Create(dstPath)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_create_file"), false))
	}
	defer func() {
		if err := csvFile.Close(); err != nil {
			log.Printf("[ERROR]: could not close CSV file, reason: %v", err)
		}
	}()

	w := csv.NewWriter(csvFile)

	report := c.Param("report")
	switch report {
	case "agents":
		return h.GenerateAgentsCSVReport(c, w, fileName)
	case "computers":
		return h.GenerateComputersCSVReport(c, w, fileName)
	case "software":
		return h.GenerateSoftwareCSVReport(c, w, fileName)
	case "antivirus":
		return h.GenerateAntivirusCSVReport(c, w, fileName)
	case "updates":
		return h.GenerateUpdatesCSVReport(c, w, fileName)
	default:
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.invalid_report_selected"), false))
	}

}

func (h *Handler) GenerateAgentsCSVReport(c echo.Context, w *csv.Writer, fileName string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	f, err := h.GetAgentFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allAgents, err := h.Model.GetAgentsByPage(p, *f, true, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_all_agents"), false))
	}

	w.Write([]string{"name", "status", "os", "version", "ip", "last_contact"})

	for _, agent := range allAgents {
		record := []string{agent.Nickname, string(agent.AgentStatus), agent.Os, agent.Edges.Release.Version, agent.IP, agent.LastContact.Format("2006-01-02T15:03:04")}
		if err := w.Write(record); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func (h *Handler) GenerateComputersCSVReport(c echo.Context, w *csv.Writer, fileName string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	f, err := h.GetComputerFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allComputers, err := h.Model.GetComputersByPage(p, *f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_all_computers"), false))
	}

	w.Write([]string{"name", "os", "version", "username", "manufacturer", "model", "serial_number"})

	for _, computer := range allComputers {
		record := []string{computer.Nickname, computer.OS, computer.Version, computer.Username, computer.Manufacturer, computer.Model, computer.Serial}
		if err := w.Write(record); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func (h *Handler) GenerateSoftwareCSVReport(c echo.Context, w *csv.Writer, fileName string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	f, err := h.GetSoftwareFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allSoftware, err := h.Model.GetAppsByPage(p, *f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_all_software"), false))
	}

	w.Write([]string{"name", "publisher", "#installations"})

	for _, software := range allSoftware {
		record := []string{software.Name, software.Publisher, strconv.Itoa(software.Count)}
		if err := w.Write(record); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func (h *Handler) GenerateAntivirusCSVReport(c echo.Context, w *csv.Writer, fileName string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	f, err := h.GetEDRFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allAntiviri, err := h.Model.GetAntiviriByPage(p, *f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_all_antiviri"), false))
	}

	w.Write([]string{"name", "os", "antivirus", "antivirus_enabled", "antivirus_updated"})

	for _, antivirus := range allAntiviri {
		record := []string{antivirus.Nickname, antivirus.OS, antivirus.Name, strconv.FormatBool(antivirus.IsActive), strconv.FormatBool(antivirus.IsUpdated)}
		if err := w.Write(record); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func (h *Handler) GenerateUpdatesCSVReport(c echo.Context, w *csv.Writer, fileName string) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	f, _, _, err := h.GetSystemUpdatesFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allSystemUpdates, err := h.Model.GetSystemUpdatesByPage(p, *f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_system_updates"), false))
	}

	w.Write([]string{"name", "os", "antivirus", "antivirus_enabled", "antivirus_updated"})

	for _, update := range allSystemUpdates {
		lastSearch := update.LastSearch.Format("2006-01-02T15:03:04")
		if update.LastSearch.IsZero() {
			lastSearch = "-"
		}

		lastInstall := update.LastInstall.Format("2006-01-02T15:03:04")
		if update.LastInstall.IsZero() {
			lastInstall = "-"
		}

		record := []string{update.Nickname, update.OS, i18n.T(c.Request().Context(), update.SystemUpdateStatus), lastSearch, lastInstall, strconv.FormatBool(update.PendingUpdates)}
		if err := w.Write(record); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_write_to_csv"), false))
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func (h *Handler) GenerateAgentsReport(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, err := h.GetAgentFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allAgents, err := h.Model.GetAgentsByPage(p, *f, true, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_all_agents"), false))
	}

	m, err := GetAgentsReport(c, allAgents)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_initiate_report"), false))
	}

	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_generate_report"), false))
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
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.nickname"), props.Text{Size: 9, Align: align.Left, Left: 3, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "Status"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.os"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.version"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "IP Address"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.last_contact"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
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
		osImage := getOperatingSystemPNG(agent.Os)

		r := row.New(4).Add(
			text.NewCol(2, agent.Nickname, props.Text{Size: 8, Left: 3, Align: align.Left}),
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

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return nil, err
	}

	f.Nickname = c.FormValue("filterByNickname")

	filteredAgentStatusOptions := []string{}
	for index := range agents_views.AgentStatus {
		value := c.FormValue(fmt.Sprintf("filterByStatusAgent%d", index))
		if value != "" {
			if value == "No Contact" {
				f.NoContact = true
			}
			filteredAgentStatusOptions = append(filteredAgentStatusOptions, value)
		}
	}
	f.AgentStatusOptions = filteredAgentStatusOptions

	availableOSes, err := h.Model.GetAgentsUsedOSes(commonInfo, f, false)
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

	appliedTags, err := h.Model.GetAppliedTags(commonInfo)
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

	tags, err := h.Model.GetAllTags(commonInfo, f)
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
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, err := h.GetComputerFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allComputers, err := h.Model.GetComputersByPage(p, *f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_all_computers"), false))
	}

	m, err := GetComputersReport(c, allComputers)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_initiate_report"), false))
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
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.nickname"), props.Text{Size: 9, Left: 3, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(1, "OS", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.version"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.username"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.manufacturer"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(1, i18n.T(c.Request().Context(), "agents.model"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, "S/N", props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
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
		osImage := getOperatingSystemPNG(computer.OS)

		r := row.New(4).Add(
			text.NewCol(2, computer.Nickname, props.Text{Size: 8, Left: 3, Align: align.Left}),
			image.NewFromFileCol(1, osImage, props.Rect{
				Center:  true,
				Percent: 75,
			}),
			text.NewCol(2, computer.Version, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(2, computer.Username, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(2, computer.Manufacturer, props.Text{Size: 7, Align: align.Left}),
			text.NewCol(1, computer.Model, props.Text{Size: 7, Align: align.Left}),
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

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return nil, err
	}

	f.Nickname = c.FormValue("filterByNickname")
	f.Username = c.FormValue("filterByUsername")
	f.Search = c.FormValue("filterBySearch")

	availableOSes, err := h.Model.GetAgentsUsedOSes(commonInfo, f, false)
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

	versions, err := h.Model.GetOSVersions(f, commonInfo)
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
	vendors, err := h.Model.GetComputerManufacturers(commonInfo, f)
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
	models, err := h.Model.GetComputerModels(f, commonInfo)
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

	tags, err := h.Model.GetAllTags(commonInfo, f)
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
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, err := h.GetEDRFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allAntiviri, err := h.Model.GetAntiviriByPage(p, *f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_all_antiviri"), false))
	}

	m, err := GetAntiviriReport(c, allAntiviri)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_initiate_report"), false))
	}

	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_generate_report"), false))
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
			text.NewCol(3, i18n.T(c.Request().Context(), "agents.nickname"), props.Text{Size: 9, Left: 3, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, "OS", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(3, i18n.T(c.Request().Context(), "Antivirus"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "antivirus.enabled"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "antivirus.updated"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
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
		osImage := getOperatingSystemPNG(antivirus.OS)

		r := row.New(4).Add(
			text.NewCol(3, antivirus.Nickname, props.Text{Size: 8, Left: 3, Align: align.Left}),
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
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, _, _, err := h.GetSystemUpdatesFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allSystemUpdates, err := h.Model.GetSystemUpdatesByPage(p, *f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_system_updates"), false))
	}

	m, err := GetSystemUpdatesReport(c, allSystemUpdates)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_initiate_report"), false))
	}

	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_generate_report"), false))
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
			text.NewCol(2, i18n.T(c.Request().Context(), "agents.nickname"), props.Text{Size: 9, Left: 3, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(1, "OS", props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(3, i18n.T(c.Request().Context(), "updates.status"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "updates.last_search"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "updates.last_install"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(2, i18n.T(c.Request().Context(), "updates.pending_updates"), props.Text{Size: 9, Align: align.Center, Style: fontstyle.Bold, Color: &props.WhiteColor}),
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
		osImage := getOperatingSystemPNG(update.OS)

		lastSearch := update.LastSearch.Format("2006-01-02T15:03:04")
		if update.LastSearch.IsZero() {
			lastSearch = "-"
		}

		lastInstall := update.LastInstall.Format("2006-01-02T15:03:04")
		if update.LastInstall.IsZero() {
			lastInstall = "-"
		}

		r := row.New(4).Add(
			text.NewCol(2, update.Nickname, props.Text{Size: 8, Left: 3, Align: align.Left}),
			image.NewFromFileCol(1, osImage, props.Rect{
				Center:  true,
				Percent: 75,
			}),
			text.NewCol(3, i18n.T(c.Request().Context(), update.SystemUpdateStatus), props.Text{Size: 8, Align: align.Left}),
			text.NewCol(2, lastSearch, props.Text{Size: 8, Align: align.Center}),
			text.NewCol(2, lastInstall, props.Text{Size: 8, Align: align.Center}),
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
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	f, err := h.GetSoftwareFilters(c)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_apply_filters"), false))
	}

	itemsPerPage, err := h.Model.GetDefaultItemsPerPage()
	if err != nil {
		log.Println("[ERROR]: could not get items per page from database")
		itemsPerPage = 5
	}

	p := partials.PaginationAndSort{}
	p.GetPaginationAndSortParams("0", "0", c.FormValue("sortBy"), c.FormValue("sortOrder"), "", itemsPerPage)

	allSoftware, err := h.Model.GetAppsByPage(p, *f, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_get_all_software"), false))
	}

	m, err := GetSoftwareReport(c, allSoftware)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_initiate_report"), false))
	}

	document, err := m.Generate()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_generate_report"), false))
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
			text.NewCol(4, i18n.T(c.Request().Context(), "apps.name"), props.Text{Size: 9, Left: 3, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(4, i18n.T(c.Request().Context(), "apps.publisher"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
			text.NewCol(4, i18n.T(c.Request().Context(), "apps.num_installations"), props.Text{Size: 9, Align: align.Left, Style: fontstyle.Bold, Color: &props.WhiteColor}),
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
			Top:   2,
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

func getMonitorEmoji() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}

	return filepath.Join(cwd, "assets", "img", "reports", "desktop_computer.png")
}

func getNICEmoji() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}

	return filepath.Join(cwd, "assets", "img", "reports", "globe_with_meridians.png")
}

func getDiskEmoji() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}

	return filepath.Join(cwd, "assets", "img", "reports", "floppy_disk.png")
}

func getShareEmoji() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}

	return filepath.Join(cwd, "assets", "img", "reports", "file_cabinet.png")
}

func getMemoryEmoji() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}

	return filepath.Join(cwd, "assets", "img", "reports", "card_file_box.png")
}

func getPrinterEmoji() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}

	return filepath.Join(cwd, "assets", "img", "reports", "printer.png")
}

func getAppEmoji() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}

	return filepath.Join(cwd, "assets", "img", "reports", "computer.png")
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

func getNeonPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "neon.png")
}

func getUbuntuPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "ubuntu.png")
}

func getSuSEPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "suse.png")
}

func getRedHatPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "redhat.png")
}

func getFedoraPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "fedora.png")
}

func getAlmaLinuxPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "almalinux.png")
}

func getRockyLinuxPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "rockylinux.png")
}

func getUnknownPNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "question.png")
}

func getApplePNG() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}
	return filepath.Join(cwd, "assets", "img", "os", "apple.png")
}

func getHDDEmoji() string {
	cwd, err := utils.GetWd()
	if err != nil {
		log.Println("[ERROR]: could not get working directory")
		return ""
	}

	return filepath.Join(cwd, "assets", "img", "reports", "minidisc.png")
}

func getOperatingSystemPNG(os string) string {
	switch os {
	case "windows":
		return getWindowsPNG()
	case "debian":
		return getDebianPNG()
	case "ubuntu":
		return getUbuntuPNG()
	case "opensuse-leap":
		return getSuSEPNG()
	case "fedora":
		return getFedoraPNG()
	case "redhat":
		return getRedHatPNG()
	case "almalinux":
		return getAlmaLinuxPNG()
	case "rocky":
		return getRockyLinuxPNG()
	case "macOS":
		return getApplePNG()
	case "neon":
		return getNeonPNG()
	default:
		return getUnknownPNG()
	}
}

func (h *Handler) GenerateComputerReport(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.computer_id_empty"), true))
	}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	fileName := uuid.NewString() + ".pdf"
	dstPath := filepath.Join(h.DownloadDir, fileName)

	m, err := h.GetComputerReport(c, agentId, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.could_not_initiate_report"), false))
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

func (h *Handler) GetComputerReport(c echo.Context, agentID string, commonInfo *partials.CommonInfo) (core.Maroto, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithOrientation(orientation.Vertical).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	header := []core.Row{
		getPageHeader(i18n.T(c.Request().Context(), "reports.computer_inventory") + " - " + commonInfo.Translator.FmtDateMedium(time.Now())),
	}

	if err := m.RegisterHeader(header...); err != nil {
		log.Println(err)
		return nil, err
	}

	hwInfo, err := h.getComputerInfo(c, agentID, commonInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	m.AddRow(4, col.New(12))
	m.AddRows(hwInfo...)

	return m, nil
}

func (h *Handler) getComputerInfo(c echo.Context, agentID string, commonInfo *partials.CommonInfo) ([]core.Row, error) {
	var r core.Row

	rows := []core.Row{}
	lightGreen := getLightGreenColor()

	hwInfo, err := h.Model.GetAgentComputerInfo(agentID, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: report error %v", err)
		return nil, err
	}

	osInfo, err := h.Model.GetAgentOSInfo(agentID, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: report error %v", err)
		return nil, err
	}

	monitorsInfo, err := h.Model.GetAgentMonitorsInfo(agentID, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: report error %v", err)
		return nil, err
	}

	ldInfo, err := h.Model.GetAgentLogicalDisksInfo(agentID, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: report error %v", err)
		return nil, err
	}

	pdInfo, err := h.Model.GetAgentPhysicalDisksInfo(agentID, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: report error %v", err)
		return nil, err
	}

	sharesInfo, err := h.Model.GetAgentSharesInfo(agentID, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: report error %v", err)
		return nil, err
	}

	printersInfo, err := h.Model.GetAgentPrintersInfo(agentID, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: report error %v", err)
		return nil, err
	}

	nicInfo, err := h.Model.GetAgentNetworkAdaptersInfo(agentID, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: report error %v", err)
		return nil, err
	}

	swInfo, err := h.Model.GetAgentAppsInfo(agentID, commonInfo)
	if err != nil {
		log.Printf("[ERROR]: report error %v", err)
		return nil, err
	}

	// Computer's name
	osImage := getOperatingSystemPNG(hwInfo.Os)
	if hwInfo.Nickname != hwInfo.Hostname {
		r = row.New(4).Add(
			text.NewCol(12, hwInfo.Nickname+" ( "+hwInfo.Hostname+" )", props.Text{Size: 9, Align: align.Left, Style: "B"}),
		)
	} else {
		r = row.New(4).Add(
			text.NewCol(12, hwInfo.Nickname, props.Text{Size: 9, Align: align.Left, Style: "B"}),
		)
	}

	rows = append(rows, r)

	// Empty row
	r = row.New(4).Add(col.New(12))
	rows = append(rows, r)

	// Manufacturer, Model and serial
	if hwInfo.Edges.Computer != nil {
		r = row.New(5).Add(
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.manufacturer"), props.Text{Size: 6, Align: align.Left, Left: 1.2, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(3, hwInfo.Edges.Computer.Manufacturer, props.Text{Size: 8, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.model"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(3, hwInfo.Edges.Computer.Model, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.serial"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(3, hwInfo.Edges.Computer.Serial, props.Text{Size: 8, Align: align.Left, Left: 1, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
		)
		rows = append(rows, r)
	}

	// Processor info
	if hwInfo.Edges.Computer != nil {
		r = row.New(5).Add(
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.processor"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(5, hwInfo.Edges.Computer.Processor, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.architecture"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(2, hwInfo.Edges.Computer.ProcessorArch, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.num_cores"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(2, strconv.Itoa(int(hwInfo.Edges.Computer.ProcessorCores)), props.Text{Size: 8, Align: align.Left, Left: 1, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
		)
		rows = append(rows, r)
	}

	// Empty row
	r = row.New(4).Add(col.New(12))
	rows = append(rows, r)

	// OS info
	if osInfo.Edges.Operatingsystem != nil {
		r = row.New(5).Add(
			image.NewFromFileCol(1, osImage, props.Rect{
				Percent: 75,
				Center:  true,
			}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(4, i18n.T(c.Request().Context(), "inventory.os.title"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
		)
		rows = append(rows, r)

		r = row.New(5).Add(
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.os.version"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(3, osInfo.Edges.Operatingsystem.Version, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.os.desc"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(3, osInfo.Edges.Operatingsystem.Description, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.os.architecture"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(3, osInfo.Edges.Operatingsystem.Arch, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
		)
		rows = append(rows, r)

		r = row.New(5).Add(
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.os.username"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(3, osInfo.Edges.Operatingsystem.Username, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(2, i18n.T(c.Request().Context(), "inventory.os.installation"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(2, commonInfo.Translator.FmtDateMedium(osInfo.Edges.Operatingsystem.InstallDate.Local()), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(1, i18n.T(c.Request().Context(), "inventory.os.last_bootup"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(3, commonInfo.Translator.FmtDateMedium(osInfo.Edges.Operatingsystem.LastBootupTime.Local())+" "+commonInfo.Translator.FmtTimeShort(osInfo.Edges.Operatingsystem.LastBootupTime.Local()), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
		)
		rows = append(rows, r)
	}

	// Empty row
	r = row.New(4).Add(col.New(12))
	rows = append(rows, r)

	// Memory info
	if hwInfo.Edges.Computer != nil {
		r = row.New(5).Add(
			image.NewFromFileCol(1, getMemoryEmoji(), props.Rect{
				Percent: 75,
				Center:  true,
			}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(2, i18n.T(c.Request().Context(), "inventory.hardware.memory"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(2, fmt.Sprintf("%d MB", hwInfo.Edges.Computer.Memory), props.Text{Size: 7, Align: align.Center, Top: 0.8}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
		)
		rows = append(rows, r)

		for _, mSlot := range hwInfo.Edges.Memoryslots {
			r = row.New(5).Add(
				text.NewCol(2, mSlot.Slot, props.Text{Size: 7, Align: align.Center, Top: 0.8}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.size"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, mSlot.Size, props.Text{Size: 7, Align: align.Center, Top: 0.8}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.mem_type"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, mSlot.Type, props.Text{Size: 7, Align: align.Center, Top: 0.8}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.speed"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, mSlot.Speed, props.Text{Size: 7, Align: align.Center, Top: 0.8}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, i18n.T(c.Request().Context(), "inventory.hardware.vendor"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(3, mSlot.Manufacturer+" "+mSlot.SerialNumber, props.Text{Size: 7, Align: align.Center, Top: 0.8}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)
		}
	}

	// Empty row
	r = row.New(4).Add(col.New(12))
	rows = append(rows, r)

	// NICS info
	if len(nicInfo.Edges.Networkadapters) > 0 {
		for index, nic := range nicInfo.Edges.Networkadapters {
			r = row.New(5).Add(
				image.NewFromFileCol(1, getNICEmoji(), props.Rect{
					Percent: 75,
					Center:  true,
				}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(4, i18n.T(c.Request().Context(), "inventory.network_adapters.report_title", strconv.Itoa(index+1)), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(1, i18n.T(c.Request().Context(), "Name"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(3, nic.Name, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.network_adapters.ip_address"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, nic.Addresses, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.network_adapters.mac_address"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, nic.MACAddress, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.network_adapters.default_gateway"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, nic.DefaultGateway, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.network_adapters.subnet"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, nic.Subnet, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.network_adapters.dhcp"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				image.NewFromFileCol(2, getCheckEmoji(nic.DhcpEnabled), props.Rect{
					Center:  true,
					Percent: 75,
				}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.network_adapters.dns"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(6, nic.DNSServers, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.network_adapters.speed"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, nic.Speed, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			// Empty row
			r = row.New(4).Add(col.New(12))
			rows = append(rows, r)
		}
	}

	// Monitors info
	if len(monitorsInfo.Edges.Monitors) > 0 {
		for index, monitor := range monitorsInfo.Edges.Monitors {
			r = row.New(5).Add(
				image.NewFromFileCol(1, getMonitorEmoji(), props.Rect{
					Percent: 75,
					Center:  true,
				}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(4, i18n.T(c.Request().Context(), "inventory.monitor.report_title", strconv.Itoa(index+1)), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.monitor.manufacturer"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, monitor.Manufacturer, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.monitor.model"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, monitor.Model, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, i18n.T(c.Request().Context(), "inventory.monitor.serial"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(3, monitor.Serial, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.monitor.week_of_manufacture"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, monitor.WeekOfManufacture, props.Text{Size: 7, Align: align.Center, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.monitor.year_of_manufacture"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, monitor.YearOfManufacture, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(4, "", props.Text{Size: 7, Align: align.Center, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: &props.WhiteColor, BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(4).Add(col.New(12))
			rows = append(rows, r)
		}
	}

	// Logical disks info
	if len(ldInfo.Edges.Logicaldisks) > 0 {
		for _, ld := range ldInfo.Edges.Logicaldisks {
			if hwInfo.Os == "windows" {
				r = row.New(5).Add(
					image.NewFromFileCol(1, getDiskEmoji(), props.Rect{
						Percent: 75,
						Center:  true,
					}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(4, i18n.T(c.Request().Context(), "inventory.logical_disk.report_label", ld.Label), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				)
				rows = append(rows, r)

				r = row.New(5).Add(
					text.NewCol(2, i18n.T(c.Request().Context(), "inventory.logical_disk.volume_name"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(2, ld.VolumeName, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(1, i18n.T(c.Request().Context(), "inventory.logical_disk.filesystem"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(3, ld.Filesystem, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(2, i18n.T(c.Request().Context(), "inventory.logical_disk.usage"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(2, fmt.Sprintf("%d %%", ld.Usage), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				)
				rows = append(rows, r)

			} else {
				r = row.New(5).Add(
					image.NewFromFileCol(1, getDiskEmoji(), props.Rect{
						Percent: 75,
						Center:  true,
					}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(4, i18n.T(c.Request().Context(), "inventory.logical_disk.report_mount_point", ld.Label), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				)
				rows = append(rows, r)

				r = row.New(5).Add(
					text.NewCol(1, i18n.T(c.Request().Context(), "inventory.logical_disk.filesystem"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(3, ld.VolumeName, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(2, i18n.T(c.Request().Context(), "inventory.logical_disk.filesystem_type"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(2, ld.Filesystem, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(2, i18n.T(c.Request().Context(), "inventory.logical_disk.usage"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
					text.NewCol(2, fmt.Sprintf("%d %%", ld.Usage), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				)
				rows = append(rows, r)

			}

			r = row.New(5).Add(
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.logical_disk.remaining_space"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, ld.RemainingSpaceInUnits, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.logical_disk.total_size"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, ld.SizeInUnits, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.logical_disk.bitlocker"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, ld.BitlockerStatus, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(4).Add(col.New(12))
			rows = append(rows, r)
		}
	}

	// Physical disks info
	if len(pdInfo.Edges.Physicaldisks) > 0 {
		r = row.New(5).Add(
			image.NewFromFileCol(1, getHDDEmoji(), props.Rect{
				Percent: 75,
				Center:  true,
			}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			text.NewCol(4, i18n.T(c.Request().Context(), "inventory.physical_disk.title"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
		)
		rows = append(rows, r)

		for _, pd := range pdInfo.Edges.Physicaldisks {
			r = row.New(5).Add(
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.physical_disk.model"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(3, pd.Model, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.physical_disk.serial"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(3, pd.SerialNumber, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, i18n.T(c.Request().Context(), "inventory.physical_disk.size"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, pd.SizeInUnits, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)
		}

		r = row.New(4).Add(col.New(12))
		rows = append(rows, r)
	}

	// Shares info
	if len(sharesInfo.Edges.Shares) > 0 {
		for index, share := range sharesInfo.Edges.Shares {
			r = row.New(5).Add(
				image.NewFromFileCol(1, getShareEmoji(), props.Rect{
					Percent: 75,
					Center:  true,
				}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(4, i18n.T(c.Request().Context(), "inventory.share.report_title", strconv.Itoa(index+1)), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(1, i18n.T(c.Request().Context(), "inventory.share.name"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(5, share.Name, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(1, i18n.T(c.Request().Context(), "inventory.share.descr"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(5, share.Description, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(1, i18n.T(c.Request().Context(), "inventory.share.path"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(5, share.Path, props.Text{Size: 7, Align: align.Center, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(6, "", props.Text{Size: 7, Align: align.Center, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: &props.WhiteColor, BorderColor: &props.BlackColor, BorderType: border.Full}),
			)

			rows = append(rows, r)

			r = row.New(4).Add(col.New(12))
			rows = append(rows, r)
		}
	}

	// Printers info
	if len(printersInfo) > 0 {
		for _, printer := range printersInfo {
			r = row.New(5).Add(
				image.NewFromFileCol(1, getPrinterEmoji(), props.Rect{
					Percent: 75,
					Center:  true,
				}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(4, i18n.T(c.Request().Context(), "inventory.printers.report_title", printer.Name), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.printers.port"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(10, printer.Port, props.Text{Size: 7, Align: align.Center, Top: 0.7}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.printers.is_default"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				image.NewFromFileCol(2, getCheckEmoji(printer.IsDefault), props.Rect{
					Center:  true,
					Percent: 75,
				}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.printers.is_network_printer"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				image.NewFromFileCol(2, getCheckEmoji(printer.IsNetwork), props.Rect{
					Center:  true,
					Percent: 75,
				}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.printers.is_shared_printer"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				image.NewFromFileCol(2, getCheckEmoji(printer.IsShared), props.Rect{
					Center:  true,
					Percent: 75,
				}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(4).Add(col.New(12))
			rows = append(rows, r)
		}
	}

	// Apps info
	if len(swInfo) > 0 {
		for _, app := range swInfo {
			r = row.New(5).Add(
				image.NewFromFileCol(1, getAppEmoji(), props.Rect{
					Percent: 75,
					Center:  true,
				}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(8, app.Name, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(3, app.Version, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(5).Add(
				text.NewCol(1, i18n.T(c.Request().Context(), "apps.publisher"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(7, app.Publisher, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, i18n.T(c.Request().Context(), "inventory.apps.installation_date"), props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BackgroundColor: lightGreen, BorderColor: &props.BlackColor, BorderType: border.Full}),
				text.NewCol(2, app.InstallDate, props.Text{Size: 7, Align: align.Left, Left: 1, Top: 1}).WithStyle(&props.Cell{BorderColor: &props.BlackColor, BorderType: border.Full}),
			)
			rows = append(rows, r)

			r = row.New(4).Add(col.New(12))
			rows = append(rows, r)
		}
	}

	return rows, nil
}

func (h *Handler) GenerateComputerODSReport(c echo.Context) error {
	fileName := uuid.NewString() + ".xlsx"

	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "reports.computer_id_empty"), true))
	}

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	hwInfo, err := h.Model.GetAgentComputerInfo(agentID, commonInfo)
	if err != nil {
		return err
	}

	osInfo, err := h.Model.GetAgentOSInfo(agentID, commonInfo)
	if err != nil {
		return err
	}

	monitorsInfo, err := h.Model.GetAgentMonitorsInfo(agentID, commonInfo)
	if err != nil {
		return err
	}

	ldInfo, err := h.Model.GetAgentLogicalDisksInfo(agentID, commonInfo)
	if err != nil {
		return err
	}

	pdInfo, err := h.Model.GetAgentPhysicalDisksInfo(agentID, commonInfo)
	if err != nil {
		return err
	}

	sharesInfo, err := h.Model.GetAgentSharesInfo(agentID, commonInfo)
	if err != nil {
		return err
	}

	printersInfo, err := h.Model.GetAgentPrintersInfo(agentID, commonInfo)
	if err != nil {
		return err
	}

	nicInfo, err := h.Model.GetAgentNetworkAdaptersInfo(agentID, commonInfo)
	if err != nil {
		return err
	}

	swInfo, err := h.Model.GetAgentAppsInfo(agentID, commonInfo)
	if err != nil {
		return err
	}

	dstPath := filepath.Join(h.DownloadDir, fileName)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	headerStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#007500"}, Pattern: 1},
		Font: &excelize.Font{Color: "#FFFFFF"},
	})
	if err != nil {
		return err
	}

	endpointNameStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#093bc4da", Size: 14, Bold: true},
	})
	if err != nil {
		return err
	}

	tagStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 12, Bold: true},
	})
	if err != nil {
		return err
	}

	// Hardware sheet
	f.SetSheetName("Sheet1", "Hardware")
	if hwInfo.Edges.Computer != nil {
		if err := f.SetCellValue("Hardware", "B2", hwInfo.Nickname); err != nil {
			return err
		}
		if err := f.SetCellStyle("Hardware", "B2", "C2", endpointNameStyle); err != nil {
			return err
		}
		if hwInfo.Nickname != hwInfo.Hostname {
			if err := f.SetCellValue("Hardware", "C2", hwInfo.Hostname); err != nil {
				return err
			}
		}

		if err := f.SetColWidth("Hardware", "B", "G", 30); err != nil {
			return err
		}
		if err := f.SetCellStyle("Hardware", "B4", "D4", headerStyle); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "B4", i18n.T(c.Request().Context(), "inventory.hardware.manufacturer")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "C4", i18n.T(c.Request().Context(), "inventory.hardware.model")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "D4", i18n.T(c.Request().Context(), "inventory.hardware.serial")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "B5", hwInfo.Edges.Computer.Manufacturer); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "C5", hwInfo.Edges.Computer.Model); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "D5", hwInfo.Edges.Computer.Serial); err != nil {
			return err
		}

		if err := f.SetCellStyle("Hardware", "B7", "D7", headerStyle); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "B7", i18n.T(c.Request().Context(), "inventory.hardware.processor")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "C7", i18n.T(c.Request().Context(), "inventory.hardware.architecture")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "D7", i18n.T(c.Request().Context(), "inventory.hardware.num_cores")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "B8", hwInfo.Edges.Computer.Processor); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "C8", hwInfo.Edges.Computer.ProcessorArch); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "D8", hwInfo.Edges.Computer.ProcessorCores); err != nil {
			return err
		}

		if err := f.SetCellStyle("Hardware", "B10", "B10", headerStyle); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "B10", i18n.T(c.Request().Context(), "inventory.hardware.memory")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "B11", fmt.Sprintf("%d MB", hwInfo.Edges.Computer.Memory)); err != nil {
			return err
		}

		if err := f.SetCellStyle("Hardware", "B12", "G12", headerStyle); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "B12", i18n.T(c.Request().Context(), "inventory.hardware.slot")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "C12", i18n.T(c.Request().Context(), "inventory.hardware.size")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "D12", i18n.T(c.Request().Context(), "inventory.hardware.mem_type")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "E12", i18n.T(c.Request().Context(), "inventory.hardware.speed")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "F12", i18n.T(c.Request().Context(), "inventory.hardware.vendor")); err != nil {
			return err
		}
		if err := f.SetCellValue("Hardware", "G12", i18n.T(c.Request().Context(), "inventory.hardware.serial")); err != nil {
			return err
		}

		for i, mSlot := range hwInfo.Edges.Memoryslots {
			if err := f.SetCellValue("Hardware", fmt.Sprintf("B%d", 13+i), mSlot.Slot); err != nil {
				return err
			}
			if err := f.SetCellValue("Hardware", fmt.Sprintf("C%d", 13+i), mSlot.Size); err != nil {
				return err
			}
			if err := f.SetCellValue("Hardware", fmt.Sprintf("D%d", 13+i), mSlot.Type); err != nil {
				return err
			}
			if err := f.SetCellValue("Hardware", fmt.Sprintf("E%d", 13+i), mSlot.Speed); err != nil {
				return err
			}
			if err := f.SetCellValue("Hardware", fmt.Sprintf("F%d", 13+i), mSlot.Manufacturer); err != nil {
				return err
			}
			if err := f.SetCellValue("Hardware", fmt.Sprintf("G%d", 13+i), mSlot.SerialNumber); err != nil {
				return err
			}
		}

	}

	// Create OS sheet.
	osSheetName := i18n.T(c.Request().Context(), "OS")
	_, err = f.NewSheet(osSheetName)
	if err != nil {
		return err
	}

	if osInfo.Edges.Operatingsystem != nil {
		if err := f.SetCellValue(osSheetName, "B2", hwInfo.Nickname); err != nil {
			return err
		}
		if err := f.SetCellStyle(osSheetName, "B2", "C2", endpointNameStyle); err != nil {
			return err
		}
		if hwInfo.Nickname != hwInfo.Hostname {
			if err := f.SetCellValue(osSheetName, "C2", hwInfo.Hostname); err != nil {
				return err
			}
		}

		if err := f.SetColWidth(osSheetName, "B", "D", 50); err != nil {
			return err
		}
		if err := f.SetCellStyle(osSheetName, "B5", "D5", headerStyle); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "B5", i18n.T(c.Request().Context(), "inventory.os.version")); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "C5", i18n.T(c.Request().Context(), "inventory.os.desc")); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "D5", i18n.T(c.Request().Context(), "inventory.os.architecture")); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "B6", osInfo.Edges.Operatingsystem.Version); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "C6", osInfo.Edges.Operatingsystem.Description); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "D6", osInfo.Edges.Operatingsystem.Arch); err != nil {
			return err
		}

		if err := f.SetCellStyle(osSheetName, "B8", "D8", headerStyle); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "B8", i18n.T(c.Request().Context(), "inventory.os.username")); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "C8", i18n.T(c.Request().Context(), "inventory.os.installation")); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "D8", i18n.T(c.Request().Context(), "inventory.os.last_bootup")); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "B9", osInfo.Edges.Operatingsystem.Username); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "C9", commonInfo.Translator.FmtDateMedium(osInfo.Edges.Operatingsystem.InstallDate.Local())); err != nil {
			return err
		}
		if err := f.SetCellValue(osSheetName, "D9", commonInfo.Translator.FmtDateMedium(osInfo.Edges.Operatingsystem.LastBootupTime.Local())+" "+commonInfo.Translator.FmtTimeShort(osInfo.Edges.Operatingsystem.LastBootupTime.Local())); err != nil {
			return err
		}
	}

	// Create Monitors sheet.
	if len(monitorsInfo.Edges.Monitors) > 0 {
		monitorsSheetName := i18n.T(c.Request().Context(), "Monitors")
		_, err = f.NewSheet(monitorsSheetName)
		if err != nil {
			return err
		}

		if err := f.SetColWidth(monitorsSheetName, "B", "D", 50); err != nil {
			return err
		}
		if err := f.SetCellValue(monitorsSheetName, "B2", hwInfo.Nickname); err != nil {
			return err
		}
		if err := f.SetCellStyle(monitorsSheetName, "B2", "C2", endpointNameStyle); err != nil {
			return err
		}
		if hwInfo.Nickname != hwInfo.Hostname {
			if err := f.SetCellValue(monitorsSheetName, "C2", hwInfo.Hostname); err != nil {
				return err
			}
		}

		for index, monitor := range monitorsInfo.Edges.Monitors {
			if err := f.SetCellStyle(monitorsSheetName, fmt.Sprintf("B%d", 4+(index*6)), fmt.Sprintf("B%d", 4+(index*6)), tagStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("B%d", 4+(index*6)), i18n.T(c.Request().Context(), "inventory.monitor.report_title", strconv.Itoa(index+1))); err != nil {
				return err
			}

			if err := f.SetCellStyle(monitorsSheetName, fmt.Sprintf("B%d", 5+(index*6)), fmt.Sprintf("D%d", 5+(index*6)), headerStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("B%d", 5+(index*6)), i18n.T(c.Request().Context(), "inventory.monitor.manufacturer")); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("C%d", 5+(index*6)), i18n.T(c.Request().Context(), "inventory.monitor.model")); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("D%d", 5+(index*6)), i18n.T(c.Request().Context(), "inventory.monitor.serial")); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("B%d", 6+(index*6)), monitor.Manufacturer); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("C%d", 6+(index*6)), monitor.Model); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("D%d", 6+(index*6)), monitor.Serial); err != nil {
				return err
			}

			if err := f.SetCellStyle(monitorsSheetName, fmt.Sprintf("B%d", 7+(index*6)), fmt.Sprintf("C%d", 7+(index*6)), headerStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("B%d", 7+(index*6)), i18n.T(c.Request().Context(), "inventory.monitor.week_of_manufacture")); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("C%d", 7+(index*6)), i18n.T(c.Request().Context(), "inventory.monitor.year_of_manufacture")); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("B%d", 8+(index*6)), monitor.WeekOfManufacture); err != nil {
				return err
			}
			if err := f.SetCellValue(monitorsSheetName, fmt.Sprintf("C%d", 8+(index*6)), monitor.YearOfManufacture); err != nil {
				return err
			}
		}
	}

	// Create Logical disks sheet.
	if len(ldInfo.Edges.Logicaldisks) > 0 {

		sheetName := i18n.T(c.Request().Context(), "Logical Disks")
		_, err = f.NewSheet(sheetName)
		if err != nil {
			return err
		}

		if err := f.SetColWidth(sheetName, "B", "D", 50); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "B2", hwInfo.Nickname); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheetName, "B2", "C2", endpointNameStyle); err != nil {
			return err
		}

		if hwInfo.Nickname != hwInfo.Hostname {
			if err := f.SetCellValue(sheetName, "C2", hwInfo.Hostname); err != nil {
				return err
			}
		}

		for index, ld := range ldInfo.Edges.Logicaldisks {
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 4+(index*6)), fmt.Sprintf("B%d", 4+(index*6)), tagStyle); err != nil {
				return err
			}
			if hwInfo.Os == "windows" {
				if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 4+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.report_label", ld.Label)); err != nil {
					return err
				}
				if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 5+(index*6)), fmt.Sprintf("D%d", 5+(index*6)), headerStyle); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 5+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.volume_name")); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 5+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.filesystem")); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 5+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.usage")); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 6+(index*6)), ld.VolumeName); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 6+(index*6)), ld.Filesystem); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 6+(index*6)), fmt.Sprintf("%d %%", ld.Usage)); err != nil {
					return err
				}
			} else {
				if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 4+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.report_mount_point", ld.Label)); err != nil {
					return err
				}
				if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 5+(index*6)), fmt.Sprintf("D%d", 5+(index*6)), headerStyle); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 5+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.filesystem")); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 5+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.filesystem_type")); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 5+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.usage")); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 6+(index*6)), ld.VolumeName); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 6+(index*6)), ld.Filesystem); err != nil {
					return err
				}
				if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 6+(index*6)), fmt.Sprintf("%d %%", ld.Usage)); err != nil {
					return err
				}

			}

			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 7+(index*6)), fmt.Sprintf("D%d", 7+(index*6)), headerStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 7+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.remaining_space")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 7+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.total_size")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 7+(index*6)), i18n.T(c.Request().Context(), "inventory.logical_disk.bitlocker")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 8+(index*6)), ld.RemainingSpaceInUnits); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 8+(index*6)), ld.SizeInUnits); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 8+(index*6)), ld.BitlockerStatus); err != nil {
				return err
			}
		}
	}

	// Create Physical disks sheet.
	if len(pdInfo.Edges.Physicaldisks) > 0 {

		sheetName := i18n.T(c.Request().Context(), "Physical Disks")
		_, err = f.NewSheet(sheetName)
		if err != nil {
			return err
		}

		if err := f.SetColWidth(sheetName, "B", "D", 50); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "B2", hwInfo.Nickname); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheetName, "B2", "C2", endpointNameStyle); err != nil {
			return err
		}

		if hwInfo.Nickname != hwInfo.Hostname {
			if err := f.SetCellValue(sheetName, "C2", hwInfo.Hostname); err != nil {
				return err
			}
		}

		if err := f.SetCellStyle(sheetName, "B4", "B4", tagStyle); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "B4", i18n.T(c.Request().Context(), "inventory.physical_disk.title")); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheetName, "B5", "D5", headerStyle); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "B5", i18n.T(c.Request().Context(), "inventory.physical_disk.model")); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "C5", i18n.T(c.Request().Context(), "inventory.physical_disk.serial")); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "D5", i18n.T(c.Request().Context(), "inventory.physical_disk.size")); err != nil {
			return err
		}

		for index, pd := range pdInfo.Edges.Physicaldisks {
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 6+(index*6)), pd.Model); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 6+(index*6)), pd.SerialNumber); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 6+(index*6)), pd.SizeInUnits); err != nil {
				return err
			}
		}
	}

	// Shares info
	if len(sharesInfo.Edges.Shares) > 0 {

		sheetName := i18n.T(c.Request().Context(), "Shares")
		_, err = f.NewSheet(sheetName)
		if err != nil {
			return err
		}

		if err := f.SetColWidth(sheetName, "B", "D", 50); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "B2", hwInfo.Nickname); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheetName, "B2", "C2", endpointNameStyle); err != nil {
			return err
		}
		if hwInfo.Nickname != hwInfo.Hostname {
			if err := f.SetCellValue(sheetName, "C2", hwInfo.Hostname); err != nil {
				return err
			}
		}

		for index, share := range sharesInfo.Edges.Shares {
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 4+(index*4)), fmt.Sprintf("B%d", 4+(index*4)), tagStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 4+(index*4)), i18n.T(c.Request().Context(), "inventory.share.report_title", strconv.Itoa(index+1))); err != nil {
				return err
			}

			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 5+(index*4)), fmt.Sprintf("D%d", 5+(index*4)), headerStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 5+(index*4)), i18n.T(c.Request().Context(), "inventory.share.name")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 5+(index*4)), i18n.T(c.Request().Context(), "inventory.share.descr")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 5+(index*4)), i18n.T(c.Request().Context(), "inventory.share.path")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 6+(index*4)), share.Name); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 6+(index*4)), share.Description); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 6+(index*4)), share.Path); err != nil {
				return err
			}
		}
	}

	// Printers info
	if len(printersInfo) > 0 {

		sheetName := i18n.T(c.Request().Context(), "Printers")
		_, err = f.NewSheet(sheetName)
		if err != nil {
			return err
		}

		if err := f.SetColWidth(sheetName, "B", "E", 50); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "B2", hwInfo.Nickname); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheetName, "B2", "C2", endpointNameStyle); err != nil {
			return err
		}
		if hwInfo.Nickname != hwInfo.Hostname {
			if err := f.SetCellValue(sheetName, "C2", hwInfo.Hostname); err != nil {
				return err
			}
		}

		for index, printer := range printersInfo {
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 4+(index*4)), fmt.Sprintf("B%d", 4+(index*4)), tagStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 4+(index*4)), i18n.T(c.Request().Context(), "inventory.printers.report_title", printer.Name)); err != nil {
				return err
			}

			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 5+(index*4)), fmt.Sprintf("E%d", 5+(index*4)), headerStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 5+(index*4)), i18n.T(c.Request().Context(), "inventory.printers.port")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 5+(index*4)), i18n.T(c.Request().Context(), "inventory.printers.is_default")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 5+(index*4)), i18n.T(c.Request().Context(), "inventory.printers.is_network_printer")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", 5+(index*4)), i18n.T(c.Request().Context(), "inventory.printers.is_shared_printer")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 6+(index*4)), printer.Port); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 6+(index*4)), printer.IsDefault); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 6+(index*4)), printer.IsNetwork); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("E%d", 6+(index*4)), printer.IsShared); err != nil {
				return err
			}

		}
	}

	// NICs info
	if len(nicInfo.Edges.Networkadapters) > 0 {

		sheetName := i18n.T(c.Request().Context(), "Network Adapters")
		_, err = f.NewSheet(sheetName)
		if err != nil {
			return err
		}

		if err := f.SetColWidth(sheetName, "B", "D", 50); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "B2", hwInfo.Nickname); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheetName, "B2", "C2", endpointNameStyle); err != nil {
			return err
		}

		if hwInfo.Nickname != hwInfo.Hostname {
			if err := f.SetCellValue(sheetName, "C2", hwInfo.Hostname); err != nil {
				return err
			}
		}

		for index, nic := range nicInfo.Edges.Networkadapters {
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 4+(index*8)), fmt.Sprintf("B%d", 4+(index*8)), tagStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 4+(index*8)), nic.Name); err != nil {
				return err
			}

			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 5+(index*8)), fmt.Sprintf("D%d", 5+(index*8)), headerStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 5+(index*8)), i18n.T(c.Request().Context(), "inventory.network_adapters.ip_address")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 5+(index*8)), i18n.T(c.Request().Context(), "inventory.network_adapters.mac_address")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 5+(index*8)), i18n.T(c.Request().Context(), "inventory.network_adapters.default_gateway")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 6+(index*8)), nic.Addresses); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 6+(index*8)), nic.MACAddress); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 6+(index*8)), nic.DefaultGateway); err != nil {
				return err
			}

			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 7+(index*8)), fmt.Sprintf("D%d", 7+(index*8)), headerStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 7+(index*8)), i18n.T(c.Request().Context(), "inventory.network_adapters.subnet")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 7+(index*8)), i18n.T(c.Request().Context(), "inventory.network_adapters.dhcp")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 7+(index*8)), i18n.T(c.Request().Context(), "inventory.network_adapters.speed")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 8+(index*8)), nic.Subnet); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 8+(index*8)), nic.DhcpEnabled); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 8+(index*8)), nic.Speed); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 9+(index*8)), fmt.Sprintf("B%d", 9+(index*8)), headerStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 9+(index*8)), i18n.T(c.Request().Context(), "inventory.network_adapters.dns")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 10+(index*8)), nic.DNSServers); err != nil {
				return err
			}
		}
	}

	// Software info
	if len(swInfo) > 0 {

		sheetName := i18n.T(c.Request().Context(), "Software")
		_, err = f.NewSheet(sheetName)
		if err != nil {
			return err
		}

		if err := f.SetColWidth(sheetName, "B", "D", 50); err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, "B2", hwInfo.Nickname); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheetName, "B2", "C2", endpointNameStyle); err != nil {
			return err
		}
		if hwInfo.Nickname != hwInfo.Hostname {
			if err := f.SetCellValue(sheetName, "C2", hwInfo.Hostname); err != nil {
				return err
			}
		}

		for index, app := range swInfo {
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 4+(index*4)), fmt.Sprintf("B%d", 4+(index*4)), tagStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 4+(index*4)), app.Name); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", 5+(index*4)), fmt.Sprintf("D%d", 5+(index*4)), headerStyle); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 5+(index*4)), i18n.T(c.Request().Context(), "Version")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 5+(index*4)), i18n.T(c.Request().Context(), "apps.publisher")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 5+(index*4)), i18n.T(c.Request().Context(), "inventory.apps.installation_date")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", 6+(index*4)), app.Version); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", 6+(index*4)), app.Publisher); err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, fmt.Sprintf("D%d", 6+(index*4)), app.InstallDate); err != nil {
				return err
			}
		}
	}

	// Save spreadsheet by the given path.
	if err := f.SaveAs(dstPath); err != nil {
		fmt.Println(err)
	}

	// Redirect to file
	url := "/download/" + fileName
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}
