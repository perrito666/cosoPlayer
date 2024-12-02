package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
)

type SpriteStack struct {
	sprites       []*AnimatedSprite
	fileCache     map[string]image.Image
	skin          *Skin
	actionHandler map[string]func() error
	draggedItem   int
}

func (s *SpriteStack) FindByID(name string) *AnimatedSprite {
	for _, sprite := range s.sprites {
		if sprite.ID == name {
			return sprite
		}
	}
	return nil
}

func (s *SpriteStack) Dragged(event *fyne.DragEvent) {
	x := int(event.Position.X / scaleFactor)
	y := int(event.Position.Y / scaleFactor)
	if s.draggedItem < 0 {
		for i := range s.sprites {
			sp := s.sprites[len(s.sprites)-i-1]
			if sp.Collision(x, y) {
				s.draggedItem = len(s.sprites) - i - 1
				break
			}
		}
		if s.draggedItem < 0 {
			return
		}
	}
	fmt.Println(s.sprites[s.draggedItem].ID)
	if s.sprites[s.draggedItem].DragAble &&
		x > s.sprites[s.draggedItem].MinDragX &&
		x < s.sprites[s.draggedItem].MaxDragX {
		s.sprites[s.draggedItem].AbsolutePositionX = x
		return
	}
}

func (s *SpriteStack) DragEnd() {
	if s.draggedItem < 0 {
		s.sprites[s.draggedItem].dePressed()
	}
	s.draggedItem = -1
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

func (s *SpriteStack) MouseDown(x, y int) {
	for i := range s.sprites {
		sp := s.sprites[len(s.sprites)-i-1]
		if sp.Collision(x, y) {
			sp.pressed()
			return
		}
	}
}

func (s *SpriteStack) DoAtPosition(x, y int) {
	for i := range s.sprites {
		sp := s.sprites[len(s.sprites)-i-1]
		if sp.Collision(x, y) {
			sp.dePressed()
			if sp.Action != "" && s.actionHandler != nil {
				if fn, ok := s.actionHandler[sp.Action]; ok {
					err := fn()
					if err != nil {
						// FIXME: Bubble this up
						fmt.Println(fmt.Errorf("error calling action %s: %v", sp.Action, err))
					}
				}
				return
			}
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
		if err := sprite.Load(s.skin, s.fileCache); err != nil {
			return fmt.Errorf("loading image in: %w", err)
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
	DownImage         *Sprite `json:"downImage"`
	ActiveImage       *Sprite `json:"activeImage"`
	Tooltip           string  `json:"tooltip"`
	ToggleAble        bool    `json:"isToggle"`
	DragAble          bool    `json:"dragAble"`
	MinDragX          int     `json:"minDrag"`
	MaxDragX          int     `json:"maxDrag"`

	// These are not part of the json, they are used to track the state of the sprite
	Pressed bool `json:"-"`
	Toggled bool `json:"-"`
}

func (s *AnimatedSprite) Collision(x, y int) bool {
	inX := x >= s.AbsolutePositionX && x < s.AbsolutePositionX+s.Image.SpriteWidth
	inY := y >= s.AbsolutePositionY && y < s.AbsolutePositionY+s.Image.SpriteHeight
	return inX && inY
}

func (s *AnimatedSprite) Load(skin *Skin, fileCache map[string]image.Image) error {
	if err := s.Image.Load(skin, fileCache); err != nil {
		return fmt.Errorf("loading Sprite: %s in Animated Sprite %s: %w", s.Image.ID, s.ID, err)
	}
	if s.DownImage != nil {
		if err := s.DownImage.Load(skin, fileCache); err != nil {
			return fmt.Errorf("loading DownSprite: %s: %w", s.DownImage.ID, err)
		}
	}
	if s.ActiveImage != nil {
		if err := s.ActiveImage.Load(skin, fileCache); err != nil {
			return fmt.Errorf("loading DownSprite: %s: %w", s.ActiveImage.ID, err)
		}
	}
	return nil
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

func (s *Sprite) Load(skin *Skin, fileCache map[string]image.Image) error {
	rawImg, ok := fileCache[s.File]

	// I suspect that, due to this was done for fat32, the skins contain uppercase filenames.
	if !ok {
		f, err := skin.Open(s.File)
		if err != nil {
			return fmt.Errorf("opening file: %s: %w", s.File, err)
		}
		defer f.Close()
		rawImg, _, err = image.Decode(f)
		if err != nil {
			return fmt.Errorf("decoding image: %s: %w", s.File, err)
		}
		fileCache[s.File] = rawImg
	}
	s.Image = rawImg
	return nil
}

func (s *AnimatedSprite) pressed() {
	s.Pressed = true
}

func (s *AnimatedSprite) dePressed() {
	s.Pressed = false
	// if this is not a toggle is a noop
	s.Toggled = !s.Toggled
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
	posX := x - s.AbsolutePositionX
	posY := y - s.AbsolutePositionY
	if s.Pressed && s.DownImage != nil {
		return s.DownImage.At(posX, posY)
	}
	if s.ToggleAble && s.Toggled && s.ActiveImage != nil {
		return s.ActiveImage.At(posX, posY)
	}
	return s.Image.At(posX, posY)
}

func (s *AnimatedSprite) DraggableSeek(percentage float64) {
	if !s.DragAble || (percentage > 1 || percentage < 0) {
		return
	}
	s.AbsolutePositionX = s.MinDragX + int(float64(s.MaxDragX-s.MinDragX)*percentage)
}

var _ image.Image = (*AnimatedSprite)(nil)
