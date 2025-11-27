package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path"
	"runtime/debug"
	"slices"
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

func (d data) format() string {
	code := d.code
	parsed, err := gojq.Parse(code)
	if err != nil {
		return ""
	}
	code = parsed.String()
	if code == "" {
		return ""
	}
	code = strings.ReplaceAll(code, `'`, `'\''`)
	code = "'" + code + "'"

	if d.raw {
		code = "-r " + code
	}

	return "jqx " + code
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

	code := strings.TrimSpace(d.code)
	if code == "" {
		code = "."
	}
	prog.Args = append(prog.Args, code)
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
	saveMsg func(string) error
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

	rt.textarea.SetHeight(4)
	rt.textarea.Placeholder = "jq..."
	rt.textarea.Focus()
	rt.textarea.Cursor.SetMode(cursor.CursorStatic)

	return rt
}
func (m model) Init() tea.Cmd {
	return textarea.Blink
}

const Margin = 4

func tlink(name, url string) string {
	return "\033]8;;" + url + "\033\\" + name + "\033]8;;\033\\"
}

func doSave(contents string) error {
	fname := fmt.Sprintf("jq-%d.txt", uptime())

	outFile := cmp.Or(os.Getenv("XDG_RUNTIME_DIR"), "/tmp")
	outFile = path.Join(outFile, fname)

	tPrintln("saved to " + tlink(outFile, "file://"+outFile))
	return os.WriteFile(outFile, []byte(contents), 0666)
}
func doExport(contents string) error {
	var fileObj map[string]string
	err := json.Unmarshal([]byte(contents), &fileObj)
	if err != nil {
		return err
	}

	keys := slices.Sorted(maps.Keys(fileObj))
	for _, k := range keys {
		err := os.WriteFile(k, []byte(fileObj[k]), 0666)
		if err != nil {
			return err
		}
	}
	tPrintln("exported to " + fmt.Sprint(keys))
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case updateMsg:
		if msg.m == nil {
			msg.m = m
		}
		return msg.m, msg.c
	case tea.WindowSizeMsg:
		width := msg.Width - Margin*2
		m.textarea.SetWidth(width)
		m.viewport.Width = width
		m.viewport.Height = msg.Height - 15
		m.viewportContent()
		return m, nil
	case tea.MouseMsg:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	case saveMsg:
		m.err = msg(m.vcontent)
		return m, nil
	case func() (string, error):
		text, err := msg()
		m.err = err
		if err == nil {
			m.vcontent = text
			m.viewportContent()
		}
		return m, nil

	// events that change the query
	case tabMsg:
		m.d.compact = !m.d.compact
		queryChanged(m.d)
		return m, nil
	default:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		m.d.code = m.textarea.Value()
		queryChanged(m.d)
		return m, cmd
	}
}

func (m *model) viewportContent() {
	m.viewport.SetContent(truncLines(m.vcontent, m.viewport.Width))
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

	hr := subtleStyle.Render(strings.Repeat("─", 8))

	mainView := lipgloss.JoinVertical(lipgloss.Center,
		hr,
		viewport,
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
			return saveMsg(doSave)
		case tea.KeyCtrlE:
			return saveMsg(doExport)
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

var (
	queryChanged func(data)
	tPrintln     func(...any)
)

func main() {
	var d data

	if !isTerminal(os.Stdin) {
		jqInput = string(must(io.ReadAll(os.Stdin)))
	}
	jqFiles = os.Args[1:]
	if _, err := d.query(); err != nil {
		d.raw = true
	}

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

	tPrintln = func(a ...any) { go p.Println(a...) }

	{
		set, wait := watcher[data]()
		queryChanged = set
		go queryThread(p.Send, wait)
	}

	m := must(p.Run())

	if !isTerminal(os.Stdout) {
		m := m.(emptyModel).Model.(model)
		os.Stdout.Write([]byte(m.vcontent))
	}
}

func queryThread(send func(tea.Msg), wait func(*data)) {
	var d data
	var logged = map[string]bool{"": true}

	for {
		rt, err := d.query()
		if err == nil {
			log := d.format()
			if !logged[log] {
				tPrintln(log)
			}
			logged[log] = true
		}

		send(func() (string, error) {
			return rt, err
		})

		wait(&d)
	}
}
