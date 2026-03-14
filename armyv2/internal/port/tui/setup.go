package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/smahovkic/agent-army/armyv2/internal/core/catalog"
	"github.com/smahovkic/agent-army/armyv2/internal/core/detector"
	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

// Setup wizard steps
const (
	stepDestination = iota
	stepTechStack
	stepPlugins
	stepSkills
	stepConfirm
	stepDone
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	starStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type selectableItem struct {
	name        string
	description string
	source      string
	selected    bool
	recommended bool
}

// SetupModel is the Bubble Tea model for the setup wizard.
type SetupModel struct {
	catalog  *catalog.Service
	manifest *types.Manifest

	step         int
	cursor       int
	cursors      map[int]int // saved cursor position per step
	destination  string      // "user" or "project"
	manifestPath string      // output path for manifest
	filter       string
	filtering    bool
	editingPath  bool   // true when editing manifest path on confirm step
	pathInput    string // buffer for path editing

	// Step data
	techItems   []selectableItem
	pluginItems []selectableItem
	skillItems  []selectableItem

	// Detected tech profiles
	detectedTech []string

	// Results
	resultManifest *types.Manifest
	completed      bool
	quitted        bool
	err            error
}

// NewSetupModel creates a new setup wizard model.
func NewSetupModel(cat *catalog.Service, manifest *types.Manifest, manifestPath string) SetupModel {
	return SetupModel{
		catalog:      cat,
		manifest:     manifest,
		manifestPath: manifestPath,
		step:         stepDestination,
		cursors:      make(map[int]int),
		destination:  "user",
	}
}

// Completed returns true if the setup wizard completed successfully.
func (m SetupModel) Completed() bool { return m.completed }

// ResultManifest returns the manifest produced by setup.
func (m SetupModel) ResultManifest() *types.Manifest { return m.resultManifest }

// ManifestPath returns the (possibly modified) manifest output path.
func (m SetupModel) ManifestPath() string { return m.manifestPath }

// previousStep returns the step before the current one, accounting for
// the user vs project flow (user flow skips stepTechStack).
func (m SetupModel) previousStep() int {
	switch m.step {
	case stepTechStack:
		return stepDestination
	case stepPlugins:
		if m.destination == "project" {
			return stepTechStack
		}
		return stepDestination
	case stepSkills:
		return stepPlugins
	case stepConfirm:
		return stepSkills
	default:
		return m.step
	}
}

func (m SetupModel) Init() tea.Cmd {
	return nil
}

func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			m.quitted = true
			return m, tea.Quit
		case "q":
			if !m.filtering && !m.editingPath {
				m.quitted = true
				return m, tea.Quit
			}
		case "left":
			if !m.filtering && !m.editingPath && m.step != stepDestination && m.step != stepDone {
				m.cursors[m.step] = m.cursor
				prev := m.previousStep()
				m.step = prev
				m.cursor = m.cursors[prev]
				m.filter = ""
				m.filtering = false
				return m, nil
			}
		}

		if m.filtering {
			return m.handleFilterInput(msg)
		}

		switch m.step {
		case stepDestination:
			return m.updateDestination(msg)
		case stepTechStack:
			return m.updateMultiSelect(msg, &m.techItems, stepPlugins)
		case stepPlugins:
			return m.updateMultiSelect(msg, &m.pluginItems, stepSkills)
		case stepSkills:
			return m.updateMultiSelect(msg, &m.skillItems, stepConfirm)
		case stepConfirm:
			return m.updateConfirm(msg)
		case stepDone:
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m SetupModel) updateDestination(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 1 {
			m.cursor++
		}
	case "enter", "right":
		m.cursors[stepDestination] = m.cursor
		if m.cursor == 0 {
			m.destination = "user"
			m.initPluginItems(nil)
			m.initSkillItems(nil)
			m.step = stepPlugins
		} else {
			m.destination = "project"
			m.initTechItems()
			m.step = stepTechStack
		}
		m.cursor = m.cursors[m.step]
	}
	return m, nil
}

func (m SetupModel) updateMultiSelect(msg tea.KeyMsg, items *[]selectableItem, nextStep int) (tea.Model, tea.Cmd) {
	filtered := m.filteredIndices(*items)

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(filtered)-1 {
			m.cursor++
		}
	case " ":
		if m.cursor < len(filtered) {
			idx := filtered[m.cursor]
			(*items)[idx].selected = !(*items)[idx].selected
		}
	case "/":
		m.filtering = true
		m.filter = ""
	case "a":
		for i := range *items {
			(*items)[i].selected = true
		}
	case "n":
		for i := range *items {
			(*items)[i].selected = false
		}
	case "enter", "right":
		// Move to next step
		m.cursors[m.step] = m.cursor
		if m.step == stepTechStack {
			var selectedTech []string
			for _, item := range *items {
				if item.selected {
					selectedTech = append(selectedTech, item.name)
				}
			}
			m.initPluginItems(selectedTech)
			m.initSkillItems(selectedTech)
		}
		m.step = nextStep
		m.cursor = m.cursors[nextStep]
		m.filter = ""
		m.filtering = false
	}
	return m, nil
}

