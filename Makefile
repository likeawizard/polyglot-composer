build:
	go build -o polyglot-composer cmd/polyglot-composer/main.go

build-texel:
	go build -o texel-data cmd/texel-data/main.go

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix
