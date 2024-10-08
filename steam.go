package main

// https://developer.valvesoftware.com/wiki/Steam_browser_protocol
import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/BenLubar/vdf"
	"github.com/karrick/godirwalk"
)

const (
	STEAM_HOME string = ".steam/steam"
)

var searchNames = []string{"libsteam_api.so", "libsteam_api.dylib", "steam_api.dll", "steam_api64.dll"}

func GetUsers() {
	p := path.Join(getHome(), STEAM_HOME, "userdata")
	userIds, err := os.ReadDir(p)
	if err != nil {
		panic(err)
	}
	for _, user := range userIds {
		id := user.Name()
		vdf, err := ReadVdfA(path.Join(getHome(), STEAM_HOME, "userdata", id, "config/localconfig.vdf"))
		if err != nil {
			panic(err)
		}
		// fmt.Println(data["UserLocalConfigStore"].(map[string]interface{})["friends"].(map[string]interface{})["PersonaName"])
		USED(vdf)
	}
}

func ReadShortcuts() (*vdf.Node, error) {
	vdf, err := ReadVdfB("/home/hax/.steam/steam/userdata/169122681/config/shortcuts.vdf")
	if err != nil {
		return nil, err
	}
	return vdf, nil
}

func GetApps() {
	p := path.Join(getHome(), STEAM_HOME, "userdata/169122681/config/localconfig.vdf")
	root, err := ReadVdfA(p)
	if err != nil {
		panic(err)
	}
	// fmt.Println(*root.FirstByName("broadcast").FirstValue())
	// fmt.Println(*root.FirstByName("broadcast").FirstValue())
	(root.FirstChild().FirstValue().SetInt(2))
	// fmt.Println(*root.FirstByName("broadcast").FirstValue())
	// txt, _ := root.MarshalText()
	// fmt.Println(string(txt[0:100]))
	// parser := vdf.NewParser(file)
	// vdf, _ := parser.Parse()
	// apps := vdf["UserLocalConfigStore"].(map[string]interface{})["Software"].(map[string]interface{})["Valve"].(map[string]interface{})["Steam"].(map[string]interface{})["apps"].(map[string]interface{})
	// for id, _app := range apps {
	// 	app := _app.(map[string]interface{})
	// 	a := App{
	// 		// Playtime: app["Playtime"].(string),
	// 	}
	// 	pp.Println(a)
	// 	pp.Println(app["Playtime"])
	// 	println(id)
	// 	pp.Println(app)
	// }
}

func getLibraries() []Library {
	libVdf, err := ReadVdfA(path.Join(getHome(), STEAM_HOME, "steamapps/libraryfolders.vdf"))
	if err != nil {
		panic(err)
	}

	var libraries []Library
	libVdf = libVdf.FirstChild()
	for libVdf != nil {
		apps := map[int32]uint64{}
		app := libVdf.FirstByName("apps").FirstChild()
		for app != nil {
			key, err := strconv.Atoi(app.Name())
			if err != nil {
				panic(err)
			}
			apps[int32(key)] = app.Uint64()
			app = app.NextChild()
		}
		libraries = append(libraries, Library{
			Path:                     libVdf.FirstByName("path").String(),
			Label:                    libVdf.FirstByName("label").String(),
			ContentId:                libVdf.FirstByName("contentid").Uint64(),
			TotalSize:                libVdf.FirstByName("totalsize").Uint64(),
			UpdateCleanBytesTally:    libVdf.FirstByName("update_clean_bytes_tally").Uint64(),
			TimeLastUpdateCorruption: libVdf.FirstByName("time_last_update_corruption").Int(),
			Apps:                     apps,
		})
		libVdf = libVdf.NextChild()
	}
	return libraries
}

func getAppStates() []AppState {
	libs := getLibraries()
	var states []AppState
	for _, lib := range libs {
		manifests := []string{}
		manif, err := filepath.Glob(filepath.Join(lib.Path, "steamapps", "appmanifest_*.acf"))
		if err != nil {
			panic(manif)
		}
		manifests = append(manifests, manif...)
		for _, match := range manifests {
			vdf, err := ReadVdfA(match)
			if err != nil {
				panic(err)
			}
			states = append(states, NewAppState(vdf, &lib))
		}
	}
	return states
}

