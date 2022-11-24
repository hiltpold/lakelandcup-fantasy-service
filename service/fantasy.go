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

	// TODO: check if userId exisits

	// check if league already exists

	if findLeague := s.R.DB.Where(&models.League{LeagueName: req.LeagueName}).First(&league); findLeague.Error == nil {
		return &pb.LeagueResponse{
			Status: http.StatusConflict,
			Error:  "League already exists",
		}, nil
	}

	league.UserID = req.UserId
	league.LeagueName = req.LeagueName
	league.FoundationYear = req.FoundationYear
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

func (s *Server) CreateFranchise(ctx context.Context, req *pb.FranchiseRequest) (*pb.FranchiseResponse, error) {
	var league models.League
	var franchise models.Franchise

	// check if franchise exists in this league
	if findLeague := s.R.DB.Where(&models.League{ID: uuid.MustParse(req.LeagueId)}).First(&league); findLeague.RowsAffected == 0 {
		return &pb.FranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Franchise cannot be created, provided leagueId (%s) does not exist in this league", req.LeagueId),
		}, nil
	} else {
		logrus.Info(fmt.Sprintf("existing league %+v", findLeague))
	}
	// check if franchise name already taken in this

	if findFranchise := s.R.DB.Where(&models.Franchise{ID: uuid.MustParse(req.LeagueId), FranchiseName: req.FranchiseName}).First(&franchise); findFranchise.Error == nil {
		return &pb.FranchiseResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("Franchise cannot be created, franchise with name (%s) already exisits in this league", req.FranchiseName),
		}, nil
	}

	franchise.FranchiseName = req.FranchiseName
	franchise.FoundationYear = req.FoundationYear
	franchise.LeagueID = uuid.MustParse(req.LeagueId)

	logrus.Info(franchise.ID)
	if createFranchise := s.R.DB.Create(&franchise); createFranchise.Error != nil {
		return &pb.FranchiseResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Creating new franchise for league (%s) failed", req.LeagueId),
		}, nil
	}

	return &pb.FranchiseResponse{
		Status:      http.StatusCreated,
		FranchiseId: franchise.ID.String(),
	}, nil
}

func (s *Server) GetLeagueById(ctx context.Context, req *pb.GetLeagueByIdRequest) (*pb.QueryResponse, error) {
	var leagues []models.League

	// check if franchise exists in this league
	s.R.DB.Preload("Franchises").Find(&leagues)
	logrus.Info(leagues)

	return &pb.QueryResponse{
		Status: http.StatusCreated,
		Result: "",
	}, nil
}
