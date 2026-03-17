package imagepreview

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

func RenderANSI(img image.Image, maxWidth int, maxHeight int) (string, error) {
	resized := resizeNearest(img, maxWidth, maxHeight*2)
	bounds := resized.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return "", fmt.Errorf("empty image")
	}

	var b strings.Builder
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			top := color.RGBAModel.Convert(resized.At(x, y)).(color.RGBA)
			bottom := top
			if y+1 < bounds.Max.Y {
				bottom = color.RGBAModel.Convert(resized.At(x, y+1)).(color.RGBA)
			}
			fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", top.R, top.G, top.B, bottom.R, bottom.G, bottom.B)
		}
		b.WriteString("\x1b[0m\n")
	}
	return b.String(), nil
}

func resizeNearest(img image.Image, maxWidth int, maxHeight int) image.Image {
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return img
	}

	scaleW := float64(maxWidth) / float64(srcW)
	scaleH := float64(maxHeight) / float64(srcH)
	scale := min(scaleW, scaleH)
	if scale <= 0 {
		scale = 1
	}
	if scale > 1 {
		scale = 1
	}
	dstW := max(1, int(float64(srcW)*scale))
	dstH := max(1, int(float64(srcH)*scale))

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		sy := bounds.Min.Y + y*srcH/dstH
		for x := 0; x < dstW; x++ {
			sx := bounds.Min.X + x*srcW/dstW
			dst.Set(x, y, img.At(sx, sy))
		}
	}
	return dst
}