type Library struct {
	Path                     string
	Label                    string
	ContentId                uint64
	TotalSize                uint64
	UpdateCleanBytesTally    uint64
	TimeLastUpdateCorruption int32
	Apps                     map[int32]uint64
}

type StateFlag int32

func (f StateFlag) String() string {
	switch f {
	case 0:
		return "invalid"
	case 1:
		return "uninstalled"
	case 2:
		return "update required"
	case 4:
		return "fully installed"
	case 2 << 2:
		return "encrypted"
	case 2 << 3:
		return "locked"
	case 2 << 4:
		return "files missing"
	case 2 << 5:
		return "app running"
	case 2 << 6:
		return "files corrupt"
	case 2 << 7:
		return "update running"
	case 2 << 8:
		return "update paused"
	case 2 << 9:
		return "update started"
	case 2 << 10:
		return "uninstalling"
	case 2 << 11:
		return "backup running"
	case 2 << 12:
		return "reconfiguring"
	case 2 << 13:
		return "validating"
	case 2 << 14:
		return "adding files"
	case 2 << 15:
		return "preallocating"
	case 2 << 16:
		return "downloading"
	case 2 << 17:
		return "staging"
	case 2 << 18:
		return "committing"
	case 2 << 19:
		return "update stopping"
	default:
		return "unknown"
	}
}

type CrackStatus int

const (
	Unknown = iota
	Uncracked
	Cracked
)

type AppState struct {
	FromLibrary *Library
	CrackStatus

	Appid                           int32
	Universe                        int32
	Name                            string
	StateFlags                      StateFlag
	Installdir                      string
	LastUpdated                     int32
	LastPlayed                      int32
	SizeOnDisk                      uint64
	StaginSize                      int32
	Buildid                         int32
	LastOwner                       uint64
	UpdateResult                    int32
	BytesToDownload                 int32
	BytesDownloaded                 int32
	BytesToStage                    int32
	BytesStaged                     int32
	TargetBuildId                   int32
	AutoUpdateBehavior              int32
	AllowOtherDownloadsWhileRunning int32
	ScheduledAutoUpdate             int32
}

func (app AppState) Run() error {
	return exec.Command("xdg-open", fmt.Sprintf("steam://rungameid/%d", app.Appid)).Run()
}

// TODO: maybe search them by name? warframe had its name and launcher swapped
func NewAppState(vdf *vdf.Node, lib *Library) AppState {
	state := AppState{}
	state.FromLibrary = lib
	node := vdf.FirstChild()
	state.Appid = node.Int()
	node = node.NextChild()
	state.Universe = node.Int()
	node = node.NextChild()
	state.Name = node.String()
	node = node.NextChild()
	state.StateFlags = StateFlag(node.Int())
	node = node.NextChild()
	state.Installdir = node.String()
	node = node.NextChild()
	state.LastUpdated = node.Int()
	node = node.NextChild()
	state.LastPlayed = node.Int()
	node = node.NextChild()
	state.SizeOnDisk = node.Uint64()
	node = node.NextChild()
	state.StaginSize = node.Int()
	node = node.NextChild()
	state.Buildid = node.Int()
	node = node.NextChild()
	state.LastOwner = node.Uint64()
	node = node.NextChild()
	state.UpdateResult = node.Int()
	node = node.NextChild()
	state.BytesToDownload = node.Int()
	node = node.NextChild()
	state.BytesDownloaded = node.Int()
	node = node.NextChild()
	state.BytesToStage = node.Int()
	node = node.NextChild()
	state.BytesStaged = node.Int()
	node = node.NextChild()
	state.TargetBuildId = node.Int()
	node = node.NextChild()
	state.AutoUpdateBehavior = node.Int()
	node = node.NextChild()
	state.AllowOtherDownloadsWhileRunning = node.Int()
	node = node.NextChild()
	state.ScheduledAutoUpdate = node.Int()
	return state
}

// TODO: make some kind of registry in each library dir to cache these things
func (a *AppState) IsCracked() bool {
	if a.CrackStatus != Unknown {
		return a.CrackStatus == Cracked
	}

	appPath := path.Join(a.FromLibrary.Path, "steamapps/common", a.Installdir)
	result := false
	if _, err := os.Stat(appPath); err != nil {
		return false
	}

	godirwalk.Walk(appPath, &godirwalk.Options{
		Callback: func(path string, file *godirwalk.Dirent) error {
			_, name := filepath.Split(path)
			if name == "cream_api.ini" {
				result = true
				return godirwalk.SkipThis
			}
			return nil
		},
		Unsorted: true,
	})

	if result {
		a.CrackStatus = Cracked
	} else {
		a.CrackStatus = Uncracked
	}
	return result
}

