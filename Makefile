BINARY_NAME=RandomSoundPlayer
SRC_DIR=./cmd/RandomSoundPlayer
RES_OUT=cmd/RandomSoundPlayer/rsrc
LDFLAGS=-ldflags "-H windowsgui"
BUILD_DIR=bin

run:
	go run ./cmd/RandomSoundPlayer

build:
	go-winres make --out $(RES_OUT)
	go build -ldflags "-H windowsgui" -o bin/RandomSoundPlayer.exe ./cmd/RandomSoundPlayer

build-all: clean
	go-winres make --arch="amd64,arm64,arm,386" --out $(RES_OUT)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe $(SRC_DIR)
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_arm64.exe $(SRC_DIR)
	GOOS=windows GOARCH=arm   go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_arm.exe   $(SRC_DIR)
	GOOS=windows GOARCH=386   go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_386.exe   $(SRC_DIR)

	rm -f $(SRC_DIR)/rsrc*.syso $(SRC_DIR)/rsrc*.rc $(RES_OUT)*.syso $(RES_OUT)*.rc 2>/dev/null || true

clean:
	rm -rf $(BUILD_DIR)/*
	rm -f $(SRC_DIR)/rsrc*.syso $(SRC_DIR)/rsrc*.rc $(RES_OUT)*.syso $(RES_OUT)*.rc 2>/dev/null || true