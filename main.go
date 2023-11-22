package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const scaleFactor = 2

type Backgound struct {
	underlyingImage image.Image
	stack           *SpriteStack
}

func (b *Backgound) ColorModel() color.Model {
	return b.underlyingImage.ColorModel()
}

func (b *Backgound) Bounds() image.Rectangle {
	return b.underlyingImage.Bounds()
}

func (b *Backgound) At(x, y int) color.Color {
	if colorAt := b.stack.DrawAtPosition(x, y); colorAt != nil {
		return colorAt
	}
	return b.underlyingImage.At(x, y)
}

type SpriteStack struct {
	sprites   []*Sprite
	fileCache map[string]image.Image
	prefix    string
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
		s := s.sprites[len(s.sprites)-i-1]
		if s.Collision(x, y) {
			if s.action != nil {
				s.action()
				return
			}
			fmt.Printf("Tapped: %s\n", s.ID)
			return
		}
	}
}

func (s *SpriteStack) UnmarshalJSON(data []byte) error {
	var tgt []*Sprite
	if err := json.Unmarshal(data, &tgt); err != nil {
		return err
	}

	for i, sprite := range tgt {
		tgt[i].boundingRectangle = image.Rect(0, 0, sprite.BoundingRectangleX, sprite.BoundingRectangleY)
		if err := sprite.Load(s.prefix, s.fileCache); err != nil {
			return err
		}
	}

	s.sprites = tgt
	return nil
}

type Sprite struct {
	ID                 string `json:"id,omitempty"`
	AbsX               int    `json:"absX"`
	AbsY               int    `json:"absY"`
	underlyingImage    image.Image
	boundingRectangle  image.Rectangle
	BoundingRectangleX int    `json:"boundingRectangleX"`
	BoundingRectangleY int    `json:"boundingRectangleY"`
	OffsetX            int    `json:"offsetX,omitempty"`
	OffsetY            int    `json:"offsetY,omitempty"`
	File               string `json:"file,omitempty"`
	action             func()
}

func (s *Sprite) Collision(x, y int) bool {
	inX := x > s.AbsX && x < s.AbsX+s.boundingRectangle.Dx()
	inY := y > s.AbsY && y < s.AbsY+s.boundingRectangle.Dy()
	return inX && inY
}

func (s *Sprite) Load(prefix string, fileCache map[string]image.Image) error {
	rawImg, ok := fileCache[s.File]
	if !ok {
		f, err := os.Open(filepath.Join(prefix, s.File))
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
	s.underlyingImage = rawImg
	return nil
}

func (s *Sprite) ColorModel() color.Model {
	return s.ColorModel()
}

func (s *Sprite) Bounds() image.Rectangle {
	return s.boundingRectangle
}

func (s *Sprite) At(x, y int) color.Color {
	return s.underlyingImage.At(x+s.OffsetX-s.AbsX, y+s.OffsetY-s.AbsY)
}

var _ image.Image = (*Sprite)(nil)

func stackFromFromDefinitions() (*SpriteStack, error) {
	stack := SpriteStack{
		prefix:    "./skins/tmp",
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

type bgWidget struct {
	widget.BaseWidget
	ci   *canvas.Image
	bg   *Backgound
	x, y float32
	w    fyne.Window
}

func (item *bgWidget) Dragged(event *fyne.DragEvent) {
	item.x = event.AbsolutePosition.X
	item.y = event.AbsolutePosition.Y

}

func (item *bgWidget) DragEnd() {
	fmt.Println("DragEnd")
}

var _ fyne.Tappable = (*bgWidget)(nil)

func (item *bgWidget) Tapped(event *fyne.PointEvent) {
	x := int(event.Position.X / scaleFactor)
	y := int(event.Position.Y / scaleFactor)
	fmt.Printf("Tapped: %d, %d\n", x, y)
	item.bg.stack.DoAtPosition(x, y)
}

var _ fyne.Draggable = (*bgWidget)(nil)

func newBgWidget(rawImg *Backgound) *bgWidget {
	img := canvas.NewImageFromImage(rawImg)
	img.ScaleMode = canvas.ImageScalePixels
	item := &bgWidget{
		ci: img,
		bg: rawImg,
	}

	item.ExtendBaseWidget(item)
	item.Resize(item.ci.Size())

	return item
}

func (item *bgWidget) CreateRenderer() fyne.WidgetRenderer {
	cnt := container.New(layout.NewStackLayout(),
		item.ci,
	)
	return widget.NewSimpleRenderer(cnt)
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

	mainWindowF, err := os.Open("./skins/tmp/MAIN.BMP")
	if err != nil {
		panic(err)
	}
	mainWindowIMG, _, err := image.Decode(mainWindowF)
	defer mainWindowF.Close()

	// Load sprites
	stack, err := stackFromFromDefinitions()
	if err != nil {
		panic(err)
	}

	mainWindowBG := &Backgound{
		underlyingImage: mainWindowIMG,
		stack:           stack,
	}

	stack.handle("closeButton", func() { w.Close() })

	w.SetContent(newBgWidget(mainWindowBG))

	w.ShowAndRun()
}
