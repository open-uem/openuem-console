package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/views/computers_views"
	"github.com/open-uem/openuem-console/internal/views/partials"
)

func (h *Handler) BitLockerInfo(c echo.Context, successMessage string) error {
	var err error

	commonInfo, err := h.GetCommonInfo(c)
	if err != nil {
		return err
	}

	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	agent, err := h.Model.GetAgentById(agentID, commonInfo)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_get_agent"), false))
	}

	ld, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	p := partials.PaginationAndSort{}

	tenantID, err := strconv.Atoi(commonInfo.TenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "tenants.could_not_convert_to_int", err.Error()), true))
	}
	settings, err := h.Model.GetNetbirdSettings(tenantID)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "netbird.could_not_get_settings", err.Error()), true))
	}

	netbird := settings.AccessToken != ""

	offline := h.IsAgentOffline(c)

	// refreshTime, err := h.Model.GetDefaultRefreshTime()
	// if err != nil {
	// 	log.Println("[ERROR]: could not get refresh time from database")
	// 	refreshTime = 5
	// }

	refreshTime := 30

	return RenderView(c, computers_views.InventoryIndex(" | BitLocker", computers_views.BitLockerInfo(c, p, agent, ld, successMessage, commonInfo, refreshTime, netbird, offline), commonInfo))
}

func (h *Handler) BitLockerProgressInfo(c echo.Context, successMessage string) error {
	// if offline don't check for progress
	offline := h.IsAgentOffline(c)
	if offline {
		return h.BitLockerInfo(c, successMessage)
	}

	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	blOperation := nats.BitLockerOp{
		Operation: nats.BitLockerInfoAction,
		Volume:    volume,
	}

	data, err := json.Marshal(blOperation)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_marshall_bitlocker_operation"), false))
	}

	msg, err := h.NATSConnection.Request(fmt.Sprintf("agent.bitlocker.%s", agentID), data, 1*time.Minute)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_bitlocker_operation"), false))
	}

	// save agent data
	blOperationResponse := nats.BitLockerOp{}
	if err := json.Unmarshal(msg.Data, &blOperationResponse); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_unmarshall_bitlocker_operation"), false))
	}

	if err := h.Model.SaveBitLockerInfo(agentID, volume, blOperationResponse); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_info"), false))
	}

	// refresh the info
	return h.BitLockerInfo(c, successMessage)
}

func (h *Handler) BitLockerEncryptDisk(c echo.Context) error {
	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	ld, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	if ld.BitlockerOperationInProgress != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_previous_operation"), false))
	}

	blOperation := nats.BitLockerOp{
		Operation: nats.BitLockerEncryptAction,
		Volume:    volume,
	}

	passphrase := c.FormValue("bitlocker-passphrase")

	// Set passphrase for fixed disks
	if ld.VolumeType == computers_views.VolumeTypeFixedDisk {
		blOperation.Passphrase = passphrase

		if err := h.Model.SaveBitLockerPassphrase(agentID, volume, passphrase); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_could_not_save_passphrase", err), false))
		}
	} else {
		if len(passphrase) < 6 || len(passphrase) > 20 {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_tpm_pin_length"), false))
		}
		blOperation.Passphrase = passphrase

		if err := h.Model.SaveBitLockerPassphrase(agentID, volume, passphrase); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_could_not_save_passphrase", err), false))
		}
	}

	data, err := json.Marshal(blOperation)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_marshall_bitlocker_operation"), false))
	}

	if _, err := h.NATSConnection.Request(fmt.Sprintf("agent.bitlocker.%s", agentID), data, 1*time.Minute); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_bitlocker_operation"), false))
	}

	// Save operation in progress
	if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerEncryptAction); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
	}

	// refresh the info
	return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_encryption_request", volume))
}

