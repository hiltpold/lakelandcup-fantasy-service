package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hiltpold/lakelandcup-fantasy-service/models"
	"github.com/hiltpold/lakelandcup-fantasy-service/service/pb"
	"github.com/hiltpold/lakelandcup-fantasy-service/storage"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Server struct {
	R storage.Repository
	// https://github.com/grpc/grpc-go/issues/3794:
	pb.UnimplementedFantasyServiceServer
}

func (s *Server) CreateLeague(ctx context.Context, req *pb.LeagueRequest) (*pb.LeagueResponse, error) {
	var league models.League
	const leagueName = "Lakelandcup"
	// only lakelandcup league can be created
	if req.Name != leagueName {
		return &pb.LeagueResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Wrong league name. Only league %q can be created.", leagueName),
		}, nil
	}

	// check if league already exists
	if findLeague := s.R.DB.Where(&models.League{Name: req.Name}).First(&league); findLeague.Error == nil {
		return &pb.LeagueResponse{
			Status: http.StatusConflict,
			Error:  "League already exists",
		}, nil
	}

	// TODO: Error Handling MustParse

	league.Name = req.Name
	league.Admin = req.Admin
	league.AdminID = uuid.MustParse(req.AdminID)
	league.Commissioner = req.Commissioner
	league.CommissionerID = uuid.MustParse(req.CommissionerID)
	league.FoundationYear = req.FoundationYear
	league.MaxFranchises = int(req.MaxFranchises)
	league.MaxProspects = int(req.MaxProspects)
	league.DraftRightsGoalie = int(req.DraftRightsGoalie)
	league.DraftRightsSkater = int(req.DraftRightsSkater)
	league.DraftRounds = int(req.DraftRounds)
	league.Franchises = []models.Franchise{}

	if createLeague := s.R.DB.Create(&league); createLeague.Error != nil {
		return &pb.LeagueResponse{
			Status: http.StatusForbidden,
			Error:  "Creating new league failed",
		}, nil
	}

	return &pb.LeagueResponse{
		Status:   http.StatusCreated,
		LeagueId: league.ID.String(),
	}, nil
}

func (s *Server) UpdateLeague(ctx context.Context, req *pb.LeagueUpdateRequest) (*pb.LeagueResponse, error) {
	var league models.League
	const leagueName = "Lakelandcup"

	// check if league already exists
	if findLeague := s.R.DB.Where(&models.League{ID: uuid.MustParse(req.Id), Name: leagueName}).First(&league); findLeague.Error != nil {
		return &pb.LeagueResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("League does't exists or is not called %q", leagueName),
		}, nil
	}

	league.ID = uuid.MustParse(req.Id)
	league.Name = req.League.Name
	league.Admin = req.League.Admin
	league.AdminID = uuid.MustParse(req.League.AdminID)
	league.Commissioner = req.League.Commissioner
	league.CommissionerID = uuid.MustParse(req.League.CommissionerID)
	league.FoundationYear = req.League.FoundationYear
	league.MaxFranchises = int(req.League.MaxFranchises)
	league.MaxProspects = int(req.League.MaxProspects)
	league.DraftRightsGoalie = int(req.League.DraftRightsGoalie)
	league.DraftRightsSkater = int(req.League.DraftRightsSkater)
	league.DraftRounds = int(req.League.DraftRounds)
	league.Franchises = []models.Franchise{}

	if updateLeague := s.R.DB.Save(&league); updateLeague.Error != nil {
		return &pb.LeagueResponse{
			Status: http.StatusForbidden,
			Error:  "Updating league failed",
		}, nil
	}

	return &pb.LeagueResponse{
		Status:   http.StatusCreated,
		LeagueId: league.ID.String(),
	}, nil
}

