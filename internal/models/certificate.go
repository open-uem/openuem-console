package models

import (
	"context"
	"time"

	"github.com/doncicuto/openuem_ent"
	"github.com/doncicuto/openuem_ent/certificate"
)

func (m *Model) GetCertificateByUID(uid string) (*openuem_ent.Certificate, error) {
	return m.Client.Certificate.Query().Where(certificate.UID(uid)).Only(context.Background())
}

func (m *Model) RevokeCertificate(cert *openuem_ent.Certificate, info string, reason int) error {
	return m.Client.Revocation.Create().SetID(cert.ID).SetExpiry(cert.Expiry).SetRevoked(time.Now()).SetReason(reason).SetInfo(info).Exec(context.Background())
}

func (m *Model) DeleteCertificate(serial int64) error {
	return m.Client.Certificate.DeleteOneID(serial).Exec(context.Background())
}
