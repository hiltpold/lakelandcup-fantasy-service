package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type League struct {
	ID             uuid.UUID   `json:"leagueId" gorm:"primaryKey"`
	UserID         string      `json:"userId" gorm:"type:uuid"`
	LeagueName     string      `json:"leagueName" gorm:"type:string"`
	FoundationYear string      `json:"foundationYear" gorm:"type:string"`
	Franchises     []Franchise `json:"franchises" gorm:"foreignKey:LeagueID"`
	Prospects      []Prospect  `json:"prospects" gorm:"foreignKey:LeagueID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (league *League) BeforeCreate(db *gorm.DB) error {
	league.ID = uuid.New()
	league.CreatedAt = time.Now().Local()
	return nil
}

func (league *League) BeforeUpdate(db *gorm.DB) error {
	league.UpdatedAt = time.Now().Local()
	return nil
}
