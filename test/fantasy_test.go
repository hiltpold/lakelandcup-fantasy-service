package test

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hiltpold/lakelandcup-fantasy-service/conf"
	"github.com/hiltpold/lakelandcup-fantasy-service/models"
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

// generate test data
const wrongLeagueId = "00000000-0000-0000-0000-000000000000"
const userId = "00000000-0000-0000-0000-000000000000"
const leagueName = "TestLeague"
const foundationYear = "2022"
const maxFranchises = 1
const userId2 = "11111111-1111-1111-1111-111111111111"
const leagueName2 = "TestLeague2"
const maxFranchises2 = 3

const franchiseName = "TestFranchise"
const franchiseFoundationYear = "2022"
const franchiseName2 = "TestFranchise2"
const franchiseName3 = "TestFranchise3"

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

func createLeague(userId string, leagueName string, foundationYear string, maxFranchises int) (*pb.LeagueResponse, error) {
	req := pb.LeagueRequest{UserId: userId, LeagueName: leagueName, FoundationYear: foundationYear, MaxFranchises: int32(maxFranchises)}
	resp, err := client.CreateLeague(ctx, &req)
	return resp, err
}

func createFranchise(leagueId string, franchiseOwner string, franchiseName string, foundationYear string) (*pb.FranchiseResponse, error) {
	req := pb.FranchiseRequest{LeagueId: leagueId, FranchiseOwner: franchiseOwner, FranchiseName: franchiseName, FoundationYear: foundationYear}
	resp, err := client.CreateFranchise(ctx, &req)
	return resp, err
}

func TestLeagueCreation(t *testing.T) {
	resp, err := createLeague(userId, leagueName, foundationYear, maxFranchises)
	if err != nil {
		t.Errorf("League creation failed: %v", err)
	}

	// Log result
	t.Log("-----------------------")
	t.Log("Create League Response:")
	t.Logf("%+v", resp)
	t.Log("-----------------------")

	// Test league creation
	if resp.Status != 201 {
		t.Errorf("Output %q not equal to expected %q", resp.Status, 201)
	}

	if lId, e := uuid.Parse(resp.LeagueId); e != nil {
		t.Errorf("Output [%q] expected to be an uuid", lId)
	}

	resp2, err2 := createLeague(userId, leagueName, foundationYear, maxFranchises)
	if err2 != nil {
		t.Errorf("League creation failed: %v", err2)
	}

	// Log result
	t.Log("-----------------------")
	t.Log("Create League Response:")
	t.Logf("%+v", resp2)
	t.Log("-----------------------")

	// Test league creation
	if resp2.Status != 409 {
		t.Errorf("Output %d not equal to expected %d", resp2.Status, 409)
	}
	expectedError := "League already exists"
	if resp2.Error != expectedError {
		t.Errorf("Output [%q] not equal to expected [%q]", resp.Error, expectedError)
	}

	// Clean up
	db.Where("league_founder = ?", userId).Delete(&models.League{})
	db.Where("league_founder = ?", userId2).Delete(&models.League{})
}

