package user

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	PasswordHash string `json:"-"`
	Role         Role   `json:"role"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type StudentProfile struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	RealName  string `json:"real_name,omitempty"`
	StudentNo string `json:"student_no,omitempty"`
	College   string `json:"college,omitempty"`
	ClassName string `json:"class_name,omitempty"`
}

type StudentProfileCreateInput struct {
	UserID    int64
	RealName  string
	StudentNo string
	College   string
	ClassName string
}

type Role string

const (
	RoleStudent  Role = "STUDENT"
	RoleTeacher  Role = "TEACHER"
	RoleMerchant Role = "MERCHANT"
	RoleAdmin    Role = "ADMIN"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleStudent, RoleTeacher, RoleMerchant, RoleAdmin:
		return true
	default:
		return false
	}
}

func (r Role) CanRegister() bool {
	return r == RoleStudent || r == RoleTeacher
}

type CreateInput struct {
	Username     string
	Email        string
	Phone        string
	PasswordHash string
	Role         Role
	Status       string
}

type ImportUserError struct {
	Line     int    `json:"line"`
	Username string `json:"username"`
	Reason   string `json:"reason"`
}

type ImportUserResult struct {
	Total        int               `json:"total"`
	SuccessCount int               `json:"success_count"`
	FailedCount  int               `json:"failed_count"`
	Errors       []ImportUserError `json:"errors"`
}

type ImportUserRow struct {
	Username  string
	Password  string
	Email     string
	Phone     string
	Role      Role
	RealName  string
	StudentNo string
	College   string
	ClassName string
}

type AdminUserFilter struct {
	Keyword  string
	Role     string
	Status   string
	Page     int
	PageSize int
}

type AdminUserProfile struct {
	RealName  string `json:"real_name,omitempty"`
	StudentNo string `json:"student_no,omitempty"`
	College   string `json:"college,omitempty"`
	ClassName string `json:"class_name,omitempty"`
}

type AdminUserItem struct {
	ID        int64            `json:"id"`
	Username  string           `json:"username"`
	Email     string           `json:"email,omitempty"`
	Phone     string           `json:"phone,omitempty"`
	Role      Role             `json:"role"`
	Status    string           `json:"status"`
	CreatedAt string           `json:"created_at,omitempty"`
	Profile   AdminUserProfile `json:"profile,omitempty"`
}

type AdminUserListResult struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Items    []AdminUserItem `json:"items"`
}

type AdminUserDetailResult struct {
	AdminUserItem
	UpdatedAt string `json:"updated_at,omitempty"`
}
