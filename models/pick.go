package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Pick struct {
	ID               uuid.UUID `json:"id" gorm:"primaryKey"`
	DraftYear        string    `json:"draftYear" gorm:"type:integer"`
	DraftRound       string    `json:"draftRound" gorm:"type:integer"`
	DraftPickOverall string    `json:"draftPickOverall" gorm:"type:integer"`
	DraftPickInRound string    `json:"draftPickInRound" gorm:"type:integer"`
	ProspectID       uuid.UUID `json:"prospectID" gorm:"foreignKey:PickID"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
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