func TestFranchiseCreation(t *testing.T) {
	// Create League
	lResp, lErr := createLeague(userId, leagueName, foundationYear, maxFranchises)
	if lErr != nil {
		t.Errorf("League creation failed: %v", lErr)
	}

	// Log result
	t.Log("-----------------------")
	t.Log("Create League Response:")
	t.Logf("%+v", lResp)
	t.Log("-----------------------")

	// Create Franchise 1
	fResp, fErr := createFranchise(lResp.LeagueId, userId, franchiseName, franchiseFoundationYear)
	if fErr != nil {
		t.Errorf("League creation failed: %v", fErr)
	}

	// Log result
	t.Log("----------------------------")
	t.Log("Create Franchise 1 Response:")
	t.Logf("%+v", fResp)
	t.Log("----------------------------")

	// Test franchise 1 creation
	if fResp.Status != 201 {
		t.Errorf("Http Status %d not equal to expected status %d", fResp.Status, 201)
	}

	if fId, e := uuid.Parse(fResp.FranchiseId); e != nil {
		t.Errorf("Output %q expected to be an uuid", fId)
	}

	// Create Franchise in non existing league
	fResp2, fErr2 := createFranchise(wrongLeagueId, userId, franchiseName2, franchiseFoundationYear)
	if fErr2 != nil {
		t.Errorf("League creation failed: %v", fErr2)
	}

	// Log result
	t.Log("----------------------------")
	t.Log("Create Franchise 2 Response:")
	t.Logf("%+v", fResp2)
	t.Log("----------------------------")

	// Test franchise 2 creation
	if fResp2.Status != 409 {
		t.Errorf("Http Status %d not equal to expected status %d", fResp2.Status, 409)
	}
	expectedError2 := fmt.Sprintf("Franchise cannot be created, provided leagueId (%s) does not exist", wrongLeagueId)
	if fResp2.Error != expectedError2 {
		t.Errorf("Output %q not equal to expected %q", fResp2.Error, expectedError2)
	}

	// Create Franchise in non existing league
	fResp3, fErr3 := createFranchise(lResp.LeagueId, userId, franchiseName2, franchiseFoundationYear)
	if fErr3 != nil {
		t.Errorf("League creation failed: %v", fErr3)
	}

	// Log result
	t.Log("----------------------------")
	t.Log("Create Franchise 3 Response:")
	t.Logf("%+v", fResp3)
	t.Log("----------------------------")

	// Test franchise 3 creation
	if fResp3.Status != 409 {
		t.Errorf("Http Status %d not equal to expected status %d", fResp3.Status, 409)
	}
	expectedError3 := fmt.Sprintf("Franchise cannot be created, maximum number of franchises already created (%d)", maxFranchises)
	if fResp3.Error != expectedError3 {
		t.Errorf("Output %q not equal to expected %q", fResp3.Error, expectedError3)
	}

	// Create Franchiser 4 existing franchise in league
	fResp4, fErr4 := createFranchise(lResp.LeagueId, userId, franchiseName, franchiseFoundationYear)
	if fErr4 != nil {
		t.Errorf("League creation failed: %v", fErr4)
	}

	// Log result
	t.Log("----------------------------")
	t.Log("Create Franchise 3 Response:")
	t.Logf("%+v", fResp4)
	t.Log("----------------------------")

	// Test franchise 2 creation
	if fResp4.Status != 409 {
		t.Errorf("Http Status %d not equal to expected status %d", fResp4.Status, 409)
	}
	expectedError4 := fmt.Sprintf("Franchise cannot be created, franchise with name (%s) already exisits in this league", franchiseName)
	if fResp4.Error != expectedError4 {
		t.Errorf("Output %q not equal to expected %q", fResp4.Error, expectedError4)
	}

	// Clean up
	db.Where("franchise_owner = ?", userId).Delete(&models.Franchise{})
	db.Where("league_founder = ?", userId).Delete(&models.League{})

}

func TestGetLeague(t *testing.T) {
	// Create League 1
	lResp, lErr := createLeague(userId, leagueName, foundationYear, maxFranchises)
	if lErr != nil {
		t.Errorf("League creation failed: %v", lErr)
	}

	// Log result
	t.Log("-----------------------")
	t.Log("Create League 1 Response:")
	t.Logf("%+v", lResp)
	t.Log("-----------------------")

	// Create Franchise 1
	fResp, fErr := createFranchise(lResp.LeagueId, userId, franchiseName, franchiseFoundationYear)
	if fErr != nil {
		t.Errorf("League creation failed: %v", fErr)
	}

	// Log result
	t.Log("----------------------------")
	t.Log("Create Franchise 1 Response:")
	t.Logf("%+v", fResp)
	t.Log("----------------------------")

	getLeagueReq := pb.GetLeagueRequest{LeagueId: lResp.LeagueId}
	resp, err := client.GetLeague(ctx, &getLeagueReq)
	if err != nil {
		t.Fatalf("Get league failed: %v", err)
	}
	t.Log("---------------------------------------------")
	log.Printf("Get League Response: %v", resp)
	t.Log("---------------------------------------------")

	// TODO: Better test
	if resp.Status != http.StatusAccepted {
		t.Errorf("Http Status %d not equal to expected status %d", resp.Status, http.StatusAccepted)
	}

	if resp.Result.ID != lResp.LeagueId {
		t.Errorf("Output %q not equal to expected %q", resp.Result.ID, lResp.LeagueId)
	}

	db.Where("franchise_owner = ?", userId).Delete(&models.Franchise{})
	db.Where("league_founder = ?", userId).Delete(&models.League{})
}

