package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type SoundPack struct {
	Name   string
	Sounds []Sound
}

func LoadSoundPackAt(packPath string) (SoundPack, error) {
	stat, err := os.Stat(packPath)
	if err != nil {
		return SoundPack{}, err
	}
	if !stat.IsDir() {
		return SoundPack{}, fmt.Errorf("Invalid path to sound pack, sound pack path should point to directory, got \"%v\"", packPath)
	}

	entries, err := os.ReadDir(packPath)
	if err != nil {
		return SoundPack{}, err
	}

	sounds := make([]Sound, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		soundPath := filepath.Join(packPath, entry.Name())
		sound, err := LoadSoundAt(soundPath)
		if err != nil {
			slog.Warn("One of the sounds wasn't loaded", "soundPath", soundPath, "err", err)
			continue
		}
		sounds = append(sounds, sound)
	}
	return SoundPack{Name: stat.Name(), Sounds: sounds}, nil
}
