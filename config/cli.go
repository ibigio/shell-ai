package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	greyStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("240"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

// type item string

// func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(menuItemModel)
	if !ok {
		return
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(i.title))
}

type setMenuMsg struct {
	menu func() menuModel
}

type menuItemModel struct {
	list.Item
	title     string
	selectCmd tea.Cmd
}

func (i menuItemModel) FilterValue() string { return i.title }

type menuModel struct {
	title string
	items []menuItemModel
}

func (m menuModel) ListItems() []list.Item {
	menuItems := m.items
	listItems := make([]list.Item, len(menuItems))

	for i, item := range menuItems {
		listItems[i] = item
	}

	return listItems
}

type model struct {
	menu menuModel
	list list.Model

	appConfig AppConfig

	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) handleSetMenu(menu menuModel) (tea.Model, tea.Cmd) {
	m.menu = menu
	m.list.Title = menu.title
	m.list.SetItems(menu.ListItems())
	return m, tea.Quit
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyCtrlD:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			i, _ := m.list.SelectedItem().(menuItemModel)
			return m, i.selectCmd
		}
	case setMenuMsg:
		m.menu = msg.menu()
		m.list = listFromMenu(m.menu)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("Not hungry? That’s cool.")
	}
	return "\n" + m.list.View()
}

func listFromMenu(m menuModel) list.Model {
	l := list.New(m.ListItems(), itemDelegate{}, 20, listHeight)
	l.Title = m.title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	return l
}

func (m model) mainMenuModel() menuModel {

	return menuModel{
		title: "ShellAI Config",
		items: []menuItemModel{
			{
				title:     "Change Default Model (" + m.appConfig.Preferences.DefaultModel + ")",
				selectCmd: func() tea.Msg { return setMenuMsg{menu: m.defaultModelMenu} },
			},
			{
				title:     "Configure Models",
				selectCmd: func() tea.Msg { return tea.KeyCtrlC },
			},
			{
				title:     "Edit Config File (.shell-ai/config.yaml)",
				selectCmd: func() tea.Msg { return tea.KeyCtrlC },
			},
			{
				title:     "Quit",
				selectCmd: func() tea.Msg { return tea.KeyCtrlC },
			},
		},
	}
}

func (m model) defaultModelMenu() menuModel {

	var modelItems []menuItemModel
	for _, model := range m.appConfig.Models {
		model := model
		modelItems = append(modelItems, menuItemModel{
			title:     model.ModelName,
			selectCmd: func() tea.Msg { return tea.KeyCtrlC },
		})
	}

	return menuModel{
		title: "ShellAI Config > Default Model",
		items: modelItems,
	}
}

func RunConfigProgram() {

	appConfig, err := LoadAppConfig()
	if err != nil {
		panic(err)
	}

	// mainMenu := mainMenuModel()
	// l := listFromMenu(mainMenu)

	m := model{
		appConfig: appConfig,
	}
	m.menu = m.mainMenuModel()
	m.list = listFromMenu(m.menu)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
