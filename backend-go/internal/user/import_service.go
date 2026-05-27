package user

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var importUserHeaders = []string{"username", "password", "email", "phone", "role", "real_name", "student_no", "college", "class_name"}

type ImportService struct {
	repo *Repository
}

func NewImportService(repo *Repository) *ImportService {
	return &ImportService{repo: repo}
}

func (s *ImportService) ImportCSV(ctx context.Context, reader io.Reader) (ImportUserResult, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true

	header, err := csvReader.Read()
	if err != nil {
		return ImportUserResult{}, fmt.Errorf("read csv header: %w", err)
	}
	if err := validateImportHeaders(header); err != nil {
		return ImportUserResult{}, err
	}

	result := ImportUserResult{Errors: make([]ImportUserError, 0)}
	seenUsernames := make(map[string]struct{})
	seenEmails := make(map[string]struct{})
	seenPhones := make(map[string]struct{})
	seenStudentNos := make(map[string]struct{})

	lineNo := 1
	for {
		record, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return result, fmt.Errorf("read csv line %d: %w", lineNo+1, err)
		}
		lineNo++
		result.Total++

		row, parseErr := parseImportRow(record)
		if parseErr != nil {
			result.Errors = append(result.Errors, ImportUserError{Line: lineNo, Username: "", Reason: parseErr.Error()})
			continue
		}

		if err := s.importOne(ctx, row, seenUsernames, seenEmails, seenPhones, seenStudentNos); err != nil {
			result.Errors = append(result.Errors, ImportUserError{Line: lineNo, Username: row.Username, Reason: err.Error()})
			continue
		}

		result.SuccessCount++
	}

	result.FailedCount = len(result.Errors)
	return result, nil
}

func (s *ImportService) importOne(ctx context.Context, row ImportUserRow, seenUsernames, seenEmails, seenPhones, seenStudentNos map[string]struct{}) error {
	row.Username = strings.TrimSpace(row.Username)
	row.Password = strings.TrimSpace(row.Password)
	row.Email = strings.TrimSpace(row.Email)
	row.Phone = strings.TrimSpace(row.Phone)
	row.RealName = strings.TrimSpace(row.RealName)
	row.StudentNo = strings.TrimSpace(row.StudentNo)
	row.College = strings.TrimSpace(row.College)
	row.ClassName = strings.TrimSpace(row.ClassName)
	if row.Username == "" {
		return errors.New("用户名不能为空")
	}
	if !strings.EqualFold(string(row.Role), string(RoleStudent)) && !strings.EqualFold(string(row.Role), string(RoleTeacher)) {
		return errors.New("角色只能是 STUDENT 或 TEACHER")
	}
	row.Role = Role(strings.ToUpper(string(row.Role)))
	if row.Password == "" {
		row.Password = "123456"
	}

	key := strings.ToLower(row.Username)
	if _, exists := seenUsernames[key]; exists {
		return errors.New("同一文件内用户名重复")
	}
	if exists, err := s.repo.UsernameExists(ctx, s.repo.db, row.Username); err != nil {
		return err
	} else if exists {
		return errors.New("用户名已存在")
	}

	if row.Email != "" {
		emailKey := strings.ToLower(row.Email)
		if _, exists := seenEmails[emailKey]; exists {
			return errors.New("同一文件内邮箱重复")
		}
		if exists, err := s.repo.EmailExists(ctx, s.repo.db, row.Email); err != nil {
			return err
		} else if exists {
			return errors.New("邮箱已存在")
		}
	}

	if row.Phone != "" {
		phoneKey := row.Phone
		if _, exists := seenPhones[phoneKey]; exists {
			return errors.New("同一文件内手机号重复")
		}
		if exists, err := s.repo.PhoneExists(ctx, s.repo.db, row.Phone); err != nil {
			return err
		} else if exists {
			return errors.New("手机号已存在")
		}
	}

	if row.StudentNo != "" {
		studentKey := row.StudentNo
		if _, exists := seenStudentNos[studentKey]; exists {
			return errors.New("同一文件内学号重复")
		}
		if exists, err := s.repo.StudentNoExists(ctx, s.repo.db, row.StudentNo); err != nil {
			return err
		} else if exists {
			return errors.New("学号已存在")
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(row.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	exists, err := s.repo.UsernameExists(ctx, tx, row.Username)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("用户名已存在")
	}

	if row.Email != "" {
		exists, err = s.repo.EmailExists(ctx, tx, row.Email)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("邮箱已存在")
		}
	}

	if row.Phone != "" {
		exists, err = s.repo.PhoneExists(ctx, tx, row.Phone)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("手机号已存在")
		}
	}

	if row.StudentNo != "" {
		exists, err = s.repo.StudentNoExists(ctx, tx, row.StudentNo)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("学号已存在")
		}
	}

	createdUser, err := s.repo.CreateWithExecutor(ctx, tx, CreateInput{
		Username:     row.Username,
		Email:        row.Email,
		Phone:        row.Phone,
		PasswordHash: string(passwordHash),
		Role:         row.Role,
		Status:       "ACTIVE",
	})
	if err != nil {
		return err
	}

	if err := s.repo.CreateStudentProfile(ctx, tx, StudentProfileCreateInput{
		UserID:    createdUser.ID,
		RealName:  row.RealName,
		StudentNo: row.StudentNo,
		College:   row.College,
		ClassName: row.ClassName,
	}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	seenUsernames[key] = struct{}{}
	if row.Email != "" {
		seenEmails[strings.ToLower(row.Email)] = struct{}{}
	}
	if row.Phone != "" {
		seenPhones[row.Phone] = struct{}{}
	}
	if row.StudentNo != "" {
		seenStudentNos[row.StudentNo] = struct{}{}
	}

	return nil
}

func validateImportHeaders(header []string) error {
	if len(header) != len(importUserHeaders) {
		return fmt.Errorf("csv表头不正确，期望 %d 列", len(importUserHeaders))
	}
	for i, value := range header {
		current := strings.TrimSpace(strings.TrimPrefix(value, "\ufeff"))
		if !strings.EqualFold(current, importUserHeaders[i]) {
			return fmt.Errorf("csv表头不正确，第 %d 列应为 %s", i+1, importUserHeaders[i])
		}
	}
	return nil
}

func parseImportRow(record []string) (ImportUserRow, error) {
	if len(record) != len(importUserHeaders) {
		return ImportUserRow{}, fmt.Errorf("列数不正确，期望 %d 列", len(importUserHeaders))
	}

	role := strings.ToUpper(strings.TrimSpace(record[4]))
	if role == "" {
		role = string(RoleStudent)
	}

	return ImportUserRow{
		Username:  strings.TrimSpace(record[0]),
		Password:  strings.TrimSpace(record[1]),
		Email:     strings.TrimSpace(record[2]),
		Phone:     strings.TrimSpace(record[3]),
		Role:      Role(role),
		RealName:  strings.TrimSpace(record[5]),
		StudentNo: strings.TrimSpace(record[6]),
		College:   strings.TrimSpace(record[7]),
		ClassName: strings.TrimSpace(record[8]),
	}, nil
}
