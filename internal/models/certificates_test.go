package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/open-uem/openuem-console/internal/views/filters"
	"github.com/open-uem/openuem-console/internal/views/partials"
	"github.com/open-uem/openuem_ent"
	"github.com/open-uem/openuem_ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ocsp"
)

type CertificatesTestSuite struct {
	suite.Suite
	t     enttest.TestingT
	model Model
	p     partials.PaginationAndSort
}

func (suite *CertificatesTestSuite) SetupTest() {
	client := enttest.Open(suite.t, "sqlite3", "file:ent?mode=memory&_fk=1")
	suite.model = Model{Client: client}

	for i := 0; i <= 6; i++ {
		err := client.Certificate.Create().
			SetID(int64(i)).
			SetType("console").
			SetDescription(fmt.Sprintf("description%d", i)).
			SetExpiry(time.Now()).
			SetUID(fmt.Sprintf("user%d", i)).
			Exec(context.Background())
		assert.NoError(suite.T(), err)
	}
}

func (suite *CertificatesTestSuite) TestGetCertificateByUID() {
	certificate, err := suite.model.GetCertificateByUID("user1")
	assert.NoError(suite.T(), err, "should get certificate by uid")
	assert.Equal(suite.T(), "description1", certificate.Description, "should get certificate for user1")

	certificate, err = suite.model.GetCertificateByUID("user7")
	assert.Error(suite.T(), err, "should not get certificate by uid")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *CertificatesTestSuite) TestGetCertificateBySerial() {
	certificate, err := suite.model.GetCertificateBySerial("5")
	assert.NoError(suite.T(), err, "should get certificate by serial")
	assert.Equal(suite.T(), "description5", certificate.Description, "should get certificate for serial 5")

	certificate, err = suite.model.GetCertificateBySerial("7")
	assert.Error(suite.T(), err, "should not get certificate by serial")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *CertificatesTestSuite) TestRevokeCertificate() {
	err := suite.model.RevokeCertificate(&openuem_ent.Certificate{
		ID: int64(1),
	}, "test", ocsp.CessationOfOperation)
	assert.NoError(suite.T(), err, "should revoke certificate")
}

func (suite *CertificatesTestSuite) TestDeleteCertificate() {
	err := suite.model.DeleteCertificate(int64(5))
	assert.NoError(suite.T(), err, "should delete certificate by serial")

	count, err := suite.model.CountAllCertificates(filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should count all certificates")
	assert.Equal(suite.T(), 6, count, "should count 6 certificates")

	err = suite.model.DeleteCertificate(int64(8))
	assert.Error(suite.T(), err, "should not delete non existent certificate")
	assert.Equal(suite.T(), true, openuem_ent.IsNotFound(err), "should raise not found error")
}

func (suite *CertificatesTestSuite) TestCountAllCertificates() {
	count, err := suite.model.CountAllCertificates(filters.CertificateFilter{TypeOptions: []string{"console"}})
	assert.NoError(suite.T(), err, "should count all certificates")
	assert.Equal(suite.T(), 7, count, "should count 7 certificates")

	count, err = suite.model.CountAllCertificates(filters.CertificateFilter{TypeOptions: []string{"nats"}})
	assert.NoError(suite.T(), err, "should count all certificates")
	assert.Equal(suite.T(), 0, count, "should count 0 certificates")

	count, err = suite.model.CountAllCertificates(filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should count all certificates")
	assert.Equal(suite.T(), 7, count, "should count 7 certificates")

	count, err = suite.model.CountAllCertificates(filters.CertificateFilter{Description: "description5"})
	assert.NoError(suite.T(), err, "should count all certificates")
	assert.Equal(suite.T(), 1, count, "should count 1 certificate")

	count, err = suite.model.CountAllCertificates(filters.CertificateFilter{ExpiryFrom: "2024-01-01", ExpiryTo: "2034-01-01"})
	assert.NoError(suite.T(), err, "should count all certificates")
	assert.Equal(suite.T(), 7, count, "should count 7 certificate")

	count, err = suite.model.CountAllCertificates(filters.CertificateFilter{Username: "user5"})
	assert.NoError(suite.T(), err, "should count all certificates")
	assert.Equal(suite.T(), 1, count, "should count 1 certificate")
}

func (suite *CertificatesTestSuite) TestCountCertificatesAboutToexpire() {
	count, err := suite.model.CountCertificatesAboutToexpire()
	assert.NoError(suite.T(), err, "should count certificates about to expire")
	assert.Equal(suite.T(), 7, count, "should count 7 certificates about to expire")
}

func (suite *CertificatesTestSuite) TestGetCertificatesByPage() {
	items, err := suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(6-i), item.ID)
	}

	suite.p.SortBy = "serial"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(i), item.ID)
	}

	suite.p.SortBy = "serial"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(6-i), item.ID)
	}

	suite.p.SortBy = "type"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(i), item.ID)
	}

	suite.p.SortBy = "type"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(i), item.ID)
	}

	suite.p.SortBy = "description"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(i), item.ID)
	}

	suite.p.SortBy = "description"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(6-i), item.ID)
	}

	suite.p.SortBy = "expiry"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(i), item.ID)
	}

	suite.p.SortBy = "expiry"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(6-i), item.ID)
	}

	suite.p.SortBy = "username"
	suite.p.SortOrder = "asc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(i), item.ID)
	}

	suite.p.SortBy = "username"
	suite.p.SortOrder = "desc"
	items, err = suite.model.GetCertificatesByPage(suite.p, filters.CertificateFilter{})
	assert.NoError(suite.T(), err, "should get certificates by page")
	for i, item := range items {
		assert.Equal(suite.T(), int64(6-i), item.ID)
	}
}

func (suite *CertificatesTestSuite) TestGetCertificatesTypes() {
	certificatesFound, err := suite.model.GetCertificatesTypes()
	assert.NoError(suite.T(), err, "should get certificates types")
	assert.Equal(suite.T(), []string{"console"}, certificatesFound, "should get console type")
}

func TestCertificatesTestSuite(t *testing.T) {
	suite.Run(t, new(CertificatesTestSuite))
}
