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
)

type Server struct {
	R storage.Repository
	// #https://github.com/grpc/grpc-go/issues/3794:
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
	//const leagueName = "Lakelandcup"

	// check if league already exists
	if findLeague := s.R.DB.Where(&models.League{ID: uuid.MustParse(req.Id)}).First(&league); findLeague.Error != nil {
		return &pb.LeagueResponse{
			Status: http.StatusConflict,
			Error:  "League does't exists",
		}, nil
	}

	logrus.Info(req)

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

func (s *Server) GetLeagueFranchisePairs(ctx context.Context, req *pb.GetLeagueFranchisePairsRequest) (*pb.GetLeagueFranchisePairsResponse, error) {
	var leagues []models.League
	var res []*pb.LeagueFranchisePair

	s.R.DB.Preload("Franchises").Find(&leagues)

	for _, l := range leagues {
		if len(l.Franchises) > 0 {
			for _, f := range l.Franchises {
				if f.OwnerID.String() == req.UserId {
					res = append(res, &pb.LeagueFranchisePair{LeagueID: f.LeagueID.String(), FranchiseID: f.ID.String()})
				}
			}
		} else {
			if l.AdminID.String() == req.UserId {
				res = append(res, &pb.LeagueFranchisePair{LeagueID: l.ID.String(), FranchiseID: ""})
			}
		}
	}

	return &pb.GetLeagueFranchisePairsResponse{
		Status: http.StatusAccepted,
		Result: res,
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
			tmpFranchise.OwnerID = f.OwnerID.String()
			tmpFranchise.OwnerName = f.OwnerName
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
		tmpFranchise := pb.Franchise{}
		if len(l.Franchises) > 0 {
			for _, f := range l.Franchises {
				tmpFranchise.ID = f.ID.String()
				tmpFranchise.OwnerID = f.OwnerID.String()
				tmpFranchise.OwnerName = f.OwnerName
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
	franchise.OwnerID = uuid.MustParse(req.OwnerID)
	franchise.OwnerName = req.OwnerName
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

	findFranchise := s.R.DB.Preload("Prospects").First(&franchise, "id = ?", req.FranchiseId)

	if findFranchise.Error != nil {
		return &pb.GetFranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Getting franchiseId (%s) failed: %v", req.FranchiseId, findFranchise.Error),
		}, nil
	}

	if findFranchise.RowsAffected == 0 {
		return &pb.GetFranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("franchiseId (%s) does not exist", req.FranchiseId),
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
		OwnerID:        franchise.OwnerID.String(),
		OwnerName:      franchise.OwnerName,
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
			OwnerID:        f.OwnerID.String(),
			OwnerName:      f.OwnerName,
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

	for _,pReq:= range req.Prospects {
		var prospect models.Prospect
		if findProspect := s.R.DB.Where(&models.Prospect{FullName: pReq.FullName, Birthdate: pReq.Birthdate,DraftYear: pReq.DraftYear, NhlDraftPickOverall: pReq.NhlDraftPickOverall}).First(&prospect); findProspect.Error != nil {
			prospect.FullName = pReq.FullName
			prospect.FirstName = pReq.FirstName
			prospect.LastName = pReq.LastName
			prospect.NhlTeam = pReq.NhlTeam
			prospect.Birthdate = pReq.Birthdate
			prospect.Height = pReq.Height
			prospect.Weight = pReq.Weight
			prospect.DraftYear = pReq.DraftYear
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
		Status:     http.StatusCreated,
		ProspectIds: []string{},
	}, nil
}


func (s *Server) CreateProspect(ctx context.Context, req *pb.CreateProspectRequest) (*pb.CreateProspectResponse, error) {
	var prospect models.Prospect

	lId := uuid.MustParse(req.Prospect.LeagueID)
	fId := uuid.MustParse(req.Prospect.FranchiseID)

	if findProspect := s.R.DB.Where(&models.Prospect{FullName: req.Prospect.FullName, Birthdate: req.Prospect.Birthdate, LeagueID: &lId}).First(&prospect); findProspect.Error == nil {
		return &pb.CreateProspectResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Prospect already exists in this league (%q)", req.Prospect.FullName),
		}, nil
	}

	prospect.FullName = req.Prospect.FullName
	prospect.FirstName = req.Prospect.FirstName
	prospect.LastName = req.Prospect.LastName
	prospect.Birthdate = req.Prospect.Birthdate
	prospect.LeagueID = &lId
	prospect.FranchiseID = &fId

	if createProspect := s.R.DB.Create(&prospect); createProspect.Error != nil {
		return &pb.CreateProspectResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Creating prospects failed %q", createProspect.Error),
		}, nil
	}

	return &pb.CreateProspectResponse{
		Status:     http.StatusCreated,
		ProspectID: prospect.ID.String(),
	}, nil
}
