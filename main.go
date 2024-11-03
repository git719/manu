package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/eiannone/keyboard"
)

var games []string

const (
	prgname      = "manu"
	prgver       = "1.0.1"
	reverseColor = "\033[7m"
	resetColor   = "\033[0m"
)

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func moveCursorToTop() {
	fmt.Printf("\033[H")
}

func displayMenu(selected int) {
	moveCursorToTop()
	fmt.Printf("\n  Select game or option with arrow keys and press ENTER:\n\n")
	for i, game := range games {
		if i == selected {
			fmt.Printf("     %s%s%s  \n", reverseColor, strings.TrimSpace(game), resetColor)
		} else {
			fmt.Printf("     %s  \n", strings.TrimSpace(game))
		}
	}
}

func runGame(game string) {
	cmd := exec.Command("mame", "-skip_gameinfo", strings.TrimSpace(game))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func powerOff() {
	fmt.Println("Powering off...")
	err := syscall.Exec("/sbin/poweroff", []string{"poweroff"}, os.Environ())
	if err != nil {
		fmt.Println("Failed to power off:", err)
	}
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

func main() {
	if err := keyboard.Open(); err != nil {
		fmt.Println("Error opening keyboard:", err)
		return
	}
	defer keyboard.Close()

	// Load games from the ROMs directory
	loadGamesFromDirectory()

	clearScreen()
	fmt.Printf("\033[?25l")
	defer fmt.Printf("\033[?25h")

	selected := 0
	displayMenu(selected)

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Println("Error reading key:", err)
			return
		}

		oldSelected := selected

		switch key {
		case keyboard.KeyArrowUp:
			selected--
			if selected < 0 {
				selected = len(games) - 1
			}
		case keyboard.KeyArrowDown:
			selected++
			if selected >= len(games) {
				selected = 0
			}
		case keyboard.KeyEnter:
			switch strings.TrimSpace(games[selected]) {
			case "<exit>":
				fmt.Printf("\n\n")
				return
			case "<poweroff>":
				powerOff()
				return
			default:
				runGame(games[selected])
			}
		case keyboard.KeyEsc:
			fmt.Printf("\n\n")
			return
		default:
			if char == 'q' || char == 'Q' {
				fmt.Printf("\n\n")
				return
			}
		}

		if oldSelected != selected {
			displayMenu(selected)
		}
	}
}
