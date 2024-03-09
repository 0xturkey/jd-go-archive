package model

import (
	"time"

	guuid "github.com/google/uuid"
)

type TaskStatus string

const (
	Scheduled  TaskStatus = "SCHEDULED"
	InProgress TaskStatus = "IN_PROGRESS"
	Completed  TaskStatus = "COMPLETED"
	Fail       TaskStatus = "FAIL"
	Unknown    TaskStatus = "UNKNOWN"
)

type Network string

const (
	EVM    Network = "EVM"
	COSMOS Network = "COSMOS"
	SVM    Network = "SVM"
)

type TaskType string

const (
	EthTx TaskType = "ETH_TX"
)

type Task struct {
	ID                 guuid.UUID  `gorm:"primaryKey" json:"id"`
	UserRefer          guuid.UUID  `json:"-"`
	WalletRefer        guuid.UUID  `json:"-"`
	TaskStatus         TaskStatus  `gorm:"type:varchar(100)" json:"task_status"`
	TaskType           TaskType    `gorm:"type:varchar(100)" json:"task_type"`
	ScheduledAt        time.Time   `json:"scheduled_at"`
	RpcUrl             string      `json:"rpc_url"`
	EncodedTransaction string      `json:"encoded_transaction"`
	TransactionID      *string     `json:"transaction_id" gorm:"default:null"`
	DependentTaskID    *guuid.UUID `gorm:"type:uuid;default:null"`
	DependentTask      *Task       `gorm:"foreignKey:DependentTaskID;references:ID;constraint:OnUpdate:CASCADE, OnDelete:SET NULL;" json:"-"`
	Dependencies       []Task      `gorm:"foreignKey:DependentTaskID"`
	CreatedAt          int64       `gorm:"autoCreateTime" json:"-" `
	UpdatedAt          int64       `gorm:"autoUpdateTime:milli" json:"-"`
}
