package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pupload/pupload/internal/models"

	filepicker "github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	rw "github.com/mattn/go-runewidth"
)

// Edge is an input with a presigned S3 URL.
type Edge struct {
	Name     string
	InputURL string // S3/presigned URL to PUT to
	FilePath string // selected local file path
	Status   string // PENDING, UPLOADING, DONE, ERROR
}

// Node represents a worker node.
type Node struct {
	Name     string
	Pipeline string
	Status   string // e.g. "IDLE", "RUNNING"
}

type section int

const (
	sectionInputs section = iota
	sectionNodes
)

type view int

const (
	viewDashboard view = iota
	viewFilePicker
)

// Messages

type uploadResultMsg struct {
	Index int
	Err   error
}

type flowRunUpdateMsg struct {
	FlowRun models.FlowRun
	Err     error
}

type pollTickMsg struct{}

// Model

type model struct {
	Inputs []Edge
	Nodes  []Node

	flowRun models.FlowRun

	focusSection  section
	selectedInput int
	selectedNode  int

	activeView      view
	filePicker      filepicker.Model
	fileTargetIndex int // which input we're uploading for

	width  int
	height int
}

// ----- Initialisation -----

func initialModel(flowrun models.FlowRun) model {
	inputs := make([]Edge, 0, len(flowrun.WaitingURLs))
	for _, input := range flowrun.WaitingURLs {
		edge := Edge{
			Name:     input.Artifact.EdgeName,
			InputURL: input.PutURL,
			Status:   "PENDING",
		}
		inputs = append(inputs, edge)
	}

	nodes := buildNodesFromFlowRun(flowrun)

	fp := filepicker.New()
	if wd, err := os.Getwd(); err == nil {
		fp.CurrentDirectory = wd
	}
	fp.SetHeight(10) // overridden by WindowSizeMsg later

	return model{
		Inputs: inputs,
		Nodes:  nodes,

		flowRun:       flowrun,
		focusSection:  sectionInputs,
		selectedInput: 0,
		selectedNode:  0,

		activeView:      viewDashboard,
		filePicker:      fp,
		fileTargetIndex: -1,
	}
}

// rebuild Nodes from FlowRun.NodeState
func buildNodesFromFlowRun(fr models.FlowRun) []Node {
	nodes := make([]Node, 0, len(fr.NodeState))
	for name, n := range fr.NodeState {
		nodes = append(nodes, Node{
			Name:     name,
			Pipeline: "", // fill if you have pipeline info
			Status:   string(n.Status),
		})
	}
	return nodes
}

func (m model) Init() tea.Cmd {
	// start polling if we have an ID (alt screen handles clearing)
	if m.flowRun.ID == "" {
		return nil
	}
	return pollFlowRunCmd(m.flowRun.ID)
}

