BIN_PATH := cgnet
IMAGE := 10.0.0.240/cgnet-exporter
TAG := $(shell git describe --tags --always)

MANIFEST_DIR := manifests/deploy
MANIFEST := $(MANIFEST_DIR)/all-in-one.yaml

VERSION := $(shell git describe --tags --always --dirty)
BINDATA := bpf/bpf-packr.go

build: $(BINDATA)
	go build -i -ldflags "-X github.com/kinvolk/cgnet/cmd.version=$(VERSION)" github.com/kinvolk/cgnet/cmd/cgnet

linux: $(BINDATA)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s" -o $(BIN_PATH) github.com/kinvolk/cgnet/cmd/cgnet 

$(BINDATA):
	@make -C bpf/

image: linux
	docker build -t $(IMAGE):$(TAG) .

manifest:
	@make -C $(MANIFEST_DIR) clean
	@make -C $(MANIFEST_DIR)

clean:
	rm -rf $(BIN_PATH)
	@make -C bpf/ clean

.PHONY: clean build linux image manifest
