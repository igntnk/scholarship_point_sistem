package service

import (
	"context"
	"errors"
	"github.com/igntnk/scholarship_point_system/controllers/requests"
	"github.com/igntnk/scholarship_point_system/controllers/responses"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/repository"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"gopkg.in/yaml.v3"
	"os"
)

type adminGroupRole struct {
	GroupName string `yaml:"group_name"`
	RoleName  string `yaml:"role_name"`
}

const (
	CacheAdminGroupFile = "./config/perm.yaml"
)

type PermissionService interface {
	CheckUserHasPermission(ctx context.Context, userUUID string, url string) (ok bool, err error)
	ActualizeResources(ctx context.Context, resources map[string]struct{}) error
	ActualizeAdminGroupAndRole(ctx context.Context, adminUUID, groupName, roleName string) error
	GetResourceList(ctx context.Context) ([]responses.Resource, error)
	CreateRole(ctx context.Context, role requests.Role) (string, error)
	CreateGroup(ctx context.Context, group requests.Group) (string, error)
	UpdateRoleWithMembers(ctx context.Context, role requests.Role) error
	UpdateGroupWithRolesAndResources(ctx context.Context, group requests.Group) error
	DeleteRole(ctx context.Context, uuid string) error
	DeleteGroup(ctx context.Context, uuid string) error
	GetRoleByUUID(ctx context.Context, uuid string) (responses.Role, error)
	GetRoleList(ctx context.Context) ([]responses.SimpleRole, error)
	GetRoleListWithPagination(ctx context.Context, limit, offset int) ([]responses.SimpleRole, int, error)
	GetGroupByUUID(ctx context.Context, uuid string) (responses.Group, error)
	GetGroupList(ctx context.Context) ([]responses.SimpleGroup, error)
	GetGroupListWithPagination(ctx context.Context, limit, offset int) ([]responses.SimpleGroup, int, error)
}

type permissionService struct {
	permissionRepo repository.PermissionRepository
}

func NewPermissionService(permissionRepo repository.PermissionRepository) PermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
	}
}

func (s *permissionService) CheckUserHasPermission(ctx context.Context, userUUID string, url string) (ok bool, err error) {
	err = s.permissionRepo.CheckUserHasPermission(ctx, userUUID, url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, errors.Join(err, unexpected.RequestErr)
	}
	return true, nil
}

func (s *permissionService) ActualizeResources(ctx context.Context, resources map[string]struct{}) error {

	oldResources, err := s.permissionRepo.GetResourceList(ctx)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	removeResources := make([]string, 0)
	createResources := make([]string, 0)

	for _, oldResource := range oldResources {
		_, ok := resources[oldResource.Value]
		if !ok {
			removeResources = append(removeResources, oldResource.UUID)
			continue
		}

		delete(resources, oldResource.Value)
	}

	for resource, _ := range resources {
		createResources = append(createResources, resource)
	}

	return s.permissionRepo.RemoveAndCreateResources(ctx, createResources, removeResources)
}

func (s *permissionService) ActualizeAdminGroupAndRole(ctx context.Context, adminUUID, groupName, roleName string) error {

	fileData, err := os.ReadFile(CacheAdminGroupFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return errors.Join(err, unexpected.InternalErr, errors.New("can't read permission cache file"))
		}
	}

	oldCfg := adminGroupRole{}
	if err = yaml.Unmarshal(fileData, &oldCfg); err != nil {
		return errors.Join(err, unexpected.InternalErr, errors.New("can't unmarshal permission cache file"))
	}

	var roleUUID string
	if oldCfg.RoleName == "" {
		roleUUID, err = s.permissionRepo.CreateRole(ctx, models.Role{
			Name: roleName,
			Members: []models.RoleMember{
				{
					UUID: adminUUID,
				},
			},
		})
		if err != nil {
			return err
		}

		oldCfg.RoleName = roleName
	}

	if oldCfg.RoleName != roleName {
		role, err := s.permissionRepo.GetRoleByName(ctx, oldCfg.RoleName)
		if err != nil {
			return err
		}

		roleUUID = role.UUID

		members := make([]requests.RoleMember, 0, len(role.Members)+1)
		hasAdmin := false
		for _, member := range role.Members {
			if member.UUID == adminUUID {
				hasAdmin = true
			}

			members = append(members, requests.RoleMember{
				UUID: member.UUID,
			})
		}

		if !hasAdmin {
			members = append(members, requests.RoleMember{
				UUID: adminUUID,
			})
		}

		reqRole := requests.Role{
			UUID:    roleUUID,
			Name:    roleName,
			Members: members,
		}

		err = s.UpdateRoleWithMembers(ctx, reqRole)
		if err != nil {
			return err
		}

		oldCfg.RoleName = roleUUID
	}

	if oldCfg.GroupName == "" {
		_, err = s.permissionRepo.CreateGroup(ctx, models.Group{
			Name: roleName,
			Roles: []models.SimpleRole{
				{
					UUID: roleUUID,
				},
			},
		})
		if err != nil {
			return err
		}

		oldCfg.GroupName = groupName
	}

	if oldCfg.GroupName != groupName {
		group, err := s.permissionRepo.GetGroupByName(ctx, oldCfg.GroupName)
		if err != nil {
			return err
		}

		roles := make([]requests.Role, 0, len(group.Roles)+1)
		hasAdmin := false
		for _, role := range group.Roles {
			if role.UUID == roleUUID {
				hasAdmin = true
			}

			roles = append(roles, requests.Role{
				UUID: role.UUID,
			})
		}

		if !hasAdmin {
			roles = append(roles, requests.Role{
				UUID: roleUUID,
			})
		}

		reqGroup := requests.Group{
			UUID:  group.UUID,
			Name:  groupName,
			Roles: roles,
		}

		err = s.UpdateGroupWithRolesAndResources(ctx, reqGroup)
		if err != nil {
			return err
		}

		oldCfg.GroupName = groupName
	}

	newCfg, err := yaml.Marshal(oldCfg)
	if err != nil {
		return errors.Join(err, unexpected.InternalErr, errors.New("can't marshal permission cache file"))
	}
	err = os.WriteFile(CacheAdminGroupFile, newCfg, 0600)
	if err != nil {
		return errors.Join(err, unexpected.InternalErr, errors.New("can't write permission cache file"))
	}

	return nil
}