func (h *Handler) BitLockerDecryptDisk(c echo.Context) error {
	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	ld, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	if ld.BitlockerOperationInProgress != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_previous_operation"), false))
	}

	blOperation := nats.BitLockerOp{
		Operation: nats.BitLockerDecryptAction,
		Volume:    volume,
	}

	data, err := json.Marshal(blOperation)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_marshall_bitlocker_operation"), false))
	}

	if _, err := h.NATSConnection.Request(fmt.Sprintf("agent.bitlocker.%s", agentID), data, 1*time.Minute); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_bitlocker_operation"), false))
	}

	// Save operation in progress
	if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerDecryptAction); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
	}

	// refresh the info
	return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_decryption_request", volume))
}

func (h *Handler) BitLockerResumeSuspendProtection(c echo.Context, resume bool) error {
	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	ld, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	if ld.BitlockerOperationInProgress != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_previous_operation"), false))
	}

	blOperation := nats.BitLockerOp{
		Volume: volume,
	}

	if resume {
		blOperation.Operation = nats.BitLockerResumeAction
	} else {
		blOperation.Operation = nats.BitLockerSuspendAction
	}

	if ld.VolumeType != computers_views.VolumeTypeSystem {
		if resume {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_resume_request_invalid_device"), false))
		} else {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_suspend_request_invalid_device"), false))
		}
	}

	data, err := json.Marshal(blOperation)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_marshall_bitlocker_operation"), false))
	}

	if _, err := h.NATSConnection.Request(fmt.Sprintf("agent.bitlocker.%s", agentID), data, 1*time.Minute); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_bitlocker_operation"), false))
	}

	// Save operation in progress
	if resume {
		if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerResumeAction); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
		}
	} else {
		if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerSuspendAction); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
		}
	}

	if resume {
		return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_resume_request", volume))
	} else {
		return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_suspend_request", volume))
	}
}

func (h *Handler) BitLockerAutoUnlock(c echo.Context, enable bool) error {
	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	ld, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	if ld.BitlockerOperationInProgress != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_previous_operation"), false))
	}

	blOperation := nats.BitLockerOp{
		Volume: volume,
	}

	if enable {
		blOperation.Operation = nats.BitLockerEnableAutoUnlockAction
	} else {
		blOperation.Operation = nats.BitLockerDisableAutoUnlockAction
		blOperation.ExternalKeyVolumeKeyProtectorID = ld.BitlockerExternalKeyVolumeKeyProtectorID
	}

	if ld.VolumeType == computers_views.VolumeTypeSystem {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_autounlock_invalid_device"), false))
	}

	data, err := json.Marshal(blOperation)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_marshall_bitlocker_operation"), false))
	}

	if _, err := h.NATSConnection.Request(fmt.Sprintf("agent.bitlocker.%s", agentID), data, 1*time.Minute); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_bitlocker_operation"), false))
	}

	// Save operation in progress
	if enable {
		if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerEnableAutoUnlockAction); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
		}
	} else {
		if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerDisableAutoUnlockAction); err != nil {
			return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
		}
	}

	if enable {
		return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_autounlock_enable_request", volume))
	} else {
		return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_autounlock_disable_request", volume))
	}
}

func (h *Handler) BitLockerChangePassphrase(c echo.Context) error {
	var err error
	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	passphrase := c.FormValue("passphrase")
	if passphrase == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_change_passphrase_required"), false))
	}

	ld, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	if ld.BitlockerOperationInProgress != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_previous_operation"), false))
	}

	blOperation := nats.BitLockerOp{
		Volume:                         volume,
		Operation:                      nats.BitLockerChangePassphraseAction,
		Passphrase:                     passphrase,
		PassphraseVolumeKeyProtectorID: ld.BitlockerPassphraseVolumeKeyProtectorID,
	}

	if ld.VolumeType == computers_views.VolumeTypeSystem {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_change_passphrase_invalid_device"), false))
	}

	data, err := json.Marshal(blOperation)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_marshall_bitlocker_operation"), false))
	}

	if _, err := h.NATSConnection.Request(fmt.Sprintf("agent.bitlocker.%s", agentID), data, 1*time.Minute); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_bitlocker_operation"), false))
	}

	// Save operation in progress
	if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerChangePassphraseAction); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
	}

	return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_change_passphrase_request", volume))
}

