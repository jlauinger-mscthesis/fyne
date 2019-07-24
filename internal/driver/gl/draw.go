package gl

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"
)

func rectInnerCoords(size fyne.Size, pos fyne.Position, fill canvas.ImageFill, aspect float32) (fyne.Size, fyne.Position) {
	if fill == canvas.ImageFillContain || fill == canvas.ImageFillOriginal {
		// change pos and size accordingly

		viewAspect := float32(size.Width) / float32(size.Height)

		newWidth, newHeight := size.Width, size.Height
		widthPad, heightPad := 0, 0
		if viewAspect > aspect {
			newWidth = int(float32(size.Height) * aspect)
			widthPad = (size.Width - newWidth) / 2
		} else if viewAspect < aspect {
			newHeight = int(float32(size.Width) / aspect)
			heightPad = (size.Height - newHeight) / 2
		}

		return fyne.NewSize(newWidth, newHeight), fyne.NewPos(pos.X+widthPad, pos.Y+heightPad)
	}

	return size, pos
}

// rectCoords calculates the openGL coordinate space of a rectangle
func (c *glCanvas) rectCoords(size fyne.Size, pos fyne.Position, frame fyne.Size,
	fill canvas.ImageFill, aspect float32, pad int) ([]float32, uint32, uint32) {
	size, pos = rectInnerCoords(size, pos, fill, aspect)

	xPos := float32(pos.X-pad) / float32(frame.Width)
	x1 := -1 + xPos*2
	x2Pos := float32(pos.X+size.Width+pad) / float32(frame.Width)
	x2 := -1 + x2Pos*2

	yPos := float32(pos.Y-pad) / float32(frame.Height)
	y1 := 1 - yPos*2
	y2Pos := float32(pos.Y+size.Height+pad) / float32(frame.Height)
	y2 := 1 - y2Pos*2

	points := []float32{
		// coord x, y, x texture x, y
		x1, y2, 0, 0.0, 1.0, // top left
		x1, y1, 0, 0.0, 0.0, // bottom left
		x2, y2, 0, 1.0, 1.0, // top right
		x2, y1, 0, 1.0, 0.0, // bottom right
	}

	vao, vbo := c.glCreateBuffer(points)
	return points, vao, vbo
}

func (c *glCanvas) freeCoords(vao, vbo uint32) {
	c.glFreeBuffer(vao, vbo)
}

func (c *glCanvas) drawWidget(wid fyne.Widget, pos fyne.Position, frame fyne.Size) {
	if widget.Renderer(wid).BackgroundColor() == color.Transparent {
		return
	}

	points, vao, vbo := c.rectCoords(wid.Size(), pos, frame, canvas.ImageFillStretch, 0.0, 0)
	texture := getTexture(wid, c.newGlRectTexture)

	c.glDrawTexture(texture, points, 1.0)
	c.freeCoords(vao, vbo)
}

func (c *glCanvas) drawCircle(circle *canvas.Circle, pos fyne.Position, frame fyne.Size) {
	points, vao, vbo := c.rectCoords(circle.Size(), pos, frame, canvas.ImageFillStretch, 0.0, vectorPad)
	texture := getTexture(circle, c.newGlCircleTexture)

	c.glDrawTexture(texture, points, 1.0)
	c.freeCoords(vao, vbo)
}

func (c *glCanvas) drawLine(line *canvas.Line, pos fyne.Position, frame fyne.Size) {
	points, vao, vbo := c.rectCoords(line.Size(), pos, frame, canvas.ImageFillStretch, 0.0, vectorPad)
	texture := getTexture(line, c.newGlLineTexture)

	c.glDrawTexture(texture, points, 1.0)
	c.freeCoords(vao, vbo)
}

func (c *glCanvas) drawImage(img *canvas.Image, pos fyne.Position, frame fyne.Size) {
	texture := getTexture(img, c.newGlImageTexture)
	if texture == 0 {
		return
	}

	aspect := aspects[img.Resource]
	if aspect == 0 {
		aspect = aspects[img]
	}
	points, vao, vbo := c.rectCoords(img.Size(), pos, frame, img.FillMode, aspect, 0)
	c.glDrawTexture(texture, points, float32(img.Alpha()))
	c.freeCoords(vao, vbo)
}

func (c *glCanvas) drawRaster(img *canvas.Raster, pos fyne.Position, frame fyne.Size) {
	texture := getTexture(img, c.newGlRasterTexture)
	if texture == 0 {
		return
	}

	points, vao, vbo := c.rectCoords(img.Size(), pos, frame, canvas.ImageFillStretch, 0.0, 0)
	c.glDrawTexture(texture, points, float32(img.Alpha()))
	c.freeCoords(vao, vbo)
}

func (c *glCanvas) drawGradient(o fyne.CanvasObject, texCreator func(fyne.CanvasObject) uint32, pos fyne.Position, frame fyne.Size) {
	texture := getTexture(o, texCreator)
	if texture == 0 {
		return
	}

	points, vao, vbo := c.rectCoords(o.Size(), pos, frame, canvas.ImageFillStretch, 0.0, 0)
	c.glDrawTexture(texture, points, 1.0)
	c.freeCoords(vao, vbo)
}

func (c *glCanvas) drawRectangle(rect *canvas.Rectangle, pos fyne.Position, frame fyne.Size) {
	points, vao, vbo := c.rectCoords(rect.Size(), pos, frame, canvas.ImageFillStretch, 0.0, 0)
	texture := getTexture(rect, c.newGlRectTexture)

	c.glDrawTexture(texture, points, 1.0)
	c.freeCoords(vao, vbo)
}

func (c *glCanvas) drawText(text *canvas.Text, pos fyne.Position, frame fyne.Size) {
	if text.Text == "" {
		return
	}

	size := text.MinSize()
	containerSize := text.Size()
	switch text.Alignment {
	case fyne.TextAlignTrailing:
		pos = fyne.NewPos(pos.X+containerSize.Width-size.Width, pos.Y)
	case fyne.TextAlignCenter:
		pos = fyne.NewPos(pos.X+(containerSize.Width-size.Width)/2, pos.Y)
	}

	if text.Size().Height > text.MinSize().Height {
		pos = fyne.NewPos(pos.X, pos.Y+(text.Size().Height-text.MinSize().Height)/2)
	}

	points, vao, vbo := c.rectCoords(size, pos, frame, canvas.ImageFillStretch, 0.0, 0)
	texture := getTexture(text, c.newGlTextTexture)

	c.glDrawTexture(texture, points, 1.0)
	c.freeCoords(vao, vbo)
}

func (c *glCanvas) drawObject(o fyne.CanvasObject, pos fyne.Position, frame fyne.Size) {
	if !o.Visible() {
		return
	}
	canvasMutex.Lock()
	canvases[o] = c
	canvasMutex.Unlock()
	switch obj := o.(type) {
	case *canvas.Circle:
		c.drawCircle(obj, pos, frame)
	case *canvas.Line:
		c.drawLine(obj, pos, frame)
	case *canvas.Image:
		c.drawImage(obj, pos, frame)
	case *canvas.Raster:
		c.drawRaster(obj, pos, frame)
	case *canvas.Rectangle:
		c.drawRectangle(obj, pos, frame)
	case *canvas.Text:
		c.drawText(obj, pos, frame)
	case *canvas.LinearGradient:
		c.drawGradient(obj, c.newGlLinearGradientTexture, pos, frame)
	case *canvas.RadialGradient:
		c.drawGradient(obj, c.newGlRadialGradientTexture, pos, frame)
	case fyne.Widget:
		c.drawWidget(obj, pos, frame)
	}
}