func (s *Server) GetLeague(ctx context.Context, req *pb.GetLeagueRequest) (*pb.GetLeagueResponse, error) {
	var league models.League
	var leagueRes *pb.League
	var franchisesRes []*pb.Franchise

	findLeague := s.R.DB.Preload("Franchises").First(&league, "id = ?", req.LeagueId)

	if findLeague.Error != nil {
		return &pb.GetLeagueResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Getting LeagueID (%s) failed", req.LeagueId),
		}, nil
	}

	if findLeague.RowsAffected == 0 {
		return &pb.GetLeagueResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("LeagueID (%s) does not exist", req.LeagueId),
		}, nil

	}

	tmpFranchise := pb.Franchise{}
	if len(league.Franchises) > 0 {
		for _, f := range league.Franchises {
			tmpFranchise.ID = f.ID.String()
			tmpFranchise.OwnerID = f.UserID.String()
			tmpFranchise.OwnerName = f.UserName
			tmpFranchise.Name = f.Name
			tmpFranchise.FoundationYear = f.FoundationYear
			franchisesRes = append(franchisesRes, &tmpFranchise)

		}
	} else {
		franchisesRes = append(franchisesRes, &pb.Franchise{})
	}

	leagueRes = &pb.League{
		ID:                league.ID.String(),
		Name:              league.Name,
		Admin:             league.Admin,
		AdminID:           league.AdminID.String(),
		Commissioner:      league.Commissioner,
		CommissionerID:    league.CommissionerID.String(),
		FoundationYear:    league.FoundationYear,
		MaxFranchises:     int32(league.MaxFranchises),
		MaxProspects:      int32(league.MaxProspects),
		DraftRightsGoalie: int32(league.DraftRightsGoalie),
		DraftRightsSkater: int32(league.DraftRightsSkater),
		Franchises:        franchisesRes,
	}

	return &pb.GetLeagueResponse{
		Status: http.StatusAccepted,
		Result: leagueRes,
	}, nil

}

func (s *Server) GetLeagues(ctx context.Context, req *pb.GetLeaguesRequest) (*pb.GetLeaguesResponse, error) {
	var leagues []models.League
	var leagueRes []*pb.League
	var franchisesRes []*pb.Franchise

	s.R.DB.Preload("Franchises").Find(&leagues).Limit(1000)

	for _, l := range leagues {
		if len(l.Franchises) > 0 {
			for _, f := range l.Franchises {
				tmpFranchise := pb.Franchise{}
				tmpFranchise.ID = f.ID.String()
				tmpFranchise.OwnerID = f.UserID.String()
				tmpFranchise.OwnerName = f.UserName
				tmpFranchise.Name = f.Name
				tmpFranchise.FoundationYear = f.FoundationYear
				franchisesRes = append(franchisesRes, &tmpFranchise)

			}
		} else {
			franchisesRes = append(franchisesRes, &pb.Franchise{})
		}
		tmpLeague := pb.League{
			ID:                l.ID.String(),
			Name:              l.Name,
			Admin:             l.Admin,
			AdminID:           l.AdminID.String(),
			Commissioner:      l.Commissioner,
			CommissionerID:    l.CommissionerID.String(),
			FoundationYear:    l.FoundationYear,
			MaxFranchises:     int32(l.MaxFranchises),
			MaxProspects:      int32(l.MaxProspects),
			DraftRightsGoalie: int32(l.DraftRightsGoalie),
			DraftRightsSkater: int32(l.DraftRightsSkater),
			Franchises:        franchisesRes,
		}
		leagueRes = append(leagueRes, &tmpLeague)

	}
	logrus.Info(leagueRes)
	return &pb.GetLeaguesResponse{
		Status: http.StatusAccepted,
		Result: leagueRes,
	}, nil

}

