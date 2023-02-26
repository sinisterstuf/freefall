// Copyright 2022 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package assets

import (
	"embed"
	"encoding/json"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"path"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/tinne26/etxt"
)

//go:embed *.png
var assets embed.FS

// Frame is a single frame of an animation, usually a sub-image of a larger
// image containing several frames
type Frame struct {
	Duration int           `json:"duration"`
	Position FramePosition `json:"frame"`
}

// FramePosition represents the position of a frame, including the top-left
// coordinates and its dimensions (width and height)
type FramePosition struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// FrameTags contains tag data about frames to identify different parts of an
// animation, e.g. idle animation, jump animation frames etc.
type FrameTags struct {
	Name      string `json:"name"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Direction string `json:"direction"`
}

// Frames is a slice of frames used to create sprite animation
type Frames []Frame

// SpriteMeta contains sprite meta-data, basically everything except frame data
type SpriteMeta struct {
	ImageName string      `json:"image"`
	FrameTags []FrameTags `json:"frameTags"`
}

// SpriteSheet is the root-node of sprite data, it contains frames and meta data
// about them
type SpriteSheet struct {
	Sprite Frames     `json:"frames"`
	Meta   SpriteMeta `json:"meta"`
	Image  *ebiten.Image
}

// Load a sprite image and associated meta-data given a file name (without
// extension)
func loadSprite(name string) *SpriteSheet {
	name = path.Join("assets", "sprites", name)
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name + ".json")
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	var ss SpriteSheet
	json.Unmarshal(data, &ss)
	if err != nil {
		log.Fatal(err)
	}

	ss.Image = LoadImage(name + ".png")

	return &ss
}

// Load an image from embedded FS into an ebiten Image object
func LoadImage(name string) *ebiten.Image {
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name)
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	raw, err := png.Decode(file)
	if err != nil {
		log.Fatalf("error decoding file %s as PNG: %v\n", name, err)
	}

	return ebiten.NewImageFromImage(raw)
}

// Sound stores and plays all the sound variants for one single soundType
type Sound struct {
	Audio      []*audio.Player
	LastPlayed int
	Volume     float64
}

// AddMusic adds one new music to the soundType
func (s *Sound) AddMusic(f string, sampleRate int, context *audio.Context) {
	s.Audio = append(s.Audio, NewMusicPlayer(LoadSoundFile(f+".ogg", sampleRate), context))
}

// AddSound adds one new sound to the soundType
func (s *Sound) AddSound(f string, sampleRate int, context *audio.Context, v ...int) {
	var filename string

	variants := 1
	if len(v) > 0 {
		variants = v[0]
	}

	for i := 0; i < variants; i++ {
		if variants == 1 {
			filename = f + ".ogg"
		} else {
			filename = f + "-" + strconv.Itoa(i+1) + ".ogg"
		}

		s.Audio = append(s.Audio, NewSoundPlayer(LoadSoundFile(filename, sampleRate), context))
	}
}

// SetVolume sets the volume of the audio
func (s *Sound) SetVolume(v float64) {
	if v >= 0 && v <= 1 {
		s.Volume = v
	}
}

// Play plays the audio or a random one if there are more
func (s *Sound) Play() {
	length := len(s.Audio)
	index := 0

	if length == 0 {
		return
	} else if length > 1 {
		index = rand.Intn(length)
	}

	s.PlayVariant(index)
}

// PlayVariant plays the selected audio
func (s *Sound) PlayVariant(i int) {
	s.LastPlayed = i
	s.Audio[i].SetVolume(s.Volume)
	s.Audio[i].Rewind()
	s.Audio[i].Play()
}

// Pause pauses the audio being played
func (s *Sound) Pause() {
	s.Audio[s.LastPlayed].Pause()
}

// Sounds is a slice of sounds
type Sounds []*Sound

// NewMusicPlayer loads a sound into an audio player that can be used to play it
// as an infinite loop of music without any additional setup required
func NewMusicPlayer(music *vorbis.Stream, context *audio.Context) *audio.Player {
	musicLoop := audio.NewInfiniteLoop(music, music.Length())
	musicPlayer, err := audio.NewPlayer(context, musicLoop)
	if err != nil {
		log.Fatalf("error making music player: %v\n", err)
	}
	return musicPlayer
}

// NewSoundPlayer loads a sound into an audio player that can be used to play it
// without any additional setup required
func NewSoundPlayer(audioFile *vorbis.Stream, context *audio.Context) *audio.Player {
	audioPlayer, err := audio.NewPlayer(context, audioFile)
	if err != nil {
		log.Fatalf("error making audio player: %v\n", err)
	}
	return audioPlayer
}

// Load an OGG Vorbis sound file with 44100 sample rate and return its stream
func LoadSoundFile(name string, sampleRate int) *vorbis.Stream {
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name)
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	music, err := vorbis.DecodeWithSampleRate(sampleRate, file)
	if err != nil {
		log.Fatalf("error decoding file %s as Vorbis: %v\n", name, err)
	}

	return music
}

// Load a font for use with etxt specified by font name
func LoadFont(name string) *etxt.Font {
	font, fname, err := etxt.ParseEmbedFontFrom(name, assets)
	if err != nil {
		log.Fatalf("error parsing font %s: %v", name, err)
	}

	log.Println("loaded font:", fname)
	return font
}
