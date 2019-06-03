
ROOT=$(shell pwd)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
MAINDIR=$(ROOT)/cmd
BUILDOUT=$(ROOT)/bin
BINARY_NAME=epVizSrv
LOG_FILENAME=log
CONFIG_FILENAME=config.json

up: deps build start

build: 
	mkdir -p $(BUILDOUT)
	rm -f $(BUILDOUT)/$(LOG_FILENAME)
	cd $(MAINDIR) && $(GOBUILD) -o $(BUILDOUT)/$(BINARY_NAME) -v && cd $(ROOT)
	chmod 777 $(BUILDOUT)/$(BINARY_NAME)
	cp $(MAINDIR)/$(CONFIG_FILENAME) $(BUILDOUT)

start:
	cd $(BUILDOUT) && $(BUILDOUT)/$(BINARY_NAME) && cd $(ROOT)

clean:
	$(GOCLEAN)
	rm -R $(BUILDOUT)

deps:
	dep ensure -v

