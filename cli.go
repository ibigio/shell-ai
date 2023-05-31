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

const (
	codePrefix     = "[code]"
	markdownPrefix = "[markdown]"
)

type queryModel struct {
	prompt            string
	rawResponse       string
	extractedResponse string
	formattedResponse string
}

type model struct {
	apiKey    string
	textInput textinput.Model
	spinner   spinner.Model

	queries      []queryModel
	cachedOutput string

	isLoading        bool
	isReceivingInput bool
	isCopying        bool

	drawBuffer string

	runWithArgs bool
	err         error
}

type loadMsg struct{}
type responseMsg struct{ response string }
type drawOutputMsg struct{}

func drawOutputCommand() tea.Msg {
	time.Sleep(50 * time.Millisecond)
	return drawOutputMsg{}
}

func callAPI() tea.Msg {
	time.Sleep(1 * time.Second)
	return responseMsg{response: "Hello World!"}
}

func extractCodeBlock(s string) (string, bool) {
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "```") && strings.HasSuffix(trimmed, "```") {
		// There might be a language hint after the first ```
		// Example: ```go
		// We should remove this if it's present
		content := strings.TrimPrefix(trimmed, "```")
		content = strings.TrimSuffix(content, "```")
		// Find newline after the first ```
		newlinePos := strings.Index(content, "\n")
		if newlinePos != -1 {
			// Check if there's a word immediately after the first ```
			if content[0:newlinePos] == strings.TrimSpace(content[0:newlinePos]) {
				// If so, remove that part from the content
				content = content[newlinePos+1:]
			}
		}
		// Strip the final newline, if present
		if content[len(content)-1] == '\n' {
			content = content[:len(content)-1]
		}
		return content, true
	}
	return s, false
}

func ParseInput(input string) (string, bool) {
	var isCode bool
	var content string

	if strings.HasPrefix(input, codePrefix) {
		isCode = true
		content = strings.TrimSpace(input[len(codePrefix):])
	} else if strings.HasPrefix(input, markdownPrefix) {
		isCode = false
		content = strings.TrimSpace(input[len(markdownPrefix):])
	} else {
		content = input
	}

	return content, isCode
}

func makeQuery(apiKey string, queries []queryModel) (string, error) {
	// turn queries into Messages array
	messages := []Message{
		{Role: "system", Content: "You are a terminal assistant. Turn the natural language instructions into a terminal command. Always output only the command, unless the user is asking a question, in which case answer it very briefly and well."},
		{Role: "user", Content: "print my local ip address on a mac"},
		{Role: "assistant", Content: "```bash\nifconfig | grep \"inet \" | grep -v 127.0.0.1 | awk '{print $2}'\n```"},
	}
	for _, q := range queries {
		messages = append(messages, Message{Role: "user", Content: q.prompt})
		if q.rawResponse != "" {
			messages = append(messages, Message{Role: "assistant", Content: q.rawResponse})
		}
	}
	completion, err := QueryOpenAIAssistant(apiKey, messages)
	if err != nil {
		return "", err
	}
	return completion, nil
}

func callMakeQuery(apiKey string, queries []queryModel) tea.Cmd {
	return func() tea.Msg {
		completion, err := makeQuery(apiKey, queries)
		if err != nil {
			return responseMsg{response: fmt.Sprintf("Error: %v", err)}
		}
		return responseMsg{response: completion}
	}
}

func initialModel(prompt string, apiKey string) model {
	ti := textinput.New()
	ti.Placeholder = "Enter natural langauge query"
	ti.Focus()
	ti.CharLimit = 256

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	runWithArgs := prompt != ""

	model := model{
		apiKey:    apiKey,
		textInput: ti,
		spinner:   s,

		queries:      []queryModel{},
		cachedOutput: "",

		isLoading:        false,
		isReceivingInput: true,
		isCopying:        false,

		runWithArgs: false,
		err:         nil,
	}

	if runWithArgs {
		model.runWithArgs = true
		model.queries = append(model.queries, queryModel{prompt: prompt})
		model.isReceivingInput = false
		model.isLoading = true
	}
	return model
}

