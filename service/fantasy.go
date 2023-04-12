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
)

type Server struct {
	R storage.Repository
	// https://github.com/grpc/grpc-go/issues/3794:
	pb.UnimplementedFantasyServiceServer
}

func (s *Server) Trade(ctx context.Context, req *pb.TradeRequest) (*pb.DefaultResponse, error) {
	//logrus.Info(fmt.Sprintf("%+v", req))
	var transaction = s.R.DB.Transaction(func(tx *gorm.DB) error {
		//
		var firstFranchise models.Franchise
		var secondFranchise models.Franchise

		firstFranchiseID, err := uuid.Parse(req.First.FranchiseID)
		if err != nil {
			logrus.Error(fmt.Sprintf("could not parse first FranchiseID %v", req.First.FranchiseID))
		}

		secondFranchiseID, err := uuid.Parse(req.Second.FranchiseID)
		if err != nil {
			logrus.Error(fmt.Sprintf("could not parse second FranchiseID %v", req.Second.FranchiseID))
		}

		firstFranchise = models.Franchise{ID: firstFranchiseID}
		if findFirstFranchise := tx.First(&firstFranchise); findFirstFranchise.Error != nil {
			return fmt.Errorf("error while querying franchise with ID %v. Error: %+v", firstFranchiseID, findFirstFranchise.Error)
		}

		secondFranchise = models.Franchise{ID: secondFranchiseID}
		if findSecondFranchise := tx.First(&secondFranchise); findSecondFranchise.Error != nil {
			return fmt.Errorf("error while querying franchise with ID %v. Error: %+v", secondFranchiseID, findSecondFranchise.Error)
		}

		// update first picks
		for _, firstPId := range req.First.Picks {
			pId, err := uuid.Parse(firstPId)
			if err != nil {
				logrus.Error(fmt.Sprintf("could not parse PickID %v", pId))
			}
			//
			var pick = models.Pick{ID: pId}
			findPick := tx.First(&pick)

			if findPick.Error != nil {
				return fmt.Errorf("error while querying pick with ID %v. Error %+v", pId, findPick.Error)
			}

			if findPick.RowsAffected == 1 {
				// update ownership
				pick.LastOwnerID = &firstFranchiseID
				pick.LastOwnerName = firstFranchise.Name

				pick.OwnerID = &secondFranchiseID
				pick.OwnerName = secondFranchise.Name

				tx.Save(&pick)
			}
		}

		// update second picks
		for _, secondPId := range req.Second.Picks {
			pId, err := uuid.Parse(secondPId)
			if err != nil {
				logrus.Error(fmt.Sprintf("could not parse PickID %v", pId))
			}
			//
			var pick = models.Pick{ID: pId}
			findPick := tx.First(&pick)

			if findPick.Error != nil {
				return fmt.Errorf("error while querying pick with ID %v. Error %+v", pId, findPick.Error)
			}

			if findPick.RowsAffected == 1 {
				// update ownership
				pick.LastOwnerID = &secondFranchiseID
				pick.LastOwnerName = secondFranchise.Name

				pick.OwnerID = &firstFranchiseID
				pick.OwnerName = firstFranchise.Name

				tx.Save(&pick)
			}
		}

		// update first prospects
		for _, firstPId := range req.First.Prospects {
			pId, err := uuid.Parse(firstPId)
			if err != nil {
				logrus.Error(fmt.Sprintf("could not parse PickID %v", pId))
			}
			//
			var prospect = models.Prospect{ID: pId}
			findProspect := tx.First(&prospect)

			if findProspect.Error != nil {
				return fmt.Errorf("error while querying pick with ID %v. Error %+v", pId, findProspect.Error)
			}

			if findProspect.RowsAffected == 1 {
				prospect.FranchiseID = &secondFranchiseID
				tx.Save(&prospect)
			}

		}

		// update second prospects
		for _, secondPId := range req.Second.Prospects {
			pId, err := uuid.Parse(secondPId)
			if err != nil {
				logrus.Error(fmt.Sprintf("could not parse PickID %v", pId))
			}
			//
			var prospect = models.Prospect{ID: pId}
			findProspect := tx.First(&prospect)

			if findProspect.Error != nil {
				return fmt.Errorf("error while querying pick with ID %v. Error %+v", pId, findProspect.Error)
			}

			if findProspect.RowsAffected == 1 {
				prospect.FranchiseID = &firstFranchiseID
				tx.Save(&prospect)
			}

		}

		//
		return nil
	})

	if transaction != nil {
		return &pb.DefaultResponse{
			Status: http.StatusConflict,
			Error:  transaction.Error(),
		}, nil

	}

	return &pb.DefaultResponse{
		Status: http.StatusOK,
	}, nil
}
