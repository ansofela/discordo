package imagepreview

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
)

func RenderKitty(img image.Image, maxWidth int, maxHeight int) (string, error) {
	resized := resizeNearest(img, maxWidth, maxHeight*2)
	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return "", fmt.Errorf("encode png: %w", err)
	}
	payload := base64.StdEncoding.EncodeToString(buf.Bytes())
	// Keep output as escape sequence text so caller can print it directly.
	return "\x1b_Gf=100,a=T,t=d;" + payload + "\x1b\\", nil
}