func (s *permissionService) GetResourceList(ctx context.Context) ([]responses.Resource, error) {
	modelResources, err := s.permissionRepo.GetResourceList(ctx)
	if err != nil {
		return nil, errors.Join(err, unexpected.RequestErr)
	}

	responseResources := make([]responses.Resource, len(modelResources))
	for i, resource := range modelResources {
		responseResources[i] = responses.Resource{
			UUID:  resource.UUID,
			Value: resource.Value,
		}
	}

	return responseResources, nil
}

func (s *permissionService) CreateRole(ctx context.Context, role requests.Role) (string, error) {

	modelMember := make([]models.RoleMember, len(role.Members))
	for i, roleMember := range role.Members {
		modelMember[i] = models.RoleMember{
			UUID: roleMember.UUID,
		}
	}
	modelRole := models.Role{
		Name:    role.Name,
		Members: modelMember,
	}
	uuid, err := s.permissionRepo.CreateRole(ctx, modelRole)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (s *permissionService) CreateGroup(ctx context.Context, group requests.Group) (string, error) {
	modelRoles := make([]models.SimpleRole, len(group.Roles))
	for i, role := range group.Roles {
		modelRoles[i] = models.SimpleRole{
			UUID: role.UUID,
		}
	}

	modelResources := make([]models.Resource, len(modelRoles))
	for i, role := range modelRoles {
		modelResources[i] = models.Resource{
			UUID: role.UUID,
		}
	}

	modelGroup := models.Group{
		UUID:      group.UUID,
		Name:      group.Name,
		Roles:     modelRoles,
		Resources: modelResources,
	}

	uuid, err := s.permissionRepo.CreateGroup(ctx, modelGroup)
	if err != nil {
		return "", err
	}
	return uuid, nil
}

func (s *permissionService) UpdateRoleWithMembers(ctx context.Context, role requests.Role) error {

	dbRole, err := s.permissionRepo.GetRoleByUUID(ctx, role.UUID)
	if err != nil {
		return err
	}

	oldMemberMap := make(map[string]struct{})
	for _, member := range dbRole.Members {
		oldMemberMap[member.UUID] = struct{}{}
	}

	createMembers := make([]string, 0)
	removeMembers := make([]string, 0)
	for _, roleMember := range role.Members {
		if _, ok := oldMemberMap[roleMember.UUID]; !ok {
			createMembers = append(createMembers, roleMember.UUID)
			continue
		}
		delete(oldMemberMap, roleMember.UUID)
	}

	for member, _ := range oldMemberMap {
		removeMembers = append(removeMembers, member)
	}

	updateRole := models.UpdateRole{
		UUID:              role.UUID,
		Name:              role.Name,
		CreateMemberUUIDs: createMembers,
		DeleteMemberUUIDs: removeMembers,
	}

	return s.permissionRepo.UpdateRoleWithMembers(ctx, updateRole)
}

func (s *permissionService) UpdateGroupWithRolesAndResources(ctx context.Context, group requests.Group) error {
	dbGroup, err := s.permissionRepo.GetGroupByUUID(ctx, group.UUID)
	if err != nil {
		return err
	}

	// Found role to create and remove
	oldRoleMap := make(map[string]struct{})
	for _, role := range dbGroup.Roles {
		oldRoleMap[role.UUID] = struct{}{}
	}

	createRoleUUIDs := make([]string, 0)
	removeRoleUUIDs := make([]string, 0)
	for _, role := range group.Roles {
		if _, ok := oldRoleMap[role.UUID]; !ok {
			createRoleUUIDs = append(createRoleUUIDs, role.UUID)
			continue
		}
		delete(oldRoleMap, role.UUID)
	}

	for role, _ := range oldRoleMap {
		removeRoleUUIDs = append(removeRoleUUIDs, role)
	}

	// Found resources to create and remove
	oldResourceMap := make(map[string]struct{})
	for _, resource := range dbGroup.Resources {
		oldResourceMap[resource.UUID] = struct{}{}
	}

	createResourceUUIDs := make([]string, 0)
	removeResourceUUIDs := make([]string, 0)
	for _, resource := range group.Resources {
		if _, ok := oldResourceMap[resource.UUID]; !ok {
			createResourceUUIDs = append(createResourceUUIDs, resource.UUID)
			continue
		}
		delete(oldResourceMap, resource.UUID)
	}

	for resource, _ := range oldResourceMap {
		removeResourceUUIDs = append(removeResourceUUIDs, resource)
	}

	updateGroup := models.UpdateGroup{
		UUID:                group.UUID,
		Name:                group.Name,
		CreateRoleUUIDs:     createRoleUUIDs,
		DeleteRoleUUIDs:     removeRoleUUIDs,
		CreateResourceUUIDs: createResourceUUIDs,
		DeleteResourceUUIDs: removeResourceUUIDs,
	}

	return s.permissionRepo.UpdateGroupWithRolesAndResources(ctx, updateGroup)
}

func (s *permissionService) DeleteRole(ctx context.Context, uuid string) error {
	return s.permissionRepo.DeleteRole(ctx, uuid)
}

func (s *permissionService) DeleteGroup(ctx context.Context, uuid string) error {
	return s.permissionRepo.DeleteGroup(ctx, uuid)
}

func (s *permissionService) GetRoleByUUID(ctx context.Context, uuid string) (responses.Role, error) {
	modelRole, err := s.permissionRepo.GetRoleByUUID(ctx, uuid)
	if err != nil {
		return responses.Role{}, err
	}

	roleMembers := make([]responses.RoleMember, len(modelRole.Members))

	for i, roleMember := range modelRole.Members {
		roleMembers[i] = responses.RoleMember{
			UUID:       roleMember.UUID,
			Name:       roleMember.Name,
			SecondName: roleMember.SecondName,
			Patronymic: roleMember.Patronymic,
		}
	}

	return responses.Role{
		UUID:    modelRole.UUID,
		Name:    modelRole.Name,
		Members: roleMembers,
	}, nil
}

func (s *permissionService) GetRoleList(ctx context.Context) ([]responses.SimpleRole, error) {
	modelRoles, err := s.permissionRepo.GetRoleList(ctx)
	if err != nil {
		return nil, err
	}

	simpleRoles := make([]responses.SimpleRole, len(modelRoles))
	for i, role := range modelRoles {
		simpleRoles[i] = responses.SimpleRole{
			UUID: role.UUID,
			Name: role.Name,
		}
	}

	return simpleRoles, nil
}

func (s *permissionService) GetRoleListWithPagination(ctx context.Context, limit, offset int) ([]responses.SimpleRole, int, error) {
	modelRoles, totalRecords, err := s.permissionRepo.GetRoleListWithPagination(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	simpleRoles := make([]responses.SimpleRole, len(modelRoles))
	for i, role := range modelRoles {
		simpleRoles[i] = responses.SimpleRole{
			UUID: role.UUID,
			Name: role.Name,
		}
	}
	return simpleRoles, totalRecords, nil
}

func (s *permissionService) GetGroupByUUID(ctx context.Context, uuid string) (responses.Group, error) {
	modelGroup, err := s.permissionRepo.GetGroupByUUID(ctx, uuid)
	if err != nil {
		return responses.Group{}, err
	}

	roles := make([]responses.SimpleRole, len(modelGroup.Roles))
	for i, role := range modelGroup.Roles {
		roles[i] = responses.SimpleRole{
			UUID: role.UUID,
			Name: role.Name,
		}
	}

	resources := make([]responses.Resource, len(modelGroup.Resources))
	for i, resource := range modelGroup.Resources {
		resources[i] = responses.Resource{
			UUID:  resource.UUID,
			Value: resource.Value,
		}
	}

	return responses.Group{
		UUID:      modelGroup.UUID,
		Name:      modelGroup.Name,
		Roles:     roles,
		Resources: resources,
	}, nil
}

func (s *permissionService) GetGroupList(ctx context.Context) ([]responses.SimpleGroup, error) {
	modelGroups, err := s.permissionRepo.GetGroupList(ctx)
	if err != nil {
		return nil, err
	}

	simpleGroups := make([]responses.SimpleGroup, len(modelGroups))
	for i, group := range modelGroups {
		simpleGroups[i] = responses.SimpleGroup{
			UUID: group.UUID,
			Name: group.Name,
		}
	}

	return simpleGroups, nil
}

func (s *permissionService) GetGroupListWithPagination(ctx context.Context, limit, offset int) ([]responses.SimpleGroup, int, error) {
	modelGroups, totalRecords, err := s.permissionRepo.GetGroupListWithPagination(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	simpleGroups := make([]responses.SimpleGroup, len(modelGroups))
	for i, group := range modelGroups {
		simpleGroups[i] = responses.SimpleGroup{
			UUID: group.UUID,
			Name: group.Name,
		}
	}

	return simpleGroups, totalRecords, nil
}
