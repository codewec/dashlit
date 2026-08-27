package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
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

	ID           string    `bun:"id,pk,type:text" json:"id"`
	Username     string    `bun:"username,unique,notnull" json:"username"`
	PasswordHash *string   `bun:"password_hash" json:"-"`
	Role         Role      `bun:"role,notnull,default:'user'" json:"role"`
	OIDCSubject  *string   `bun:"oidc_subject" json:"-"`
	OIDCIssuer   *string   `bun:"oidc_issuer" json:"-"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updatedAt"`
}

type ThemeColors struct {
	Background        string `json:"background,omitempty"`
	Surface           string `json:"surface,omitempty"`
	SurfaceHover      string `json:"surfaceHover,omitempty"`
	Primary           string `json:"primary,omitempty"`
	PrimaryForeground string `json:"primaryForeground,omitempty"`
	Accent            string `json:"accent,omitempty"`
	Text              string `json:"text,omitempty"`
	TextMuted         string `json:"textMuted,omitempty"`
	Border            string `json:"border,omitempty"`
}

type DashboardTheme struct {
	Mode   string `json:"mode"`
	Colors *struct {
		Light *ThemeColors `json:"light,omitempty"`
		Dark  *ThemeColors `json:"dark,omitempty"`
	} `json:"colors,omitempty"`
}

type Dashboard struct {
	bun.BaseModel `bun:"table:dashboards,alias:d"`

	ID          string          `bun:"id,pk,type:text" json:"id"`
	OwnerID     string          `bun:"owner_id,notnull" json:"ownerId"`
	Name        string          `bun:"name,notnull" json:"name"`
	Slug        string          `bun:"slug,unique,notnull" json:"slug"`
	Description string          `bun:"description,notnull,default:''" json:"description"`
	Icon        string          `bun:"icon,notnull,default:''" json:"icon"`
	IconDark    string          `bun:"icon_dark,notnull,default:''" json:"iconDark"`
	Layout      Layout          `bun:"layout,notnull,default:'rows'" json:"layout"`
	Width       Width           `bun:"width,notnull,default:'default'" json:"width"`
	Privacy     Privacy         `bun:"privacy,notnull,default:'private'" json:"privacy"`
	CleanMode   bool            `bun:"clean_mode,notnull,default:false" json:"cleanMode"`
	IsMain      bool            `bun:"is_main,notnull,default:false" json:"isMain"`
	Theme       *DashboardTheme `bun:"theme,type:json" json:"theme,omitempty"`
	CreatedAt   time.Time       `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt   time.Time       `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updatedAt"`

	Owner  *User    `bun:"rel:belongs-to,join:owner_id=id" json:"owner,omitempty"`
	Groups []*Group `bun:"rel:has-many,join:id=dashboard_id" json:"groups,omitempty"`
}

type Group struct {
	bun.BaseModel `bun:"table:groups,alias:g"`

	ID          string    `bun:"id,pk,type:text" json:"id"`
	DashboardID string    `bun:"dashboard_id,notnull" json:"dashboardId"`
	Title       string    `bun:"title,notnull" json:"title"`
	Description string    `bun:"description,notnull,default:''" json:"description"`
	Icon        string    `bun:"icon,notnull,default:''" json:"icon"`
	IconDark    string    `bun:"icon_dark,notnull,default:''" json:"iconDark"`
	ItemSize    ItemSize  `bun:"item_size,notnull,default:'1x1'" json:"itemSize"`
	Position    int       `bun:"position,notnull,default:0" json:"position"`
	Collapsed   bool      `bun:"collapsed,notnull,default:false" json:"collapsed"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updatedAt"`

	Items []*Item `bun:"rel:has-many,join:id=group_id" json:"items,omitempty"`
}

type Item struct {
	bun.BaseModel `bun:"table:items,alias:i"`

	ID          string    `bun:"id,pk,type:text" json:"id"`
	GroupID     string    `bun:"group_id,notnull" json:"groupId"`
	Title       string    `bun:"title,notnull" json:"title"`
	Description string    `bun:"description" json:"description"`
	URL         string    `bun:"url,notnull" json:"url"`
	Icon        string    `bun:"icon,notnull" json:"icon"`
	IconDark    string    `bun:"icon_dark,notnull,default:''" json:"iconDark"`
	Position    int       `bun:"position,notnull,default:0" json:"position"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updatedAt"`
}

type UploadedIcon struct {
	bun.BaseModel `bun:"table:uploaded_icons,alias:ui"`

	ID           string    `bun:"id,pk,type:text" json:"id"`
	Filename     string    `bun:"filename,notnull" json:"filename"`
	OriginalName string    `bun:"original_name,notnull" json:"originalName"`
	Mime         string    `bun:"mime,notnull" json:"mime"`
	Size         int64     `bun:"size,notnull" json:"size"`
	OwnerID      string    `bun:"owner_id,notnull" json:"ownerId"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"createdAt"`
}
