package config

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"q/types"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 12

var (
	greyStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("240"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Faint(true).Margin(1, 0, 2, 4)
)

// type item string

// func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(menuItem)
	if !ok {
		return
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {

			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	text := fn(i.title)
	if i.data != "" {
		text = fmt.Sprintf("%s %s", text, greyStyle.Render("("+i.data+")"))
	}

	fmt.Fprint(w, text)
}

func quit() tea.Cmd {
	return func() tea.Msg {
		return quitMsg{}
	}
}

type quitMsg struct{}

type setMenuMsg struct {
	state state
	menu  menuFunc
}

type setStateMsg struct {
	state state
}

func setMenu(menu menuFunc) tea.Cmd {
	return func() tea.Msg { return setMenuMsg{menu: menu} }
}

type backMsg struct{}

func back() tea.Cmd {
	return func() tea.Msg { return backMsg{} }
}

type updateConfigMsg struct {
	appConfig AppConfig
}

type editorFinishedMsg struct{ err error }

func openEditor() tea.Cmd {
	fullPath, err := FullFilePath()
	if err != nil {
		return tea.Cmd(func() tea.Msg { return editorFinishedMsg{err} })
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	c := exec.Command(editor, fullPath) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

type configSavedMsg struct{}

func saveConfig(config AppConfig) tea.Cmd {
	return func() tea.Msg {
		SaveAppConfig(config)
		return configSavedMsg{}
	}
}

func updateConfig(config AppConfig) tea.Cmd {
	return func() tea.Msg { return updateConfigMsg{config} }
}

type setDefaultModelMsg struct {
	model string
}

func setDefaultModel(model string) tea.Cmd {
	return func() tea.Msg { return setDefaultModelMsg{model} }
}

// func saveConfig(appConfig AppConfig) tea.Cmd {
// 	return func() tea.Msg {
// 		return saveConfigMsg{}
// 	}
// }

type menuItem struct {
	title     string
	selectCmd tea.Cmd
	data      string
}

func (i menuItem) FilterValue() string { return i.title }

type menuFunc func(config AppConfig) list.Model

type menuModel struct {
	title             string
	items             []menuItem
	lastSelectedIndex int
}

func (m menuModel) ListItems() []list.Item {
	menuItems := m.items
	listItems := make([]list.Item, len(menuItems))

	for i, item := range menuItems {
		listItems[i] = item
	}

	return listItems
}

type page int

const (
	ListPage page = iota
)

type state struct {
	page      page
	menu      menuFunc
	listIndex int
	model     string
}

type model struct {
	state state

	list list.Model

	dirty     bool
	backstack []state

	appConfig AppConfig

	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case quitMsg:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyCtrlD {
			return m, quit()
		}
	case backMsg:
		if len(m.backstack) > 0 {
			m.state = m.backstack[len(m.backstack)-1]
			m.backstack = m.backstack[:len(m.backstack)-1]
			m.list = m.state.menu(m.appConfig)
			m.list.Select(m.state.listIndex)
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'q' {
			m.quitting = true
			return m, tea.Quit
		}
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			return m, quit()

		case tea.KeyEsc:
			if len(m.backstack) > 0 {
				return m, back()
			}
			return m, quit()

		case tea.KeyEnter:
			i, _ := m.list.SelectedItem().(menuItem)
			return m, i.selectCmd

		case tea.KeyBackspace:
			return m, back()
		}
	case quitMsg:
		m.quitting = true
		return m, tea.Quit

	case setMenuMsg:
		m.backstack = append(m.backstack, m.state)
		m.list = msg.menu(m.appConfig)

	case setDefaultModelMsg:
		m.appConfig.Preferences.DefaultModel = msg.model
		// fmt.Println("Config:", m.appConfig.Preferences.DefaultModel)
		return m, tea.Sequence(saveConfig(m.appConfig), back())

	case editorFinishedMsg:
		if msg.err != nil {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	if !m.quitting {
		m.list, cmd = m.list.Update(msg)
	}
	m.state.listIndex = m.list.Index()
	return m, cmd
}

func (m *model) handleSelect() {

}

func (m model) View() string {
	if m.quitting {
		return ""
		// return quitTextStyle.Render("Changes saved to ~/.shell-ai/config.yaml")
	}
	return "\n" + m.list.View()
}

func listFromMenu(m menuModel) list.Model {
	l := list.New(m.ListItems(), itemDelegate{}, 20, listHeight)
	l.Title = m.title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.SetWidth(100)
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.SetShowHelp(false)
	l.Select(m.lastSelectedIndex)
	return l
}

func defaultList(title string, items []menuItem) list.Model {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	l := list.New(listItems, itemDelegate{}, 20, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.SetWidth(100)
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.SetShowHelp(false)
	return l
}

// func (m menuModel)

func mainMenu(appConfig AppConfig) list.Model {
	items := []menuItem{
		{
			title:     "Change Default Model",
			data:      appConfig.Preferences.DefaultModel,
			selectCmd: setMenu(defaultModelSelectMenu),
		},
		// {
		// 	title:     "Configure Models",
		// 	selectCmd: setMenu(configureModelsMenu),
		// },
		{
			title:     "Edit Config File",
			data:      "~/.shell-ai/config.yaml",
			selectCmd: openEditor(),
		},
		{
			title:     "Quit",
			data:      "esc",
			selectCmd: quit(),
		},
	}
	return defaultList("ShellAI Config", items)
}

func defaultModelSelectMenu(appConfig AppConfig) list.Model {
	var modelItems []menuItem
	for _, model := range appConfig.Models {
		model := model
		modelItems = append(modelItems, menuItem{
			title:     model.ModelName,
			selectCmd: tea.Sequence(setDefaultModel(model.ModelName), back()),
		})
	}
	return defaultList("Choose Default Model", modelItems)
}

func configureModelsMenu(appConfig AppConfig) list.Model {
	var modelItems []menuItem
	for _, model := range appConfig.Models {
		model := model
		modelItems = append(modelItems, menuItem{
			title:     model.ModelName,
			selectCmd: setMenu(modelDetailsMenu(model)),
		})
	}
	modelItems = append(modelItems, menuItem{
		title: "Add Model",
		// selectCmd: a
	})
	return defaultList("Configure Models", modelItems)
}

func modelDetailsMenu(modelConfig types.ModelConfig) menuFunc {
	return func(c AppConfig) list.Model {
		return modelDetailsForModelMenu(c, modelConfig)
	}
}

func modelDetailsForModelMenu(appConfig AppConfig, modelConfig types.ModelConfig) list.Model {
	items := []menuItem{
		{
			title: "Name: " + modelConfig.ModelName,
		},
		{
			title: "Endpoint: " + modelConfig.Endpoint,
		},
		{
			title: "Auth: " + modelConfig.Auth,
		},
		{
			title: "Auth: " + modelConfig.Auth,
		},
		{
			title: "Prompt",
		},
	}
	return defaultList(modelConfig.ModelName, items)
}

func RunConfigProgram() {

	appConfig, err := LoadAppConfig()
	if err != nil {
		panic(err)
	}

	m := model{
		appConfig: appConfig,
		list:      mainMenu(appConfig),
		state:     state{page: ListPage, menu: mainMenu},
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
