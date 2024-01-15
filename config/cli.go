package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"q/types"
	"q/util"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 12

var (
	styleRed          = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
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
	fullPath, err := FullFilePath(configFilePath)
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

func openBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		util.OpenBrowser(url)
		return nil
	}
}

func openGithubRepo() tea.Cmd {
	return func() tea.Msg {
		util.OpenBrowser("https://github.com/ibigio/shell-ai?tab=readme-ov-file#contributing")
		return nil
	}
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
		m.state = state{page: ListPage, menu: msg.menu}

	case setDefaultModelMsg:
		m.appConfig.Preferences.DefaultModel = msg.model
		// fmt.Println("Config:", m.appConfig.Preferences.DefaultModel)
		return m, tea.Sequence(saveConfig(m.appConfig), back())

	case editorFinishedMsg:
		if msg.err != nil {
			return m, quit()
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
		{
			title:     "Edit Config File",
			data:      "~/.shell-ai/config.yaml",
			selectCmd: openEditor(),
		},
		{
			title:     "Configure Models",
			selectCmd: setMenu(configureModelsMenu),
		},
		{
			title:     "Contribute",
			selectCmd: openBrowser("https://github.com/ibigio/shell-ai#contributing"),
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
		title:     "Add Model",
		data:      "coming soon!",
		selectCmd: openBrowser("https://github.com/ibigio/shell-ai#custom-model-configuration-new"),
	})
	modelItems = append(modelItems, menuItem{
		title:     "Install Model",
		data:      "coming soon!",
		selectCmd: openBrowser("https://github.com/ibigio/shell-ai#custom-model-configuration-new"),
	})
	return defaultList("Configure Models (coming soon!)", modelItems)
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
	return defaultList(modelConfig.ModelName+"(editing coming soon!)", items)
}

func PrintConfigErrorMessage(err error) {
	maxWidth := util.GetTermSafeMaxWidth()
	styleRed := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).PaddingLeft(2)
	styleDim := lipgloss.NewStyle().Faint(true).Width(maxWidth).PaddingLeft(2)

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)

	msg1 := styleRed.Render("Failed to load config file.")

	filePath, _ := FullFilePath(configFilePath)
	msg2 := styleDim.Render(err.Error())
	revertConfigCmd := "q config revert"
	resetConfigCmd := "q config reset"

	// Concatenate message string with backticks
	message_string := fmt.Sprintf(
		"---\n"+
			"# Options:\n\n"+
			"1. Run `%s` to load the automatic backup - you're welcome.\n"+
			"2. Nuke it. Run `%s` to reset the config to default.\n"+
			"3. DIY - take a look at the config and fix the errors. It's at:\n\n"+
			" `%s`\n\n",
		revertConfigCmd, resetConfigCmd, filePath)

	msg3, _ := r.Render(message_string)

	fmt.Printf("\n%s\n\n%s%s", msg1, msg2, msg3)
}

func handleConfigResets(args []string) {
	if len(args) < 2 {
		return
	}

	greyStylePadded := greyStyle.PaddingLeft(2)
	reader := bufio.NewReader(os.Stdin)

	warningMessage, confirmationMessage := getMessages(args[1], greyStylePadded)
	fmt.Print("\n" + styleRed.PaddingLeft(2).Render(warningMessage) + "\n\n" + confirmationMessage + " ")

	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "yes" || response == "y" {
		handleResetOrRevert(args[1])
	} else {
		fmt.Println("\n" + styleRed.PaddingLeft(2).Render("Operation cancelled.\n"))
	}
	os.Exit(0)
}

func getMessages(arg string, greyStylePadded lipgloss.Style) (string, string) {
	warningMessage := "WARNING: You are about to "
	confirmationMessage := greyStylePadded.Render("Do you want to continue? (y/N):")

	switch arg {
	case "reset":
		warningMessage += "reset the config file to the default."
	case "revert":
		warningMessage += "revert the config file to the last working automatic backup."
	}

	return warningMessage, confirmationMessage
}

func handleResetOrRevert(arg string) {
	var (
		err     error
		message string
	)

	switch arg {
	case "reset":
		err = ResetAppConfigToDefault()
		message = "Config reset to default.\n"
	case "revert":
		err = RevertAppConfigToBackup()
		message = "Config reverted to backup.\n"
	}

	if err == nil {
		fmt.Println("\n" + greyStyle.PaddingLeft(2).Render(message))
	} else {
		fmt.Println("\n" + styleRed.PaddingLeft(2).Render("Operation failed.\n"))
		fmt.Println("\n" + styleRed.PaddingLeft(2).Render(fmt.Sprintf("Error: %s\n", err)))
	}
}

func RunConfigProgram(args []string) {

	handleConfigResets(args)

	appConfig, err := LoadAppConfig()
	if err != nil {
		PrintConfigErrorMessage(err)
		os.Exit(1)
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
