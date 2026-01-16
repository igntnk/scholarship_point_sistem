package repository

import (
	"context"
	"errors"
	"github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/igntnk/scholarship_point_system/db"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"github.com/igntnk/scholarship_point_system/service/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PermissionRepository interface {
	CheckUserHasPermission(ctx context.Context, userUUID string, url string) error
	RemoveAndCreateResources(ctx context.Context, create, remove []string) error
	GetResourceList(ctx context.Context) ([]models.Resource, error)
	CreateRole(ctx context.Context, role models.Role) (string, error)
	CreateGroup(ctx context.Context, group models.Group) (string, error)
	UpdateRoleWithMembers(ctx context.Context, role models.UpdateRole) error
	UpdateGroupWithRolesAndResources(ctx context.Context, group models.UpdateGroup) error
	DeleteRole(ctx context.Context, uuid string) error
	DeleteGroup(ctx context.Context, uuid string) error
	GetRoleByUUID(ctx context.Context, uuid string) (models.Role, error)
	GetRoleList(ctx context.Context) ([]models.SimpleRole, error)
	GetRoleListWithPagination(ctx context.Context, limit, offset int) ([]models.SimpleRole, int, error)
	GetGroupByUUID(ctx context.Context, uuid string) (models.Group, error)
	GetGroupList(ctx context.Context) ([]models.SimpleGroup, error)
	GetGroupListWithPagination(ctx context.Context, limit, offset int) ([]models.SimpleGroup, int, error)
}

type permissionRepository struct {
	queries   *db.Queries
	txCreator db.TxCreator
}

func NewPermissionRepository(pool pgxv5.Tr) PermissionRepository {
	return &permissionRepository{
		queries:   db.New(pool),
		txCreator: db.NewTxCreator(pool),
	}
}

func (r *permissionRepository) CheckUserHasPermission(ctx context.Context, userUUID string, url string) error {

	pgUUID, err := ParseToPgUUID(userUUID)
	if err != nil {
		return err
	}

	pgResource, err := ParseToPgText(url)
	if err != nil {
		return err
	}

	args := db.CheckHasPermissionParams{
		Uuid:  pgUUID,
		Value: pgResource,
	}
	_, err = r.queries.CheckHasPermission(ctx, args)
	return err
}

