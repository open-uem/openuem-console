package handlers

import (
	"archive/zip"
	"crypto/rsa"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	model "github.com/open-uem/openuem-console/internal/models/servers"
	"github.com/open-uem/openuem-console/internal/views/computers_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/utils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type CheckedItemsForm struct {
	Cwd         string   `form:"cwd" query:"cwd"`
	Dst         string   `form:"dst" query:"dst"`
	FolderCheck []string `form:"check-folder" query:"check-folder"`
	FileCheck   []string `form:"check-file" query:"check-file"`
	Parent      string   `form:"parent" query:"parent"`
}

func (h *Handler) BrowseLogicalDisk(c echo.Context) error {

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	// Get form values
	action := c.FormValue("action")
	cwd := c.FormValue("cwd")
	parent := c.FormValue("parent")
	dst := c.FormValue("dst")

	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return err
	}

	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer client.Close()
	defer sshConn.Close()

	if cwd == "" {
		cwd = `C:\`
	}

	if parent == "" {
		parent = filepath.Dir(cwd)
	}

	if action == "down" {
		if dst != "" {
			parent = cwd
			cwd = filepath.Join(cwd, dst)
		}
	}

	if action == "up" {
		cwd = parent
		parent = filepath.Dir(cwd)
	}

	files, err := client.ReadDir(cwd)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	sortFiles(files)
	p := partials.PaginationAndSort{}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	return RenderView(c, computers_views.InventoryIndex(" | File Browser", computers_views.SFTPHome(p, h.SessionManager, h.Version, latestServerRelease.Version, agent, cwd, parent, files, detectRemoteAgents)))
}

func (h *Handler) NewFolder(c echo.Context) error {
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return err
	}

	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer client.Close()
	defer sshConn.Close()

	// Get form values
	cwd := c.FormValue("cwd")
	itemName := c.FormValue("itemName")

	if cwd == "" {
		return fmt.Errorf("current working directory cannot be empty")
	}

	if itemName == "" {
		return fmt.Errorf("folder name cannot be empty")
	}

	path := filepath.Join(cwd, itemName)
	if err := client.Mkdir(path); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	return h.BrowseLogicalDisk(c)
}

func (h *Handler) DeleteItem(c echo.Context) error {

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return err
	}
	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer client.Close()
	defer sshConn.Close()

	// Get form values
	cwd := c.FormValue("cwd")
	itemName := c.FormValue("itemName")
	parent := c.FormValue("parent")

	if cwd == "" {
		return fmt.Errorf("current working directory cannot be empty")
	}

	if itemName == "" {
		return fmt.Errorf("file/folder name cannot be empty")
	}

	path := filepath.Join(cwd, itemName)
	if err := client.RemoveAll(path); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	files, err := client.ReadDir(cwd)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	sortFiles(files)
	p := partials.PaginationAndSort{}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	return RenderView(c, computers_views.InventoryIndex(" | File Browser", computers_views.SFTPHome(p, h.SessionManager, h.Version, latestServerRelease.Version, agent, cwd, parent, files, detectRemoteAgents)))
}

func (h *Handler) RenameItem(c echo.Context) error {

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return err
	}

	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer client.Close()
	defer sshConn.Close()

	// Get form values
	cwd := c.FormValue("cwd")
	currentName := c.FormValue("currentName")
	newName := c.FormValue("newName")
	parent := c.FormValue("parent")

	if cwd == "" {
		return RenderError(c, partials.ErrorMessage("current working directory cannot be empty", false))
	}

	if currentName == "" {
		return RenderError(c, partials.ErrorMessage("current name cannot be empty", false))
	}

	if newName == "" {
		return RenderError(c, partials.ErrorMessage("current name cannot be empty", false))
	}

	currentPath := filepath.Join(cwd, currentName)
	newPath := filepath.Join(cwd, newName)

	if err := client.Rename(currentPath, newPath); err != nil {
		return RenderError(c, partials.ErrorMessage("current name cannot be empty", false))
	}

	files, err := client.ReadDir(cwd)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	sortFiles(files)

	p := partials.PaginationAndSort{}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	return RenderView(c, computers_views.InventoryIndex(" | File Browser", computers_views.SFTPHome(p, h.SessionManager, h.Version, latestServerRelease.Version, agent, cwd, parent, files, detectRemoteAgents)))
}

func (h *Handler) DeleteMany(c echo.Context) error {
	removeForm := new(CheckedItemsForm)
	if err := c.Bind(removeForm); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	items := slices.Concat(removeForm.FolderCheck, removeForm.FileCheck)

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return err
	}

	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer client.Close()
	defer sshConn.Close()

	cwd := removeForm.Cwd
	if cwd == "" {
		return RenderError(c, partials.ErrorMessage("cwd cannot be empty", false))
	}

	for _, item := range items {
		path := filepath.Join(removeForm.Cwd, item)
		if err := client.RemoveAll(path); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}

	files, err := client.ReadDir(cwd)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	sortFiles(files)
	p := partials.PaginationAndSort{}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	return RenderView(c, computers_views.InventoryIndex(" | File Browser", computers_views.SFTPHome(p, h.SessionManager, h.Version, latestServerRelease.Version, agent, cwd, removeForm.Parent, files, detectRemoteAgents)))
}

func (h *Handler) UploadFile(c echo.Context) error {
	// Get form values
	parent := c.FormValue("parent")

	cwd := c.FormValue("cwd")
	if cwd == "" {
		return RenderError(c, partials.ErrorMessage("cwd cannot be empty", false))
	}

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return err
	}

	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer client.Close()
	defer sshConn.Close()

	path := filepath.Join(cwd, file.Filename)
	dst, err := client.Create(path)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer dst.Close()

	// Copy
	if _, err = dst.ReadFrom(src); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Get stat info
	_, err = client.Stat(path)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	files, err := client.ReadDir(cwd)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	sortFiles(files)

	p := partials.PaginationAndSort{}

	latestServerRelease, err := model.GetLatestServerReleaseFromAPI(h.ServerReleasesFolder)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), true))
	}

	detectRemoteAgents, err := h.Model.GetDefaultDetectRemoteAgents()
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "settings.could_not_get_detect_remote_agents_setting"), true))
	}

	return RenderView(c, computers_views.SFTPHome(p, h.SessionManager, h.Version, latestServerRelease.Version, agent, cwd, parent, files, detectRemoteAgents))
}

func (h *Handler) DownloadFile(c echo.Context) error {

	// Get form values
	cwd := c.FormValue("cwd")
	if cwd == "" {
		return RenderError(c, partials.ErrorMessage("cwd cannot be empty", false))
	}

	file := c.FormValue("itemName")
	if cwd == "" {
		return RenderError(c, partials.ErrorMessage("file name cannot be empty", false))
	}
	remoteFile := filepath.Join(cwd, file)

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return err
	}

	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer client.Close()
	defer sshConn.Close()

	dstPath := filepath.Join(h.DownloadDir, file)
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer dstFile.Close()

	srcFile, err := client.OpenFile(remoteFile, (os.O_RDONLY))
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Redirect to file
	url := "/download/" + filepath.Base(dstFile.Name())
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func (h *Handler) DownloadFolderAsZIP(c echo.Context) error {

	// Get form values
	cwd := c.FormValue("cwd")
	if cwd == "" {
		return RenderError(c, partials.ErrorMessage("cwd cannot be empty", false))
	}

	folder := c.FormValue("itemName")
	if cwd == "" {
		return RenderError(c, partials.ErrorMessage("folder name cannot be empty", false))
	}
	remoteFolder := filepath.Join(cwd, folder)

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return err
	}

	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer client.Close()
	defer sshConn.Close()

	file, err := os.CreateTemp(h.DownloadDir, "openuem")
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	w := zip.NewWriter(file)

	if err := addFiles(client, w, remoteFolder, ""); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	if err := w.Close(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := file.Close(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := os.Rename(file.Name(), file.Name()+".zip"); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Redirect to ZIP file
	url := "/download/" + filepath.Base(file.Name()+".zip")
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func (h *Handler) DownloadManyAsZIP(c echo.Context) error {

	// Get form values
	deleteForm := new(CheckedItemsForm)
	if err := c.Bind(deleteForm); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	if deleteForm.Cwd == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.cwd_cannot_be_empty"), false))
	}

	items := slices.Concat(deleteForm.FolderCheck, deleteForm.FileCheck)
	if len(items) == 0 {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_items_were_checked"), false))
	}

	agentId := c.Param("uuid")
	if agentId == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	agent, err := h.Model.GetAgentById(agentId)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	key, err := utils.ReadPEMPrivateKey(h.SFTPKeyPath)
	if err != nil {
		return err
	}

	client, sshConn, err := connectWithSFTP(agent.IP, key, agent.SftpPort, agent.Os)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}
	defer client.Close()
	defer sshConn.Close()

	file, err := os.CreateTemp(h.DownloadDir, "openuem")
	if err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	w := zip.NewWriter(file)

	for _, item := range items {
		path := filepath.Join(deleteForm.Cwd, item)
		if err := addFiles(client, w, path, ""); err != nil {
			return RenderError(c, partials.ErrorMessage(err.Error(), false))
		}
	}
	if err := w.Close(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := file.Close(); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	if err := os.Rename(file.Name(), file.Name()+".zip"); err != nil {
		return RenderError(c, partials.ErrorMessage(err.Error(), false))
	}

	// Redirect to ZIP file
	url := "/download/" + filepath.Base(file.Name()+".zip")
	c.Response().Header().Set("HX-Redirect", url)

	return c.String(http.StatusOK, "")
}

func (h *Handler) Download(c echo.Context) error {
	fileName := c.Param("filename")
	if fileName == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	path := filepath.Join(h.DownloadDir, fileName)
	return c.Attachment(path, fileName)
}

func addFiles(client *sftp.Client, w *zip.Writer, basePath, baseInZip string) error {
	// Check if is file or directory
	entry, err := client.Open(basePath)
	if err != nil {
		return err
	}
	defer entry.Close()

	fileInfo, err := entry.Stat()
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		// Open the Directory
		files, err := client.ReadDir(basePath)
		if err != nil {
			return err
		}
		baseInZip := filepath.Join(baseInZip, filepath.Base(basePath), "/")

		for _, file := range files {
			if !file.IsDir() {
				filePath := filepath.Join(basePath, file.Name())
				if err := addFiles(client, w, filePath, baseInZip); err != nil {
					return err
				}
			} else {
				filePath := filepath.Join(basePath, file.Name(), "/")
				if err := addFiles(client, w, filePath, baseInZip); err != nil {
					return err
				}
			}
		}
	} else {
		// Add file to the archive.
		zipPath := filepath.Join(baseInZip, filepath.Base(entry.Name()))
		f, err := w.Create(zipPath)
		if err != nil {
			return err
		}

		_, err = entry.WriteTo(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func sortFiles(files []fs.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir() {
			if files[j].IsDir() {
				return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
			}
			return true
		}

		if files[j].IsDir() {
			return false
		}
		return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
	})

}

func connectWithSFTP(IPAddress string, key *rsa.PrivateKey, sftpPort, os string) (*sftp.Client, *ssh.Client, error) {
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, nil, err
	}

	user := ""
	if os == "windows" {
		user = "NT AUTHORITY\\SYSTEM"
	} else {
		user = "root"
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", IPAddress+":"+sftpPort, config)
	if err != nil {
		return nil, nil, err
	}

	sftpConn, err := sftp.NewClient(conn)
	if err != nil {
		return nil, nil, err
	}

	return sftpConn, conn, nil
}
