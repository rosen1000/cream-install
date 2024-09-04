package main

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"time"

	_ "github.com/charmbracelet/bubbles/help"
	_ "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func InitModel() Model {
	return Model{
		choices:  []string{"test1", "test2", "test3"},
		selected: make(map[int]struct{}),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "w", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "s", "down":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}
func (m Model) View() string {
	s := "Which game to touch?\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		cheched := " "
		if _, ok := m.selected[i]; ok {
			cheched = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, cheched, choice)
	}

	return s
}

type Mode int

const (
	NormalM Mode = iota
	SearchM
	EditM
	RunM
	ExitM
)

func (m Mode) String() string {
	switch m {
	case NormalM:
		return "normal"
	case SearchM:
		return "search"
	}
	return ""
}

type GameModel struct {
	games    []AppState
	display  []*AppState
	scroll   int
	Selected *AppState
	vp       viewport.Model
	mode     Mode
	search   string
	cursor   int
}

func NewGameModel(games []AppState) GameModel {
	width, height, err := term.GetSize(os.Stdin.Fd())
	if err != nil {
		panic(err)
	}

	slices.SortFunc(games, func(i, j AppState) int {
		return int(j.LastPlayed) - int(i.LastPlayed)
	})

	vp := viewport.New(width-4, height-2)
	display := []*AppState{}
	for _, game := range games {
		display = append(display, &game)
	}
	return GameModel{games: games, display: display, vp: vp}
}

func (m GameModel) Init() tea.Cmd {
	return nil
}
func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	needSearch := false
	switch m.mode {
	case NormalM:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				m.mode = ExitM
				return m, tea.Quit
			case "w", "up":
				if m.scroll > 0 {
					m.scroll--
					if m.scroll < len(m.games)-m.vp.Height/2 {
						m.vp.LineUp(1)
					}
				}
			case "s", "down":
				if m.scroll < len(m.games)-1 {
					m.scroll++
					if m.scroll > m.vp.Height/2 {
						m.vp.LineDown(1)
					}
				}
			case " ", "enter":
				m.Selected = m.display[m.scroll]
				m.mode = RunM
				return m, tea.Quit
			case "?", "ctrl+f":
				m.mode = SearchM
				m.vp.Height -= 2
			case "e":
				m.Selected = m.display[m.scroll]
				m.mode = EditM
				return m, tea.Quit
			}
		}
	case SearchM:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc", "enter":
				m.mode = NormalM
				m.search = ""
				m.cursor = 0
				m.vp.Height += 2
				return m, nil
			case "backspace":
				needSearch = true
				if m.cursor > 0 {
					m.search = m.search[:m.cursor-1] + m.search[m.cursor:]
					m.cursor--
				}
			case "delete":
				needSearch = true
				if m.cursor < len(m.search) {
					m.search = m.search[:m.cursor] + m.search[m.cursor+1:]
				}
			case "ctrl+a", "home":
				m.cursor = 0
			case "ctrl+e", "end":
				m.cursor = len(m.search)
			case "left":
				if m.cursor > 0 {
					m.cursor--
				}
			case "right":
				if m.cursor < len(m.search) {
					m.cursor++
				}
			case "ctrl+left":
				for i := m.cursor - 2; i >= 0; i-- {
					if m.search[i] == ' ' {
						m.cursor = i + 1
						break
					}
					if i == 0 {
						m.cursor = i
					}
				}
			case "ctrl+right":
				for i := m.cursor + 1; i < len(m.search); i++ {
					if m.search[i] == ' ' {
						m.cursor = i + 1
						break
					}
					if i == len(m.search)-1 {
						m.cursor = i + 1
					}
				}
			default:
				needSearch = true
				str := msg.String()
				if len(str) == 1 {
					m.search = m.search[:m.cursor] + str + m.search[m.cursor:]
					m.cursor++
				}
			}
		}
	}

	// Search/filter games
	if needSearch {
		ranked := fuzzy.RankFindFold(m.search, func() []string {
			names := []string{}
			for _, game := range m.games {
				names = append(names, game.Name)
			}
			return names
		}())
		m.display = []*AppState{}
		sort.Sort(ranked)
		for _, rank := range ranked {
			// if rank.Distance >= 0 {
			// 	break
			// }
			m.display = append(m.display, &m.games[rank.OriginalIndex])
		}
		if len(m.search) > 0 && len(m.display) == 0 {
			for _, game := range m.games {
				m.display = append(m.display, &game)
			}
		}
	}

	// Rerender main view
	str := ""
	if len(m.display) == 0 {
		str = "No games found!"
	}
	for i, game := range m.display {
		cursor := ' '
		if i == m.scroll {
			cursor = '>'
		}
		str += fmt.Sprintf("\000%c %s\n", cursor, game.Name)
	}
	m.vp.SetContent(str)

	return m, nil
}
func (m GameModel) View() string {
	bottomBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.Border{Bottom: "─", Left: "│", Right: "│", BottomLeft: "╰", BottomRight: "╯"}, false, true, true).
		BorderForeground(lipgloss.Color("62")).
		PaddingLeft(2).
		SetString(">")
	topBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.Border{Top: "─", Bottom: "─", Left: "│", Right: "│", BottomLeft: "├", BottomRight: "┤", TopLeft: "╭", TopRight: "╮"}).
		BorderForeground(lipgloss.Color("62")).
		PaddingLeft(2)
	cursorStyle := lipgloss.NewStyle().Background(lipgloss.Color("#ffffff")).Foreground(lipgloss.Color("00"))
	boxStyle := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).PaddingLeft(2)

	search := viewport.New(m.vp.Width-2, 1)
	searchStr := m.search[:m.cursor]
	if len(m.search) == m.cursor {
		searchStr += cursorStyle.Render(" ")
	} else if len(m.search) > 0 {
		searchStr += cursorStyle.Render(string(m.search[m.cursor]))
		searchStr += m.search[m.cursor+1:]
	}
	search.SetContent(searchStr)

	var window string
	if m.mode == NormalM {
		window = boxStyle.Render(m.vp.View())
	} else if m.mode == SearchM {
		window = lipgloss.JoinVertical(
			lipgloss.Top,
			topBoxStyle.Render(m.vp.View()),
			bottomBoxStyle.Render(search.View()),
		)
	}

	return window
	// return lipgloss.NewStyle().
	// 	Border(lipgloss.NormalBorder()).
	// 	Padding(1).
	// 	Render(m.vp.View())
}

