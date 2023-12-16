package main

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2/dialog"
	"image"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const scaleFactor = 2
const skinsPrefix = "./skins"

func stackFromFromDefinitions() (*SpriteStack, error) {
	stack := SpriteStack{
		prefix:        skinsPrefix,
		skinName:      "default",
		fileCache:     map[string]image.Image{},
		actionHandler: map[string]func(){},
	}
	f, err := os.Open("./sprites.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&stack); err != nil {
		return nil, err
	}
	return &stack, nil

}

func main() {
	a := app.New()
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
	stack, err := stackFromFromDefinitions()
	if err != nil {
		panic(err)
	}

	textLayer := &TextLayer{}
	ts := &TextSprite{
		Text:              "CLICK EJECT BUTTON",
		File:              "skin/text.bmp",
		StrLen:            27,
		Marquee:           false,
		RenderedText:      nil,
		Image:             nil,
		AbsolutePositionX: 110,
		AbsolutePositionY: 28,
	}
	if err := ts.Load(skinsPrefix, "default"); err != nil {
		panic(err)
	}
	textLayer.sprites = append(textLayer.sprites, ts)

	timeM := &TextSprite{
		Text:              "00",
		File:              "skin/numbers.bmp",
		Numeric:           true,
		CharSpacing:       1,
		StrLen:            2,
		Marquee:           false,
		RenderedText:      nil,
		Image:             nil,
		AbsolutePositionX: 50,
		AbsolutePositionY: 26,
	}
	if err := timeM.Load(skinsPrefix, "default"); err != nil {
		panic(err)
	}
	timeS := &TextSprite{
		Text:              "00",
		File:              "skin/numbers.bmp",
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

	stack.register("close", func() { w.Close() })
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

	stack.register("STOP", func() {
		player.Stop()
	})

	stack.register("PLAY", func() {
		err = player.Play()
		if err != nil {
			panic(err)
		}
	})
	stack.register("PAUSE", func() {
		player.TogglePause()
	})
	stack.register("EJECT", func() {
		dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
			defer uri.Close()
			player.LoadFile(uri.URI().Path())
			fileName := filepath.Base(uri.URI().Path())
			ts.Set(strings.ToUpper(fileName))
			widget.Refresh()
		}, w).Show()
		//NewFileOpen(callback func(fyne.URIReadCloser, error), parent fyne.Window) *FileDialog
	})

	w.ShowAndRun()
}