func (s *Server) CreateFranchise(ctx context.Context, req *pb.FranchiseRequest) (*pb.FranchiseResponse, error) {
	var league models.League
	var franchise models.Franchise
	// check if franchise exists in this league
	if findLeague := s.R.DB.Preload("Franchises").First(&league, "id = ?", uuid.MustParse(req.LeagueId)); findLeague.RowsAffected == 0 {
		return &pb.FranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Franchise cannot be created, provided leagueId (%s) does not exist", req.LeagueId),
		}, nil
	}

	// check if franchise name already taken in this league
	if findFranchise := s.R.DB.Where(&models.Franchise{LeagueID: uuid.MustParse(req.LeagueId), Name: req.Name}).First(&franchise); findFranchise.Error == nil {
		return &pb.FranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Franchise cannot be created, franchise with name (%s) already exisits in this league", req.Name),
		}, nil
	}

	// check if maximum franchises already satisfied
	if len(league.Franchises) >= league.MaxFranchises {
		return &pb.FranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Franchise cannot be created, maximum number of franchises already created (%d)", league.MaxFranchises),
		}, nil
	}

	franchise.Name = req.Name
	franchise.UserID = uuid.MustParse(req.OwnerID)
	franchise.UserName = req.OwnerName
	franchise.FoundationYear = req.FoundationYear
	franchise.LeagueID = uuid.MustParse(req.LeagueId)

	if createFranchise := s.R.DB.Create(&franchise); createFranchise.Error != nil {
		return &pb.FranchiseResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Creating new franchise for league (%s) failed: %v", req.LeagueId, createFranchise.Error),
		}, nil
	}

	return &pb.FranchiseResponse{
		Status:      http.StatusCreated,
		FranchiseId: franchise.ID.String(),
	}, nil
}

func (s *Server) GetFranchise(ctx context.Context, req *pb.GetFranchiseRequest) (*pb.GetFranchiseResponse, error) {
	var franchise models.Franchise
	var franchiseRes *pb.Franchise
	var prospectRes []*pb.Prospect

	findFranchise := s.R.DB.Preload("Prospects").First(&franchise, "id = ?", req.FranchiseID)

	if findFranchise.Error != nil {
		return &pb.GetFranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Getting franchiseId (%s) failed: %v", req.FranchiseID, findFranchise.Error),
		}, nil
	}

	if findFranchise.RowsAffected == 0 {
		return &pb.GetFranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("franchiseId (%s) does not exist", req.FranchiseID),
		}, nil

	}

	tmpProspect := pb.Prospect{}
	if len(franchise.Prospects) > 0 {
		for _, p := range franchise.Prospects {
			tmpProspect.ID = p.ID.String()
			tmpProspect.FullName = p.FullName
			tmpProspect.FirstName = p.FirstName
			tmpProspect.LastName = p.LastName
			tmpProspect.FranchiseID = p.FranchiseID.String()
		}
	} else {
		prospectRes = append(prospectRes, &pb.Prospect{})
	}
	franchiseRes = &pb.Franchise{
		ID:             franchise.ID.String(),
		OwnerID:        franchise.UserID.String(),
		OwnerName:      franchise.UserName,
		Name:           franchise.Name,
		FoundationYear: franchise.FoundationYear,
		Prospects:      prospectRes,
	}

	return &pb.GetFranchiseResponse{
		Status: http.StatusAccepted,
		Result: franchiseRes,
	}, nil

}

func (s *Server) GetLeagueFranchises(ctx context.Context, req *pb.GetLeagueRequest) (*pb.GetLeagueFranchisesResponse, error) {
	var franchises []models.Franchise
	var franchiseRes []*pb.Franchise
	var prospectRes []*pb.Prospect

	findFranchises := s.R.DB.Preload("Prospects").Find(&franchises, "league_id = ?", req.LeagueId).Limit(1000)

	if findFranchises.Error != nil {
		return &pb.GetLeagueFranchisesResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Getting franchises for league (%s) failed: %v", req.LeagueId, findFranchises.Error),
		}, nil
	}

	for _, f := range franchises {
		//tmpFranchise := &pb.Franchise{}
		if len(f.Prospects) > 0 {
			tmpProspect := &pb.Prospect{}
			for _, p := range f.Prospects {
				tmpProspect.ID = p.ID.String()
				tmpProspect.FullName = p.FullName
				tmpProspect.FirstName = p.FirstName
				tmpProspect.LastName = p.LastName
				tmpProspect.FranchiseID = p.FranchiseID.String()
				// append
				prospectRes = append(prospectRes, tmpProspect)
			}
		} else {
			prospectRes = append(prospectRes, &pb.Prospect{})
		}
		tmpFranchise := &pb.Franchise{
			ID:             f.ID.String(),
			OwnerID:        f.UserID.String(),
			OwnerName:      f.UserName,
			Name:           f.Name,
			FoundationYear: f.FoundationYear,
			Prospects:      prospectRes,
		}
		franchiseRes = append(franchiseRes, tmpFranchise)
	}
	return &pb.GetLeagueFranchisesResponse{
		Status: http.StatusAccepted,
		Result: franchiseRes}, nil

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

