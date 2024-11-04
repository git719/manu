package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var fontPaths = map[string]string{
	"linux":  "/usr/share/fonts/truetype/ubuntu/UbuntuMono-R.ttf",
	"darwin": "/System/Library/Fonts/SFNSMono.ttf", // macOS
}
var games []string

// var selected int
// const (
// 	prgname = "manu"
// 	prgver  = "2.0.0"
// )

func loadGamesFromDirectory() {
	romDir := filepath.Join(os.Getenv("HOME"), ".mame/roms")
	files, err := os.ReadDir(romDir)
	if err != nil {
		fmt.Println("Error reading ROM directory:", err)
		os.Exit(1)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".zip") {
			games = append(games, strings.TrimSuffix(file.Name(), ".zip"))
		}
	}

	// Append default commands
	games = append(games, "<exit>", "<poweroff>")
}

func runGame(game string) {
	cmd := exec.Command("mame", "-skip_gameinfo", strings.TrimSpace(game))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func getFontPath() string {
	osType := runtime.GOOS
	if path, ok := fontPaths[osType]; ok {
		return path
	}
	fmt.Println("No font path found for OS:", osType)
	os.Exit(1)
	return ""
}

func sdlMain() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		return err
	}
	defer ttf.Quit()

	window, err := sdl.CreateWindow("Game Selector", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return err
	}
	defer renderer.Destroy()

	// Load the font based on the operating system
	fontPath := getFontPath()
	font, err := ttf.OpenFont(fontPath, 24) // Adjust font size as needed
	if err != nil {
		return err
	}
	defer font.Close()

	loadGamesFromDirectory()
	selected := 0

	running := true
	for running {
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		for i, game := range games {
			x, y := int32(50), int32(50+(i*30))

			// Highlight selected game
			if i == selected {
				renderer.SetDrawColor(100, 100, 255, 255)
				renderer.FillRect(&sdl.Rect{X: x - 10, Y: y, W: 200, H: 25})
			}

			// Render text for each game
			surface, err := font.RenderUTF8Solid(game, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err != nil {
				return err
			}
			texture, err := renderer.CreateTextureFromSurface(surface)
			surface.Free()
			if err != nil {
				return err
			}
			defer texture.Destroy()

			textRect := sdl.Rect{X: x, Y: y, W: surface.W, H: surface.H}
			renderer.Copy(texture, nil, &textRect)
		}

		renderer.Present()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					switch e.Keysym.Sym {
					case sdl.K_UP:
						selected = (selected - 1 + len(games)) % len(games)
					case sdl.K_DOWN:
						selected = (selected + 1) % len(games)
					case sdl.K_RETURN:
						if games[selected] == "<exit>" {
							running = false
						} else if games[selected] == "<poweroff>" {
							exec.Command("/sbin/poweroff").Run()
							running = false
						} else {
							runGame(games[selected])
						}
					case sdl.K_ESCAPE:
						running = false
					}
				}
			}
		}
	}

	return nil
}

func main() {
	if err := sdlMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
