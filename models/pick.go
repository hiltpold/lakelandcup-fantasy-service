package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Pick struct {
	ID               uuid.UUID  `json:"id" gorm:"primaryKey"`
	DraftYear        string     `json:"draftYear" gorm:"type:integer"`
	DraftRound       string     `json:"draftRound" gorm:"type:integer"`
	DraftPickOverall *string    `json:"draftPickOverall" gorm:"type:integer;default:null"`
	DraftPickInRound *string    `json:"draftPickInRound" gorm:"type:integer;default:null"`
	ProspectID       *uuid.UUID `json:"prospectID" gorm:"foreignKey:ProspectID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	OwnerID          *uuid.UUID `json:"ownerID" gorm:"foreignKey:FranchiseID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL"`
	OwnerName        string     `json:"ownerName"`
	LastOwnerID      *uuid.UUID `json:"lastOwnerID" gorm:"foreignKey:FranchiseID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL"`
	LastOwnerName    string     `json:"lastOwnerName"`
	OriginID         *uuid.UUID `json:"originID" gorm:"foreignKey:FranchiseID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL"`
	OriginName       string     `json:"originName"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        time.Time
}

func (pick *Pick) BeforeCreate(db *gorm.DB) error {
	pick.ID = uuid.New()
	pick.CreatedAt = time.Now().Local()
	return nil
}

func (pick *Pick) BeforeUpdate(db *gorm.DB) error {
	pick.UpdatedAt = time.Now().Local()
	return nil
}
