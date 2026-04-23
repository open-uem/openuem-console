package handlers

import (
	"errors"
	"strings"

	"github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	turnstile "github.com/meyskens/go-turnstile"
	"github.com/open-uem/utils"
)

func (h *Handler) TurnstileCheckChallenge(c echo.Context, cfTurnStileResponse string, tsSecretKey string) error {
	var err error

	if tsSecretKey == "" {
		return errors.New(i18n.T(c.Request().Context(), "settings.turnstile_empty_secret_key"))
	}

	if h.EncryptionMasterKey == "" {
		return errors.New(i18n.T(c.Request().Context(), "settings.turnstile_empty_encryption_master_key"))
	}

	tsSecretKey, err = utils.DecryptSensitiveField(tsSecretKey, h.EncryptionMasterKey)
	if err != nil {
		return errors.New(i18n.T(c.Request().Context(), "settings.turnstile_secret_key_cannot_be_decrypted", err))
	}

	// check if turnstile challenge passed
	ts := turnstile.New(tsSecretKey)
	resp, err := ts.Verify(cfTurnStileResponse, c.RealIP())
	if err != nil {
		return errors.New(i18n.T(c.Request().Context(), "settings.turnstile_could_not_check_challenge_response", err))
	}

	if !resp.Success {
		return errors.New(i18n.T(c.Request().Context(), "settings.turnstile_challenge_did_not_pass", strings.Join(resp.ErrorCodes, ",")))
	}

	return nil
}
