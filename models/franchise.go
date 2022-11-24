package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Franchise struct {
	ID             uuid.UUID `json:"id" gorm:"primaryKey"`
	FranchiseName  string    `json:"franchiseName" gorm:"type:string"`
	FoundationYear string    `json:"foundationYear" gorm:"type:string"`
	LeagueID       uuid.UUID `json:"leagueId" gorm:"foreignKey:FranchiseID"`
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
