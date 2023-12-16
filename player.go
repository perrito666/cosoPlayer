package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

type Player struct {
	currentSong       string
	otoContext        *oto.Context
	player            *oto.Player
	currentSongLength int64
	tickAction        func(elapsed, total uint64) error
	lastStart         time.Time
	pauseTime         time.Time
	playChan          chan struct{}
}

var singlePlayer *Player

const sampleRate = 44100

func NewPlayer(tickAction func(elapsed, total uint64) error) (*Player, error) {
	if singlePlayer != nil {
		return singlePlayer, nil
	}
	// Prepare an Oto context (this will use your default audio device) that will
	// play all our sounds. Its configuration can't be changed later.

	op := &oto.NewContextOptions{}

	// Usually 44100 or 48000. Other values might cause distortions in Oto
	op.SampleRate = sampleRate

	// Number of channels (aka locations) to play sounds from. Either 1 or 2.
	// 1 is mono sound, and 2 is stereo (most speakers are stereo).
	op.ChannelCount = 2

	// Format of the source. go-mp3's format is signed 16bit integers.
	op.Format = oto.FormatSignedInt16LE

	// Remember that you should **not** create more than one context
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("oto.NewContext failed: %w", err)
	}
	// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
	<-readyChan
	singlePlayer = &Player{
		currentSong: "",
		otoContext:  otoCtx,
		tickAction:  tickAction,
		playChan:    make(chan struct{}),
	}
	return singlePlayer, nil
}

func (p *Player) PlayerLoop() {
	println("loop invoked")
	for range p.playChan {
		println("loop")
		for {
			if p.player == nil || !p.player.IsPlaying() {
				break
			}
			err := p.tickAction(uint64(time.Since(p.lastStart).Seconds()), uint64(p.currentSongLength))
			if err != nil {
				p.player.Pause()
			}
			time.Sleep(time.Second)
		}
	}
	println("end loop")
}

const sampleSize = 4

func (p *Player) LoadFile(song string) error {
	if p.player != nil {
		if err := p.player.Close(); err != nil {
			return fmt.Errorf("closing previous player: %w", err)
		}
	}
	fileBytes, err := os.ReadFile(song)
	if err != nil {
		return fmt.Errorf("reading %q failed: %w", song, err)
	}

	// Convert the pure bytes into a reader object that can be used with the mp3 decoder
	fileBytesReader := bytes.NewReader(fileBytes)

	// Decode file
	decodedMp3, err := mp3.NewDecoder(fileBytesReader)
	if err != nil {
		return fmt.Errorf("mp3.NewDecoder failed: :%w", err)
	}
	p.currentSongLength = (decodedMp3.Length() / sampleSize) / sampleRate

	// Create a new 'player' that will handle our sound. Paused by default.
	p.player = singlePlayer.otoContext.NewPlayer(decodedMp3)
	p.currentSong = song
	return nil
}

func (p *Player) Stop() error {
	if p.player == nil {
		return nil
	}
	err := p.player.Close()
	p.player = nil
	return err
}

func (p *Player) Play() error {
	if p.currentSong == "" || (p.player != nil && p.player.IsPlaying()) {
		return nil
	}
	if !p.pauseTime.IsZero() {
		p.TogglePause()
		return nil
	}
	// Play starts playing the sound and returns without waiting for it (Play() is async).
	p.player.Play()
	p.lastStart = time.Now()
	p.playChan <- struct{}{}
	return nil
}

func (p *Player) TogglePause() {
	if p.currentSong == "" {
		return
	}
	if p.player.IsPlaying() {
		p.pauseTime = time.Now()
		p.player.Pause()
		return
	}
	p.lastStart = time.Now().Add(-(p.pauseTime.Sub(p.lastStart)))
	fmt.Printf("Pause time: %#v\n", p.pauseTime)
	fmt.Printf("Play start time: %#v\n", p.lastStart)
	fmt.Printf("Pause - start diff: %#v\n", p.pauseTime.Sub(p.lastStart))
	p.player.Play()
	p.playChan <- struct{}{}
	p.pauseTime = time.Time{}
}
