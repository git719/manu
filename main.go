package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var games []string

//go:embed assets/fonts/DejaVuSansMono.ttf
var fontData []byte

const (
	prgname = "manu"
	prgver  = "2.0.0"
)

// Prints program usage
func printUsage() {
	fmt.Printf(prgname + " v" + prgver + "\n" +
		"MAME selector menu - https://github.com/git719/manu\n" +
		"Usage: " + prgname + " [options]\n" +
		"  Run with no options to read ROMs under '$HOME/.mame/roms/' and present a selection menu\n" +
		"  -?, -h, --help                    Print this usage page\n")
	os.Exit(0)
}

func getEmbeddedFontPath() (string, error) {
	// Create a temporary file to store the embedded font
	tempFontPath := filepath.Join(os.TempDir(), "DejaVuSansMono.ttf")
	err := os.WriteFile(tempFontPath, fontData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write embedded font to file: %v", err)
	}
	return tempFontPath, nil
}

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

func sdlMain() error {
	// Load games from directory
	loadGamesFromDirectory()

	// Initialize SDL and SDL_ttf
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}
	defer sdl.Quit()

	sdl.ShowCursor(sdl.DISABLE) // Hide the mouse cursor

	if err := ttf.Init(); err != nil {
		return err
	}
	defer ttf.Quit()

	// Set up window and renderer
	window, err := sdl.CreateWindow("Game Selector", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED) // sdl.RENDERER_SOFTWARE ?
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %v\n", err)
		os.Exit(1)
	}
	defer renderer.Destroy()

	// Get the path to the embedded font
	fontPath, err := getEmbeddedFontPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get embedded font path: %v\n", err)
		os.Exit(1)
	}

	// Open the font from the embedded file
	font, err := ttf.OpenFont(fontPath, 24) // Adjust font size as needed
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load font: %v\n", err)
		os.Exit(1)
	}
	defer font.Close()

	// Get window dimensions and convert them to int
	winWidth, winHeight := window.GetSize()
	winWidthInt, winHeightInt := int(winWidth), int(winHeight)
	itemHeight := 30                           // Space between each menu item
	totalMenuHeight := len(games) * itemHeight // Total height of all menu items

	// Calculate starting positions for centered menu
	startY := (winHeightInt - totalMenuHeight) / 2 // Center vertically
	startX := winWidthInt/2 - 100                  // Center horizontally, adjust -100 for padding

	running := true
	selected := 0

	for running {
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		// Render each game in the centered position
		for i, game := range games {
			x, y := int32(startX), int32(startY+(i*itemHeight))

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

		// Handle events
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
							// Quit SDL session before launching the game
							sdl.Quit()
							ttf.Quit()

							runGame(games[selected])

							os.Exit(0)
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
	numberOfArguments := len(os.Args[1:])
	if numberOfArguments > 0 {
		printUsage()
	}

	if err := sdlMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
