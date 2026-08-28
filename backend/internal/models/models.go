package models

import "github.com/uptrace/bun"

type Role string
type AuthMethod string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

const (
	AuthMethodPassword AuthMethod = "password"
	AuthMethodOIDC     AuthMethod = "oidc"
)

type Privacy string

const (
	PrivacyPublic  Privacy = "public"
	PrivacyPrivate Privacy = "private"
	PrivacyUsers   Privacy = "users"
)

type Layout string

const (
	LayoutRows    Layout = "rows"
	LayoutColumns Layout = "columns"
	LayoutMasonry Layout = "masonry"
)

type Width string

const (
	WidthDefault Width = "default"
	WidthWide    Width = "wide"
)

type ItemSize string

const (
	Size1x1 ItemSize = "1x1"
	Size1x2 ItemSize = "1x2"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           string     `bun:"id,pk,type:text" json:"id"`
	Username     string     `bun:"username,unique,notnull" json:"username"`
	PasswordHash *string    `bun:"password_hash" json:"-"`
	Role         Role       `bun:"role,notnull,default:'user'" json:"role"`
	OIDCSubject  *string    `bun:"oidc_subject" json:"-"`
	OIDCIssuer   *string    `bun:"oidc_issuer" json:"-"`
	AuthMethod   AuthMethod `bun:"-" json:"authMethod,omitempty"`
}

type Dashboard struct {
	bun.BaseModel `bun:"table:dashboards,alias:d"`

	ID          string  `bun:"id,pk,type:text" json:"id"`
	OwnerID     string  `bun:"owner_id,notnull" json:"ownerId"`
	Name        string  `bun:"name,notnull" json:"name"`
	Slug        string  `bun:"slug,unique,notnull" json:"slug"`
	Description string  `bun:"description,notnull,default:''" json:"description"`
	Icon        string  `bun:"icon,notnull,default:''" json:"icon"`
	IconDark    string  `bun:"icon_dark,notnull,default:''" json:"iconDark"`
	Layout      Layout  `bun:"layout,notnull,default:'rows'" json:"layout"`
	Width       Width   `bun:"width,notnull,default:'default'" json:"width"`
	Privacy     Privacy `bun:"privacy,notnull,default:'private'" json:"privacy"`
	CleanMode   bool    `bun:"clean_mode,notnull,default:false" json:"cleanMode"`
	IsMain      bool    `bun:"is_main,notnull,default:false" json:"isMain"`
	IsDefault   bool    `bun:"is_default,notnull,default:false" json:"isDefault"`

	Owner  *User    `bun:"rel:belongs-to,join:owner_id=id" json:"owner,omitempty"`
	Groups []*Group `bun:"rel:has-many,join:id=dashboard_id" json:"groups,omitempty"`
}

type Group struct {
	bun.BaseModel `bun:"table:groups,alias:g"`

	ID          string   `bun:"id,pk,type:text" json:"id"`
	DashboardID string   `bun:"dashboard_id,notnull" json:"dashboardId"`
	Title       string   `bun:"title,notnull" json:"title"`
	Description string   `bun:"description,notnull,default:''" json:"description"`
	Icon        string   `bun:"icon,notnull,default:''" json:"icon"`
	IconDark    string   `bun:"icon_dark,notnull,default:''" json:"iconDark"`
	ItemSize    ItemSize `bun:"item_size,notnull,default:'1x1'" json:"itemSize"`
	Position    int      `bun:"position,notnull,default:0" json:"position"`

	Items []*Item `bun:"rel:has-many,join:id=group_id" json:"items,omitempty"`
}

type Item struct {
	bun.BaseModel `bun:"table:items,alias:i"`

	ID           string `bun:"id,pk,type:text" json:"id"`
	GroupID      string `bun:"group_id,notnull" json:"groupId"`
	Title        string `bun:"title,notnull" json:"title"`
	Description  string `bun:"description" json:"description"`
	URL          string `bun:"url,notnull" json:"url"`
	Icon         string `bun:"icon,notnull" json:"icon"`
	IconDark     string `bun:"icon_dark,notnull,default:''" json:"iconDark"`
	PingEnabled  bool   `bun:"ping_enabled,notnull,default:false" json:"pingEnabled"`
	PingOnlyDown bool   `bun:"ping_only_down,notnull,default:false" json:"pingOnlyDown"`
	Position     int    `bun:"position,notnull,default:0" json:"position"`
}

type UploadedIcon struct {
	bun.BaseModel `bun:"table:uploaded_icons,alias:ui"`

	ID       string `bun:"id,pk,type:text" json:"id"`
	Filename string `bun:"filename,notnull" json:"filename"`
	Mime     string `bun:"mime,notnull" json:"mime"`
	OwnerID  string `bun:"owner_id,notnull" json:"ownerId"`
}