// Edit page for a single game
type EditModel struct {
	vp    viewport.Model
	game  *AppState
	table table.Model
}

func NewEditModel(game *AppState) EditModel {
	width, height, err := term.GetSize(os.Stdin.Fd())
	if err != nil {
		panic(err)
	}
	vp := viewport.New(width-4, height-4)
	tabl := table.New(
		table.WithColumns([]table.Column{{Title: "Keys", Width: 12}, {Title: "Values", Width: 32}}),
		table.WithRows([]table.Row{
			{"App Id", strconv.Itoa(int(game.Appid))},
			{"Name", game.Name},
			{"Install dir", game.Installdir},
			{"Last Played", time.Unix(int64(game.LastPlayed), 0).Format("15:04:05 02/01/06")},
			{"Last Updated", time.Unix(int64(game.LastUpdated), 0).Format("15:04:05 02/01/06")},
			{"State Flags", strconv.Itoa(int(game.StateFlags))},
		}),
		table.WithFocused(false),
		table.WithHeight(height-3),
		table.WithWidth(width-4),
	)
	st := table.DefaultStyles()
	st.Selected = st.Cell.UnsetPaddingLeft()
	tabl.SetStyles(st)
	return EditModel{vp: vp, game: game, table: tabl}
}

func (m EditModel) Init() tea.Cmd { return nil }
func (m EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.table, _ = m.table.Update(msg)
	return m, nil
}
func (m EditModel) View() string {
	boxStyle := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).PaddingLeft(2)

	return boxStyle.Render(m.table.View())
}
