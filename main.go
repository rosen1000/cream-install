package main

import (
	"flag"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	// tea "github.com/charmbracelet/bubbletea"
	tea "github.com/charmbracelet/bubbletea"
	// "github.com/k0kubun/pp"
	// "gopkg.in/ini.v1"
)

var debug bool

func main() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.Parse()

	states := getAppStates()
	screen := BrowseScr
	var selectedGame *AppState

loop:
	for {
		switch screen {
		case BrowseScr:
			_model, _ := tea.NewProgram(NewGameModel(states)).Run()
			model := _model.(GameModel)
			ClearScreen()
			switch model.mode {
			case EditM:
				screen = EditScr
				selectedGame = model.Selected
			case RunM:
				if model.Selected != nil && model.Selected.Appid != 0 {
					if !debug {
						model.Selected.Run()
					}
					fmt.Printf("Starting %s!\n", model.Selected.Name)
				} else {
					fmt.Println("Canceled")
				}
				fallthrough
			case ExitM:
				screen = ExitScr
			}
		case EditScr:
			_model, _ := tea.NewProgram(NewEditModel(selectedGame)).Run()
			_ = _model
			ClearScreen()
			screen = BrowseScr
		case ExitScr:
			break loop
		}
	}

}

type Screen int

const (
	BrowseScr Screen = iota
	EditScr
	ExitScr
)

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// Returns all .dll paths for all platforms
func Scan(path string) []string {
	found := []string{}

	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, "steam_api.dll") ||
			strings.HasSuffix(path, "steam_api64.dll") ||
			strings.HasSuffix(path, "libsteam_api.so") ||
			strings.HasSuffix(path, "libsteam_api.dylib") {
			found = append(found, path)
		}
		return nil
	})

	return found
}

func USED(arg any) { _ = arg }

// if _, err := tea.NewProgram(initModel()).Run(); err != nil {
// }

// if len(os.Args) < 2 {
// 	panic("Need 1 arg")
// }
// path := os.Args[1]

// files := scan(path)
// printArray(files)

// getUsers()

// _, err := readShortcuts()
// if err != nil {
// 	panic(err)
// }

// getApps()
