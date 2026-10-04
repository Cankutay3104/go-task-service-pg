run:
	go run cmd/api/main.go

test:
	go test -v -race ./...

clean:
	rm -rf bin/

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/task-service cmd/api/main.go