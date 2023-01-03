package service

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/hiltpold/lakelandcup-fantasy-service/models"
	"github.com/hiltpold/lakelandcup-fantasy-service/service/pb"
	"github.com/hiltpold/lakelandcup-fantasy-service/storage"
)

type Server struct {
	R storage.Repository
	// #https://github.com/grpc/grpc-go/issues/3794:
	pb.UnimplementedFantasyServiceServer
}

func (s *Server) CreateLeague(ctx context.Context, req *pb.LeagueRequest) (*pb.LeagueResponse, error) {
	var league models.League

	// TODO: check if userId exisits

	// check if league already exists

	if findLeague := s.R.DB.Where(&models.League{LeagueName: req.LeagueName}).First(&league); findLeague.Error == nil {
		return &pb.LeagueResponse{
			Status: http.StatusConflict,
			Error:  "League already exists",
		}, nil
	}

	league.LeagueFounder = uuid.MustParse(req.UserId)
	league.LeagueName = req.LeagueName
	league.FoundationYear = req.FoundationYear
	league.MaxFranchises = int(req.MaxFranchises)
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

func (s *Server) GetLeagueFranchisePairs(ctx context.Context, req *pb.GetLeagueFranchisePairsRequest) (*pb.GetLeagueFranchisePairsResponse, error) {
	var leagues []models.League
	var res []*pb.LeagueFranchisePair

	s.R.DB.Preload("Franchises").Find(&leagues)

	for _, l := range leagues {
		if len(l.Franchises) > 0 {
			for _, f := range l.Franchises {
				if f.FranchiseOwner.String() == req.UserId {
					res = append(res, &pb.LeagueFranchisePair{LeagueID: f.LeagueID.String(), FranchiseID: f.ID.String()})
				}
			}
		} else {
			if l.LeagueFounder.String() == req.UserId {
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
			Error:  fmt.Sprintf("Getting leagueId (%s) failed", req.LeagueId),
		}, nil
	}

	if findLeague.RowsAffected == 0 {
		return &pb.GetLeagueResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("leagueId (%s) does not exist", req.LeagueId),
		}, nil

	}

	tmpFranchise := pb.Franchise{}
	if len(league.Franchises) > 0 {
		for _, f := range league.Franchises {
			tmpFranchise.ID = f.ID.String()
			tmpFranchise.FranchisOwner = f.FranchiseOwner.String()
			tmpFranchise.FranchiseName = f.FranchiseName
			tmpFranchise.FoundationYear = f.FoundationYear
			franchisesRes = append(franchisesRes, &tmpFranchise)

		}
	} else {
		franchisesRes = append(franchisesRes, &pb.Franchise{})
	}
	leagueRes = &pb.League{
		ID:             league.ID.String(),
		LeagueFounder:  league.LeagueFounder.String(),
		LeagueName:     league.LeagueName,
		FoundationYear: league.FoundationYear,
		MaxFranchises:  int32(league.MaxFranchises),
		Franchises:     franchisesRes,
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
				tmpFranchise.FranchisOwner = f.FranchiseOwner.String()
				tmpFranchise.FranchiseName = f.FranchiseName
				tmpFranchise.FoundationYear = f.FoundationYear
				franchisesRes = append(franchisesRes, &tmpFranchise)

			}
		} else {
			franchisesRes = append(franchisesRes, &pb.Franchise{})
		}
		tmpLeague := pb.League{
			ID:             l.ID.String(),
			LeagueFounder:  l.LeagueFounder.String(),
			LeagueName:     l.LeagueName,
			FoundationYear: l.FoundationYear,
			MaxFranchises:  int32(l.MaxFranchises),
			Franchises:     franchisesRes,
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
	if findFranchise := s.R.DB.Where(&models.Franchise{LeagueID: uuid.MustParse(req.LeagueId), FranchiseName: req.FranchiseName}).First(&franchise); findFranchise.Error == nil {
		return &pb.FranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Franchise cannot be created, franchise with name (%s) already exisits in this league", req.FranchiseName),
		}, nil
	}

	// check if maximum franchises already satisfied
	if len(league.Franchises) >= league.MaxFranchises {
		return &pb.FranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Franchise cannot be created, maximum number of franchises already created (%d)", league.MaxFranchises),
		}, nil
	}

	franchise.FranchiseName = req.FranchiseName
	franchise.FranchiseOwner = uuid.MustParse(req.FranchiseOwner)
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
		FranchisOwner:  franchise.FranchiseOwner.String(),
		FranchiseName:  franchise.FranchiseName,
		FoundationYear: franchise.FoundationYear,
		Prospects:      prospectRes,
	}

	return &pb.GetFranchiseResponse{
		Status: http.StatusAccepted,
		Result: franchiseRes,
	}, nil

}

func (s *Server) CreateUndraftedProspects(ctx context.Context, req *pb.CreateUndraftedProspectsRequest) (*pb.CreateUndraftedProspectsResponse, error) {
	var prospects []models.Prospect

	for _, p := range req.Prospects {
		prospects = append(prospects, models.Prospect{
			FullName:  p.FullName,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Birthdate: p.Birthdate,
		})
	}

	if createProspects := s.R.DB.Create(&prospects); createProspects.Error != nil {
		return &pb.CreateUndraftedProspectsResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Creating prospects failed: %v", createProspects.Error),
		}, nil
	}

	var result []string
	for _, p := range prospects {
		result = append(result, p.ID.String())
	}

	log.Printf("%v", result)

	return &pb.CreateUndraftedProspectsResponse{
		Status:      http.StatusCreated,
		ProspectIds: result,
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