func MoveFile(from, to string) {
	_, err := os.Stat(from)
	catchErr(err)
	fromFile, err := os.Open(from)
	catchErr(err)
	toFile, err := os.Create(to)
	catchErr(err)
	_, err = fromFile.WriteTo(toFile)
	catchErr(err)
	catchErr(os.Chmod(to, 0755))
	catchErr(fromFile.Close())
	catchErr(os.Remove(from))
	catchErr(toFile.Close())
}

func (a *AppState) ApplyCrack() {
	basePath := path.Join(a.FromLibrary.Path, "steamapps/common", a.Installdir)
	godirwalk.Walk(basePath, &godirwalk.Options{
		Callback: func(path string, file *godirwalk.Dirent) error {
			if slices.Index(searchNames, file.Name()) != -1 {
				dir, _ := filepath.Split(path)
				oldNameParts := strings.Split(file.Name(), ".")
				oldName := fmt.Sprintf("%v_o.%v", oldNameParts[0], oldNameParts[1])

				MoveFile(path, filepath.Join(dir, oldName))

				var fName, iniName string
				switch file.Name() {
				// TODO: maybe check if game lib is 32 bits?? what are we the 90's?
				case "libsteam_api.so":
					fName = "log_build/linux/x64/libsteam_api.so"
					iniName = "log_build/linux/x64/cream_api.ini"
				case "libsteam_api.dylib":
					fName = "log_build/macos/libsteam_api.so"
					iniName = "log_build/macos/cream_api.ini"
				case "steam_api.dll":
					fName = "log_build/windows/steam_api.dll"
					iniName = "log_build/windows/cream_api.ini"
				case "steam_api64.dll":
					fName = "log_build/windows/steam_api64.dll"
					iniName = "log_build/windows/cream_api.ini"
				}

				f, err := cream.ReadFile(fName)
				catchErr(err)
				ini, err := cream.ReadFile(iniName)
				catchErr(err)
				catchErr(os.WriteFile(path, f, fs.ModeAppend))
				creamFile, err := os.Create(filepath.Join(dir, "cream_api.ini"))
				catchErr(err)
				creamFile.Write(ini)
				creamFile.Close()
				a.CrackStatus = Cracked
			}
			return nil
		},
		Unsorted: true,
	})
}

func (a *AppState) RemoveCrack() {
	basePath := path.Join(a.FromLibrary.Path, "steamapps/common", a.Installdir)
	godirwalk.Walk(basePath, &godirwalk.Options{
		Callback: func(path string, file *godirwalk.Dirent) error {
			if file.Name() == "cream_api.ini" {
				if err := os.Remove(path); err != nil {
					panic(err)
				}
				return nil
			}

			if slices.Index(searchNames, file.Name()) != -1 {
				dir, _ := filepath.Split(path)
				oldNameParts := strings.Split(file.Name(), ".")
				oldName := fmt.Sprintf("%v_o.%v", oldNameParts[0], oldNameParts[1])
				if err := os.Remove(path); err != nil {
					panic(err)
				}
				MoveFile(filepath.Join(dir, oldName), path)
				a.CrackStatus = Uncracked
			}

			return nil
		},
		Unsorted: true,
	})

}

type App struct {
	LastPlayed string
	Playtime   string
	Cloud      struct {
		last_sync_state string
	}
	Autocloud struct {
		lastlaunch string
		lastexit   string
	}
	BadgeData     string
	LaunchOptions string
}

// Read VDF in text format
func ReadVdfA(path string) (*vdf.Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return _readVdf(data, false)
}

// Read VDF in binary format
func ReadVdfB(path string) (*vdf.Node, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return _readVdf(data, true)
}

func _readVdf(data []byte, binary bool) (*vdf.Node, error) {
	root := &vdf.Node{}
	if binary {
		if err := root.UnmarshalBinary(data); err != nil {
			return nil, err
		}
	} else {
		if err := root.UnmarshalText(data); err != nil {
			return nil, err
		}
	}
	return root, nil
}
