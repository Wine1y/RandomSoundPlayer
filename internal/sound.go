package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

var supported_filetypes = []string{".wav", ".mp3", ".ogg", ".flac"}
var decoders = map[string]func(*os.File) (beep.StreamSeekCloser, beep.Format, error){
	".wav":  DecodeWAV,
	".mp3":  DecodeMP3,
	".ogg":  DecodeOGG,
	".flac": DecodeFLAC,
}

func DecodeWAV(file *os.File) (beep.StreamSeekCloser, beep.Format, error) {
	return wav.Decode(file)
}

func DecodeMP3(file *os.File) (beep.StreamSeekCloser, beep.Format, error) {
	return mp3.Decode(file)
}

func DecodeOGG(file *os.File) (beep.StreamSeekCloser, beep.Format, error) {
	return vorbis.Decode(file)
}

func DecodeFLAC(file *os.File) (beep.StreamSeekCloser, beep.Format, error) {
	return flac.Decode(file)
}

type UnsupportedFiletypeError struct {
	filePath string
}

func (err UnsupportedFiletypeError) Error() string {
	return fmt.Sprintf(
		"File \"%v\" has unsupported filetype \"%v\". Please use one of the supported filetypes: %v",
		err.filePath, filepath.Ext(err.filePath), supported_filetypes,
	)
}

type Sound struct {
	Name     string
	Streamer beep.StreamSeekCloser
	Format   beep.Format
}

func LoadSoundAt(soundPath string) (Sound, error) {
	if !slices.Contains(supported_filetypes, filepath.Ext(soundPath)) {
		return Sound{}, UnsupportedFiletypeError{filePath: soundPath}
	}
	fileExtension := filepath.Ext(soundPath)

	soundFile, err := os.Open(soundPath)
	if err != nil {
		return Sound{}, err
	}

	streamer, format, err := decoders[fileExtension](soundFile)
	if err != nil {
		return Sound{}, err
	}

	return Sound{Name: filepath.Base(soundPath), Format: format, Streamer: streamer}, nil
}
