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

func (s *Server) GetProspectsByFranchise(ctx context.Context, req *pb.GetFranchiseRequest) (*pb.ProspectsResponse, error) {
	var prospects []models.Prospect

	fId, err := uuid.Parse(req.FranchiseID)
	if err != nil {
		return &pb.ProspectsResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Could not parse uuid for franchise id %q.", fId)}, nil

	}

	findProspects := s.R.DB.Preload("Pick").Where(models.Prospect{FranchiseID: &fId}).Find(&prospects).Limit(1000)

	if findProspects.Error != nil {
		return &pb.ProspectsResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Could not fetch picks for franchise %q. Error: %v", fId, findProspects.Error),
		}, nil
	}
	logrus.Info(fmt.Sprintf("-> %+v", prospects))

	prospectsRes := []*pb.Prospect{}
	if findProspects.RowsAffected > 0 {
		for _, p := range prospects {
			pickId := ""
			if p.Pick != nil {
				pickId = p.Pick.ID.String()
			}
			prospectId := ""
			if p.FranchiseID != nil {
				prospectId = p.FranchiseID.String()
			}
			logrus.Info(p.Pick)
			prospectsRes = append(prospectsRes, &pb.Prospect{
				ID:          p.ID.String(),
				FullName:    p.FullName,
				FirstName:   p.FirstName,
				LastName:    p.LastName,
				NhlTeam:     p.NhlTeam,
				FranchiseID: prospectId,
				Pick:        &pb.Pick{ID: pickId, DraftYear: p.Pick.DraftYear, DraftRound: p.Pick.DraftRound, DraftPickInRound: p.Pick.DraftPickInRound, DraftPickOverall: p.Pick.DraftPickOverall},
				Birthdate:   p.Birthdate,
			})
		}

		logrus.Info(fmt.Sprintf("-> %+v", prospectsRes))

	}

	return &pb.ProspectsResponse{
		Status:    http.StatusOK,
		Prospects: prospectsRes,
	}, nil

}

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

func (s *Server) CreateProspectsBulk(ctx context.Context, req *pb.CreateProspectsBulkRequest) (*pb.CreateProspectsBulkResponse, error) {

	prospects := []models.Prospect{}

	for _, pReq := range req.Prospects {
		var prospect models.Prospect
		if findProspect := s.R.DB.Where(&models.Prospect{FullName: pReq.FullName, Birthdate: pReq.Birthdate, NhlDraftYear: pReq.DraftYear, NhlDraftPickOverall: pReq.NhlDraftPickOverall}).First(&prospect); findProspect.Error != nil {
			prospect.FullName = pReq.FullName
			prospect.FirstName = pReq.FirstName
			prospect.LastName = pReq.LastName
			prospect.NhlTeam = pReq.NhlTeam
			prospect.Birthdate = pReq.Birthdate
			prospect.Height = pReq.Height
			prospect.Weight = pReq.Weight
			prospect.NhlDraftYear = pReq.DraftYear
			prospect.NhlDraftRound = pReq.NhlDraftRound
			prospect.NhlDraftPickInRound = pReq.NhlDraftPickInRound
			prospect.NhlDraftPickOverall = pReq.NhlDraftPickOverall
			prospect.PositionCode = pReq.PositionCode
			prospects = append(prospects, prospect)
		}
	}

	if len(prospects) > 0 {
		logrus.Info(fmt.Sprintf("Prospects that will be batch inserted: %v", len(prospects)))
		if createProspects := s.R.DB.Create(&prospects); createProspects.Error != nil {
			return &pb.CreateProspectsBulkResponse{
				Status: http.StatusForbidden,
				Error:  fmt.Sprintf("Creating prospects failed %q", createProspects.Error),
			}, nil
		}
	} else {
		logrus.Info(fmt.Sprintf("Prospects that will be batch inserted: %v", len(prospects)))
	}

	return &pb.CreateProspectsBulkResponse{
		Status:      http.StatusCreated,
		ProspectIds: []string{},
	}, nil
}

func (s *Server) TextSearchProspects(ctx context.Context, req *pb.TextSearchRequest) (*pb.ProspectsResponse, error) {

	rows, err := s.R.DB.Model(&models.Prospect{}).Preload("Picks").Raw(fmt.Sprintf("SELECT * FROM fantasy.prospects WHERE to_tsvector(full_name) @@ to_tsquery('%q')", req.Text)).Rows()
	if err != nil {
		return &pb.ProspectsResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Creating prospects failed %q", err),
		}, nil
	}
	defer rows.Close()

	prospectsRes := []*pb.Prospect{}

	for rows.Next() {
		p := models.Prospect{}
		s.R.DB.ScanRows(rows, &p)
		// TODO: avoid second query, better text search in first query
		s.R.DB.Model(&p).Preload("Pick").Where(&models.Prospect{ID: p.ID}).First(&p)
		pick := pb.Pick{}
		if p.Pick != nil {
			pick = pb.Pick{
				ID:               p.Pick.ID.String(),
				DraftYear:        p.Pick.DraftYear,
				DraftRound:       p.Pick.DraftRound,
				DraftPickInRound: p.Pick.DraftPickInRound,
				DraftPickOverall: p.Pick.DraftPickOverall,
				ProspectID:       p.Pick.ProspectID.String(),
			}
		}

		lId := ""
		fId := ""

		if p.LeagueID != nil {
			lId = p.LeagueID.String()
		}

		if p.FranchiseID != nil {
			fId = p.FranchiseID.String()
		}

		pp := &pb.Prospect{
			ID:                  p.ID.String(),
			FullName:            p.FullName,
			FirstName:           p.FirstName,
			LastName:            p.LastName,
			NhlTeam:             p.NhlTeam,
			Birthdate:           p.Birthdate,
			Height:              p.Height,
			Weight:              p.Weight,
			PositionCode:        p.PositionCode,
			NhlDraftYear:        p.NhlDraftYear,
			NhlDraftRound:       p.NhlDraftRound,
			NhlPickInRound:      p.NhlDraftPickInRound,
			NhlDraftPickOverall: p.NhlDraftPickOverall,
			LeagueID:            lId,
			FranchiseID:         fId,
			Pick:                &pick,
		}
		prospectsRes = append(prospectsRes, pp)
	}

	return &pb.ProspectsResponse{
		Status:    http.StatusOK,
		Prospects: prospectsRes,
	}, nil
}
