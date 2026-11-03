package handlers

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/openuem-console/internal/taskfiles"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

const (
	taskFilesCleanupInterval = time.Hour
	// Uploaded files are kept for a day even if no task uses them yet, as the
	// task they were uploaded for may not have been saved yet
	taskFilesCleanupGrace = 24 * time.Hour
)

// UploadMSIFile stores an MSI package uploaded for a task of a profile. It answers with
// the MSI path set to the URL that the agents of the profile download it from.
func (h *Handler) UploadMSIFile(c echo.Context) error {
	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	if h.TaskFiles == nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.msi_upload_not_available"), false))
	}

	profileID, err := strconv.Atoi(c.FormValue("task-profile-id"))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.msi_upload_no_profile"), false))
	}
	if _, err := h.Model.GetProfileById(profileID, commonInfo); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	header, err := c.FormFile("msi-file")
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.msi_upload_no_file"), false))
	}
	if !strings.EqualFold(filepath.Ext(header.Filename), ".msi") {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.msi_upload_not_msi"), false))
	}

	src, err := header.Open()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer src.Close()

	file, err := h.TaskFiles.Save(profileID, header.Filename, src)
	if err != nil {
		log.Printf("[ERROR]: could not save the uploaded file %s, reason: %v", header.Filename, err)
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tasks.msi_upload_failed"), false))
	}

	return RenderView(c, partials.MSIPathInput(h.agentFilesBaseURL()+file.URLPath(), i18n.T(c.Request().Context(), "tasks.msi_uploaded", file.Name)))
}

// agentFilesBaseURL is the address of the auth server, which serves the uploaded files
func (h *Handler) agentFilesBaseURL() string {
	if h.ReverseProxyServer != "" && h.ReverseProxyAuthPort != "" {
		return fmt.Sprintf("https://%s:%s", h.ReverseProxyServer, h.ReverseProxyAuthPort)
	}
	return fmt.Sprintf("https://%s:%s", h.ServerName, h.AuthPort)
}

// StartTaskFilesCleanupJob removes, every hour, the uploaded files that no task uses anymore
func (h *Handler) StartTaskFilesCleanupJob() error {
	if h.TaskFiles == nil {
		return nil
	}

	_, err := h.TaskScheduler.NewJob(
		gocron.DurationJob(taskFilesCleanupInterval),
		gocron.NewTask(h.removeUnusedTaskFiles),
	)
	return err
}

func (h *Handler) removeUnusedTaskFiles() {
	paths, err := h.Model.GetTaskFilePaths()
	if err != nil {
		log.Printf("[ERROR]: could not get the task paths to remove unused files, reason: %v", err)
		return
	}

	used := map[taskfiles.File]bool{}
	for _, p := range paths {
		if f, ok := taskfiles.ParseURL(p); ok {
			used[f] = true
		}
	}

	if err := h.TaskFiles.RemoveUnused(used, taskFilesCleanupGrace); err != nil {
		log.Printf("[ERROR]: could not remove unused task files, reason: %v", err)
	}
}
