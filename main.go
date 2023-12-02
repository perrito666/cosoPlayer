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

func stackFromFromDefinitions() (*SpriteStack, error) {
	stack := SpriteStack{
		prefix:    "./skins",
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
	textLayer.sprites = append(textLayer.sprites, &TextSprite{
		Text:              "CD TRACK 5",
		StrLen:            9,
		Marquee:           false,
		RenderedText:      nil,
		Image:             nil,
		AbsolutePositionX: 40,
		AbsolutePositionY: 60,
	})
	mainWindowBG := &Background{
		stack:     stack,
		textLayer: textLayer,
	}

	stack.register("close", func() { w.Close() })

	w.SetContent(newBgWidget(mainWindowBG))

	w.ShowAndRun()
}
