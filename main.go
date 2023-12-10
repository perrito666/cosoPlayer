package main

import (
	"encoding/json"
	"image"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
)

const scaleFactor = 2
const skinsPrefix = "./skins"

func stackFromFromDefinitions() (*SpriteStack, error) {
	stack := SpriteStack{
		prefix:    skinsPrefix,
		skinName:  "default",
		fileCache: map[string]image.Image{},
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
	// w := a.NewWindow("It really whips the guanaco's ass!!!")

	drv, ok := a.Driver().(desktop.Driver)
	if !ok {
		panic("driver is not a driver")
	}
	w := drv.CreateSplashWindow()
	w.SetFixedSize(true)
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
		Text:              "CD TRACK 5.MP3",
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
	mainWindowBG := &Background{
		stack:     stack,
		textLayer: textLayer,
	}

	stack.register("close", func() { w.Close() })

	w.SetContent(newBgWidget(mainWindowBG))

	w.ShowAndRun()
}
