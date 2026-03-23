BINARY=mini-url
DB=urls.db

.PHONY: build run test clean migrate

build:
	go build -o $(BINARY) ./cmd/mini-url

run: build
	./$(BINARY)

test:
	go test ./... 

clean:
	rm -f $(BINARY) $(DB)

migrate:
	# sqlite migrations are simple; ensure DB exists and schema is created by init
	go run ./cmd/mini-url
