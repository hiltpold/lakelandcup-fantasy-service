package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hiltpold/lakelandcup-fantasy-service/service/pb"
	"github.com/sirupsen/logrus"
)

func (s *Server) CreateOrUpdatePicks(ctx context.Context, req *pb.CreateOrUpdatePicksRequest) (*pb.DefaultResponse, error) {

	logrus.Info(fmt.Sprintf("%v", req))

	return &pb.DefaultResponse{
		Status: http.StatusCreated,
	}, nil
}
