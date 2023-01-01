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
	Birthdate   string     `json:"birthdate" gorm:"not null;type:string"`
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
