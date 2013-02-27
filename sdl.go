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
		log.Fatal("failed to init sdl", sdl.GetError())
		return nil
	}

	if ttf.Init() != 0 {
		log.Fatal("failed to init ttf", sdl.GetError())
		return nil
	}

	var ctx sdlCtx
	if err := ctx.setSurface(w, h); err != nil {
		log.Fatal(err)
	}

	var vInfo = sdl.GetVideoInfo()
	log.Println("HW_available = ", vInfo.HW_available)
	log.Println("WM_available = ", vInfo.WM_available)
	log.Println("Video_mem = ", vInfo.Video_mem, "kb")

	title := "Subtitle Player"
	icon := "" // path/to/icon
	sdl.WM_SetCaption(title, icon)

	sdl.EnableUNICODE(1)

	go func() {
	EVENT_LOOP:
		for {
			err := ctx.handelEvent()
			if err != nil {
				fmt.Println(err)
				break EVENT_LOOP
			}
		}
		log.Println("sdl: exit event loop")
	}()

	return &ctx
}

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

func (c *sdlCtx) DisplayScript(script *srt.Script) {
	c.displayScript(script, true, false)
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

func (c *sdlCtx) SetFont(path string, size int) error {
	c.font = ttf.OpenFont(path, size)
	if c.font == nil {
		errMsg := fmt.Sprintf("failed to open font from %s: %s",
			path, sdl.GetError())
		return errors.New(errMsg)
	}
	c.FontSize = size
	c.lineHeight = c.font.LineSkip()
	/* ctx.font.SetStyle(ttf.STYLE_UNDERLINE) */
	return nil
}

func (c *sdlCtx) setSurface(w, h int) error {
	log.Printf("setSurface to %dx%d", w, h)
	c.surface = sdl.SetVideoMode(w, h, 32, sdl.RESIZABLE) /* sdl.FULLSCREEN */
	if c.surface == nil {
		errMsg := fmt.Sprintf("sdl: failed to set video to %dx%d: %s",
			w, h, sdl.GetError())
		return errors.New(errMsg)
	}

	c.w, c.h = w, h
	if c.currScript != nil {
		c.displayScript(c.currScript, false, true)
	}

	return nil
}

func (c *sdlCtx) handelEvent() error {
	select {
	case event := <-sdl.Events:
		/* log.Printf("%#v\n", event) */
		switch e := event.(type) {
		case sdl.QuitEvent:
			return errors.New("sdl: received QuitEvent")
		case sdl.KeyboardEvent:
			log.Printf("Sim:%08x, Mod:%04x, Unicode:%02x\n",
				e.Keysym.Sym, e.Keysym.Mod, e.Keysym.Unicode)
		case sdl.ResizeEvent:
			if err := c.setSurface(int(e.W), int(e.H)); err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

func (c *sdlCtx) displayScript(script *srt.Script,
	andClear bool, forceUpdate bool) {
	if forceUpdate == false && c.currScript == script {
		return
	}
	c.currScript = script

	log.Printf("display %s", script.Text)

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

	if andClear == false {
		return
	}

	timer := time.NewTimer(script.Duration())
	<-timer.C
	if c.currScript == script {
		c.surface.FillRect(nil, 0 /* BG_COLOR */)
		c.surface.Flip()
	}
}
