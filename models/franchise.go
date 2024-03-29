package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Franchise struct {
	ID             uuid.UUID  `json:"id" gorm:"primaryKey"`
	Name           string     `json:"name" gorm:"not null;type:string"`
	UserID         uuid.UUID  `json:"userId" gorm:"not null;type:uuid;"`
	UserName       string     `json:"userName" gorm:"not null;type:string;"`
	FoundationYear string     `json:"foundationYear" gorm:"not null;type:string"`
	LeagueID       uuid.UUID  `json:"leagueId" gorm:"not null;foreignKey:FranchiseID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL"`
	Prospects      []Prospect `json:"prospects" gorm:"foreignKey:FranchiseID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL"`
	Picks          []Pick     `json:"picks" gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE;OnDelete:SET NULL"`
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
