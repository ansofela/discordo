package imagepreview

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"runtime"
	"strings"
	"sync"

	_ "golang.org/x/image/webp"
)

type Backend uint8

const (
	BackendNone Backend = iota
	BackendKitty
	BackendANSI
)

type RenderConfig struct {
	Enable    bool
	MaxWidth  int
	MaxHeight int
}

type Previewer struct {
	cfg   RenderConfig
	mu    sync.RWMutex
	cache map[string]string
}

func New(cfg RenderConfig) *Previewer {
	return &Previewer{cfg: cfg, cache: make(map[string]string)}
}

func DetectBackend() Backend {
	if !isLikelyTTY() {
		return BackendNone
	}

	term := strings.ToLower(os.Getenv("TERM"))
	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))

	if strings.Contains(term, "kitty") || strings.Contains(termProgram, "kitty") || strings.Contains(termProgram, "wezterm") || strings.Contains(termProgram, "ghostty") {
		return BackendKitty
	}
	if strings.Contains(term, "xterm") || strings.Contains(term, "screen") || strings.Contains(term, "tmux") || strings.Contains(termProgram, "apple_terminal") || strings.Contains(termProgram, "iterm") {
		return BackendANSI
	}

	return BackendNone
}

func (p *Previewer) CanRenderInline() bool {
	return p.cfg.Enable && DetectBackend() != BackendNone
}

func (p *Previewer) Render(path string, termWidth int, termHeight int) (string, Backend, error) {
	backend := DetectBackend()
	if !p.cfg.Enable || backend == BackendNone {
		return "", BackendNone, nil
	}

	cacheKey := fmt.Sprintf("%d:%s:%d:%d", backend, path, termWidth, termHeight)
	p.mu.RLock()
	if cached, ok := p.cache[cacheKey]; ok {
		p.mu.RUnlock()
		return cached, backend, nil
	}
	p.mu.RUnlock()

	f, err := os.Open(path)
	if err != nil {
		return "", backend, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", backend, err
	}

	width, height := p.limitSize(termWidth, termHeight)
	var out string
	switch backend {
	case BackendKitty:
		out, err = RenderKitty(img, width, height)
	case BackendANSI:
		out, err = RenderANSI(img, width, height)
	}
	if err != nil {
		return "", backend, err
	}

	p.mu.Lock()
	p.cache[cacheKey] = out
	p.mu.Unlock()

	return out, backend, nil
}

func (p *Previewer) limitSize(termWidth int, termHeight int) (int, int) {
	width := p.cfg.MaxWidth
	height := p.cfg.MaxHeight
	if termWidth > 0 && (width == 0 || termWidth < width) {
		width = termWidth
	}
	if termHeight > 0 && (height == 0 || termHeight < height) {
		height = termHeight
	}
	if width <= 0 {
		width = 64
	}
	if height <= 0 {
		height = 32
	}
	if width < 8 {
		width = 8
	}
	if height < 4 {
		height = 4
	}
	return width, height
}

func isLikelyTTY() bool {
	if os.Getenv("TERM") == "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return true
}
