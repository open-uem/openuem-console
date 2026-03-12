package models

import (
	"context"
	"fmt"

	ent "github.com/open-uem/ent"
	"github.com/open-uem/ent/tenant"
	"github.com/open-uem/ent/user"
	"github.com/open-uem/ent/usertenant"
)

// UserTenantRole represents the role a user has within a tenant
type UserTenantRole string

const (
	UserTenantRoleAdmin    UserTenantRole = "admin"    // Can manage everything including users
	UserTenantRoleOperator UserTenantRole = "operator" // Can manage settings but NOT users
	UserTenantRoleUser     UserTenantRole = "user"     // Read-only access
)

// AssignUserToTenant assigns a user to a tenant with the specified role
func (m *Model) AssignUserToTenant(userID string, tenantID int, role UserTenantRole, isDefault bool) error {
	// Check if assignment already exists
	exists, err := m.Client.UserTenant.Query().
		Where(
			usertenant.UserID(userID),
			usertenant.TenantID(tenantID),
		).Exist(context.Background())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("user %s is already assigned to tenant %d", userID, tenantID)
	}

	// If this should be the default, remove default from other assignments
	if isDefault {
		err = m.Client.UserTenant.Update().
			Where(usertenant.UserID(userID)).
			SetIsDefault(false).
			Exec(context.Background())
		if err != nil {
			return err
		}
	}

	return m.Client.UserTenant.Create().
		SetUserID(userID).
		SetTenantID(tenantID).
		SetRole(usertenant.Role(role)).
		SetIsDefault(isDefault).
		Exec(context.Background())
}

// RemoveUserFromTenant removes a user from a tenant
func (m *Model) RemoveUserFromTenant(userID string, tenantID int) error {
	_, err := m.Client.UserTenant.Delete().
		Where(
			usertenant.UserID(userID),
			usertenant.TenantID(tenantID),
		).Exec(context.Background())
	return err
}

// UpdateUserTenantRole updates the role of a user within a tenant
func (m *Model) UpdateUserTenantRole(userID string, tenantID int, role UserTenantRole) error {
	return m.Client.UserTenant.Update().
		Where(
			usertenant.UserID(userID),
			usertenant.TenantID(tenantID),
		).
		SetRole(usertenant.Role(role)).
		Exec(context.Background())
}

// SetUserDefaultTenant sets the default tenant for a user
func (m *Model) SetUserDefaultTenant(userID string, tenantID int) error {
	// First, remove default from all user's tenant assignments
	err := m.Client.UserTenant.Update().
		Where(usertenant.UserID(userID)).
		SetIsDefault(false).
		Exec(context.Background())
	if err != nil {
		return err
	}

	// Set the new default
	return m.Client.UserTenant.Update().
		Where(
			usertenant.UserID(userID),
			usertenant.TenantID(tenantID),
		).
		SetIsDefault(true).
		Exec(context.Background())
}

// GetUserTenants returns all tenants a user has access to
func (m *Model) GetUserTenants(userID string) ([]*ent.Tenant, error) {
	userTenants, err := m.Client.UserTenant.Query().
		Where(usertenant.UserID(userID)).
		WithTenant().
		All(context.Background())
	if err != nil {
		return nil, err
	}

	tenants := make([]*ent.Tenant, 0, len(userTenants))
	for _, ut := range userTenants {
		if ut.Edges.Tenant != nil {
			tenants = append(tenants, ut.Edges.Tenant)
		}
	}
	return tenants, nil
}

// GetUserTenantsWithRoles returns all tenant assignments for a user including roles
func (m *Model) GetUserTenantsWithRoles(userID string) ([]*ent.UserTenant, error) {
	return m.Client.UserTenant.Query().
		Where(usertenant.UserID(userID)).
		WithTenant().
		All(context.Background())
}

// GetUserDefaultTenant returns the default tenant for a user
func (m *Model) GetUserDefaultTenant(userID string) (*ent.Tenant, error) {
	ut, err := m.Client.UserTenant.Query().
		Where(
			usertenant.UserID(userID),
			usertenant.IsDefault(true),
		).
		WithTenant().
		Only(context.Background())
	if err != nil {
		// If no default is set, return the first tenant
		ut, err = m.Client.UserTenant.Query().
			Where(usertenant.UserID(userID)).
			WithTenant().
			First(context.Background())
		if err != nil {
			return nil, err
		}
	}
	return ut.Edges.Tenant, nil
}