func (s *Server) TextSearchProspects(ctx context.Context, req *pb.TextSearchRequest) (*pb.TextSearchProspectsResponse, error) {

	rows, err := s.R.DB.Model(&models.Prospect{}).Preload("Picks").Raw(fmt.Sprintf("SELECT * FROM fantasy.prospects WHERE to_tsvector(full_name) @@ to_tsquery('%q')", req.Text)).Rows()
	if err != nil {
		return &pb.TextSearchProspectsResponse{
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

	return &pb.TextSearchProspectsResponse{
		Status:    http.StatusOK,
		Prospects: prospectsRes,
	}, nil
}

func (s *Server) Trade(ctx context.Context, req *pb.TradeRequest) (*pb.DefaultResponse, error) {
	/*
		var picks []models.Pick
		fromFranchiseID := uuid.MustParse(req.FromFranchiseID)
		toFranchiseID := uuid.MustParse(req.ToFranchiseID)
	*/

	var transaction = s.R.DB.Transaction(func(tx *gorm.DB) error {
		// handle picks
		var picks = []models.Pick{}
		for _, p := range req.Picks {

			pId, err := uuid.Parse(p.ProspectID)
			if err != nil {
				logrus.Error(fmt.Sprintf("Could not parse %v", p.ProspectID))
			}

			oId, err := uuid.Parse(p.OwnerID)
			if err != nil {
				logrus.Error(fmt.Sprintf("Could not parse %v", p.OwnerID))
			}

			loId, err := uuid.Parse(p.LastOwnerID)
			if err != nil {
				logrus.Error(fmt.Sprintf("Could not parse %v", p.LastOwnerID))
			}

			orId, err := uuid.Parse(p.OriginID)
			if err != nil {
				logrus.Error(fmt.Sprintf("Could not parse %v", p.OriginID))
			}

			picks = append(picks, models.Pick{
				DraftYear:        p.DraftYear,
				DraftRound:       p.DraftRound,
				DraftPickInRound: p.DraftPickInRound,
				DraftPickOverall: p.DraftPickOverall,
				ProspectID:       &pId,
				OwnerID:          &oId,
				OwnerName:        p.OriginName,
				LastOwnerID:      &loId,
				LastOwnerName:    p.LastOwnerName,
				OriginID:         &orId,
				OriginName:       p.OriginName,
			})

		}

		// Update columns to new value on `id` conflict
		createOrUpdatePicks := s.R.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "draft_year"}, {Name: "draft_round"}, {Name: "draft_pick_in_round"}, {Name: "draft_pick_overall"}, {Name: "origin_id"}, {Name: "origin_name"}}, // key colume
			DoUpdates: clause.AssignmentColumns([]string{"prospect_id", "owner_id", "owner_name", "last_owner_id", "last_owner_name"}),                                                       // column needed to be updated
		}).Create(&picks)

		if createOrUpdatePicks.Error != nil {
			return createOrUpdatePicks.Error
		}

		// handle prospects
		var prospects = []models.Prospect{}
		for _, p := range req.Prospects {
			lPtr := uuid.MustParse(p.LeagueID)
			fPtr := uuid.MustParse(p.FranchiseID)

			prospects = append(prospects, models.Prospect{
				FullName:            p.FullName,
				FirstName:           p.FirstName,
				LastName:            p.LastName,
				NhlTeam:             p.NhlTeam,
				Birthdate:           p.Birthdate,
				Height:              p.Height,
				Weight:              p.Weight,
				NhlDraftYear:        p.NhlDraftYear,
				NhlDraftRound:       p.NhlDraftRound,
				NhlDraftPickInRound: p.NhlPickInRound,
				NhlDraftPickOverall: p.NhlDraftPickOverall,
				PositionCode:        p.PositionCode,
				LeagueID:            &lPtr,
				FranchiseID:         &fPtr,
			})
		}

		// Update columns to new value on `id` conflict
		createOrUpdateProspects := s.R.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "full_name"}, {Name: "first_name"}, {Name: "last_name"}, {Name: "nhl_team"}, {Name: "birthdate"}, {Name: "league_id"}, {Name: "franchise_id"}}, // key colume
			DoUpdates: clause.AssignmentColumns([]string{"franchise_id"}),                                                                                                                    // column needed to be updated
		}).Create(&prospects)

		if createOrUpdateProspects.Error != nil {
			return createOrUpdatePicks.Error
		}

		// return nil will commit the whole transaction
		return nil
	})

	if transaction != nil {
		return &pb.DefaultResponse{
			Status: http.StatusOK,
			Error:  transaction.Error(),
		}, nil

	}

	return &pb.DefaultResponse{
		Status: http.StatusOK,
	}, nil

}

