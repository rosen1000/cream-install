package main

// https://developer.valvesoftware.com/wiki/Steam_browser_protocol
import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	vdf2 "github.com/BenLubar/vdf"
	"github.com/andygrunwald/vdf"
	"github.com/k0kubun/pp"
	bvdf "github.com/wakeful-cloud/vdf"
)

const (
	STEAM_HOME string = ".steam/steam"
)

func GetUsers() {
	p := path.Join(getHome(), STEAM_HOME, "userdata")
	userIds, err := os.ReadDir(p)
	if err != nil {
		panic(err)
	}
	for _, user := range userIds {
		id := user.Name()
		file, err := os.Open(path.Join(getHome(), STEAM_HOME, "userdata", id, "config/localconfig.vdf"))
		if err != nil {
			panic(err)
		}
		parser := vdf.NewParser(file)
		data, err := parser.Parse()
		if err != nil {
			panic(err)
		}
		USED(data)
		// fmt.Println(data["UserLocalConfigStore"].(map[string]interface{})["friends"].(map[string]interface{})["PersonaName"])
	}
}

func ReadShortcuts() (any, error) {
	file, err := os.Open("/home/hax/.steam/steam/userdata/169122681/config/shortcuts.vdf")
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	m, err := bvdf.ReadVdf(data)
	if err != nil {
		return nil, err
	}
	pp.Print(m)
	return nil, nil
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
	manifests := []string{}
	libs := getLibraries()
	for _, lib := range libs {
		manif, err := filepath.Glob(filepath.Join(lib.Path, "steamapps", "appmanifest_*.acf"))
		if err != nil {
			panic(manif)
		}
		manifests = append(manifests, manif...)
	}
	var states []AppState
	for _, match := range manifests {
		vdf, err := ReadVdfA(match)
		if err != nil {
			panic(err)
		}
		states = append(states, NewAppState(vdf))
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

type AppState struct {
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
func NewAppState(vdf *vdf2.Node) AppState {
	state := AppState{}
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

type App struct {
	LastPlayed string
	Playtime   string
	Cloud      struct {
		last_sync_state string
	} `vdf:"cloud"`
	Autocloud struct {
		lastlaunch string
		lastexit   string
	} `vdf:"autocloud"`
	BadgeData     string
	LaunchOptions string
}

// Read VDF in text format
func ReadVdfA(path string) (*vdf2.Node, error) {
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
func ReadVdfB(path string) (*vdf2.Node, error) {
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

func _readVdf(data []byte, binary bool) (*vdf2.Node, error) {
	root := &vdf2.Node{}
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
