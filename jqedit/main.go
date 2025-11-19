package main

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itchyny/gojq"
	"github.com/myaaaaaaaaa/go-jqx"
)

const sampleJSON = `{"a":5,"b":"c","d":["e",true,-11.5,null,"f"],"g":{"h":"i","j":"k"},"l":null}`

var jqInput string
var jqFiles []string

type data struct {
	code string

	compact bool
	raw     bool
}

var logged = map[string]bool{}

func (d data) logScript() tea.Cmd {
	code := d.code
	code = must(gojq.Parse(code)).String()
	code = strings.ReplaceAll(code, `'`, `'\''`)
	code = "'" + code + "'"

	if d.raw {
		code = "-r " + code
	}

	if logged[code] {
		return nil
	}
	logged[code] = true
	return tea.Printf("    %s", code)
}
func (d data) query() (string, error) {
	var output bytes.Buffer

	filesVarInput := false
	inputString := jqInput
	if inputString == "" {
		inputString = sampleJSON
		if len(jqFiles) > 0 {
			filesVarInput = true
		}
	}

	prog := jqx.Program{
		Stdin:   bytes.NewBufferString(inputString),
		Println: func(s string) { fmt.Fprintln(&output, s) },
		Open:    func(f string) (fs.File, error) { return os.Open(f) },

		StdinIsTerminal:  filesVarInput,
		StdoutIsTerminal: !d.compact,
	}

	if d.raw {
		prog.Args = append(prog.Args, "-r")
	}
	prog.Args = append(prog.Args, " "+d.code)
	prog.Args = append(prog.Args, jqFiles...)

	var err error
	func() {
		defer func() {
			r := recover()
			if r != nil {
				err = fmt.Errorf("query panic: %v", r)
			}
		}()
		_, err = prog.Main()
	}()
	rt := output.String()

	if err == nil {
		switch rt {
		case "null\n":
			err = fmt.Errorf("error: query returned null")
		case "":
			err = fmt.Errorf("error: query returned empty")
		}
	}

	return rt, err
}

type (
	saveMsg struct{}
	tabMsg  struct{}
)
type updateMsg struct {
	m tea.Model
	c tea.Cmd
}

func init() {
	textarea.DefaultKeyMap.WordForward = key.NewBinding(key.WithKeys("ctrl+right"))
	textarea.DefaultKeyMap.WordBackward = key.NewBinding(key.WithKeys("ctrl+left"))
}

type model struct {
	textarea textarea.Model
	viewport viewport.Model
	vcontent string

	d data

	err error
}

func newModel(d data) model {
	rt := model{
		textarea: textarea.New(),
		viewport: viewport.New(10, 10),
		d:        d,
	}

	rt.textarea.SetHeight(3)
	rt.textarea.Placeholder = "jq..."
	rt.textarea.Focus()
	rt.textarea.Cursor.SetMode(cursor.CursorStatic)

	return rt
}
func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
	)
}

const Margin = 8

func tlink(name, url string) string {
	return "\033]8;;" + url + "\033\\" + name + "\033]8;;\033\\"
}

var doSave = func(contents string) (msg string, err error) {
	fname := fmt.Sprintf("jq-%d.txt", uptime())

	outFile := cmp.Or(os.Getenv("XDG_RUNTIME_DIR"), "/tmp")
	outFile = path.Join(outFile, fname)

	msg = "saved to " + tlink(outFile, "file://"+outFile)
	err = os.WriteFile(outFile, []byte(contents), 0666)
	return
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	oldData := m.d

	switch msg := msg.(type) {
	case updateMsg:
		if msg.m == nil {
			msg.m = m
		}
		return msg.m, msg.c
	case tea.WindowSizeMsg:
		width := msg.Width - Margin*2
		m.textarea.SetWidth(width * 3 / 4)
		m.viewport.Width = width
		m.viewport.Height = msg.Height / 2
		return m, nil
	case tea.MouseMsg:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case saveMsg:
		var result string
		result, m.err = doSave(m.vcontent)
		if m.err == nil {
			return m, tea.Println("    " + result)
		}
		return m, nil
	case tabMsg:
		m.d.compact = !m.d.compact
	}

	var cmd tea.Cmd
	var text string
	m.textarea, cmd = m.textarea.Update(msg)

	m.d.code = strings.TrimSpace(m.textarea.Value())
	if m.d.code == "" {
		m.d.code = "."
	}

	if m.d == oldData {
		goto abortUpdate
	}

	text, m.err = m.d.query()
	if m.err != nil {
		goto abortUpdate
	}

	cmd = tea.Batch(cmd, m.d.logScript())
	m.viewport.SetContent(text)
	m.vcontent = text

abortUpdate:
	return m, cmd
}

var (
	errorStyle = lipgloss.Style{}.
			Bold(true).
			Foreground(lipgloss.Color("#880000"))
	subtleStyle = lipgloss.Style{}.
			Foreground(lipgloss.Color("#cccccc"))
)

func (m model) View() string {
	viewport := m.viewport.View()

	err := ""
	if m.err != nil {
		err = m.err.Error()
	}

	hr := subtleStyle.Render("────")
	vr := subtleStyle.Render("│\n│")

	mainView := lipgloss.JoinVertical(lipgloss.Center,
		hr,
		lipgloss.JoinHorizontal(lipgloss.Center,
			vr,
			viewport,
			vr,
		),
		hr,

		m.textarea.View(),
	)

	return lipgloss.JoinHorizontal(lipgloss.Center, strings.Repeat(" ", Margin-1), mainView) +
		"\n" + errorStyle.Render(err)
}

type emptyModel struct{ tea.Model }

func (emptyModel) View() string { return "" }

////////////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////////////

func uptime() (rt int) {
	rt = int(time.Now().Unix())

	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return
	}

	data, _, _ = bytes.Cut(data, []byte(" "))
	f, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return
	}

	return int(f)
}

func isTerminal(f fs.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&fs.ModeCharDevice != 0
}

func msgFilter(m tea.Model, msg tea.Msg) tea.Msg {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC, tea.KeyCtrlQ:
			return updateMsg{
				emptyModel{m},
				tea.Quit,
			}
		case tea.KeyCtrlS:
			return saveMsg{}
		case tea.KeyTab:
			return tabMsg{}
		}
	}
	return msg
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func main() {
	d := data{
		code: ".",
	}

	if !isTerminal(os.Stdin) {
		jqInput = string(must(io.ReadAll(os.Stdin)))
	}
	jqFiles = os.Args[1:]
	if _, err := d.query(); err != nil {
		d.raw = true
	}

	d.code = "#placeholder"

	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stderr))

	buildInfo, _ := debug.ReadBuildInfo()
	if buildInfo != nil {
		s := "jqedit " + buildInfo.Main.Version + " built with " + buildInfo.GoVersion
		s = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#888888")).
			Render(s)
		fmt.Fprintln(os.Stderr, s)
	}

	p := tea.NewProgram(newModel(d),
		tea.WithFilter(msgFilter),
		tea.WithMouseCellMotion(),
		tea.WithOutput(os.Stderr),
	)

	m := must(p.Run())

	if !isTerminal(os.Stdout) {
		m := m.(emptyModel).Model.(model)
		os.Stdout.Write([]byte(m.vcontent))
	}
}
