package model

import (
	guuid "github.com/google/uuid"
)

type Wallet struct {
	ID        guuid.UUID `gorm:"primaryKey" json:"id"`
	TurnkeyID string     `json:"turnkey_id"`
	UserRefer guuid.UUID `json:"-"`
	User      *User      `gorm:"foreignKey:UserRefer"`
	Tasks     []Task     `gorm:"foreignKey:WalletRefer; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;" json:"-"`
	CreatedAt int64      `gorm:"autoCreateTime" json:"-" `
	UpdatedAt int64      `gorm:"autoUpdateTime:milli" json:"-"`
}
