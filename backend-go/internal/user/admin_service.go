package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAdminUserNotFound      = errors.New("admin user not found")
	ErrAdminUserSelfModify    = errors.New("cannot modify current admin")
	ErrAdminUserInvalidStatus = errors.New("invalid user status")
	ErrAdminUserInvalidRole   = errors.New("invalid user role")
	ErrDuplicateName          = errors.New("duplicate name")
)

type AdminService struct {
	repo *Repository
}

func NewAdminService(repo *Repository) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) ListUsers(ctx context.Context, filter AdminUserFilter) (AdminUserListResult, error) {
	items, total, err := s.repo.ListAdminUsers(ctx, filter)
	if err != nil {
		return AdminUserListResult{}, err
	}
	return AdminUserListResult{
		Total:    total,
		Page:     normalizePage(filter.Page),
		PageSize: normalizePageSize(filter.PageSize),
		Items:    items,
	}, nil
}

func (s *AdminService) GetUser(ctx context.Context, id int64) (AdminUserDetailResult, error) {
	item, err := s.repo.GetAdminUserByID(ctx, id)
	if err != nil {
		return AdminUserDetailResult{}, normalizeAdminUserError(err)
	}
	return item, nil
}

func (s *AdminService) UpdateStatus(ctx context.Context, currentUserID, targetUserID int64, status string) (AdminUserDetailResult, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	if !isAllowedStatus(status) {
		return AdminUserDetailResult{}, ErrAdminUserInvalidStatus
	}
	if currentUserID == targetUserID {
		return AdminUserDetailResult{}, ErrAdminUserSelfModify
	}
	item, err := s.repo.UpdateUserStatus(ctx, targetUserID, status)
	if err != nil {
		return AdminUserDetailResult{}, normalizeAdminUserError(err)
	}
	return item, nil
}

func (s *AdminService) UpdateRole(ctx context.Context, currentUserID, targetUserID int64, role string) (AdminUserDetailResult, error) {
	role = strings.ToUpper(strings.TrimSpace(role))
	parsedRole := Role(role)
	if !parsedRole.IsValid() || parsedRole == RoleAdmin || parsedRole == Role("") {
		return AdminUserDetailResult{}, ErrAdminUserInvalidRole
	}
	if currentUserID == targetUserID {
		return AdminUserDetailResult{}, ErrAdminUserSelfModify
	}
	item, err := s.repo.UpdateUserRole(ctx, targetUserID, parsedRole)
	if err != nil {
		return AdminUserDetailResult{}, normalizeAdminUserError(err)
	}
	return item, nil
}

func (s *AdminService) ResetPassword(ctx context.Context, targetUserID int64, newPassword string) (AdminUserDetailResult, string, error) {
	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" {
		newPassword = "123456"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return AdminUserDetailResult{}, "", fmt.Errorf("hash password: %w", err)
	}
	item, err := s.repo.ResetUserPassword(ctx, targetUserID, string(hash))
	if err != nil {
		return AdminUserDetailResult{}, "", normalizeAdminUserError(err)
	}
	return item, newPassword, nil
}

func normalizeAdminUserError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAdminUserNotFound
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return ErrDuplicateName
		}
	}
	return err
}

func isAllowedStatus(status string) bool {
	switch status {
	case "ACTIVE", "DISABLED", "PENDING":
		return true
	default:
		return false
	}
}

func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 10
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}
