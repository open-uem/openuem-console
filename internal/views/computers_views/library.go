package computers_views

import "github.com/open-uem/openuem-console/internal/models"

// DeploymentPackageSource returns the software library as the source of a deployment
// when its package is one of the library's packages. Deployments don't record the
// source, so the handler passes the library's package ids along with them
func DeploymentPackageSource(packageID string, libraryPackages map[string]bool) string {
	if libraryPackages[packageID] {
		return models.LibraryPackageSource
	}
	return ""
}
