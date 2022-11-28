package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Prospect struct {
	ID               uuid.UUID `json:"id" gorm:"primaryKey"`
	FullName         string    `json:"fullName" gorm:"type:string"`
	FirstName        string    `json:"firstName" gorm:"type:string"`
	LastName         string    `json:"lastName" gorm:"type:string"`
	DraftYear        string    `json:"draftYear" gorm:"type:integer"`
	DraftRound       string    `json:"draftRound" gorm:"type:integer"`
	DraftPickOverall string    `json:"draftPickOverall" gorm:"type:integer"`
	DraftPickInRound string    `json:"draftPickInRound" gorm:"type:integer"`
	LeagueID         uuid.UUID `json:"leagueID" gorm:"foreignKey:ProspectID"`
	FranchiseID      uuid.UUID `json:"franchiseID" gorm:"foreignKey:ProspectID"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (prospect *Prospect) BeforeCreate(db *gorm.DB) error {
	prospect.ID = uuid.New()
	prospect.CreatedAt = time.Now().Local()
	return nil
}

func (prospect *Prospect) BeforeUpdate(db *gorm.DB) error {
	prospect.UpdatedAt = time.Now().Local()
	return nil
}
