GOOS		:= linux
GOARCH		:= amd64
GO111MODULE	:= on

MAIN		:= admin_client

.PHONY:		clean

all:		$(MAIN)
		@echo '$(MAIN)' has been started

$(MAIN):	run

# .go files are reformatted to conform to gofmt standards
fmt:
		GOOS=$(GOOS) GOARCH=$(GOARCH) gofmt -d -e -s -w *.go

lint:		fmt
		GOOS=$(GOOS) GOARCH=$(GOARCH) golint -set_exit_status *.go

vet:		lint
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go vet *.go

init:		vet
		@rm -f go.mod
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go mod init admin-client

test:		init
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go test -coverprofile=coverage.txt -covermode=atomic -v .
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go tool cover -html=coverage.txt -o coverage.html

build:		test
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go build -o ../../compiled/$(MAIN) main.go  server_validate.go

run:		build
		../../compiled/$(MAIN)

clean:
		rm -f go.mod go.sum ../../compiled/$(MAIN) coverage.html coverage.txt
