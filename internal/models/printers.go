package models

import (
	"context"

	"github.com/open-uem/ent/printer"
)

func (m *Model) CountDifferentPrinters() (int, error) {
	return m.Client.Printer.Query().Select(printer.FieldName).Unique(true).Count(context.Background())
}
