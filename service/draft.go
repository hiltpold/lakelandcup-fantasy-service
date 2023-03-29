package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hiltpold/lakelandcup-fantasy-service/models"
	"github.com/hiltpold/lakelandcup-fantasy-service/service/pb"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (s *Server) UndraftProspect(ctx context.Context, req *pb.DraftRequest) (*pb.DefaultResponse, error) {
	var pick models.Pick
	var prospect models.Prospect
	var transaction = s.R.DB.Transaction(func(tx *gorm.DB) error {

		// parse id to uuid
		pickId, err := uuid.Parse(req.PickID)
		if err != nil {
			return fmt.Errorf("could not parse PickID %v", req.PickID)
		}

		prospectId, err := uuid.Parse(req.ProspectID)
		if err != nil {
			return fmt.Errorf("could not parse ProspectID%v", req.ProspectID)
		}

		// pick
		findPick := tx.Model(&pick).Where(&models.Pick{ID: pickId}).First(&pick)
		logrus.Info(fmt.Sprintf("%v", pick))
		if findPick.Error != nil {
			return findPick.Error
		}

		if findPick.RowsAffected == 0 {
			return fmt.Errorf("could not find any pick with ID %v", req.PickID)
		}

		if pick.ProspectID == nil {
			return fmt.Errorf("nothing to delete for pick with ID %v no prospect assigned", pickId)
		}

		if *pick.ProspectID != prospectId {
			return fmt.Errorf("nothing to delete for pick with ID %v. pick was never assigned to prospect %v", pickId, prospectId)

		}

		// prospect
		findProspect := tx.Model(&prospect).Preload("Pick").Where(&models.Prospect{ID: prospectId}).First(&prospect)
		logrus.Info(fmt.Sprintf("%+v", prospect))

		if findProspect.Error != nil {
			return findProspect.Error
		}

		if prospect.Pick == nil {
			return fmt.Errorf("nothing to delete for prospect with ID %v no pick assigned", prospectId)
		}
		if prospect.Pick.ID != pickId {
			return fmt.Errorf("nothing to delete for prospect with ID %v. prospect was not picked with pick %v", prospectId, pickId)

		}

		// update both
		/*
			// this syntax does not work
			updateProspect := tx.Model(&models.Prospect{}).Where(&models.Prospect{ID: prospectId}).Updates(models.Prospect{LeagueID: nil, FranchiseID: nil, Pick: nil})
			if updateProspect.Error != nil {
				return updateProspect.Error
			}
			updatePick := tx.Model(&models.Pick{}).Where(&models.Pick{ID: pickId}).Updates(models.Pick{ProspectID: nil})
			if updatePick.Error != nil {
				return updatePick.Error
			}
		*/

		prospect.LeagueID = nil
		prospect.FranchiseID = nil
		prospect.Pick = nil

		tx.Save(&prospect)

		pick.ProspectID = nil

		tx.Save(&pick)

		// return nil will commit the whole transaction
		return nil
	})

	if transaction != nil {
		return &pb.DefaultResponse{
			Status: http.StatusConflict,
			Error:  transaction.Error(),
		}, nil

	}

	return &pb.DefaultResponse{
		Status:  http.StatusOK,
		Message: "prospect was successfully undrafted",
	}, nil
}

func (s *Server) DraftProspect(ctx context.Context, req *pb.DraftRequest) (*pb.DefaultResponse, error) {
	var pick models.Pick
	var prospect models.Prospect
	var transaction = s.R.DB.Transaction(func(tx *gorm.DB) error {

		// parse id to uuid
		pickId, err := uuid.Parse(req.PickID)
		if err != nil {
			return fmt.Errorf("could not parse PickID %v", req.PickID)
		}

		prospectId, err := uuid.Parse(req.ProspectID)
		if err != nil {
			return fmt.Errorf("could not parse ProspectID%v", req.ProspectID)
		}

		franchiseId, err := uuid.Parse(req.FranchiseID)
		if err != nil {
			return fmt.Errorf("could not parse FranchisetID%v", req.FranchiseID)
		}

		leagueId, err := uuid.Parse(req.LeagueID)
		if err != nil {
			return fmt.Errorf("could not parse LeaguetID%v", req.LeagueID)
		}

		// pick
		findPick := tx.Model(&pick).Where(&models.Pick{ID: pickId}).First(&pick)
		logrus.Info(fmt.Sprintf("%v", pick))
		if findPick.Error != nil {
			return findPick.Error
		}

		if findPick.RowsAffected == 0 {
			return fmt.Errorf("could not find any pick with ID %v", req.PickID)
		}

		if pick.ProspectID != nil {
			return fmt.Errorf("pick with ID %v is already assigned to prospect %v", pickId, prospectId)
		}

		// prospect
		findProspect := tx.Model(&prospect).Preload("Pick").Where(&models.Prospect{ID: prospectId}).First(&prospect)
		logrus.Info(fmt.Sprintf("%+v", prospect))
		if findProspect.Error != nil {
			return findProspect.Error
		}

		if prospect.Pick != nil {
			return fmt.Errorf("prospect with ID %v is already assigned to pick %v", prospectId, prospect.Pick.ID)
		}

		// update both

		updateProspect := tx.Model(&prospect).Where(&models.Prospect{ID: prospectId}).Updates(models.Prospect{LeagueID: &leagueId, FranchiseID: &franchiseId})
		if updateProspect.Error != nil {
			return updateProspect.Error
		}

		updatePick := tx.Model(&pick).Where(&models.Pick{ID: pickId}).Update("prospect_id", prospectId)
		if updatePick.Error != nil {
			return updatePick.Error
		}

		// return nil will commit the whole transaction
		return nil
	})

	if transaction != nil {
		return &pb.DefaultResponse{
			Status: http.StatusConflict,
			Error:  transaction.Error(),
		}, nil

	}

	return &pb.DefaultResponse{
		Status:  http.StatusOK,
		Message: "prospect was successfully drafted",
	}, nil

}
