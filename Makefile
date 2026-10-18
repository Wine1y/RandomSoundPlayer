run:
	go run ./cmd/RandomSoundPlayer

build:
	go-winres make --out cmd/RandomSoundPlayer/rsrc
	go build -ldflags "-H windowsgui" -o bin/RandomSoundPlayer.exe ./cmd/RandomSoundPlayer