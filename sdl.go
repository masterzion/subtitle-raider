package main

import (
	"./srt"
	"errors"
	"fmt"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/ttf"
	"log"
	"time"
)

const (
	DEFAULT_FONT_PATH = "/usr/share/fonts/truetype/nanum/NanumGothicBold.ttf"
	DEFAULT_FONT_SIZE = 70
)

var (
	BG_COLOR   = sdl.Color{0, 0, 0, 0}
	TEXT_COLOR = sdl.Color{255, 255, 255, 0}
)

type sdlCtx struct {
	surface *sdl.Surface
	w, h    int

	currScript *srt.Script
	font       *ttf.Font
	lineHeight int

	FontSize int
}

func NewSdlContext(w, h int) *sdlCtx {
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		log.Fatal(sdl.GetError())
		return nil
	}

	var ctx sdlCtx
	ctx.surface = sdl.SetVideoMode(w, h, 32, sdl.RESIZABLE) /* sdl.FULLSCREEN */
	if ctx.surface == nil {
		log.Fatal(sdl.GetError())
		return nil
	}

	var vInfo = sdl.GetVideoInfo()
	log.Println("HW_available = ", vInfo.HW_available)
	log.Println("WM_available = ", vInfo.WM_available)
	log.Println("Video_mem = ", vInfo.Video_mem, "kb")

	sdl.EnableUNICODE(1)

	if ttf.Init() != 0 {
		log.Fatal("failed to init ttf", sdl.GetError())
	}

	title := "Subtitle Player"
	icon := "" // path/to/icon
	sdl.WM_SetCaption(title, icon)

	go func() {
		for {
			select {
			case event := <-sdl.Events:
				/* log.Printf("%#v\n", event) */
				switch e := event.(type) {
				case sdl.QuitEvent:
					return
				case sdl.KeyboardEvent:
					log.Printf("Sim:%08x, Mod:%04x\n",
						e.Keysym.Sym, e.Keysym.Mod)
				case sdl.ResizeEvent:
					screen := sdl.SetVideoMode(int(e.W), int(e.H),
						32, sdl.RESIZABLE)
					if screen == nil {
						log.Fatal(sdl.GetError())
					}
					ctx.surface = screen
				}
			}
		}
	}()

	return &ctx
}

func (c *sdlCtx) Clear() {
	if c.currScript == nil {
		return
	}
	log.Println("clear")
	c.surface.FillRect(nil, 0 /* BG_COLOR */)
	c.surface.Flip()
	c.currScript = nil
}

func (c *sdlCtx) DisplayScript(script *srt.Script) {
	if c.currScript == script {
		return
	}
	c.currScript = script

	log.Printf("display %s", script.Text)
	timer := time.NewTimer(script.Duration())

	if c.font == nil {
		log.Println("set default font")
		err := c.SetFont(DEFAULT_FONT_PATH, DEFAULT_FONT_SIZE)
		if err != nil {
			log.Fatal("failed to set default font")
			return
		}
	}

	// w, h, err := c.font.SizeUTF8(script.Text)
	// if err != 0 {
	// 	log.Fatal("Failed to get size of the font")
	// }

	glypse := ttf.RenderUTF8_Blended(c.font, script.Text, TEXT_COLOR)
	c.surface.FillRect(nil, 0 /* BG_COLOR */)
	c.surface.Blit(&sdl.Rect{0, 0, 0, 0}, glypse, nil)
	c.surface.Flip()

	<-timer.C
	if c.currScript == script {
		c.surface.FillRect(nil, 0 /* BG_COLOR */)
		c.surface.Flip()
	}
}

func (c *sdlCtx) SetFont(path string, size int) error {
	c.font = ttf.OpenFont(path, size)
	if c.font == nil {
		errMsg := fmt.Sprintf("failed to open font from %s: %s",
			path, sdl.GetError())
		return errors.New(errMsg)
	}
	c.FontSize = size
	c.lineHeight = c.font.LineSkip()
	return nil
}

/* ctx.font.SetStyle(ttf.STYLE_UNDERLINE) */
func (c *sdlCtx) Release() {
	if c.font != nil {
		c.font.Close()
	}
	if c.surface != nil {
		c.surface.Free()
	}
	ttf.Quit()
	sdl.Quit()
}
