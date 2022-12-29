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

func (s *Server) GetAllLeaguesForUser(ctx context.Context, req *pb.GetAllLeaguesForUserRequest) (*pb.LeagueResultResponse, error) {
	var leagues []models.League
	var leagueRes []*pb.League
	var franchisesRes []*pb.Franchise

	s.R.DB.Preload("Franchises").Find(&leagues)
	log.Printf("LOG %v", leagues)
	for _, l := range leagues {
		tmpFranchise := pb.Franchise{}
		if len(l.Franchises) > 0 {
			for _, f := range l.Franchises {
				if f.FranchiseOwner.String() == req.UserId {
					tmpFranchise.ID = f.ID.String()
					tmpFranchise.FranchisOwner = f.FranchiseOwner.String()
					tmpFranchise.FranchiseName = f.FranchiseName
					tmpFranchise.FoundationYear = f.FoundationYear
					franchisesRes = append(franchisesRes, &tmpFranchise)

				}
			}
		} else {
			franchisesRes = append(franchisesRes, &pb.Franchise{})
		}
		if l.LeagueFounder.String() == req.UserId || len(l.Franchises) > 0 {
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
	}
	return &pb.LeagueResultResponse{
		Status: http.StatusCreated,
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
