package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Repository struct {
	db *sql.DB
}

type dbExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type rowScanner interface {
	Scan(dest ...any) error
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByLogin(ctx context.Context, identifier string) (User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, phone, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE username = ? OR email = ? OR phone = ?
		LIMIT 1`, identifier, identifier, identifier)

	var item User
	if err := row.Scan(
		&item.ID,
		&item.Username,
		&item.Email,
		&item.Phone,
		&item.PasswordHash,
		&item.Role,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return User{}, err
	}

	return item, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (User, error) {
	return r.findByIDWithExecutor(ctx, r.db, id)
}

func (r *Repository) Create(ctx context.Context, input CreateInput) (User, error) {
	item, err := r.CreateWithExecutor(ctx, r.db, input)
	if err != nil {
		return User{}, err
	}
	return *item, nil
}

func (r *Repository) CreateWithExecutor(ctx context.Context, executor dbExecutor, input CreateInput) (*User, error) {
	res, err := executor.ExecContext(ctx, `
		INSERT INTO users (username, email, phone, password_hash, role, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, input.Username, nullIfEmpty(input.Email), nullIfEmpty(input.Phone), input.PasswordHash, string(input.Role), input.Status)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	item, err := r.findByIDWithExecutor(ctx, executor, id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) findByIDWithExecutor(ctx context.Context, executor dbExecutor, id int64) (User, error) {
	row := executor.QueryRowContext(ctx, `
		SELECT id, username, email, phone, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE id = ?`, id)

	var item User
	if err := row.Scan(
		&item.ID,
		&item.Username,
		&item.Email,
		&item.Phone,
		&item.PasswordHash,
		&item.Role,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return User{}, err
	}

	return item, nil
}

func (r *Repository) CreateStudentProfile(ctx context.Context, executor dbExecutor, input StudentProfileCreateInput) error {
	_, err := executor.ExecContext(ctx, `
		INSERT INTO student_profiles (user_id, real_name, student_no, college, class_name)
		VALUES (?, ?, ?, ?, ?)
	`, input.UserID, nullIfEmpty(input.RealName), nullIfEmpty(input.StudentNo), nullIfEmpty(input.College), nullIfEmpty(input.ClassName))
	if err != nil {
		return fmt.Errorf("insert student profile: %w", err)
	}
	return nil
}

func (r *Repository) UsernameExists(ctx context.Context, executor dbExecutor, username string) (bool, error) {
	return r.exists(ctx, executor, `SELECT 1 FROM users WHERE username = ? LIMIT 1`, username)
}

func (r *Repository) EmailExists(ctx context.Context, executor dbExecutor, email string) (bool, error) {
	return r.exists(ctx, executor, `SELECT 1 FROM users WHERE email = ? LIMIT 1`, email)
}

func (r *Repository) PhoneExists(ctx context.Context, executor dbExecutor, phone string) (bool, error) {
	return r.exists(ctx, executor, `SELECT 1 FROM users WHERE phone = ? LIMIT 1`, phone)
}

func (r *Repository) StudentNoExists(ctx context.Context, executor dbExecutor, studentNo string) (bool, error) {
	return r.exists(ctx, executor, `SELECT 1 FROM student_profiles WHERE student_no = ? LIMIT 1`, studentNo)
}

func (r *Repository) ListAdminUsers(ctx context.Context, filter AdminUserFilter) ([]AdminUserItem, int64, error) {
	whereParts := make([]string, 0, 4)
	args := make([]any, 0, 8)

	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		whereParts = append(whereParts, "(u.username LIKE ? OR u.email LIKE ? OR u.phone LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like, like)
	}
	if role := strings.TrimSpace(filter.Role); role != "" {
		whereParts = append(whereParts, "u.role = ?")
		args = append(args, role)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		whereParts = append(whereParts, "u.status = ?")
		args = append(args, status)
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = " WHERE " + strings.Join(whereParts, " AND ")
	}

	var total int64
	countQuery := `
		SELECT COUNT(1)
		FROM users u
		LEFT JOIN student_profiles sp ON sp.user_id = u.id` + whereClause
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin users: %w", err)
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT u.id, u.username, u.email, u.phone, u.role, u.status, u.created_at,
		       sp.real_name, sp.student_no, sp.college, sp.class_name
		FROM users u
		LEFT JOIN student_profiles sp ON sp.user_id = u.id` + whereClause + `
		ORDER BY u.id DESC
		LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin users: %w", err)
	}
	defer rows.Close()

	items := make([]AdminUserItem, 0)
	for rows.Next() {
		item, err := scanAdminUserItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate admin users: %w", err)
	}
	return items, total, nil
}

func (r *Repository) GetAdminUserByID(ctx context.Context, id int64) (AdminUserDetailResult, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.email, u.phone, u.role, u.status, u.created_at,
		       sp.real_name, sp.student_no, sp.college, sp.class_name, u.updated_at
		FROM users u
		LEFT JOIN student_profiles sp ON sp.user_id = u.id
		WHERE u.id = ?`, id)

	return scanAdminUserDetail(row)
}

func (r *Repository) UpdateUserStatus(ctx context.Context, id int64, status string) (AdminUserDetailResult, error) {
	if _, err := r.db.ExecContext(ctx, `UPDATE users SET status = ? WHERE id = ?`, status, id); err != nil {
		return AdminUserDetailResult{}, fmt.Errorf("update user status: %w", err)
	}
	return r.GetAdminUserByID(ctx, id)
}

func (r *Repository) UpdateUserRole(ctx context.Context, id int64, role Role) (AdminUserDetailResult, error) {
	if _, err := r.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE id = ?`, string(role), id); err != nil {
		return AdminUserDetailResult{}, fmt.Errorf("update user role: %w", err)
	}
	return r.GetAdminUserByID(ctx, id)
}

