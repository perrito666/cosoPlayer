package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func mainWindow(a fyne.App, skin *Skin) (fyne.Window, error) {
	w := a.NewWindow("It really whips the guanaco's ass!!!")
	//drv, ok := a.Driver().(desktop.Driver)
	//if !ok {
	//panic("driver is not a driver")
	//}
	//w := drv.CreateSplashWindow()
	//w.SetFixedSize(true)
	w.SetMaster()
	w.SetPadded(false)

	w.Resize(fyne.Size{
		Width:  275 * scaleFactor,
		Height: 116 * scaleFactor,
	})

	// Load sprites
	stack, err := stackFromFromDefinitions(skin)
	if err != nil {
		return nil, fmt.Errorf("loading stack: %w", err)
	}

	textLayer := &TextLayer{}
	ts := &TextSprite{
		Text:              "CLICK EJECT BUTTON",
		File:              "text.bmp",
		StrLen:            27,
		Marquee:           false,
		RenderedText:      nil,
		Image:             nil,
		AbsolutePositionX: 110,
		AbsolutePositionY: 28,
	}
	if err := ts.Load(skin); err != nil {
		panic(err)
	}
	textLayer.sprites = append(textLayer.sprites, ts)

	timeM := &TextSprite{
		Text:              "00",
		File:              "numbers.bmp",
		Numeric:           true,
		CharSpacing:       1,
		StrLen:            2,
		Marquee:           false,
		RenderedText:      nil,
		Image:             nil,
		AbsolutePositionX: 50,
		AbsolutePositionY: 26,
	}
	if err := timeM.Load(skin); err != nil {
		panic(err)
	}
	timeS := &TextSprite{
		Text:              "00",
		File:              "numbers.bmp",
		Numeric:           true,
		CharSpacing:       1,
		StrLen:            2,
		Marquee:           false,
		RenderedText:      nil,
		Image:             timeM.Image, // it's the same image
		AbsolutePositionX: 80,
		AbsolutePositionY: 26,
	}

	textLayer.sprites = append(textLayer.sprites, timeS)
	textLayer.sprites = append(textLayer.sprites, timeM)

	mainWindowBG := &Background{
		stack:     stack,
		textLayer: textLayer,
	}

	stack.register("close", func() error {
		w.Close()
		return nil
	})
	widget := newBgWidget(mainWindowBG)
	w.SetContent(widget)

	player, err := NewPlayer(func(elapsed, total uint64) error {
		mins := elapsed / 60
		seconds := elapsed - (mins * 60)
		timeM.Set(fmt.Sprintf("%02d", mins))
		timeS.Set(fmt.Sprintf("%02d", seconds))
		perc := (float64(elapsed) * 100.0) / float64(total)
		stack.FindByID("player.slider.seek").DraggableSeek(perc / 100)
		widget.Refresh()
		return nil
	})
	if err != nil {
		panic(err)
	}

	go player.PlayerLoop()

	stack.register("STOP", func() error {
		return player.Stop()
	})

	stack.register("PLAY", func() error {
		err = player.Play()
		if err != nil {
			return fmt.Errorf("playing: %w", err)
		}
		return nil
	})
	stack.register("PAUSE", func() error {
		player.TogglePause()
		return nil
	})
	stack.register("EJECT", func() error {
		dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
			defer uri.Close()
			player.LoadFile(uri.URI().Path())
			fileName := filepath.Base(uri.URI().Path())
			ts.Set(strings.ToUpper(fileName))
			widget.Refresh()
		}, w).Show()
		return nil
	})
	return w, nil
}