func (m SetupModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingPath {
		return m.handlePathInput(msg)
	}
	switch msg.String() {
	case "y", "Y", "enter":
		m.buildResultManifest()
		m.completed = true
		m.step = stepDone
		return m, tea.Quit
	case "n", "N":
		// Go back to skills instead of quitting
		m.cursors[m.step] = m.cursor
		prev := m.previousStep()
		m.step = prev
		m.cursor = m.cursors[prev]
		return m, nil
	case "d":
		m.editingPath = true
		m.pathInput = m.manifestPath
		return m, nil
	}
	return m, nil
}

func (m SetupModel) handlePathInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.manifestPath = m.pathInput
		m.editingPath = false
	case "esc":
		m.editingPath = false
	case "backspace":
		if len(m.pathInput) > 0 {
			m.pathInput = m.pathInput[:len(m.pathInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.pathInput += msg.String()
		}
	}
	return m, nil
}

func (m SetupModel) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.filtering = false
		m.cursor = 0
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.filter += msg.String()
			m.cursor = 0
		}
	}
	return m, nil
}

func (m SetupModel) View() string {
	if m.quitted {
		return "Setup cancelled.\n"
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render("armyv2 setup") + "\n\n")

	// Show progress indicator for navigable steps
	if m.step != stepDone {
		s.WriteString(m.viewProgress())
	}

	switch m.step {
	case stepDestination:
		s.WriteString(m.viewDestination())
	case stepTechStack:
		s.WriteString(m.viewMultiSelect("Detected tech stack (adjust as needed):", m.techItems))
	case stepPlugins:
		s.WriteString(m.viewMultiSelect("Select plugins to install:", m.pluginItems))
	case stepSkills:
		s.WriteString(m.viewMultiSelect("Select skills to install:", m.skillItems))
	case stepConfirm:
		s.WriteString(m.viewConfirm())
	case stepDone:
		s.WriteString(m.viewDone())
	}

	return s.String()
}

func (m SetupModel) viewProgress() string {
	type stepInfo struct {
		step  int
		label string
	}

	var steps []stepInfo
	if m.destination == "project" {
		steps = []stepInfo{
			{stepDestination, "Destination"},
			{stepTechStack, "Tech Stack"},
			{stepPlugins, "Plugins"},
			{stepSkills, "Skills"},
			{stepConfirm, "Confirm"},
		}
	} else {
		steps = []stepInfo{
			{stepDestination, "Destination"},
			{stepPlugins, "Plugins"},
			{stepSkills, "Skills"},
			{stepConfirm, "Confirm"},
		}
	}

	currentIdx := 0
	for i, s := range steps {
		if s.step == m.step {
			currentIdx = i
			break
		}
	}

	var dots strings.Builder
	for i := range steps {
		if i > 0 {
			dots.WriteString(" ")
		}
		if i <= currentIdx {
			dots.WriteString("●")
		} else {
			dots.WriteString("○")
		}
	}

	return dimStyle.Render(fmt.Sprintf("%s  Step %d of %d: %s",
		dots.String(), currentIdx+1, len(steps), steps[currentIdx].label)) + "\n\n"
}

func (m SetupModel) viewDestination() string {
	var s strings.Builder
	s.WriteString("Where are you setting up Claude Code?\n\n")

	options := []string{
		"User-level (global defaults for all projects)",
		"Project-level (for current project)",
	}

	for i, opt := range options {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("❯ ")
		}
		s.WriteString(cursor + opt + "\n")
	}

	s.WriteString("\n" + helpStyle.Render("↑/↓ navigate · →/enter select"))
	return s.String()
}

func (m SetupModel) viewMultiSelect(title string, items []selectableItem) string {
	var s strings.Builder
	s.WriteString(title + "\n")

	if m.filtering {
		s.WriteString(dimStyle.Render("Filter: ") + m.filter + "█\n")
	}
	s.WriteString("\n")

	filtered := m.filteredIndices(items)

	for vi, idx := range filtered {
		item := items[idx]
		cursor := "  "
		if vi == m.cursor {
			cursor = cursorStyle.Render("❯ ")
		}

		check := "  "
		if item.selected {
			check = selectedStyle.Render("✓ ")
		}

		star := ""
		if item.recommended {
			star = " " + starStyle.Render("★")
		}

		source := ""
		if item.source != "" {
			source = " " + dimStyle.Render("("+item.source+")")
		}

		desc := ""
		if item.description != "" {
			desc = " " + dimStyle.Render("— "+item.description)
		}

		s.WriteString(cursor + check + item.name + star + source + desc + "\n")
	}

	if len(filtered) < len(items) {
		s.WriteString(dimStyle.Render(fmt.Sprintf("\n  ... %d hidden by filter", len(items)-len(filtered))))
	}

	s.WriteString("\n" + helpStyle.Render("↑/↓ navigate · space toggle · a all · n none · / filter · ← back · →/enter confirm"))
	return s.String()
}