func (r *permissionRepository) RemoveAndCreateResources(ctx context.Context, create, remove []string) (err error) {
	pgCreate := make([]pgtype.Text, len(create))
	for i, resource := range create {
		pgCreate[i], err = ParseToPgText(resource)
		if err != nil {
			return errors.Join(err, parsing.InputDataErr)
		}
	}

	pgRemove := make([]pgtype.UUID, len(remove))
	for i, resource := range remove {
		pgRemove[i], err = ParseToPgUUID(resource)
		if err != nil {
			return errors.Join(err, parsing.InputDataErr)
		}
	}

	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	batch := qtx.BatchInsertResources(ctx, pgCreate)
	batch.Exec(func(i int, e error) {
		err = errors.Join(err, e)
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	err = qtx.BatchDeleteResources(ctx, pgRemove)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	return nil
}

func (r *permissionRepository) GetResourceList(ctx context.Context) (resources []models.Resource, err error) {
	pgResources, err := r.queries.GetResourceList(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.Resource{}, nil
		}
		return nil, errors.Join(err, parsing.InputDataErr)
	}

	resources = make([]models.Resource, len(pgResources))
	for i, pgResource := range pgResources {
		resources[i] = models.Resource{
			UUID:  pgResource.Uuid.String(),
			Value: pgResource.Value.String,
		}
	}

	return resources, nil
}

func (r *permissionRepository) CreateRole(ctx context.Context, role models.Role) (string, error) {

	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgRoleUUID, err := qtx.CreateRole(ctx, role.Name)
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	args := make([]db.AddUsersToRoleParams, len(role.Members))
	for i, member := range role.Members {
		pgMemberUUID, err := ParseToPgUUID(member.UUID)
		if err != nil {
			return "", errors.Join(err, parsing.InputDataErr)
		}

		args[i] = db.AddUsersToRoleParams{
			RoleUuid: pgRoleUUID,
			UserUuid: pgMemberUUID,
		}
	}

	batch := qtx.AddUsersToRole(ctx, args)
	batch.Exec(func(i int, e error) {
		err = errors.Join(err, e)
	})
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	return pgRoleUUID.String(), nil
}

func (r *permissionRepository) CreateGroup(ctx context.Context, group models.Group) (string, error) {
	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgGroupUUID, err := qtx.CreateGroup(ctx, group.Name)
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	rolesToGroupArgs := make([]db.AddRolesToGroupParams, len(group.Roles))
	for i, role := range group.Roles {
		pgRoleUUID, err := ParseToPgUUID(role.UUID)
		if err != nil {
			return "", errors.Join(err, parsing.InputDataErr)
		}

		rolesToGroupArgs[i] = db.AddRolesToGroupParams{
			RoleUuid:  pgRoleUUID,
			GroupUuid: pgGroupUUID,
		}
	}

	resourcesToGroupArgs := make([]db.AddResourcesToGroupParams, len(group.Resources))
	for i, resource := range group.Resources {
		pgResourceUUID, err := ParseToPgUUID(resource.UUID)
		if err != nil {
			return "", errors.Join(err, parsing.InputDataErr)
		}

		resourcesToGroupArgs[i] = db.AddResourcesToGroupParams{
			ResourceUuid: pgResourceUUID,
			GroupUuid:    pgGroupUUID,
		}
	}

	rolesToGroupBatch := qtx.AddRolesToGroup(ctx, rolesToGroupArgs)
	rolesToGroupBatch.Exec(func(i int, e error) {
		err = errors.Join(err, e)
	})
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	resourcesToGroupBatch := qtx.AddResourcesToGroup(ctx, resourcesToGroupArgs)
	resourcesToGroupBatch.Exec(func(i int, e error) {
		err = errors.Join(err, e)
	})
	if err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", errors.Join(err, unexpected.RequestErr)
	}

	return pgGroupUUID.String(), nil
}

