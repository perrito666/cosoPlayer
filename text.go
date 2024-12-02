package main

import (
	"fmt"
	"image"
	"image/color"
)

type position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

const ellipse = '…'
const charWidth = 5
const charHeight = 6
const numWidth = 9
const numHeight = 13

var charMap = map[rune]position{
	'A':  {0, 0},
	'B':  {5, 0},
	'C':  {10, 0},
	'D':  {15, 0},
	'E':  {20, 0},
	'F':  {25, 0},
	'G':  {30, 0},
	'H':  {35, 0},
	'I':  {40, 0},
	'J':  {45, 0},
	'K':  {50, 0},
	'L':  {55, 0},
	'M':  {60, 0},
	'N':  {65, 0},
	'O':  {70, 0},
	'P':  {75, 0},
	'Q':  {80, 0},
	'R':  {85, 0},
	'S':  {90, 0},
	'T':  {95, 0},
	'U':  {100, 0},
	'V':  {105, 0},
	'W':  {110, 0},
	'X':  {115, 0},
	'Y':  {120, 0},
	'Z':  {125, 0},
	'"':  {130, 0},
	'@':  {135, 0},
	'0':  {0, 6},
	'1':  {5, 6},
	'2':  {10, 6},
	'3':  {15, 6},
	'4':  {20, 6},
	'5':  {25, 6},
	'6':  {30, 6},
	'7':  {35, 6},
	'8':  {40, 6},
	'9':  {45, 6},
	'…':  {50, 6},
	'.':  {55, 6},
	':':  {60, 6},
	'(':  {65, 6},
	')':  {70, 6},
	'-':  {75, 6},
	'\'': {80, 6},
	'!':  {85, 6},
	'_':  {90, 6},
	'+':  {95, 6},
	'\\': {100, 6},
	'/':  {105, 6},
	'[':  {110, 6},
	']':  {115, 6},
	'^':  {120, 6},
	'&':  {125, 6},
	'%':  {130, 6},
	',':  {135, 6},
	'=':  {140, 6},
	'$':  {145, 6},
	' ':  {145, 0},
}

var numberMap = map[rune]position{
	'0': {0, 0},
	'1': {9, 0},
	'2': {18, 0},
	'3': {27, 0},
	'4': {36, 0},
	'5': {45, 0},
	'6': {54, 0},
	'7': {63, 0},
	'8': {72, 0},
	'9': {81, 0},
	' ': {90, 0},
}

type TextSprite struct {
	Text              string `json:"text"`
	Numeric           bool
	CharSpacing       int
	File              string
	StrLen            int  `json:"strLen"`
	Marquee           bool `json:"marquee"`
	RenderedText      []position
	Image             image.Image
	AbsolutePositionX int
	AbsolutePositionY int
}

func (t *TextSprite) Set(text string) {
	if text == t.Text {
		return
	}
	t.Text = text
	t.RenderedText = []position{}
}

func (t *TextSprite) Load(skin *Skin) error {
	// I suspect that, due to this was done for fat32, the skins contain uppercase filenames.
	f, err := skin.Open(t.File)
	if err != nil {
		return fmt.Errorf("opening file: %s: %w", t.File, err)
	}
	defer f.Close()
	rawImg, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("decoding image: %s: %w", t.File, err)
	}

	t.Image = rawImg
	return nil
}

func (t *TextSprite) ColorModel() color.Model {
	return t.Image.ColorModel()
}

func (t *TextSprite) RuneWidth() int {
	if t.Numeric {
		return numWidth
	}
	return charWidth
}

func (t *TextSprite) RuneHeight() int {
	if t.Numeric {
		return numHeight
	}
	return charHeight
}

func (t *TextSprite) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: 0,
		},
		Max: image.Point{
			X: (t.RuneWidth() + t.CharSpacing) * t.StrLen,
			Y: t.RuneHeight(),
		},
	}
}

func (t *TextSprite) At(x, y int) color.Color {
	posX := x - t.AbsolutePositionX
	posY := y - t.AbsolutePositionY
	return t.DrawAtPosition(posX, posY)
}

func (t *TextSprite) DrawAtPosition(x, y int) color.Color {
	var useMap map[rune]position
	if t.Numeric {
		useMap = numberMap
	} else {
		useMap = charMap
	}
	if len(t.RenderedText) == 0 {
		spriteString := make([]position, t.StrLen)
		for i, c := range t.Text {
			if i > t.StrLen-1 {
				break
			}
			if p, ok := useMap[c]; ok {
				spriteString[i] = p
			}
		}
		if !t.Numeric {
			spriteString[t.StrLen-1] = useMap[ellipse]
		}
		t.RenderedText = spriteString
	}
	if x == 0 {
		charPos := t.RenderedText[0]
		return t.Image.At(charPos.X, charPos.Y+y)
	}
	charN := x / (t.RuneWidth() + t.CharSpacing)
	if charN > len(t.RenderedText)-1 || charN > len(t.Text)-1 {
		return nil
	}
	drawableChar := t.RenderedText[charN]
	// position for X is tricky here
	// we know the offset to the beginning of the char within the text by multiplying the charN by the width of a char + spacing
	// then we subtract that from the X position we were asked to draw at we get the position within the character
	xPosInChar := x - (t.RuneWidth()+t.CharSpacing)*charN
	if xPosInChar >= t.RuneWidth() { // this is the case where we're in the spacing
		return nil
	}
	return t.Image.At(drawableChar.X+xPosInChar, drawableChar.Y+y)
}

type TextLayer struct {
	sprites []*TextSprite
}

func (t *TextSprite) Collision(x, y int) bool {
	inX := x >= t.AbsolutePositionX && x < (t.AbsolutePositionX+t.RuneWidth()+t.CharSpacing)*t.StrLen
	inY := y >= t.AbsolutePositionY && y < t.AbsolutePositionY+t.RuneHeight()
	return inX && inY
}

func (t *TextLayer) DrawAtPosition(x, y int) color.Color {
	for i := range t.sprites {
		if t.sprites[i].Collision(x, y) {
			return t.sprites[i].At(x, y)
		}
	}
	return nil
}