// ----- Update -----

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global messages: quit, upload results, polling, window size
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case uploadResultMsg:
		if msg.Index >= 0 && msg.Index < len(m.Inputs) {
			if msg.Err != nil {
				m.Inputs[msg.Index].Status = "ERROR"
			} else {
				m.Inputs[msg.Index].Status = "DONE"
			}
		}
		return m, nil

	case flowRunUpdateMsg:
		// got latest FlowRun from server
		if msg.Err == nil {
			m.flowRun = msg.FlowRun
			m.Nodes = buildNodesFromFlowRun(m.flowRun)

			// clamp selectedNode
			if m.selectedNode >= len(m.Nodes) {
				if len(m.Nodes) == 0 {
					m.selectedNode = 0
				} else {
					m.selectedNode = len(m.Nodes) - 1
				}
			}
		}
		// schedule next tick regardless of error
		return m, pollTickCmd()

	case pollTickMsg:
		// time to poll again
		if m.flowRun.ID == "" || m.flowRun.Status == models.FLOWRUN_COMPLETE {
			return m, pollTickCmd()
		}
		return m, pollFlowRunCmd(m.flowRun.ID)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.height > 4 {
			m.filePicker.SetHeight(m.height - 4)
		}
		return m, nil
	}

	// File picker view
	if m.activeView == viewFilePicker {
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)

		if ok, path := m.filePicker.DidSelectFile(msg); ok {
			if m.fileTargetIndex >= 0 && m.fileTargetIndex < len(m.Inputs) {
				edge := &m.Inputs[m.fileTargetIndex]
				edge.FilePath = path
				edge.Status = "UPLOADING"

				uploadCmd := uploadFileCmd(m.fileTargetIndex, edge.InputURL, path)

				m.activeView = viewDashboard
				m.fileTargetIndex = -1
				return m, tea.Batch(cmd, uploadCmd)
			}
			m.activeView = viewDashboard
			m.fileTargetIndex = -1
			return m, cmd
		}

		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "esc":
				m.activeView = viewDashboard
				m.fileTargetIndex = -1
				return m, cmd
			}
		}

		return m, cmd
	}

	// Dashboard view logic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "up", "k":
			switch m.focusSection {
			case sectionInputs:
				if len(m.Inputs) == 0 {
					// nothing here, maybe jump to nodes if any
					if len(m.Nodes) > 0 {
						m.focusSection = sectionNodes
						if m.selectedNode >= len(m.Nodes) {
							m.selectedNode = len(m.Nodes) - 1
						}
					}
				} else if m.selectedInput > 0 {
					m.selectedInput--
				} else if m.selectedInput == 0 && len(m.Nodes) > 0 {
					// move to last node
					m.focusSection = sectionNodes
					m.selectedNode = len(m.Nodes) - 1
				}
			case sectionNodes:
				if len(m.Nodes) == 0 {
					if len(m.Inputs) > 0 {
						m.focusSection = sectionInputs
						if m.selectedInput >= len(m.Inputs) {
							m.selectedInput = len(m.Inputs) - 1
						}
					}
				} else if m.selectedNode > 0 {
					m.selectedNode--
				} else if m.selectedNode == 0 && len(m.Inputs) > 0 {
					// move to last input
					m.focusSection = sectionInputs
					m.selectedInput = len(m.Inputs) - 1
				}
			}

		case "down", "j":
			switch m.focusSection {
			case sectionInputs:
				if len(m.Inputs) == 0 {
					// nothing here, maybe jump to nodes if any
					if len(m.Nodes) > 0 {
						m.focusSection = sectionNodes
						m.selectedNode = 0
					}
				} else if m.selectedInput < len(m.Inputs)-1 {
					m.selectedInput++
				} else if m.selectedInput == len(m.Inputs)-1 && len(m.Nodes) > 0 {
					// move to first node
					m.focusSection = sectionNodes
					m.selectedNode = 0
				}
			case sectionNodes:
				if len(m.Nodes) == 0 {
					if len(m.Inputs) > 0 {
						m.focusSection = sectionInputs
						m.selectedInput = 0
					}
				} else if m.selectedNode < len(m.Nodes)-1 {
					m.selectedNode++
				} else if m.selectedNode == len(m.Nodes)-1 && len(m.Inputs) > 0 {
					// move to first input
					m.focusSection = sectionInputs
					m.selectedInput = 0
				}
			}

		case "left", "h":
			m.focusSection = sectionInputs
			if len(m.Inputs) > 0 && m.selectedInput >= len(m.Inputs) {
				m.selectedInput = len(m.Inputs) - 1
			}

		case "right", "l":
			m.focusSection = sectionNodes
			if len(m.Nodes) > 0 && m.selectedNode >= len(m.Nodes) {
				m.selectedNode = len(m.Nodes) - 1
			}

		case "enter":
			if m.focusSection == sectionInputs && len(m.Inputs) > 0 {
				m.activeView = viewFilePicker
				m.fileTargetIndex = m.selectedInput

				if m.height > 4 {
					m.filePicker.SetHeight(m.height - 4)
				}

				return m, m.filePicker.Init()
			}
		}
	}

	return m, nil
}

