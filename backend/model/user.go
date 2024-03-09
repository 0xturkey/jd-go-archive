package model

import (
	guuid "github.com/google/uuid"
)

type User struct {
	ID                guuid.UUID `gorm:"primaryKey" json:"-"`
	Email             string     `json:"email"`
	TkID              string     `json:"tk_id"`
	SuborganizationID string     `json:"suborganization_id"`
	Tasks             []Task     `gorm:"foreignKey:UserRefer; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;" json:"-"`
	Wallets           []Wallet   `gorm:"foreignKey:UserRefer; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;" json:"-"`
	CreatedAt         int64      `gorm:"autoCreateTime" json:"-" `
	UpdatedAt         int64      `gorm:"autoUpdateTime:milli" json:"-"`
}
