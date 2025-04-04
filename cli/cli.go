package cli

import (
	"fmt"
	"os"
	"q/config"
	"q/llm"
	. "q/types"
	"q/util"

	"runtime"
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
	client           *llm.LLMClient
	markdownRenderer *glamour.TermRenderer
	p                *tea.Program

	textInput textinput.Model
	spinner   spinner.Model

	state                 State
	query                 string
	latestCommandResponse string
	latestCommandIsCode   bool

	formattedPartialResponse string

	maxWidth int

	runWithArgs bool
	err         error
}

type responseMsg struct {
	response string
	err      error
}
type partialResponseMsg struct {
	content string
	err     error
}
type setPMsg struct{ p *tea.Program }

// === Commands === //

func makeQuery(client *llm.LLMClient, query string) tea.Cmd {
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

	if v == "" {
		if m.latestCommandResponse == "" {
			return m, tea.Quit
		}
		err := clipboard.WriteAll(m.latestCommandResponse)
		if err != nil {
			// Print a user-friendly message about missing clipboard utilities
			clipboardErrMsg := "Error: Failed to copy to clipboard. Please install 'xclip', 'xsel', or 'wl-clipboard' depending on your system."
			return m, tea.Sequence(tea.Printf(clipboardErrMsg), tea.Quit)
		}
		placeholderStyle := lipgloss.NewStyle().Faint(true)
		message := "Copied to clipboard."
		if !m.latestCommandIsCode {
			message = "Copied only code to clipboard."
		}
		message = placeholderStyle.Render(message)
		return m, tea.Sequence(tea.Printf("%s", message), tea.Quit)
	}

	m.textInput.SetValue("")
	m.query = v
	m.state = Loading
	placeholderStyle := lipgloss.NewStyle().Faint(true).Width(m.maxWidth)
	message := placeholderStyle.Render(fmt.Sprintf("> %s", v))
	return m, tea.Sequence(tea.Printf("%s", message), tea.Batch(m.spinner.Tick, makeQuery(m.client, m.query)))
}

func (m model) formatResponse(response string, isCode bool) (string, error) {

	formatted, err := m.markdownRenderer.Render(response)
	if err != nil {
		// TODO: handle error
		panic(err)
	}

	formatted = strings.TrimPrefix(formatted, "\n")
	formatted = strings.TrimSuffix(formatted, "\n")

	if !isCode {
		formatted = "\n" + formatted
	}
	return formatted, nil
}

// TODO: parse the model endpoint to infer whether it's openai, other, or local.
// for local, suggest it may not be running, and how to run it
func (m model) getConnectionError(err error) string {
	styleRed := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleGreen := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleDim := lipgloss.NewStyle().Faint(true).Width(m.maxWidth).PaddingLeft(2)

	providerName := "API"
	if m.client != nil && m.client.GetProvider() != nil {
		providerName = strings.ToUpper(m.client.GetProvider().GetProviderName())
	}

	message := fmt.Sprintf("\n  %v\n\n%v\n",
		styleRed.Render(fmt.Sprintf("Error: Failed to connect to %s.", providerName)),
		styleDim.Render(err.Error()))

	if providerName == "OPENAI" && util.IsLikelyBillingError(err.Error()) {
		message = fmt.Sprintf("%v\n  %v %v\n\n  %v%v\n\n",
			message,
			styleGreen.Render("Hint:"),
			"You may need to set up OpenAI billing. You can do so here:",
			styleGreen.Render("->"),
			styleDim.Render("https://platform.openai.com/account/billing"),
		)
	} else if providerName == "GEMINI" {
		message = fmt.Sprintf("%v\n  %v %v\n\n  %v%v\n",
			message,
			styleGreen.Render("Hint:"),
			"Ensure the Generative Language API is enabled in your Google Cloud project and the API key is valid.",
			styleGreen.Render("->"),
			styleDim.Render("https://console.cloud.google.com/apis/library/generativelanguage.googleapis.com"),
		)
	}

	return message
}

func (m model) handleResponseMsg(msg responseMsg) (tea.Model, tea.Cmd) {
	m.formattedPartialResponse = ""

	if msg.err != nil {
		m.state = RecevingInput
		message := m.getConnectionError(msg.err)
		return m, tea.Sequence(tea.Printf("%s", message), textinput.Blink)
	}

	content, isOnlyCode := util.ExtractFirstCodeBlock(msg.response)
	if content != "" {
		m.latestCommandResponse = content
	}

	formatted, err := m.formatResponse(msg.response, util.StartsWithCodeBlock(msg.response))
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
	return m, tea.Sequence(tea.Printf("%s", message), textinput.Blink)
}