func (r *permissionRepository) UpdateRoleWithMembers(ctx context.Context, role models.UpdateRole) error {
	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgRoleUUID, err := ParseToPgUUID(role.UUID)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	// Change Role Name
	err = qtx.UpdateRole(ctx, db.UpdateRoleParams{
		Name: role.Name,
		Uuid: pgRoleUUID,
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	// Add Members To Role
	addMemberParams := make([]db.AddUsersToRoleParams, len(role.CreateMemberUUIDs))
	for i, uuid := range role.CreateMemberUUIDs {
		pgMemberUUID, err := ParseToPgUUID(uuid)
		if err != nil {
			return errors.Join(err, parsing.InputDataErr)
		}

		addMemberParams[i] = db.AddUsersToRoleParams{
			RoleUuid: pgRoleUUID,
			UserUuid: pgMemberUUID,
		}
	}

	batch := qtx.AddUsersToRole(ctx, addMemberParams)
	batch.Exec(func(i int, e error) {
		err = errors.Join(err, e)
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	// Remove Members from Role
	deleteMemberParams := make([]pgtype.UUID, len(role.DeleteMemberUUIDs))
	for i, uuid := range role.DeleteMemberUUIDs {
		pgMemberUUID, err := ParseToPgUUID(uuid)
		if err != nil {
			return errors.Join(err, parsing.InputDataErr)
		}
		deleteMemberParams[i] = pgMemberUUID
	}
	if err = qtx.RemoveMembersFromRole(ctx, deleteMemberParams); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	return nil
}

func (r *permissionRepository) UpdateGroupWithRolesAndResources(ctx context.Context, group models.UpdateGroup) error {
	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Change Group Name
	pgGroupUUID, err := ParseToPgUUID(group.UUID)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	err = qtx.UpdateGroup(ctx, db.UpdateGroupParams{
		Name: group.Name,
		Uuid: pgGroupUUID,
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	// Add Roles to Group
	addRolesParams := make([]db.AddRolesToGroupParams, len(group.CreateRoleUUIDs))
	for i, uuid := range group.CreateRoleUUIDs {
		pgRoleUUID, err := ParseToPgUUID(uuid)
		if err != nil {
			return errors.Join(err, parsing.InputDataErr)
		}
		addRolesParams[i] = db.AddRolesToGroupParams{
			RoleUuid:  pgRoleUUID,
			GroupUuid: pgGroupUUID,
		}
	}
	roleBatch := qtx.AddRolesToGroup(ctx, addRolesParams)
	roleBatch.Exec(func(i int, e error) {
		err = errors.Join(err, e)
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	// Remove Roles from Group
	deleteRolesParams := make([]pgtype.UUID, len(group.DeleteRoleUUIDs))
	for i, uuid := range group.DeleteRoleUUIDs {
		pgRoleUUID, err := ParseToPgUUID(uuid)
		if err != nil {
			return errors.Join(err, parsing.InputDataErr)
		}
		deleteRolesParams[i] = pgRoleUUID
	}
	if err = qtx.RemoveRolesFromGroup(ctx, deleteRolesParams); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	// Add Resources to Group
	addResourcesParams := make([]db.AddResourcesToGroupParams, len(group.CreateResourceUUIDs))
	for i, uuid := range group.CreateResourceUUIDs {
		pgResourceUUID, err := ParseToPgUUID(uuid)
		if err != nil {
			return errors.Join(err, parsing.InputDataErr)
		}
		addResourcesParams[i] = db.AddResourcesToGroupParams{
			GroupUuid:    pgGroupUUID,
			ResourceUuid: pgResourceUUID,
		}
	}
	groupBatch := qtx.AddResourcesToGroup(ctx, addResourcesParams)
	groupBatch.Exec(func(i int, e error) {
		err = errors.Join(err, e)
	})
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	// Remove Resources from Group
	removeResourcesArg := make([]pgtype.UUID, len(group.DeleteResourceUUIDs))
	for i, uuid := range group.DeleteResourceUUIDs {
		pgResourceUUID, err := ParseToPgUUID(uuid)
		if err != nil {
			return errors.Join(err, parsing.InputDataErr)
		}
		removeResourcesArg[i] = pgResourceUUID
	}
	if err = qtx.RemoveResourcesFromGroup(ctx, removeResourcesArg); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	return nil
}

func (r *permissionRepository) DeleteRole(ctx context.Context, uuid string) error {
	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgRoleUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	if err = qtx.DeleteMembersFromDeletedRole(ctx, pgRoleUUID); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = qtx.DeleteRole(ctx, pgRoleUUID); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	return nil
}

func (r *permissionRepository) DeleteGroup(ctx context.Context, uuid string) error {
	tx, err := r.txCreator.CreateTx(ctx)
	if err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgGroupUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return errors.Join(err, parsing.InputDataErr)
	}

	if err = qtx.DeleteRolesFromDeletedGroup(ctx, pgGroupUUID); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = qtx.DeleteResourcesFromDeletedGroup(ctx, pgGroupUUID); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = qtx.DeleteGroup(ctx, pgGroupUUID); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Join(err, unexpected.RequestErr)
	}

	return nil
}

func (r *permissionRepository) GetRoleByUUID(ctx context.Context, uuid string) (models.Role, error) {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return models.Role{}, errors.Join(err, parsing.InputDataErr)
	}
	dbRole, err := r.queries.GetRoleByUUID(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Role{}, errors.Join(err, validation.NoDataFoundErr)
		}
		return models.Role{}, errors.Join(err, parsing.InputDataErr)
	}

	dbUsers, err := r.queries.GetRoleMembers(ctx, pgUUID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return models.Role{}, errors.Join(err, parsing.InputDataErr)
		}

		dbUsers = []db.SysUser{}
	}

	members := make([]models.RoleMember, len(dbUsers))
	for i, dbUser := range dbUsers {
		members[i] = models.RoleMember{
			UUID:       dbUser.Uuid.String(),
			Name:       dbUser.Name,
			SecondName: dbUser.SecondName,
			Patronymic: dbUser.Patronymic.String,
		}
	}

	return models.Role{
		UUID:    dbRole.Uuid.String(),
		Name:    dbRole.Name,
		Members: members,
	}, nil
}

func (r *permissionRepository) GetRoleList(ctx context.Context) ([]models.SimpleRole, error) {
	dbRoles, err := r.queries.GetRoleList(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.SimpleRole{}, nil
		}
		return nil, errors.Join(err, parsing.InputDataErr)
	}

	simpleRoles := make([]models.SimpleRole, len(dbRoles))
	for i, role := range dbRoles {
		simpleRoles[i] = models.SimpleRole{
			UUID: role.Uuid.String(),
			Name: role.Name,
		}
	}
	return simpleRoles, nil
}

func (r *permissionRepository) GetRoleListWithPagination(ctx context.Context, limit, offset int) ([]models.SimpleRole, int, error) {
	args := db.GetRoleListWithPaginationParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
	dbRoles, err := r.queries.GetRoleListWithPagination(ctx, args)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.SimpleRole{}, 0, nil
		}
		return nil, 0, errors.Join(err, parsing.InputDataErr)
	}

	totalRecords := 0
	simpleRoles := make([]models.SimpleRole, len(dbRoles))
	for i, role := range dbRoles {
		totalRecords = int(role.TotalRecords)
		simpleRoles[i] = models.SimpleRole{
			UUID: role.Uuid.String(),
			Name: role.Name,
		}
	}

	return simpleRoles, totalRecords, nil
}

