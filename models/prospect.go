package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Prospect struct {
	ID          uuid.UUID  `json:"id" gorm:"primaryKey"`
	FullName    string     `json:"fullName" gorm:"not null;type:string"`
	FirstName   string     `json:"firstName" gorm:"not null;type:string"`
	LastName    string     `json:"lastName" gorm:"not null;type:string"`
	NhlTeam    string     `json:"nhlTeamName" gorm:"not null;type:string"`
	Birthdate   string     `json:"birthdate" gorm:"not null;type:string"`
	Height   string     `json:"height" gorm:"not null;type:string"`
	Weight   string     `json:"weight" gorm:"not null;type:string"`
	DraftYear   string     `json:"nhlYear" gorm:"not null;type:string"`
	NhlDraftRound   string     `json:"nhlDraftRound" gorm:"not null;type:string"`
	NhlDraftPickOverall   string     `json:"nhlDraftPickOverall" gorm:"not null;type:string"`
	NhlDraftPickInRound   string     `json:"nhlDraftPickInRound" gorm:"not null;type:string"`
	PositionCode   string     `json:"positionCode" gorm:"not null;type:string"`
	LeagueID    *uuid.UUID `json:"leagueID" gorm:"foreignKey:ProspectID"`
	FranchiseID *uuid.UUID `json:"franchiseID" gorm:"foreignKey:ProspectID"`
	Pick        Pick       `json:"pick" gorm:"foreignKey:ProspectID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
