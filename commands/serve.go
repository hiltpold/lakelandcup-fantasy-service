package commands

import (
	"fmt"
	"net"

	"github.com/hiltpold/lakelandcup-fantasy-service/conf"
	api "github.com/hiltpold/lakelandcup-fantasy-service/service"
	"github.com/hiltpold/lakelandcup-fantasy-service/service/pb"
	"github.com/hiltpold/lakelandcup-fantasy-service/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start Fantasy Service",
	Run: func(cmd *cobra.Command, args []string) {
		runWithConfig(cmd, serve)
	},
}

func serve(c *conf.Configuration) {
	h := storage.Dial(&c.DB)

	serviceUri := fmt.Sprintf(":%s", c.API.Port)

	lis, err := net.Listen("tcp", serviceUri)

	if err != nil {
		logrus.Fatal("Failed to listen on: ", err)
	}

	logrus.Info(fmt.Sprintf("Service [%s] from app [%s] is running on [%s]", c.API.Svc, c.API.App, serviceUri))

	s := api.Server{
		R: h,
	}

	grpcServer := grpc.NewServer()

	pb.RegisterFantasyServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		logrus.Fatalln("Failed to serve:", err)
	}
}
