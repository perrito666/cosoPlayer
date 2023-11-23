package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
)

type SpriteStack struct {
	sprites       []*AnimatedSprite
	fileCache     map[string]image.Image
	prefix        string
	skinName      string
	actionHandler map[string]func()
}

func (s *SpriteStack) DrawAtPosition(x, y int) color.Color {
	for i := range s.sprites {
		s := s.sprites[len(s.sprites)-i-1]
		if s.Collision(x, y) {
			return s.At(x, y)
		}
	}
	return nil
}

func (s *SpriteStack) DoAtPosition(x, y int) {
	for i := range s.sprites {
		sp := s.sprites[len(s.sprites)-i-1]
		if sp.Collision(x, y) {
			if sp.Action != "" && s.actionHandler != nil {
				if fn, ok := s.actionHandler[sp.Action]; ok {
					fn()
				}
				return
			}
			fmt.Printf("Tapped: %s\n", sp.ID)
			return
		}
	}
}

func (s *SpriteStack) UnmarshalJSON(data []byte) error {
	var tgt []*AnimatedSprite
	if err := json.Unmarshal(data, &tgt); err != nil {
		return err
	}

	for _, sprite := range tgt {
		if err := sprite.Load(s.prefix, s.skinName, s.fileCache); err != nil {
			return err
		}
	}

	s.sprites = tgt
	return nil
}

// AnimatedSprite is a sprite that can contain two states (regular and down), the sprites for both are contained in
// image and downimage attributes respectively.
type AnimatedSprite struct {
	ID                string  `json:"id"`
	Action            string  `json:"action"`
	AbsolutePositionX int     `json:"absolutePositionX"`
	AbsolutePositionY int     `json:"absolutePositionY"`
	Image             Sprite  `json:"image"`
	DownImage         *Sprite `json:"downimage"`
	Tooltip           string  `json:"tooltip"`
}

// Sprite is a single sprite, they are used in the context of an AnimatedSprite as one of the two frames
type Sprite struct {
	ID              string      `json:"id"`
	File            string      `json:"file"`
	Image           image.Image `json:"-"`
	SpritePositionX int         `json:"spritePositionX"`
	SpritePositionY int         `json:"spritePositionY"`
	SpriteHeight    int         `json:"spriteHeight"`
	SpriteWidth     int         `json:"spriteWidth"`
}

func (s *AnimatedSprite) Collision(x, y int) bool {
	inX := x > s.AbsolutePositionX && x < s.AbsolutePositionX+s.Image.SpriteWidth
	inY := y > s.AbsolutePositionY && y < s.AbsolutePositionY+s.Image.SpriteHeight
	return inX && inY
}

func (s *AnimatedSprite) Load(prefix, skinName string, fileCache map[string]image.Image) error {
	if err := s.Image.Load(prefix, skinName, fileCache); err != nil {
		return err
	}
	if s.DownImage != nil {
		if err := s.DownImage.Load(prefix, skinName, fileCache); err != nil {
			return err
		}
	}
	return nil
}

func (s *Sprite) Load(prefix, skinName string, fileCache map[string]image.Image) error {
	rawImg, ok := fileCache[s.File]

	// I suspect that, due to this was done for fat32, the skins contain uppercase filenames.
	fName := strings.Replace(strings.ToUpper(s.File), "SKIN", skinName, 1)
	if !ok {
		f, err := os.Open(filepath.Join(prefix, fName))
		if err != nil {
			return err
		}
		defer f.Close()
		rawImg, _, err = image.Decode(f)
		if err != nil {
			return err
		}
		fileCache[s.File] = rawImg
	}
	s.Image = rawImg
	return nil
}

func (s *AnimatedSprite) ColorModel() color.Model {
	return s.Image.Image.ColorModel()
}

func (s *AnimatedSprite) Bounds() image.Rectangle {
	return s.Image.Image.Bounds()
}

func (s *Sprite) At(x, y int) color.Color {
	return s.Image.At(x+s.SpritePositionX, y+s.SpritePositionY)
}

func (s *AnimatedSprite) At(x, y int) color.Color {
	return s.Image.At(x-s.AbsolutePositionX, y-s.AbsolutePositionY)
}

var _ image.Image = (*AnimatedSprite)(nil)
