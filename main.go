package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"

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

type bgWidget struct {
	widget.BaseWidget
	ci   *canvas.Image
	bg   *Backgound
	x, y float32
	w    fyne.Window
	rdr  fyne.WidgetRenderer
}

func (item *bgWidget) MouseDown(event *desktop.MouseEvent) {
	x := int(event.Position.X / scaleFactor)
	y := int(event.Position.Y / scaleFactor)
	fmt.Printf("MouseDown: %d, %d\n", x, y)
	item.bg.stack.MouseDown(x, y)
	item.ci.Refresh()
	item.rdr.Refresh()
}

func (item *bgWidget) MouseUp(event *desktop.MouseEvent) {
	x := int(event.Position.X / scaleFactor)
	y := int(event.Position.Y / scaleFactor)
	fmt.Printf("MouseUp: %d, %d\n", x, y)
	item.bg.stack.DoAtPosition(x, y)
	item.ci.Refresh()
	item.rdr.Refresh()
}

func (item *bgWidget) Dragged(event *fyne.DragEvent) {
	item.x = event.AbsolutePosition.X
	item.y = event.AbsolutePosition.Y

}

func (item *bgWidget) DragEnd() {
	fmt.Println("DragEnd")
}

var _ fyne.Tappable = (*bgWidget)(nil)
var _ desktop.Mouseable = (*bgWidget)(nil)

func (item *bgWidget) Tapped(event *fyne.PointEvent) {
	//x := int(event.Position.X / scaleFactor)
	//y := int(event.Position.Y / scaleFactor)
	//fmt.Printf("Tapped: %d, %d\n", x, y)
	//item.bg.stack.DoAtPosition(x, y)
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
	item.rdr = widget.NewSimpleRenderer(cnt)
	return item.rdr
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

	mainWindowF, err := os.Open("./skins/default/MAIN.BMP")
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

	stack.register("close", func() { w.Close() })

	w.SetContent(newBgWidget(mainWindowBG))

	w.ShowAndRun()
}