func (m SetupModel) viewConfirm() string {
	var s strings.Builder
	s.WriteString("Summary:\n\n")

	var selPlugins, selSkills []string
	for _, p := range m.pluginItems {
		if p.selected {
			selPlugins = append(selPlugins, p.name)
		}
	}
	for _, sk := range m.skillItems {
		if sk.selected {
			selSkills = append(selSkills, sk.name)
		}
	}

	s.WriteString(fmt.Sprintf("  %s selected: %s\n",
		selectedStyle.Render(fmt.Sprintf("%d plugin(s)", len(selPlugins))),
		strings.Join(selPlugins, ", ")))
	s.WriteString(fmt.Sprintf("  %s selected: %s\n",
		selectedStyle.Render(fmt.Sprintf("%d skill(s)", len(selSkills))),
		strings.Join(selSkills, ", ")))
	s.WriteString(fmt.Sprintf("  Destination: %s (%s)\n", m.destination, tildefy(m.manifestPath)))

	if m.editingPath {
		s.WriteString("\n  " + dimStyle.Render("Path: ") + m.pathInput + "█\n")
		s.WriteString("\n  " + helpStyle.Render("enter save · esc cancel"))
	} else {
		s.WriteString("\n  " + helpStyle.Render("Proceed? [Y/n] · ← back · d edit path · enter confirm"))
	}
	return s.String()
}

// tildefy replaces the home directory prefix with ~ for display.
func tildefy(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

func (m SetupModel) viewDone() string {
	var s strings.Builder
	s.WriteString(successStyle.Render("✓ Setup complete!") + "\n\n")
	s.WriteString("  Run 'armyv2 sync' to install your selections.\n")
	return s.String()
}

// --- Initialization helpers ---

func (m *SetupModel) initTechItems() {
	if len(m.techItems) > 0 {
		return // Already initialized, preserve user selections
	}

	profiles := m.catalog.AllTechProfiles()
	m.detectedTech = detector.Detect(".", profiles)

	detectedSet := make(map[string]bool)
	for _, t := range m.detectedTech {
		detectedSet[t] = true
	}

	for name := range profiles {
		m.techItems = append(m.techItems, selectableItem{
			name:        name,
			selected:    detectedSet[name],
			recommended: detectedSet[name],
		})
	}
}

func (m *SetupModel) initPluginItems(selectedTech []string) {
	profiles := m.catalog.AllTechProfiles()
	recPlugins, _ := detector.RecommendedItems(selectedTech, profiles)
	recSet := make(map[string]bool)
	for _, p := range recPlugins {
		recSet[p] = true
	}

	if len(m.pluginItems) > 0 {
		// Already populated — update recommendation flags only, preserve selections
		for i := range m.pluginItems {
			m.pluginItems[i].recommended = recSet[m.pluginItems[i].name]
		}
		return
	}

	// First population
	for _, p := range m.catalog.AllPlugins() {
		m.pluginItems = append(m.pluginItems, selectableItem{
			name:        p.Name,
			description: p.Description,
			source:      p.Marketplace,
			selected:    recSet[p.Name],
			recommended: recSet[p.Name],
		})
	}
}

func (m *SetupModel) initSkillItems(selectedTech []string) {
	profiles := m.catalog.AllTechProfiles()
	_, recSkills := detector.RecommendedItems(selectedTech, profiles)
	recSet := make(map[string]bool)
	for _, s := range recSkills {
		recSet[s] = true
	}

	if len(m.skillItems) > 0 {
		// Already populated — update recommendation flags only, preserve selections
		for i := range m.skillItems {
			m.skillItems[i].recommended = recSet[m.skillItems[i].name]
		}
		return
	}

	// First population
	for _, s := range m.catalog.AllSkills() {
		m.skillItems = append(m.skillItems, selectableItem{
			name:        s.Name,
			description: s.Description,
			source:      s.Source,
			selected:    recSet[s.Name],
			recommended: recSet[s.Name],
		})
	}
}

func (m *SetupModel) buildResultManifest() {
	result := &types.Manifest{Version: 1}

	for _, p := range m.pluginItems {
		if p.selected {
			cp, _ := m.catalog.FindPlugin(p.name)
			result.Plugins = append(result.Plugins, types.ManifestPlugin{
				Name:        cp.Name,
				Marketplace: cp.Marketplace,
				Tags:        cp.Tags,
				Destination: m.destination,
			})
		}
	}

	for _, s := range m.skillItems {
		if s.selected {
			cs, _ := m.catalog.FindSkill(s.name)
			result.Skills = append(result.Skills, types.ManifestSkill{
				Name:        cs.Name,
				Source:      cs.Source,
				Tags:        cs.Tags,
				Destination: m.destination,
			})
		}
	}

	m.resultManifest = result
}

func (m SetupModel) filteredIndices(items []selectableItem) []int {
	if m.filter == "" {
		indices := make([]int, len(items))
		for i := range items {
			indices[i] = i
		}
		return indices
	}

	var indices []int
	lower := strings.ToLower(m.filter)
	for i, item := range items {
		if strings.Contains(strings.ToLower(item.name), lower) ||
			strings.Contains(strings.ToLower(item.source), lower) ||
			strings.Contains(strings.ToLower(item.description), lower) {
			indices = append(indices, i)
		}
	}
	return indices
}
