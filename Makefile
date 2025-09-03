

GOPATH := $(shell go env GOPATH)

LDFLAGS := -s -w

CMD := book
BIN := bin

BUILD_DEPS := clean mod fmt test

all: build

fmt:
	gofmt -s -w .

mod:
	go mod tidy

test:
	go test -v -cover

build: $(BUILD_DEPS)
	mkdir -p $(BIN)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN)/$(CMD) .

install: build
	mkdir -p $(GOPATH)/bin/
	cp $(BIN)/$(CMD) $(GOPATH)/bin/$(CMD)

clean:
	go clean -testcache
	rm -rf $(BIN)/ *.docx *.pdf
