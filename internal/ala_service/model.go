package ala_service

import (
	"time"

	"gorm.io/gorm"
)

// Alias represents a short-URL → long-URL mapping.
type Alias struct {
	AliasID                 uint           `gorm:"primaryKey;autoIncrement;column:alias_id" json:"alias_id"`
	AliasURL                string         `gorm:"uniqueIndex;column:alias_url;size:64" json:"alias_url"`
	RedirectURI             string         `gorm:"column:redirect_uri;type:text" json:"redirect_uri"`
	CreatedDateTimeUTC      time.Time      `gorm:"column:created_datetime_utc;not null;autoCreateTime" json:"created_datetime_utc"`
	LastModifiedDateTimeUTC *time.Time     `gorm:"column:last_modified_datetime_utc" json:"last_modified_datetime_utc,omitempty"`
	LastUsedDateTimeUTC     *time.Time     `gorm:"column:last_used_datetime_utc" json:"last_used_datetime_utc,omitempty"`
	DeletedAt               gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

// TableName overrides the default pluralized table name.
func (Alias) TableName() string {
	return "aliases"
}
