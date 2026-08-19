package user

import "time"

const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

type Permission string

const (
	PermissionUserRead       Permission = "user.read"
	PermissionUserManage     Permission = "user.manage"
	PermissionMemorialRead   Permission = "memorial.read"
	PermissionMemorialManage Permission = "memorial.manage"
)

type User struct {
	ID           string       `json:"id"`
	Username     string       `json:"username"`
	DisplayName  string       `json:"display_name"`
	AvatarURL    string       `json:"avatar_url"`
	PasswordHash string       `json:"-"`
	Role         string       `json:"role"`
	AllHouses    bool         `json:"all_houses"`
	HouseIDs     []string     `json:"house_ids"`
	Permissions  []Permission `json:"permissions"`
	Active       bool         `json:"active"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type Session struct {
	TokenHash string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type CreateInput struct {
	Username    string
	DisplayName string
	Password    string
	Role        string
	AllHouses  bool
	HouseIDs   []string
}

type UpdateInput struct {
	Username    string
	DisplayName string
	Role        string
	Password    string
	AllHouses  bool
	HouseIDs   []string
}

func PermissionsForRole(role string) []Permission {
	switch role {
	case RoleAdmin:
		return []Permission{
			PermissionUserRead,
			PermissionUserManage,
			PermissionMemorialRead,
			PermissionMemorialManage,
		}
	case RoleViewer:
		return []Permission{PermissionMemorialRead}
	case RoleEditor:
		return []Permission{PermissionMemorialRead, PermissionMemorialManage}
	default:
		return []Permission{}
	}
}

func HasPermission(role string, permission Permission) bool {
	for _, candidate := range PermissionsForRole(role) {
		if candidate == permission {
			return true
		}
	}
	return false
}

type LoginResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}