/*

func (s *Server) DraftProspect(ctx context.Context, req *pb.DraftProspectRequest) (*pb.DraftProspectResponse, error) {
	var pick models.Pick
	var pick2 models.Pick
	var prospect models.Prospect

	lId := uuid.MustParse(req.LeagueID)
	fId := uuid.MustParse(req.FranchiseID)
	pId := uuid.MustParse(req.ProspectID)

	if findPick := s.R.DB.Where(&models.Pick{DraftYear: req.DraftPick.DraftYear, DraftRound: req.DraftPick.DraftRound, DraftPickInRound: req.DraftPick.DraftPickInRound, DraftPickOverall: req.DraftPick.DraftPickOverall}).First(&pick); findPick.Error == nil {
		return &pb.DraftProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Pick already exists in this league (%v)", findPick),
		}, nil
	}

	if findPick := s.R.DB.Where(&models.Pick{ProspectID: pId}).First(&pick2); findPick.Error == nil {
		return &pb.DraftProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("This ProspectID was already drafted (%v)", pId),
		}, nil
	}

	pick.DraftYear = req.DraftPick.DraftYear
	pick.DraftRound = req.DraftPick.DraftRound
	pick.DraftPickInRound = req.DraftPick.DraftPickInRound
	pick.DraftPickOverall = req.DraftPick.DraftPickOverall
	pick.ProspectID = pId
	pick.Owner = fId

	if createPick := s.R.DB.Create(&pick); createPick.Error != nil {
		return &pb.DraftProspectResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Creating prospects failed %q", createPick.Error),
		}, nil
	}

	if findProspect := s.R.DB.Where(&models.Prospect{ID: pId}).First(&prospect); findProspect.Error != nil {
		return &pb.DraftProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("This ProspectID doest not exist and cannot be updated (%v)", pId),
		}, nil
	}
	prospect.LeagueID = &lId
	prospect.FranchiseID = &fId

	s.R.DB.Save(&prospect)

	return &pb.DraftProspectResponse{
		Status: http.StatusCreated,
		PickID: pick.ID.String(),
	}, nil

}

func (s *Server) UndraftProspect(ctx context.Context, req *pb.DraftProspectRequest) (*pb.DraftProspectResponse, error) {
	var pick models.Pick
	var franchise models.Franchise
	var prospect models.Prospect
	var league models.League

	lId := uuid.MustParse(req.LeagueID)
	fId := uuid.MustParse(req.FranchiseID)
	pId := uuid.MustParse(req.ProspectID)

	s.R.DB.Where(&models.Prospect{ID: pId}).First(&prospect)

	if findPick := s.R.DB.Where(&models.Pick{DraftYear: req.DraftPick.DraftYear, DraftRound: req.DraftPick.DraftRound, DraftPickInRound: req.DraftPick.DraftPickInRound, DraftPickOverall: req.DraftPick.DraftPickOverall, ProspectID: pId}).First(&pick); findPick.Error != nil {
		return &pb.DraftProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Pick does not exist in this league (%v)", pick.ID),
		}, nil
	}

	if findProspect := s.R.DB.Where(&models.Prospect{ID: pId}).First(&prospect); findProspect.Error != nil {
		return &pb.DraftProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("This ProspectID doest not exist and cannot be undrafted (%v)", pId),
		}, nil
	}

	if fId != *prospect.FranchiseID {
		return &pb.DraftProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Prospect (%v) does not belong to this Franchise (%q).", pId, prospect.FranchiseID),
		}, nil
	}

	s.R.DB.Preload("Franchises").Preload("Prospects").Preload("Franchises.Prospects").Where(&models.League{ID: lId}).First(&league)
	s.R.DB.Preload("Prospects").Where(&models.Franchise{ID: fId}).First(&franchise)
	s.R.DB.Model(&league).Association("Prospects").Delete(&prospect)
	s.R.DB.Model(&franchise).Association("Prospects").Delete(&prospect)
	s.R.DB.Model(&prospect).Association("Pick").Delete(&pick)
	s.R.DB.Model(&franchise).Association("Picks").Delete(&pick)

	return &pb.DraftProspectResponse{
		Status: http.StatusCreated,
		PickID: pick.ID.String(),
	}, nil
}
*/
/*
func (s *Server) Trade(ctx context.Context, req *pb.TradeRequest) (*pb.TradeResponse, error) {
	var pick models.Pick
	var franchise models.Franchise
	var prospect models.Prospect
	var league models.League

	lId := uuid.MustParse(req.LeagueID)
	fId := uuid.MustParse(req.FranchiseID)
	pId := uuid.MustParse(req.ProspectID)

	s.R.DB.Where(&models.Prospect{ID: pId}).First(&prospect)

	if findPick := s.R.DB.Where(&models.Pick{DraftYear: req.DraftPick.DraftYear, DraftRound: req.DraftPick.DraftRound, DraftPickInRound: req.DraftPick.DraftPickInRound, DraftPickOverall: req.DraftPick.DraftPickOverall, ProspectID: pId}).First(&pick); findPick.Error != nil {
		return &pb.DraftProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Pick does not exist in this league (%v)", pick.ID),
		}, nil
	}

	if findProspect := s.R.DB.Where(&models.Prospect{ID: pId}).First(&prospect); findProspect.Error != nil {
		return &pb.DraftProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("This ProspectID doest not exist and cannot be undrafted (%v)", pId),
		}, nil
	}

	if fId != *prospect.FranchiseID {
		return &pb.DraftProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Prospect (%v) does not belong to this Franchise (%q).", pId, prospect.FranchiseID),
		}, nil
	}

	s.R.DB.Preload("Franchises").Preload("Prospects").Preload("Franchises.Prospects").Where(&models.League{ID: lId}).First(&league)
	s.R.DB.Preload("Prospects").Where(&models.Franchise{ID: fId}).First(&franchise)
	s.R.DB.Model(&league).Association("Prospects").Delete(&prospect)
	s.R.DB.Model(&franchise).Association("Prospects").Delete(&prospect)
	s.R.DB.Model(&prospect).Association("Pick").Delete(&pick)
	s.R.DB.Model(&franchise).Association("Picks").Delete(&pick)

	return &pb.DraftProspectResponse{
		Status: http.StatusCreated,
		PickID: pick.ID.String(),
	}, nil
}
*/
