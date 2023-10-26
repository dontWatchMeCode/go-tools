run:
	go run main.go

build:
	go build -o bin/main main.go

clean:
	go get -u all  && go clean && gofmt -w .

install:
	go install