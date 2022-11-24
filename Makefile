proto:
	protoc service/pb/*.proto --go_out=. --go-grpc_out=.

build:
	go build -ldflags "-X github.com/hiltpold/lakelandcup-fantasy-service/commands.Version=`git rev-parse HEAD`"

fantasy-service-dev:
	go run main.go -c .dev.env
