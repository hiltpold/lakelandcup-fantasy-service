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

func (s *Server) CreateOrUpdatePicks(ctx context.Context, req *pb.CreateOrUpdatePicksRequest) (*pb.DefaultResponse, error) {
	var nFranchises int64

	var league models.League
	const leagueName = "Lakelandcup"

	// get league
	if findLeague := s.R.DB.Where(&models.League{ID: uuid.MustParse(req.LeagueID), Name: leagueName}).First(&league); findLeague.Error != nil {
		return &pb.DefaultResponse{
			Status: http.StatusConflict,
			Error:  fmt.Sprintf("League does't exist. Error %v", findLeague.Error),
		}, nil
	}

	// get franchise count
	s.R.DB.Model(&models.Franchise{}).Count(&nFranchises)

	for _, p := range req.Picks {
		fId := uuid.MustParse(p.FranchiseID)
		for dr := 1; dr <= league.DraftRounds; dr++ {
			var pick models.Pick

			findPick := s.R.DB.Where("origin_id = ? AND draft_year = ? AND draft_round = ?", fId, p.Year, dr).Find(&pick)

			if findPick.Error != nil {
				return &pb.DefaultResponse{
					Status: http.StatusForbidden,
					Error:  fmt.Sprintf("Error while querying for pick with origin_id %q and year %q and draft round %q. Error: %q", fId, p.Year, dr, findPick.Error),
				}, nil
			}

			if p.LotteryPosition == 0 {
				pick.DraftPickInRound = nil
				pick.DraftPickOverall = nil

			} else {
				// calculate overall pick
				overallPick := int(p.LotteryPosition) + (dr-1)*int(nFranchises)
				pInRound := fmt.Sprintf("%v", p.LotteryPosition)
				pOverall := fmt.Sprintf("%v", overallPick)
				pick.DraftPickInRound = &pInRound
				pick.DraftPickOverall = &pOverall
			}

			if findPick.RowsAffected == 0 {
				// pick does not exisit, create it
				pick.DraftYear = p.Year
				pick.DraftRound = fmt.Sprintf("%v", dr)
				pick.ProspectID = nil
				pick.OwnerID = &fId
				pick.OwnerName = p.Franchise
				pick.LastOwnerID = &fId
				pick.LastOwnerName = p.Franchise
				pick.OriginID = &fId
				pick.OriginName = p.Franchise

				// create
				if createPick := s.R.DB.Create(&pick); createPick.Error != nil {
					return &pb.DefaultResponse{
						Status: http.StatusForbidden,
						Error:  fmt.Sprintf("Creating prospects failed %q", createPick.Error),
					}, nil
				}
			} else if findPick.RowsAffected == 1 {

				// pick exists, update it
				pick.DraftYear = p.Year
				pick.DraftRound = fmt.Sprintf("%v", dr)
				// update
				s.R.DB.Save(&pick)

			} else {
				// multiple picks exist, this should not happen
				return &pb.DefaultResponse{
					Status: http.StatusForbidden,
					Error:  fmt.Sprintf("Multiple picks for origin_id %q and year %q exist, shoud not happen!", fId, p.Year),
				}, nil
			}

		}

	}

	return &pb.DefaultResponse{
		Status: http.StatusCreated,
	}, nil
}

func (s *Server) GetPicksByYear(ctx context.Context, req *pb.GetPicksRequest) (*pb.GetPicksResponse, error) {
	var picks []models.Pick
	picksRes := []*pb.Pick{}

	year := fmt.Sprintf("%v", req.Year)
	logrus.Info(year)

	if findPicks := s.R.DB.Where("draft_year = ?", year).Find(&picks).Limit(1000); findPicks.Error != nil {
		return &pb.GetPicksResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Could not find any picks %q", findPicks.Error),
		}, nil
	}

	for _, p := range picks {
		pId := ""
		if p.ProspectID != nil {
			pId = p.ProspectID.String()
		}

		pInRound := ""
		if p.DraftPickInRound != nil {
			pInRound = *p.DraftPickInRound
		}

		pOverall := ""
		if p.DraftPickOverall != nil {
			pOverall = *p.DraftPickOverall
		}

		picksRes = append(picksRes, &pb.Pick{
			ID:               p.ID.String(),
			DraftYear:        p.DraftYear,
			DraftRound:       p.DraftRound,
			DraftPickInRound: pInRound,
			DraftPickOverall: pOverall,
			ProspectID:       pId,
			OwnerID:          p.OwnerID.String(),
			OwnerName:        p.OwnerName,
			LastOwnerID:      p.LastOwnerID.String(),
			LastOwnerName:    p.LastOwnerName,
			OriginID:         p.OriginID.String(),
			OriginName:       p.OriginName,
		})
	}

	return &pb.GetPicksResponse{
		Status: http.StatusOK,
		Picks:  picksRes,
	}, nil

}

func (s *Server) GetPicksByFranchise(ctx context.Context, req *pb.GetPicksRequest) (*pb.GetPicksResponse, error) {
	var picks []models.Pick

	fId, err := uuid.Parse(req.FranchiseID)
	if err != nil {
		return &pb.GetPicksResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Could not parse uuid for franchise id %q.", fId)}, nil

	}

	if findPicks := s.R.DB.Where(models.Pick{OwnerID: &fId}).Find(&picks).Limit(1000); findPicks.Error != nil {
		return &pb.GetPicksResponse{
			Status: http.StatusForbidden,
			Error:  fmt.Sprintf("Could not fetch picks for franchise %q. Error: %v", fId, findPicks.Error),
		}, nil
	}

	picksRes := []*pb.Pick{}
	for _, p := range picks {
		pId := ""
		if p.ProspectID != nil {
			pId = p.ProspectID.String()
		}

		pInRound := ""
		if p.DraftPickInRound != nil {
			pInRound = *p.DraftPickInRound
		}

		pOverall := ""
		if p.DraftPickOverall != nil {
			pOverall = *p.DraftPickOverall
		}

		picksRes = append(picksRes, &pb.Pick{
			ID:               p.ID.String(),
			DraftYear:        p.DraftYear,
			DraftRound:       p.DraftRound,
			DraftPickInRound: pInRound,
			DraftPickOverall: pOverall,
			ProspectID:       pId,
			OwnerID:          p.OwnerID.String(),
			OwnerName:        p.OwnerName,
			LastOwnerID:      p.LastOwnerID.String(),
			LastOwnerName:    p.LastOwnerName,
			OriginID:         p.OriginID.String(),
			OriginName:       p.OriginName,
		})
	}

	return &pb.GetPicksResponse{
		Status: http.StatusOK,
		Picks:  picksRes,
	}, nil

}