func (h *Handler) BitLockerDeleteStalledOp(c echo.Context) error {
	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	_, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	// Save operation in progress
	if err := h.Model.DeleteBitLockerStalledOp(agentID, volume); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_delete_stalled_operation", err), false))
	}

	// refresh the info
	return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_decryption_request", volume))
}

func (h *Handler) BitLockerUnlockWithPassphrase(c echo.Context) error {
	var err error
	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	ld, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	blOperation := nats.BitLockerOp{
		Volume:     volume,
		Operation:  nats.BitLockerUnlockWithPassphraseAction,
		Passphrase: ld.BitlockerPassphrase,
	}

	data, err := json.Marshal(blOperation)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_marshall_bitlocker_operation"), false))
	}

	if _, err := h.NATSConnection.Request(fmt.Sprintf("agent.bitlocker.%s", agentID), data, 1*time.Minute); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_bitlocker_operation"), false))
	}

	// Save operation in progress
	if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerUnlockWithPassphraseAction); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
	}

	return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_unlock_with_passphrase_request", volume))
}

func (h *Handler) BitLockerAddPassphrase(c echo.Context) error {
	var err error
	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	passphrase := c.FormValue("passphrase")
	if passphrase == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_change_passphrase_required"), false))
	}

	ld, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	if ld.BitlockerPassphrase != "" || ld.BitlockerPassphraseVolumeKeyProtectorID != "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_add_passphrase_already_exists"), false))
	}

	blOperation := nats.BitLockerOp{
		Volume:     volume,
		Operation:  nats.BitLockerAddPassphraseAction,
		Passphrase: passphrase,
	}

	data, err := json.Marshal(blOperation)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_marshall_bitlocker_operation"), false))
	}

	if _, err := h.NATSConnection.Request(fmt.Sprintf("agent.bitlocker.%s", agentID), data, 1*time.Minute); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_bitlocker_operation"), false))
	}

	// Save operation in progress
	if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerAddPassphraseAction); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
	}

	return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_add_passphrase_request", volume))
}

func (h *Handler) BitLockerDeletePassphrase(c echo.Context) error {
	var err error
	agentID := c.Param("uuid")
	if agentID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_id"), false))
	}

	volume := c.Param("volume")
	if volume == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.no_empty_logical_disk_volume_name"), false))
	}

	ld, err := h.Model.GetLogicalDiskByLabel(agentID, volume)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.logical_disk_not_found"), false))
	}

	if ld.BitlockerPassphrase == "" || ld.BitlockerPassphraseVolumeKeyProtectorID == "" {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.bitlocker_delete_passphrase_invalid_state"), false))
	}

	blOperation := nats.BitLockerOp{
		Volume:                         volume,
		Operation:                      nats.BitLockerDeletePassphraseAction,
		PassphraseVolumeKeyProtectorID: ld.BitlockerPassphraseVolumeKeyProtectorID,
	}

	data, err := json.Marshal(blOperation)
	if err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_marshall_bitlocker_operation"), false))
	}

	if _, err := h.NATSConnection.Request(fmt.Sprintf("agent.bitlocker.%s", agentID), data, 1*time.Minute); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.could_not_send_bitlocker_operation"), false))
	}

	// Save operation in progress
	if err := h.Model.SaveBitLockerOperationInProgress(agentID, volume, nats.BitLockerDeletePassphraseAction); err != nil {
		return RenderError(c, partials.ErrorMessage(i18n.T(c.Request().Context(), "agents.not_save_bitlocker_operation_in_progress", err), false))
	}

	return h.BitLockerInfo(c, i18n.T(c.Request().Context(), "agents.bitlocker_delete_passphrase_request", volume))
}
