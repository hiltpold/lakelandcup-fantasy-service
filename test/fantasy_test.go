package test

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hiltpold/lakelandcup-fantasy-service/conf"
	"github.com/hiltpold/lakelandcup-fantasy-service/service"
	"github.com/hiltpold/lakelandcup-fantasy-service/service/pb"
	"github.com/hiltpold/lakelandcup-fantasy-service/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/gorm"
)

const bufSize = 1024 * 1024
const testConfig = ".test.env"

var lis *bufconn.Listener
var db *gorm.DB
var client pb.FantasyServiceClient
var ctx context.Context
var conn *grpc.ClientConn

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func setupServer(c *conf.Configuration) {
	h := storage.Dial(&c.DB)
	db = h.DB

	lis = bufconn.Listen(bufSize)
	s := service.Server{
		R: h,
	}
	grpcServer := grpc.NewServer()
	pb.RegisterFantasyServiceServer(grpcServer, &s)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func setupClient() (pb.FantasyServiceClient, context.Context, *grpc.ClientConn) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("Failed to dial bufnet: %v", err)
	}
	c := pb.NewFantasyServiceClient(conn)
	return c, ctx, conn
}

func setup() (pb.FantasyServiceClient, context.Context, *grpc.ClientConn) {
	c, err := conf.LoadConfig(testConfig)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("Failed to load config %s.", testConfig), err)
	}
	setupServer(c)
	return setupClient()
}

func TestMain(m *testing.M) {
	client, ctx, conn = setup()
	exitVal := m.Run()
	conn.Close()
	os.Exit(exitVal)
}

func TestLeagueCreation(t *testing.T) {
	// generate userId
	userId := uuid.Must(uuid.NewRandom()).String()
	leagueName := "TestLeague"
	foundationYear := "2022"
	// test request
	leagueReq := pb.LeagueRequest{UserId: userId, LeagueName: leagueName, FoundationYear: foundationYear}

	leagueResp, err := client.CreateLeague(ctx, &leagueReq)
	if err != nil {
		t.Fatalf("League creation failed: %v", err)
	}
	log.Printf("Response: %+v", leagueResp)

	// Test for response here.
	//assert.Equal(t, loginResp.Status, int64(404))
	//assert.Equal(t, loginResp.Error, "Incorrect email or password")

	// Clean up
	//db.Where("UserId = ?", userId).Delete(&models.League{})

}

func TestFranchiseCreation(t *testing.T) {
	// genreate league test request
	leagueReq := pb.LeagueRequest{UserId: uuid.Must(uuid.NewRandom()).String(), LeagueName: "TestFranchiseLeague", FoundationYear: "2022"}

	leagueResp, err := client.CreateLeague(ctx, &leagueReq)
	if err != nil {
		t.Fatalf("League creation failed: %v", err)
	}

	log.Printf("League response:\n%+v", leagueResp)

	// genreate franchise test request
	franchiseReq := pb.FranchiseRequest{LeagueId: leagueResp.LeagueId, FranchiseName: "TestFranchise", FoundationYear: "2022"}

	franchiseResp, err := client.CreateFranchise(ctx, &franchiseReq)
	if err != nil {
		t.Fatalf("Franchise creation failed: %v", err)
	}

	log.Printf("Franchise response:\n%+v", franchiseResp)

	// Test for response here.
	//assert.Equal(t, loginResp.Status, int64(404))
	//assert.Equal(t, loginResp.Error, "Incorrect email or password")

	// Clean up
	//db.Where(" = ?", userId).Delete(&models.League{})

}

func TestGetLeagueById(t *testing.T) {
	getLeagueByIdReq := pb.GetLeagueByIdRequest{LeagueId: "8ef73594-c688-47b6-9233-f92cbf8b75e4"}
	franchiseResp, _ := client.GetLeagueById(ctx, &getLeagueByIdReq)
	logrus.Info(franchiseResp)

}
