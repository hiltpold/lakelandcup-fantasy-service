package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Franchise struct {
	ID             uuid.UUID  `json:"id" gorm:"primaryKey"`
	FranchiseOwner uuid.UUID  `json:"franchiseOwner" gorm:"not null;type:uuid;"`
	FranchiseName  string     `json:"franchiseName" gorm:"not null;type:string"`
	FoundationYear string     `json:"foundationYear" gorm:"not null;type:string"`
	LeagueID       uuid.UUID  `json:"leagueId" gorm:"not null;foreignKey:FranchiseID"`
	Prospects      []Prospect `json:"prospects" gorm:"not null;foreignKey:FranchiseID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (franchise *Franchise) BeforeCreate(db *gorm.DB) error {
	franchise.ID = uuid.New()
	franchise.CreatedAt = time.Now().Local()
	return nil
}

func (franchise *Franchise) BeforeUpdate(db *gorm.DB) error {
	franchise.UpdatedAt = time.Now().Local()
	return nil
}