func (m model) handlePartialResponseMsg(msg partialResponseMsg) (tea.Model, tea.Cmd) {
	m.state = ReceivingResponse
	isCode := util.StartsWithCodeBlock(msg.content)
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

func initialModel(prompt string, client *llm.LLMClient) model {
	maxWidth := util.GetTermSafeMaxWidth()
	ti := textinput.New()
	ti.Placeholder = "Describe a shell command, or ask a question."
	ti.Focus()
	ti.Width = maxWidth

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	runWithArgs := prompt != ""

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(int(maxWidth)),
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

func printAPIKeyNotSetMessage(modelConfig ModelConfig) {
	authEnvVar := modelConfig.Auth
	provider := modelConfig.Provider
	if provider == "" {
		provider = "openai"
	}
	provider = strings.ToUpper(provider)

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)

	profileScriptName := ".zshrc or .bashrc"
	shellSyntax := fmt.Sprintf("\n```bash\nexport %s=[your key]\n```", authEnvVar)
	windowsShellSyntax := fmt.Sprintf("\n```powershell\n$env:%s = \"[your key]\"\n```", authEnvVar)

	if runtime.GOOS == "windows" {
		profileScriptName = "$profile"
		shellSyntax = windowsShellSyntax
	}

	styleRed := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	msg1 := styleRed.Render(fmt.Sprintf("%s environment variable not set.", authEnvVar))

	var instructions string
	switch provider {
	case "OPENAI":
		instructions = fmt.Sprintf(`
	1. Generate your API key at https://platform.openai.com/account/api-keys
	2. Add your credit card in the API (for the free trial)
	3. Set your key by running:
	%s
	4. (Recommended) Add that ^ line to your %s file.`, shellSyntax, profileScriptName)
	case "GEMINI":
		instructions = fmt.Sprintf(`
	1. Create an API key in your Google Cloud project: https://console.cloud.google.com/apis/credentials
	2. Ensure the Generative Language API is enabled: https://console.cloud.google.com/apis/library/generativelanguage.googleapis.com
	3. Set your key by running:
	%s
	4. (Recommended) Add that ^ line to your %s file.`, shellSyntax, profileScriptName)
	default:
		instructions = fmt.Sprintf(`
	1. Obtain your API key for %s.
	2. Set your key by running:
	%s
	3. (Recommended) Add that ^ line to your %s file.`, provider, shellSyntax, profileScriptName)
	}

	msg2, _ := r.Render(instructions)
	fmt.Printf("\n  %v%v\n", msg1, msg2)
}

func streamHandler(p *tea.Program) func(content string, err error) {
	return func(content string, err error) {
		p.Send(partialResponseMsg{content, err})
	}
}

func getModelConfig(appConfig config.AppConfig) (ModelConfig, error) {
	if len(appConfig.Models) == 0 {
		return ModelConfig{}, fmt.Errorf("no models available")
	}
	for _, model := range appConfig.Models {
		if model.ModelName == appConfig.Preferences.DefaultModel {
			return model, nil
		}
	}
	return appConfig.Models[0], nil
}

func runQProgram(prompt string) {
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		config.PrintConfigErrorMessage(err)
		os.Exit(1)
	}

	modelConfig, err := getModelConfig(appConfig)
	if err != nil {
		config.PrintConfigErrorMessage(err)
		os.Exit(1)
	}

	config.SaveAppConfig(appConfig)

	authToken := os.Getenv(modelConfig.Auth)
	if authToken == "" {
		printAPIKeyNotSetMessage(modelConfig)
		os.Exit(1)
	}

	orgID := os.Getenv(modelConfig.OrgID)
	modelConfig.OrgID = orgID

	c := llm.NewLLMClient(modelConfig)
	p := tea.NewProgram(initialModel(prompt, c))
	c.StreamCallback = streamHandler(p)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

var RootCmd = &cobra.Command{
	Use:   "q [request]",
	Short: "A command line interface for natural language queries",
	Run: func(cmd *cobra.Command, args []string) {
		prompt := strings.Join((args), " ")
		if len(args) > 0 && args[0] == "config" {
			config.RunConfigProgram(args)
			return
		}
		runQProgram(prompt)

	},
}
