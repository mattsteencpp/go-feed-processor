# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD)get
BINARY_NAME=main
LIBRARY_NAME=processor
PROJECT_PATH=github.com/mattsteencpp/go-feed-processor
EXE_PATH=$GOPATH/bin

#all: test build
all: build
build:
	$(GOINSTALL) $(PROJECT_PATH)/$(LIBRARY_NAME)
	$(GOINSTALL) $(PROJECT_PATH)/$(BINARY_NAME)
test: 
	$(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	rm -f $(EXE_PATH)/$(BINARY_NAME)
	rm -f $(EXE_PATH)/$(LIBRARY_NAME)
run:
	$(BINARY_NAME)