func (m model) updateCacheOutput() tea.Model {
	var s string
	for i, q := range m.queries {
		if q.prompt != "" && !(m.runWithArgs && i == 0) {
			s += fmt.Sprintf("> %s\n", q.prompt)
		}
		if q.formattedResponse != "" {
			s += q.formattedResponse
		}
	}
	m.cachedOutput = s
	fmt.Printf(s)
	return m
}

func (m model) Init() tea.Cmd {
	if m.runWithArgs {
		return tea.Batch(m.spinner.Tick, callMakeQuery(m.apiKey, m.queries))
	}
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case drawOutputMsg:
		if m.drawBuffer != "" {
			fmt.Printf("%s", m.drawBuffer)
		}
		m.drawBuffer = ""
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyCtrlD:
			return m, tea.Quit

		case tea.KeyEnter:
			if !m.isReceivingInput {
				return m, nil
			}
			v := m.textInput.Value()
			if v == "" {
				err := clipboard.WriteAll(m.queries[len(m.queries)-1].extractedResponse)
				if err != nil {
					fmt.Println("Failed to copy text to clipboard:", err)
					return m, tea.Quit
				}

				m.isCopying = true
				m.isLoading = false
				m.isReceivingInput = false
				placeholderStyle := lipgloss.NewStyle().Faint(true)
				m.drawBuffer = placeholderStyle.Render("Copied to clipboard.\n")
				return m, tea.Batch(drawOutputCommand, tea.Quit)
			}
			m.queries = append(m.queries, queryModel{prompt: v})
			m.textInput.SetValue("")
			m.isLoading = true
			m.isReceivingInput = false
			m.drawBuffer = fmt.Sprintf("> %s\n", v)
			return m, tea.Sequence(drawOutputCommand, tea.Batch(m.spinner.Tick, callMakeQuery(m.apiKey, m.queries)))
		}

	case responseMsg:
		m.isLoading = false

		m.textInput.Placeholder = "Follow up, ENTER to copy & quit, ESC to quit"

		// // set last query response
		// body, isCode := ParseInput(msg.response)
		body, isCode := extractCodeBlock(msg.response)

		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
		)
		formatted, err := r.Render(msg.response)
		if isCode {
			formatted = strings.TrimPrefix(formatted, "\n")
		}
		if err != nil {
			panic(err)
		}

		m.queries[len(m.queries)-1].rawResponse = msg.response
		m.queries[len(m.queries)-1].extractedResponse = body
		m.queries[len(m.queries)-1].formattedResponse = formatted
		m.isReceivingInput = true
		m.drawBuffer = formatted
		return m, tea.Batch(drawOutputCommand, textinput.Blink)

	case error:
		m.err = msg
		return m, nil

	}

	if m.isLoading {
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	if m.isReceivingInput {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if m.drawBuffer != "" {
		// fmt.Printf("Hi!\n")
		return ""
	}
	var s string

	// s += m.cachedOutput
	if m.isLoading {
		s += m.spinner.View()
	}
	if m.isReceivingInput {
		s += m.textInput.View()
	}
	// if m.isCopying {
	// 	placeholderStyle := lipgloss.NewStyle().Faint(true)
	// 	s += placeholderStyle.Render("Copied to clipboard.\n")
	// }
	return s
}

var rootCmd = &cobra.Command{
	Use:   "ai [request]",
	Short: "A command line interface for natural language queries",
	Run: func(cmd *cobra.Command, args []string) {
		// join args into a single string separated by spaces
		var prompt string
		if len(args) > 0 {
			prompt = args[0]
			for _, arg := range args[1:] {
				prompt += " " + arg
			}
		}
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("OPENAI_API_KEY environment variable not set")
			os.Exit(1)
		}

		p := tea.NewProgram(initialModel(prompt, apiKey))
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
