package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type League struct {
	ID             uuid.UUID   `json:"leagueId" gorm:"primaryKey"`
	LeagueFounder  uuid.UUID   `json:"userId" gorm:"not null;type:uuid"`
	LeagueName     string      `json:"leagueName" gorm:"not null;type:string"`
	FoundationYear string      `json:"foundationYear" gorm:"not null;type:string"`
	MaxFranchises  int         `json:"maxFranchise" gorm:"not null;maxFranchises:int"`
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
