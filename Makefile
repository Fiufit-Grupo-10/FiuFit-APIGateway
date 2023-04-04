.PHONY: build

default: build

unit-tests:
	./scripts/unit_tests.sh

race-condition-tests:
	go test -v -race ./...

docker-tests:
	docker-compose build
	docker-compose run --rm unit-tests
	docker-compose down

build: docker-tests

tidy:
	go mod tidy -v

down:
	docker-compose down

air:
	./scripts/air_dev.sh
