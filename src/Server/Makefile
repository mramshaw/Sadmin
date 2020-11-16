GOOS		:= linux
GOARCH		:= amd64
GO111MODULE	:= on

MAIN		:= admin_server

.PHONY:		clean

all:		$(MAIN)
		@echo '$(MAIN)' has been started

$(MAIN):	run

# .go files are reformatted to conform to gofmt standards
fmt:
		GOOS=$(GOOS) GOARCH=$(GOARCH) gofmt -d -e -s -w *.go
		GOOS=$(GOOS) GOARCH=$(GOARCH) gofmt -d -e -s -w ./application/*.go
		GOOS=$(GOOS) GOARCH=$(GOARCH) gofmt -d -e -s -w ./servers/*.go
		GOOS=$(GOOS) GOARCH=$(GOARCH) gofmt -d -e -s -w ./test/*.go

lint:		fmt
		GOOS=$(GOOS) GOARCH=$(GOARCH) golint -set_exit_status *.go
		GOOS=$(GOOS) GOARCH=$(GOARCH) golint -set_exit_status ./application/*.go
		GOOS=$(GOOS) GOARCH=$(GOARCH) golint -set_exit_status ./servers/*.go
		GOOS=$(GOOS) GOARCH=$(GOARCH) golint -set_exit_status ./test/*.go

init:		lint
		@rm -f go.mod
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go mod init admin-server
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go mod download

vet:		init
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go vet *.go
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go vet ./application/*.go
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go vet ./servers/*.go
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go vet ./test/*.go

test:		vet
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go test -coverpkg admin-server,admin-server/application,admin-server/servers -coverprofile=coverage.txt -covermode=atomic -v ./...
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go tool cover -html=coverage.txt -o coverage.html

build:		test
		GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=$(GO111MODULE) go build -o ../../compiled/$(MAIN) main.go

run:		build
		../../compiled/$(MAIN)

clean:
		rm -f go.mod go.sum ../../compiled/$(MAIN) coverage.html coverage.txt
