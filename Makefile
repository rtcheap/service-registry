test:
	go test ./...

install:
	go mod download

run:
	sh run-local.sh

run-image:
	sh run-image.sh
