package test

import (
	"fmt"

	"github.com/hiltpold/lakelandcup-fantasy-service/conf"
	"github.com/hiltpold/lakelandcup-fantasy-service/service/pb"
	"google.golang.org/grpc"
)

type ServiceClient struct {
	Client pb.FantasyServiceClient
}

func InitServiceClient(c *conf.Configuration) pb.FantasyServiceClient {
	// using WithInsecure() because no SSL running
	cc, err := grpc.Dial(c.API.Host, grpc.WithInsecure())

	if err != nil {
		fmt.Println("Could not connect:", err)
	}

	return pb.NewFantasyServiceClient(cc)
}
