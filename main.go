package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"

	"fyne.io/fyne/v2/app"
)

const scaleFactor = 2

func stackFromFromDefinitions(skin *Skin) (*SpriteStack, error) {
	stack := SpriteStack{
		skin:          skin,
		fileCache:     map[string]image.Image{},
		actionHandler: map[string]func() error{},
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
	if len(os.Args) == 1 {
		fmt.Println(errors.New("a path to a skin is expected"))
		os.Exit(1)
	}
	skin, err := skinFromPath(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	a := app.New()
	w, err := mainWindow(a, skin)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	w.ShowAndRun()
}
