package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type League struct {
	ID                uuid.UUID   `json:"leagueId" gorm:"primaryKey"`
	Name              string      `json:"name" gorm:"not null;type:string"`
	Admin             string      `json:"admin" gorm:"not null;type:string"`
	AdminID           uuid.UUID   `json:"userId" gorm:"not null;type:uuid"`
	Commissioner      string      `json:"commissioner" gorm:"not null;type:string"`
	CommissionerID    uuid.UUID   `json:"commissionerID" gorm:"not null;type:uuid"`
	FoundationYear    string      `json:"foundationYear" gorm:"not null;type:string"`
	MaxFranchises     int         `json:"maxFranchise" gorm:"not null;type:int"`
	MaxProspects      int         `json:"maxProspects" gorm:"not null;type:int"`
	DraftRightsGoalie int         `json:"DraftRightsGoalie" gorm:"not null;type:int"`
	DraftRightsSkater int         `json:"draftRightsSkater" gorm:"not null;type:int"`
	Franchises        []Franchise `json:"franchises" gorm:"foreignKey:LeagueID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Prospects         []Prospect  `json:"prospects" gorm:"foreignKey:LeagueID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
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