// ----- Commands -----

const pollInterval = 2 * time.Second

func pollTickCmd() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg {
		return pollTickMsg{}
	})
}

func pollFlowRunCmd(id string) tea.Cmd {
	return func() tea.Msg {
		fr, err := fetchFlowRun(id)
		return flowRunUpdateMsg{FlowRun: fr, Err: err}
	}
}

// uploadFileCmd performs the PUT upload in a Cmd.
func uploadFileCmd(index int, url, path string) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(path)
		if err != nil {
			return uploadResultMsg{Index: index, Err: err}
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return uploadResultMsg{Index: index, Err: err}
		}

		req, err := http.NewRequest("PUT", url, f)
		if err != nil {
			return uploadResultMsg{Index: index, Err: err}
		}
		req.ContentLength = info.Size()

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return uploadResultMsg{Index: index, Err: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return uploadResultMsg{
				Index: index,
				Err:   fmt.Errorf("upload failed: %s", resp.Status),
			}
		}

		return uploadResultMsg{Index: index, Err: nil}
	}
}

// ----- Fetching latest FlowRun from server -----

// TODO: change this to your real API base.
const flowRunAPIBase = "http://localhost:1234/api/v1/flow/status"

func fetchFlowRun(id string) (models.FlowRun, error) {
	url := fmt.Sprintf("%s/%s", flowRunAPIBase, id)

	resp, err := http.Get(url)
	if err != nil {
		return models.FlowRun{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return models.FlowRun{}, fmt.Errorf("fetch flowrun: %s", resp.Status)
	}

	var fr models.FlowRun
	if err := json.NewDecoder(resp.Body).Decode(&fr); err != nil {
		return models.FlowRun{}, err
	}
	return fr, nil
}

// ----- Layout & color helpers -----

const (
	defaultBoxWidth    = 90 // fallback before first WindowSizeMsg
	leftPadding        = 2
	rightPadding       = 2
	maxURLDisplayWidth = 40

	colorReset      = "\x1b[0m"
	colorSelectedBg = "\x1b[48;5;236m" // subtle background for selected rows
	colorBlue       = "\x1b[34m"
	colorCyan       = "\x1b[36m"
	colorGreen      = "\x1b[32m"
	colorYellow     = "\x1b[33m"
	colorRed        = "\x1b[31m"
	colorMagenta    = "\x1b[35m"
	colorBorderDim  = "\x1b[38;5;240m"   // dim grey for borders
	colorURLDimGrey = "\x1b[2;38;5;245m" // dim grey URL text

	// foreground-only reset – keeps background (for selected rows)
	colorFgReset = "\x1b[39m"
)

func (m model) boxWidth() int {
	if m.width <= 0 {
		return defaultBoxWidth
	}
	if m.width < 20 {
		return m.width
	}
	return m.width
}

func contentWidth(boxWidth int) int {
	return boxWidth - 2 - leftPadding - rightPadding
}

// padRightVisual pads/truncates based on visual (cell) width.
func padRightVisual(s string, width int) string {
	w := rw.StringWidth(s)
	if w >= width {
		return rw.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-w)
}

// line builds: "│" + left pad + padded content + right pad + "│".
func line(content string, boxWidth int) string {
	cw := contentWidth(boxWidth)
	inner := padRightVisual(content, cw)
	return "│" +
		strings.Repeat(" ", leftPadding) +
		inner +
		strings.Repeat(" ", rightPadding) +
		"│"
}

// styleBorderLine dims the left and right border characters of a content line.
func styleBorderLine(row string) string {
	runes := []rune(row)
	if len(runes) < 2 {
		return row
	}
	left := string(runes[0])
	middle := string(runes[1 : len(runes)-1])
	right := string(runes[len(runes)-1])

	return colorBorderDim + left + colorFgReset +
		middle +
		colorBorderDim + right + colorFgReset
}

// header builds a full-width header line like:
// ┌  Run ────────────────────────────────┐
func header(title string, left, right rune, boxWidth int) string {
	totalInner := boxWidth - 2              // between borders
	headerWidth := totalInner - leftPadding // excluding left padding
	label := " " + title + " "
	if rw.StringWidth(label) > headerWidth {
		label = rw.Truncate(label, headerWidth, "")
	}
	labelWidth := rw.StringWidth(label)
	dashes := headerWidth - labelWidth
	if dashes < 0 {
		dashes = 0
	}

	return string(left) +
		strings.Repeat(" ", leftPadding) +
		label +
		strings.Repeat("─", dashes) +
		string(right)
}

// highlightIfSelected wraps the whole row in background color if selected,
// and always appends a reset so color doesn't bleed.
func highlightIfSelected(row string, selected bool) string {
	if selected {
		return colorSelectedBg + row + colorReset
	}
	return row + colorReset
}

// replace the last occurrence of needle with replacement
func replaceLast(haystack, needle, replacement string) string {
	if needle == "" {
		return haystack
	}
	idx := strings.LastIndex(haystack, needle)
	if idx == -1 {
		return haystack
	}
	return haystack[:idx] + replacement + haystack[idx+len(needle):]
}

// replace the first occurrence of needle with replacement
func replaceFirst(haystack, needle, replacement string) string {
	if needle == "" {
		return haystack
	}
	idx := strings.Index(haystack, needle)
	if idx == -1 {
		return haystack
	}
	return haystack[:idx] + replacement + haystack[idx+len(needle):]
}

func colorFlowRunStatusText(s models.FlowRunStatus) string {
	switch s {
	case models.FLOWRUN_WAITING:
		return colorYellow + string(s) + colorFgReset
	case models.FLOWRUN_RUNNING:
		return colorGreen + string(s) + colorFgReset
	case models.FLOWRUN_COMPLETE:
		return colorGreen + string(s) + colorFgReset
	case models.FLOWRUN_ERROR:
		return colorRed + string(s) + colorFgReset
	case models.FLOWRUN_STOPPED:
		return colorMagenta + string(s) + colorFgReset
	default:
		return string(s)
	}
}

// Node.Status is a string (e.g. "IDLE", "RUNNING", ...).
func colorNodeStatusText(s string) string {
	switch s {
	case string(models.NODERUN_IDLE):
		return colorBlue + s + colorFgReset
	case string(models.NODERUN_READY):
		return colorCyan + s + colorFgReset
	case string(models.NODERUN_RUNNING):
		return colorGreen + s + colorFgReset
	case string(models.NODERUN_COMPLETE):
		return colorGreen + s + colorFgReset
	case string(models.NODERUN_ERROR):
		return colorRed + s + colorFgReset
	default:
		return s
	}
}

// Edge.Status: "PENDING", "UPLOADING", "DONE", "ERROR"
func colorEdgeStatusText(s string) string {
	switch s {
	case "PENDING":
		return colorYellow + s + colorFgReset
	case "UPLOADING":
		return colorCyan + s + colorFgReset
	case "DONE":
		return colorGreen + s + colorFgReset
	case "ERROR":
		return colorRed + s + colorFgReset
	default:
		return s
	}
}

func colorHeaderTitle(title string) string {
	return colorCyan + title
}

// Shorten URL for display with ellipsis.
func shortenURL(s string, max int) string {
	if max <= 0 {
		return ""
	}
	w := rw.StringWidth(s)
	if w <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	truncated := rw.Truncate(s, max-1, "")
	return truncated + "…"
}

func colorURLText(s string) string {
	return colorURLDimGrey + s + colorFgReset
}

// styleSectionHeader takes a raw header line (from header()) and a title
// and returns it with the whole border dimmed and the title highlighted.
func styleSectionHeader(raw, title string) string {
	idx := strings.Index(raw, title)
	if idx == -1 {
		return colorBorderDim + raw + colorReset
	}
	left := raw[:idx]
	right := raw[idx+len(title):]
	return colorBorderDim + left + colorHeaderTitle(title) + colorBorderDim + right + colorReset
}

// ----- View -----

func (m model) View() string {
	// File picker view
	if m.activeView == viewFilePicker {
		var b strings.Builder

		title := "Select file"
		if m.fileTargetIndex >= 0 && m.fileTargetIndex < len(m.Inputs) {
			title = fmt.Sprintf("Select file for %s", m.Inputs[m.fileTargetIndex].Name)
		}

		b.WriteString(title + "\n\n")
		b.WriteString(m.filePicker.View())
		b.WriteString("\n\n[enter] select  [esc] cancel  [q] quit\n")
		return b.String()
	}

	// Dashboard view
	var b strings.Builder
	boxWidth := m.boxWidth()
	cw := contentWidth(boxWidth)

	// ----- RUN SUMMARY SECTION -----

	runHeaderRaw := header("Run", '┌', '┐', boxWidth)
	b.WriteString(styleSectionHeader(runHeaderRaw, "Run"))
	b.WriteString("\n")

	// spacer
	row := styleBorderLine(line("", boxWidth))
	row = highlightIfSelected(row, false)
	b.WriteString(row)
	b.WriteString("\n")

	// Row 1: ID + Flow status
	{
		left := fmt.Sprintf("  ID: %s", m.flowRun.ID)
		statusPlain := string(m.flowRun.Status)
		right := statusPlain

		leftW := rw.StringWidth(left)
		rightW := rw.StringWidth(right)
		space := cw - leftW - rightW
		if space < 1 {
			space = 1
			maxLeft := cw - rightW - space
			if maxLeft < 0 {
				maxLeft = 0
			}
			left = rw.Truncate(left, maxLeft, "")
			leftW = rw.StringWidth(left)
			space = cw - leftW - rightW
			if space < 1 {
				space = 1
			}
		}

		content := left + strings.Repeat(" ", space) + right
		row := line(content, boxWidth)
		row = styleBorderLine(row)
		row = replaceLast(row, statusPlain, colorFlowRunStatusText(m.flowRun.Status))
		row = highlightIfSelected(row, false)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Row 2: Inputs summary
	{
		total := len(m.Inputs)
		var pending, uploading, done, failed int
		for _, e := range m.Inputs {
			switch e.Status {
			case "PENDING":
				pending++
			case "UPLOADING":
				uploading++
			case "DONE":
				done++
			case "ERROR":
				failed++
			}
		}
		text := fmt.Sprintf("  Inputs: %d (pending %d, uploading %d, done %d, error %d)", total, pending, uploading, done, failed)
		row := styleBorderLine(line(text, boxWidth))
		row = highlightIfSelected(row, false)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Row 3: Nodes summary
	{
		total := len(m.Nodes)
		var idle, ready, running, complete, failed int
		for _, n := range m.Nodes {
			switch n.Status {
			case string(models.NODERUN_IDLE):
				idle++
			case string(models.NODERUN_READY):
				ready++
			case string(models.NODERUN_RUNNING):
				running++
			case string(models.NODERUN_COMPLETE):
				complete++
			case string(models.NODERUN_ERROR):
				failed++
			}
		}
		text := fmt.Sprintf("  Nodes: %d (idle %d, ready %d, running %d, done %d, error %d)", total, idle, ready, running, complete, failed)
		row := styleBorderLine(line(text, boxWidth))
		row = highlightIfSelected(row, false)
		b.WriteString(row)
		b.WriteString("\n")
	}

	// spacer under Run
	row = styleBorderLine(line("", boxWidth))
	row = highlightIfSelected(row, false)
	b.WriteString(row)
	b.WriteString("\n")

	// ----- INPUTS SECTION -----

	inputsHeaderRaw := header("Inputs", '├', '┤', boxWidth)
	b.WriteString(styleSectionHeader(inputsHeaderRaw, "Inputs"))
	b.WriteString("\n")

	row = styleBorderLine(line("", boxWidth))
	row = highlightIfSelected(row, false)
	b.WriteString(row)
	b.WriteString("\n")

	for i, e := range m.Inputs {
		selected := m.focusSection == sectionInputs && i == m.selectedInput

		marker := "•"
		if selected {
			marker = "◉"
		}

		// shorten URL for display and track the plain parenthetical
		urlDisplay := shortenURL(e.InputURL, maxURLDisplayWidth)
		parenPlain := fmt.Sprintf("(%s)", urlDisplay)

		left := fmt.Sprintf("  %s %s %s", marker, e.Name, parenPlain)
		statusPlain := e.Status
		right := statusPlain

		leftW := rw.StringWidth(left)
		rightW := rw.StringWidth(right)
		space := cw - leftW - rightW
		if space < 1 {
			space = 1
			maxLeft := cw - rightW - space
			if maxLeft < 0 {
				maxLeft = 0
			}
			left = rw.Truncate(left, maxLeft, "")
			leftW = rw.StringWidth(left)
			space = cw - leftW - rightW
			if space < 1 {
				space = 1
			}
		}

		content := left + strings.Repeat(" ", space) + right
		row := line(content, boxWidth)
		row = styleBorderLine(row)

		// color status
		row = replaceLast(row, statusPlain, colorEdgeStatusText(statusPlain))
		// color URL parenthetical as dim/grey
		row = replaceLast(row, parenPlain, colorURLText(parenPlain))

		row = highlightIfSelected(row, selected)
		b.WriteString(row)
		b.WriteString("\n")
	}

	row = styleBorderLine(line("", boxWidth))
	row = highlightIfSelected(row, false)
	b.WriteString(row)
	b.WriteString("\n")

	// ----- NODES SECTION -----

	nodesHeaderRaw := header("Nodes", '├', '┤', boxWidth)
	b.WriteString(styleSectionHeader(nodesHeaderRaw, "Nodes"))
	b.WriteString("\n")

	row = styleBorderLine(line("", boxWidth))
	row = highlightIfSelected(row, false)
	b.WriteString(row)
	b.WriteString("\n")

	for i, n := range m.Nodes {
		selected := m.focusSection == sectionNodes && i == m.selectedNode

		arrow := "▶"
		if selected {
			arrow = "◉"
		}

		left := fmt.Sprintf("  %s %s (%s)", arrow, n.Name, n.Pipeline)
		statusPlain := n.Status
		right := statusPlain

		leftW := rw.StringWidth(left)
		rightW := rw.StringWidth(right)
		space := cw - leftW - rightW
		if space < 1 {
			space = 1
			maxLeft := cw - rightW - space
			if maxLeft < 0 {
				maxLeft = 0
			}
			left = rw.Truncate(left, maxLeft, "")
			leftW = rw.StringWidth(left)
			space = cw - leftW - rightW
			if space < 1 {
				space = 1
			}
		}

		content := left + strings.Repeat(" ", space) + right
		row := line(content, boxWidth)
		row = styleBorderLine(row)

		row = replaceLast(row, statusPlain, colorNodeStatusText(statusPlain))
		row = highlightIfSelected(row, selected)
		b.WriteString(row)
		b.WriteString("\n")
	}

	row = styleBorderLine(line("", boxWidth))
	row = highlightIfSelected(row, false)
	b.WriteString(row)
	b.WriteString("\n")

	// bottom border dimmed
	bottom := "└" + strings.Repeat("─", boxWidth-2) + "┘"
	b.WriteString(colorBorderDim + bottom + colorReset)

	return b.String()
}

// Entry point for testing

func TestFlowUI(fr models.FlowRun) {
	// Redirect stdin/stdout/stderr to prevent background logs from interfering
	p := tea.NewProgram(
		initialModel(fr),
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