func (r *Repository) ResetUserPassword(ctx context.Context, id int64, passwordHash string) (AdminUserDetailResult, error) {
	if _, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, id); err != nil {
		return AdminUserDetailResult{}, fmt.Errorf("reset password: %w", err)
	}
	return r.GetAdminUserByID(ctx, id)
}

func (r *Repository) exists(ctx context.Context, executor dbExecutor, query string, arg string) (bool, error) {
	var marker int
	err := executor.QueryRowContext(ctx, query, arg).Scan(&marker)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query exists: %w", err)
	}
	return true, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func scanAdminUserItem(row rowScanner) (AdminUserItem, error) {
	var item AdminUserItem
	var email, phone, createdAt, realName, studentNo, college, className sql.NullString
	if err := row.Scan(&item.ID, &item.Username, &email, &phone, &item.Role, &item.Status, &createdAt, &realName, &studentNo, &college, &className); err != nil {
		return AdminUserItem{}, err
	}
	if email.Valid {
		item.Email = email.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}
	if createdAt.Valid {
		item.CreatedAt = createdAt.String
	}
	if realName.Valid || studentNo.Valid || college.Valid || className.Valid {
		item.Profile = AdminUserProfile{
			RealName:  realName.String,
			StudentNo: studentNo.String,
			College:   college.String,
			ClassName: className.String,
		}
	}
	return item, nil
}

func scanAdminUserDetail(row rowScanner) (AdminUserDetailResult, error) {
	var item AdminUserDetailResult
	var email, phone, createdAt, updatedAt, realName, studentNo, college, className sql.NullString
	if err := row.Scan(&item.ID, &item.Username, &email, &phone, &item.Role, &item.Status, &createdAt, &realName, &studentNo, &college, &className, &updatedAt); err != nil {
		return AdminUserDetailResult{}, err
	}
	if email.Valid {
		item.Email = email.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}
	if createdAt.Valid {
		item.CreatedAt = createdAt.String
	}
	if updatedAt.Valid {
		item.UpdatedAt = updatedAt.String
	}
	if realName.Valid || studentNo.Valid || college.Valid || className.Valid {
		item.Profile = AdminUserProfile{
			RealName:  realName.String,
			StudentNo: studentNo.String,
			College:   college.String,
			ClassName: className.String,
		}
	}
	return item, nil
}
