package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hiltpold/lakelandcup-fantasy-service/models"
	"github.com/hiltpold/lakelandcup-fantasy-service/service/pb"
	"github.com/sirupsen/logrus"
)

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
