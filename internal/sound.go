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

type Sound string

func LoadSoundAt(soundPath string) (Sound, error) {
	if !slices.Contains(supported_filetypes, filepath.Ext(soundPath)) {
		return "", fmt.Errorf(
			"File \"%v\" has unsupported filetype \"%v\". Please use one of the supported filetypes: %v",
			soundPath, filepath.Ext(soundPath), supported_filetypes,
		)
	}
	return Sound(soundPath), nil
}

func (sound Sound) Name() string {
	return filepath.Base(string(sound))
}

func (sound Sound) Decode() (beep.StreamSeekCloser, beep.Format, error) {
	fileExtension := filepath.Ext(string(sound))

	soundFile, err := os.Open(string(sound))
	if err != nil {
		return nil, beep.Format{}, err
	}

	streamer, format, err := decoders[fileExtension](soundFile)
	if err != nil {
		return nil, beep.Format{}, err
	}

	return streamer, format, nil
}
