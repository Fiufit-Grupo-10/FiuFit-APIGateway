.PHONY: build

default: build

tests:
	./scripts/unit_tests.sh

race-condition-tests:
	go test -v -race ./...

docker-tests:
	docker-compose build
	docker-compose run --rm unit-tests
	docker-compose down

build: app

app: tests
	go build -o main ./cmd/main.go

build-docker: docker-tests
	docker build -t apigateway

tidy:
	go mod tidy -v

down:
	docker-compose down

air:
	./scripts/air_dev.sh

clean:
	rm -v ./main
