package chat

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"

	imagepreview "github.com/ayn2op/discordo/internal/image_preview"
	"github.com/ayn2op/tview/list"
)

func TestRenderInlineImageANSIBackend(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "")

	imgPath := writeTestImage(t)

	ml := &messagesList{
		Model:          list.NewModel(),
		imagePreviewer: imagepreview.New(imagepreview.RenderConfig{Enable: true, MaxWidth: 64, MaxHeight: 32}),
	}
	ml.SetRect(0, 0, 80, 24)

	rendered, backend, err := ml.renderInlineImage(imgPath)
	if err != nil {
		t.Fatalf("renderInlineImage returned error: %v", err)
	}
	if !rendered {
		t.Fatalf("expected image to render inline")
	}
	if backend != "ansi" {
		t.Fatalf("expected ansi backend, got %q", backend)
	}
}

func TestRenderInlineImageDisabledPreview(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("TERM_PROGRAM", "")

	imgPath := writeTestImage(t)

	ml := &messagesList{
		Model:          list.NewModel(),
		imagePreviewer: imagepreview.New(imagepreview.RenderConfig{Enable: false, MaxWidth: 64, MaxHeight: 32}),
	}
	ml.SetRect(0, 0, 80, 24)

	rendered, backend, err := ml.renderInlineImage(imgPath)
	if err != nil {
		t.Fatalf("renderInlineImage returned error: %v", err)
	}
	if rendered {
		t.Fatalf("expected image not to render inline when feature is disabled")
	}
	if backend != "" {
		t.Fatalf("expected empty backend label when not rendered, got %q", backend)
	}
}

func writeTestImage(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "inline-preview.png")

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 50), G: uint8(y * 50), B: 180, A: 255})
		}
	}

	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatalf("create image file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	return imgPath
}

func TestIsImageAttachment(t *testing.T) {
	tests := []struct {
		name       string
		attachment discord.Attachment
		want       bool
	}{
		{
			name: "content type image",
			attachment: discord.Attachment{
				Filename:    "file.bin",
				ContentType: "image/png",
			},
			want: true,
		},
		{
			name: "filename extension fallback",
			attachment: discord.Attachment{
				Filename: "preview.jpeg",
			},
			want: true,
		},
		{
			name: "url extension fallback",
			attachment: discord.Attachment{
				Filename: "download",
				URL:      "https://cdn.example.com/path/img.webp?token=abc",
			},
			want: true,
		},
		{
			name: "non-image",
			attachment: discord.Attachment{
				Filename:    "notes.txt",
				ContentType: "text/plain",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isImageAttachment(tt.attachment); got != tt.want {
				t.Fatalf("isImageAttachment() = %v, want %v", got, tt.want)
			}
		})
	}
}