func (r *permissionRepository) GetGroupByUUID(ctx context.Context, uuid string) (models.Group, error) {
	pgUUID, err := ParseToPgUUID(uuid)
	if err != nil {
		return models.Group{}, errors.Join(err, parsing.InputDataErr)
	}
	dbGroup, err := r.queries.GetGroupByUUID(ctx, pgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Group{}, errors.Join(err, validation.NoDataFoundErr)
		}
		return models.Group{}, errors.Join(err, parsing.InputDataErr)
	}

	dbRoles, err := r.queries.GetGroupRoles(ctx, pgUUID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return models.Group{}, errors.Join(err, validation.NoDataFoundErr)
		}

		dbRoles = []db.AuthRole{}
	}

	dbResources, err := r.queries.GetGroupResources(ctx, pgUUID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return models.Group{}, errors.Join(err, validation.NoDataFoundErr)
		}
		dbResources = []db.Resource{}
	}

	roles := make([]models.SimpleRole, len(dbRoles))
	for i, role := range dbRoles {
		roles[i] = models.SimpleRole{
			UUID: role.Uuid.String(),
			Name: role.Name,
		}
	}

	resources := make([]models.Resource, len(dbResources))
	for i, resource := range dbResources {
		resources[i] = models.Resource{
			UUID:  resource.Uuid.String(),
			Value: resource.Value.String,
		}
	}

	group := models.Group{
		UUID:      dbGroup.Uuid.String(),
		Name:      dbGroup.Name,
		Roles:     roles,
		Resources: resources,
	}

	return group, nil
}

func (r *permissionRepository) GetGroupList(ctx context.Context) ([]models.SimpleGroup, error) {
	dbGroups, err := r.queries.GetGroupList(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.SimpleGroup{}, nil
		}
		return nil, errors.Join(err, parsing.InputDataErr)
	}

	simpleGroups := make([]models.SimpleGroup, len(dbGroups))
	for i, dbGroup := range dbGroups {
		simpleGroups[i] = models.SimpleGroup{
			UUID: dbGroup.Uuid.String(),
			Name: dbGroup.Name,
		}
	}

	return simpleGroups, nil
}

func (r *permissionRepository) GetGroupListWithPagination(ctx context.Context, limit, offset int) ([]models.SimpleGroup, int, error) {
	args := db.GetGroupListWithPaginationParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
	dbGroups, err := r.queries.GetGroupListWithPagination(ctx, args)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.SimpleGroup{}, 0, nil
		}
		return nil, 0, errors.Join(err, parsing.InputDataErr)
	}

	totalRecords := 0
	simpleGroups := make([]models.SimpleGroup, len(dbGroups))
	for i, dbGroup := range dbGroups {
		totalRecords = int(dbGroup.TotalRecords)
		simpleGroups[i] = models.SimpleGroup{
			UUID: dbGroup.Uuid.String(),
			Name: dbGroup.Name,
		}
	}
	return simpleGroups, totalRecords, nil
}