// UserHasAccessToTenant checks if a user has access to a specific tenant
func (m *Model) UserHasAccessToTenant(userID string, tenantID int) (bool, error) {
	return m.Client.UserTenant.Query().
		Where(
			usertenant.UserID(userID),
			usertenant.TenantID(tenantID),
		).Exist(context.Background())
}

// GetUserRoleInTenant returns the role of a user in a specific tenant
func (m *Model) GetUserRoleInTenant(userID string, tenantID int) (UserTenantRole, error) {
	ut, err := m.Client.UserTenant.Query().
		Where(
			usertenant.UserID(userID),
			usertenant.TenantID(tenantID),
		).Only(context.Background())
	if err != nil {
		return "", err
	}
	return UserTenantRole(ut.Role), nil
}

// IsUserTenantAdmin checks if a user is an admin in a specific tenant
func (m *Model) IsUserTenantAdmin(userID string, tenantID int) (bool, error) {
	role, err := m.GetUserRoleInTenant(userID, tenantID)
	if err != nil {
		return false, err
	}
	return role == UserTenantRoleAdmin, nil
}

// GetTenantUsers returns all users assigned to a tenant
func (m *Model) GetTenantUsers(tenantID int) ([]*ent.User, error) {
	userTenants, err := m.Client.UserTenant.Query().
		Where(usertenant.TenantID(tenantID)).
		WithUser().
		All(context.Background())
	if err != nil {
		return nil, err
	}

	users := make([]*ent.User, 0, len(userTenants))
	for _, ut := range userTenants {
		if ut.Edges.User != nil {
			users = append(users, ut.Edges.User)
		}
	}
	return users, nil
}

// GetTenantUsersWithRoles returns all user assignments for a tenant including roles
func (m *Model) GetTenantUsersWithRoles(tenantID int) ([]*ent.UserTenant, error) {
	return m.Client.UserTenant.Query().
		Where(usertenant.TenantID(tenantID)).
		WithUser().
		All(context.Background())
}

// GetMainTenant returns the main tenant (the one with the lowest ID)
func (m *Model) GetMainTenant() (*ent.Tenant, error) {
	return m.Client.Tenant.Query().
		Order(ent.Asc(tenant.FieldID)).
		First(context.Background())
}

// GetTenantsWhereUserIsAdmin returns all tenants where the user has admin role
func (m *Model) GetTenantsWhereUserIsAdmin(userID string) ([]*ent.Tenant, error) {
	userTenants, err := m.Client.UserTenant.Query().
		Where(
			usertenant.UserID(userID),
			usertenant.RoleEQ(usertenant.RoleAdmin),
		).
		WithTenant().
		All(context.Background())
	if err != nil {
		return nil, err
	}

	tenants := make([]*ent.Tenant, 0, len(userTenants))
	for _, ut := range userTenants {
		if ut.Edges.Tenant != nil {
			tenants = append(tenants, ut.Edges.Tenant)
		}
	}
	return tenants, nil
}

// IsMainTenant checks if a tenant is the main tenant (lowest ID)
func (m *Model) IsMainTenant(tenantID int) (bool, error) {
	mainTenant, err := m.GetMainTenant()
	if err != nil {
		return false, err
	}
	return mainTenant.ID == tenantID, nil
}

// IsMainTenantAdmin checks if a user is an admin in the main tenant
func (m *Model) IsMainTenantAdmin(userID string) (bool, error) {
	mainTenant, err := m.GetMainTenant()
	if err != nil {
		return false, err
	}

	role, err := m.GetUserRoleInTenant(userID, mainTenant.ID)
	if err != nil {
		return false, nil // User not assigned to main tenant
	}
	return role == UserTenantRoleAdmin, nil
}

// GetTenantsForUser returns all tenants the user is explicitly assigned to
func (m *Model) GetTenantsForUser(userID string) ([]*ent.Tenant, error) {
	return m.GetUserTenants(userID)
}

// GetUsersNotInTenant returns all users that are NOT assigned to the given tenant
func (m *Model) GetUsersNotInTenant(tenantID int) ([]*ent.User, error) {
	// Get user IDs already in this tenant
	existingUTs, err := m.Client.UserTenant.Query().
		Where(usertenant.TenantID(tenantID)).
		All(context.Background())
	if err != nil {
		return nil, err
	}

	existingUserIDs := make([]string, 0, len(existingUTs))
	for _, ut := range existingUTs {
		existingUserIDs = append(existingUserIDs, ut.UserID)
	}

	// Query all users NOT in that list
	query := m.Client.User.Query()
	if len(existingUserIDs) > 0 {
		query.Where(user.IDNotIn(existingUserIDs...))
	}
	return query.All(context.Background())
}
