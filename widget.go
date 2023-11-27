package main

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Background Holds the background image and implements a thin wrapper to implement image.Image
type Background struct {
	stack *SpriteStack
}

func (b *Background) ColorModel() color.Model {
	return b.stack.sprites[0].ColorModel()
}

func (b *Background) Bounds() image.Rectangle {
	return b.stack.sprites[0].Bounds()
}

func (b *Background) At(x, y int) color.Color {
	if colorAt := b.stack.DrawAtPosition(x, y); colorAt != nil {
		return colorAt
	}
	return b.stack.sprites[0].At(x, y)
}

type bgWidget struct {
	widget.BaseWidget
	ci   *canvas.Image
	bg   *Background
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
	item.bg.stack.Dragged(event)
	item.ci.Refresh()
	item.rdr.Refresh()
}

func (item *bgWidget) DragEnd() {
	item.bg.stack.DragEnd()
}

var _ desktop.Mouseable = (*bgWidget)(nil)
var _ fyne.Draggable = (*bgWidget)(nil)

func newBgWidget(rawImg *Background) *bgWidget {
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
