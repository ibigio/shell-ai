package cli

import (
	"fmt"
	"os"
	"q/openai"
	"strings"

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
	ReceivingResponse
)

type model struct {
	client           *openai.OpenAIClient
	markdownRenderer *glamour.TermRenderer
	p                *tea.Program

	textInput textinput.Model
	spinner   spinner.Model

	state                 State
	query                 string
	latestCommandResponse string
	latestCommandIsCode   bool

	partialResponse          string
	formattedPartialResponse string

	maxWidth int

	runWithArgs bool
	err         error
}

type loadMsg struct{}
type responseMsg struct {
	response string
	err      error
}
type queryMsg struct{}
type partialResponseMsg struct {
	content string
	err     error
}
type setPMsg struct{ p *tea.Program }

// === Commands === //

func makeQuery(client *openai.OpenAIClient, query string) tea.Cmd {
	return func() tea.Msg {
		response, err := client.Query(query)
		return responseMsg{response: response, err: err}
	}
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
		placeholderStyle := lipgloss.NewStyle().Faint(true)
		message := "Copied to clipboard."
		if !m.latestCommandIsCode {
			message = "Copied only code to clipboard."
		}
		message = placeholderStyle.Render(message)
		return m, tea.Sequence(tea.Printf(message), tea.Quit)
	}
	// Input, run query.
	m.textInput.SetValue("")
	m.query = v
	m.state = Loading
	placeholderStyle := lipgloss.NewStyle().Faint(true)
	message := placeholderStyle.Render(fmt.Sprintf("> %s", v))
	return m, tea.Sequence(tea.Printf(message), tea.Batch(m.spinner.Tick, makeQuery(m.client, m.query)))
}

func (m model) formatResponse(response string, isCode bool) (string, error) {

	// format nicely
	formatted, err := m.markdownRenderer.Render(response)
	if err != nil {
		// TODO: handle error
		panic(err)
	}

	// trim preceding and trailing newlines
	formatted = strings.TrimPrefix(formatted, "\n")
	formatted = strings.TrimSuffix(formatted, "\n")

	// Add newline for non-code blocks (hacky)
	if !isCode {
		formatted = "\n" + formatted
	}
	return formatted, nil
}

func (m model) handleResponseMsg(msg responseMsg) (tea.Model, tea.Cmd) {
	m.formattedPartialResponse = ""

	// really shitty error handling but it's better than nothing
	if msg.err != nil {
		m.state = RecevingInput
		styleRed := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		styleDim := lipgloss.NewStyle().Faint(true)

		message := fmt.Sprintf("\n  %v\n\n  %v\n",
			styleRed.Render("Error: Failed to connect to OpenAI."),
			styleDim.Render(msg.err.Error()))
		return m, tea.Sequence(tea.Printf(message), textinput.Blink)
	}

	// parse out the code block
	content, isOnlyCode := extractFirstCodeBlock(msg.response)
	if content != "" {
		m.latestCommandResponse = content
	}

	formatted, err := m.formatResponse(msg.response, startsWithCodeBlock(msg.response))
	if err != nil {
		// TODO: handle error
		panic(err)
	}

	m.textInput.Placeholder = "Follow up, ENTER to copy & quit, CTRL+C to quit"
	if !isOnlyCode {
		m.textInput.Placeholder = "Follow up, ENTER to copy (code only), CTRL+C to quit"
	}
	if m.latestCommandResponse == "" {
		m.textInput.Placeholder = "Follow up, ENTER or CTRL+C to quit"
	}

	m.state = RecevingInput
	m.latestCommandIsCode = isOnlyCode
	message := formatted
	return m, tea.Sequence(tea.Printf(message), textinput.Blink)
}

func (m model) handlePartialResponseMsg(msg partialResponseMsg) (tea.Model, tea.Cmd) {
	m.state = ReceivingResponse
	isCode := startsWithCodeBlock(msg.content)
	formatted, err := m.formatResponse(msg.content, isCode)
	if err != nil {
		// TODO: handle error
		panic(err)
	}
	m.formattedPartialResponse = formatted
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
	switch m.state {
	case Loading:
		return m.spinner.View()
	case RecevingInput:
		return m.textInput.View()
	case ReceivingResponse:
		return m.formattedPartialResponse + "\n"
	}
	return ""
}

// === Initial Model Setup === //

func initialModel(prompt string, client *openai.OpenAIClient) model {
	maxWidth := 100

	ti := textinput.New()
	ti.Placeholder = "Describe a shell command, or ask a question."
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
		latestCommandIsCode:   false,
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

var RootCmd = &cobra.Command{
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
		c := openai.NewClient(apiKey, modelOverride)
		p := tea.NewProgram(initialModel(prompt, c))
		c.StreamCallback = streamHandler(p)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}
