package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hiltpold/lakelandcup-fantasy-service/models"
	"github.com/hiltpold/lakelandcup-fantasy-service/service/pb"
)

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
