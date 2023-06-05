package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type State int

const (
	Loading State = iota
	RecevingInput
	Copying
)

type model struct {
	client           *OpenAIClient
	markdownRenderer *glamour.TermRenderer
	p                *tea.Program

	textInput textinput.Model
	spinner   spinner.Model

	state                 State
	query                 string
	latestCommandResponse string

	partialResponse          string
	formattedPartialResponse string

	drawBuffer string
	maxWidth   int

	runWithArgs bool
	err         error
}

type loadMsg struct{}
type responseMsg struct {
	response string
	err      error
}
type queryMsg struct{}
type drawOutputMsg struct{}
type partialResponseMsg struct {
	content string
	err     error
}
type setPMsg struct{ p *tea.Program }

// === Commands === //

func makeQuery(client *OpenAIClient, query string) tea.Cmd {
	return func() tea.Msg {
		response, err := client.Query(query)
		return responseMsg{response: response, err: err}
	}
}

func drawOutputCommand() tea.Msg {
	// wait for the screen to clear
	// (this is kind of hacky but it works)
	time.Sleep(50 * time.Millisecond)
	return drawOutputMsg{}
}

// === Msg Handlers === //

func (m model) handleKeyEnter() (tea.Model, tea.Cmd) {
	if m.state != RecevingInput {
		return m, nil
	}
	v := m.textInput.Value()

	// No input, copy and quit.
	if v == "" {
		if m.latestCommandResponse == "" {
			return m, tea.Quit
		}
		err := clipboard.WriteAll(m.latestCommandResponse)
		if err != nil {
			fmt.Println("Failed to copy text to clipboard:", err)
			return m, tea.Quit
		}
		m.state = Copying
		return m, tea.Quit
	}
	// Input, run query.
	m.textInput.SetValue("")
	m.query = v
	m.state = Loading
	placeholderStyle := lipgloss.NewStyle().Faint(true)
	m.drawBuffer = placeholderStyle.Render(fmt.Sprintf("> %s\n", v))
	return m, tea.Sequence(drawOutputCommand, tea.Batch(m.spinner.Tick, makeQuery(m.client, m.query)))
}

func (m model) formatResponse(response string, isCode bool) (string, error) {

	// format nicely
	formatted, err := m.markdownRenderer.Render(response)
	if err != nil {
		// TODO: handle error
		panic(err)
	}

	// trim trailing newlines
	formatted = strings.TrimPrefix(formatted, "\n")

	// Add newline for non-code blocks (hacky)
	if !isCode {
		formatted = "\n" + formatted
	}
	return formatted, nil
}

func (m model) handleResponseMsg(msg responseMsg) (tea.Model, tea.Cmd) {
	m.textInput.Placeholder = "Follow up, ENTER to copy & quit, ESC to quit"
	m.formattedPartialResponse = ""

	// really shitty error handling but it's better than nothing
	if msg.err != nil {
		m.state = RecevingInput
		styleRed := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		styleDim := lipgloss.NewStyle().Faint(true)

		m.drawBuffer = fmt.Sprintf("\n  %v\n\n  %v\n\n",
			styleRed.Render("Error: Failed to connect to OpenAI."),
			styleDim.Render(msg.err.Error()))
		return m, tea.Sequence(drawOutputCommand, textinput.Blink)
	}

	// parse out the code block
	code, isCode := extractCodeBlock(msg.response)
	m.latestCommandResponse = code

	formatted, err := m.formatResponse(msg.response, isCode)
	if err != nil {
		// TODO: handle error
		panic(err)
	}

	m.state = RecevingInput
	m.latestCommandResponse = code
	m.drawBuffer = formatted
	return m, tea.Sequence(drawOutputCommand, textinput.Blink)
}

func (m model) handlePartialResponseMsg(msg partialResponseMsg) (tea.Model, tea.Cmd) {
	formatted, err := m.formatResponse(msg.content, false)
	if err != nil {
		// TODO: handle error
		panic(err)
	}
	m.formattedPartialResponse = formatted

	return m, nil
}

func (m model) handleDrawOutputMsg() (tea.Model, tea.Cmd) {
	if m.drawBuffer != "" {
		fmt.Printf("%s", m.drawBuffer)
	}
	m.drawBuffer = ""
	return m, nil
}

// === Init, Update, View === //

func (m model) Init() tea.Cmd {
	if m.runWithArgs {
		return tea.Batch(m.spinner.Tick, makeQuery(m.client, m.query))
	}
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyCtrlD:
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleKeyEnter()
		}

	case drawOutputMsg:
		return m.handleDrawOutputMsg()

	case responseMsg:
		return m.handleResponseMsg(msg)

	case partialResponseMsg:
		return m.handlePartialResponseMsg(msg)

	case setPMsg:
		m.p = msg.p
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}
	// Update spinner or cursor.
	switch m.state {
	case Loading:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case RecevingInput:
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	var str string
	if m.formattedPartialResponse != "" {
		str += m.formattedPartialResponse
	}
	if m.state == Copying {
		placeholderStyle := lipgloss.NewStyle().Faint(true)
		return placeholderStyle.Render("Copied to clipboard.\n")
	}
	if m.drawBuffer != "" {
		return ""
	}
	switch m.state {
	case Loading:
		str += m.spinner.View()
	case RecevingInput:
		str += m.textInput.View()
	}
	return str
}

// === Initial Model Setup === //

func initialModel(prompt string, client *OpenAIClient) model {
	maxWidth := 100

	ti := textinput.New()
	ti.Placeholder = "Enter natural langauge query"
	ti.Focus()
	ti.CharLimit = maxWidth

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	runWithArgs := prompt != ""

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)
	model := model{
		client:                client,
		markdownRenderer:      r,
		textInput:             ti,
		spinner:               s,
		state:                 RecevingInput,
		query:                 "",
		latestCommandResponse: "",
		drawBuffer:            "",
		maxWidth:              maxWidth,
		runWithArgs:           false,
		err:                   nil,
	}

	if runWithArgs {
		model.runWithArgs = true
		model.state = Loading
		model.query = prompt
	}
	return model
}

// === Main === //

func printAPIKeyNotSetMessage() {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)

	styleRed := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	msg1 := styleRed.Render("OPENAI_API_KEY environment variable not set.")
	msg2, _ := r.Render(`
	1. Generate your API key at https://platform.openai.com/account/api-keys
	2. Add your credit card in the API (for the free trial)
	2. Set your key by running:
	` + "\n```bash\nexport OPENAPI_API_KEY=[your key]\n```")

	fmt.Printf("\n  %v%v", msg1, msg2)
}

func streamHandler(p *tea.Program) func(content string, err error) {
	return func(content string, err error) {
		p.Send(partialResponseMsg{content, err})
	}
}

var rootCmd = &cobra.Command{
	Use:   "q [request]",
	Short: "A command line interface for natural language queries",
	Run: func(cmd *cobra.Command, args []string) {
		// join args into a single string separated by spaces
		prompt := strings.Join((args), " ")
		apiKey := os.Getenv("OPENAI_API_KEY")
		modelOverride := os.Getenv("OPENAI_MODEL_OVERRIDE")
		if apiKey == "" {
			printAPIKeyNotSetMessage()
			os.Exit(1)
		}
		c := NewClient(apiKey, modelOverride)
		p := tea.NewProgram(initialModel(prompt, c))
		c.streamCallback = streamHandler(p)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