func TestGetLeagueFranchisePairs(t *testing.T) {
	// Create League 1
	lResp, lErr := createLeague(userId, leagueName, foundationYear, maxFranchises2)
	if lErr != nil {
		t.Errorf("League creation failed: %v", lErr)
	}

	// Log result
	t.Log("-----------------------")
	t.Log("Create League Response:")
	t.Logf("%+v", lResp)
	t.Log("-----------------------")

	// Create league 2
	lResp2, lErr2 := createLeague(userId2, leagueName2, foundationYear, maxFranchises2)
	if lErr2 != nil {
		t.Errorf("League creation failed: %v", lErr2)
	}

	// Log result
	t.Log("-----------------------")
	t.Log("Create League Response:")
	t.Logf("%+v", lResp2)
	t.Log("-----------------------")

	// Create franchise 1 in league 1
	fResp, fErr := createFranchise(lResp.LeagueId, userId, franchiseName, franchiseFoundationYear)
	if fErr != nil {
		t.Errorf("League creation failed: %v", fErr)
	}

	// Log result
	t.Log("----------------------------")
	t.Log("Create Franchise 1 Response:")
	t.Logf("%+v", fResp)
	t.Log("----------------------------")

	// Create franchise 2 in league 1
	fResp2, fErr2 := createFranchise(lResp.LeagueId, userId2, franchiseName2, franchiseFoundationYear)
	if fErr2 != nil {
		t.Errorf("League creation failed: %v", fErr2)
	}

	// Log result
	t.Log("----------------------------")
	t.Log("Create Franchise 2 Response:")
	t.Logf("%+v", fResp2)
	t.Log("----------------------------")

	// Get leagues for user 1
	getLeaguesForUserReq := pb.GetLeagueFranchisePairsRequest{UserId: userId}
	result1, lfErr1 := client.GetLeagueFranchisePairs(ctx, &getLeaguesForUserReq)
	if lfErr1 != nil {
		t.Fatalf("Get all leagues failed: %v", lfErr1)
	}
	log.Printf("Get Leagues Response: %v", result1.Result)

	// Test franchise
	if result1.Status != http.StatusAccepted {
		t.Errorf("Http Status %d not equal to expected status %d", result1.Status, http.StatusAccepted)
	}
	actualLeagueId1 := result1.Result[0].LeagueID
	actualFranchiseId1 := result1.Result[0].FranchiseID

	if actualLeagueId1 != lResp.LeagueId {
		t.Errorf("LeagueId %q not equal to expected %q", actualLeagueId1, lResp.LeagueId)
	}

	if actualFranchiseId1 != fResp.FranchiseId {
		t.Errorf("FranchiseId %q not equal to expected %q", actualFranchiseId1, fResp.FranchiseId)
	}

	// Create franchise 3 in league 1
	fResp3, fErr3 := createFranchise(lResp.LeagueId, userId, franchiseName3, franchiseFoundationYear)
	if fErr3 != nil {
		t.Errorf("League creation failed: %v", fErr3)
	}

	// Log result
	t.Log("----------------------------")
	t.Log("Create Franchise 2 Response:")
	t.Logf("%+v", fResp3)
	t.Log("----------------------------")

	// Get leagues for user 1
	result2, lfErr2 := client.GetLeagueFranchisePairs(ctx, &getLeaguesForUserReq)
	if lfErr2 != nil {
		t.Fatalf("Get all leagues failed: %v", lfErr2)
	}
	t.Log("---------------------------------------------")
	log.Printf("Get Leagues Response: %v", result2.Result)
	t.Log("---------------------------------------------")

	// Get leagues for user 2
	getLeaguesForUserReq2 := pb.GetLeagueFranchisePairsRequest{UserId: userId2}
	result3, lfErr3 := client.GetLeagueFranchisePairs(ctx, &getLeaguesForUserReq2)
	if lfErr2 != nil {
		t.Fatalf("Get all leagues failed: %v", lfErr3)
	}
	t.Log("---------------------------------------------")
	log.Printf("Get Leagues Response: %v", result3.Result)
	t.Log("---------------------------------------------")

	if result3.Status != http.StatusAccepted {
		t.Errorf("Http Status %d not equal to expected status %d", result3.Status, http.StatusAccepted)
	}
	actualLeagueId3 := result3.Result[1].LeagueID
	actualFranchiseId3 := result3.Result[1].FranchiseID

	if actualLeagueId3 != lResp2.LeagueId {
		t.Errorf("LeagueId %q not equal to expected %q", actualLeagueId3, lResp2.LeagueId)
	}

	if actualFranchiseId3 != "" {
		t.Errorf("FranchiseId %q not equal to expected %q", actualFranchiseId3, "")
	}

	// Clean up
	db.Where("franchise_owner = ?", userId).Delete(&models.Franchise{})
	db.Where("franchise_owner = ?", userId2).Delete(&models.Franchise{})
	db.Where("league_founder = ?", userId).Delete(&models.League{})
	db.Where("league_founder = ?", userId2).Delete(&models.League{})

}
