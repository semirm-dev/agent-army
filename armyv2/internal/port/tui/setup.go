package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/smahovkic/agent-army/armyv2/internal/core/catalog"
	"github.com/smahovkic/agent-army/armyv2/internal/core/detector"
	"github.com/smahovkic/agent-army/armyv2/internal/core/orchestrator"
	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

// Setup wizard steps
const (
	stepDestination = iota
	stepTechStack
	stepPlugins
	stepSkills
	stepConfirm
	stepInstalling
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
	selected    bool
	recommended bool
}

// SetupModel is the Bubble Tea model for the setup wizard.
type SetupModel struct {
	catalog      *catalog.Service
	orchestrator *orchestrator.Orchestrator
	manifest     *types.Manifest

	step        int
	cursor      int
	destination string // "user" or "project"
	filter      string
	filtering   bool

	// Step data
	techItems   []selectableItem
	pluginItems []selectableItem
	skillItems  []selectableItem

	// Detected tech profiles
	detectedTech []string

	// Results
	resultManifest *types.Manifest
	installResult  *orchestrator.Result
	completed      bool
	quitted        bool
	err            error
}

// NewSetupModel creates a new setup wizard model.
func NewSetupModel(cat *catalog.Service, manifest *types.Manifest, orch *orchestrator.Orchestrator) SetupModel {
	return SetupModel{
		catalog:      cat,
		orchestrator: orch,
		manifest:     manifest,
		step:         stepDestination,
		destination:  "user",
	}
}

// Completed returns true if the setup wizard completed successfully.
func (m SetupModel) Completed() bool { return m.completed }

// ResultManifest returns the manifest produced by setup.
func (m SetupModel) ResultManifest() *types.Manifest { return m.resultManifest }

type installDoneMsg struct {
	result orchestrator.Result
}

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
	case installDoneMsg:
		m.installResult = &msg.result
		m.step = stepDone
		m.completed = true
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step != stepInstalling {
				m.quitted = true
				return m, tea.Quit
			}
		case "left":
			if !m.filtering && m.step != stepDestination && m.step != stepInstalling && m.step != stepDone {
				m.step = m.previousStep()
				m.cursor = 0
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
		m.cursor = 0
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
		m.cursor = 0
		m.filter = ""
		m.filtering = false
	}
	return m, nil
}

func (m SetupModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter", "right":
		m.buildResultManifest()
		m.step = stepInstalling
		return m, m.runInstall()
	case "n", "N":
		// Go back to skills instead of quitting
		m.step = m.previousStep()
		m.cursor = 0
		return m, nil
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
	if m.step != stepInstalling && m.step != stepDone {
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
	case stepInstalling:
		s.WriteString(m.viewInstalling())
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

		desc := ""
		if item.description != "" {
			desc = " " + dimStyle.Render("— "+item.description)
		}

		s.WriteString(cursor + check + item.name + star + desc + "\n")
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

	s.WriteString(fmt.Sprintf("  %s to install: %s\n",
		selectedStyle.Render(fmt.Sprintf("%d plugin(s)", len(selPlugins))),
		strings.Join(selPlugins, ", ")))
	s.WriteString(fmt.Sprintf("  %s to install: %s\n",
		selectedStyle.Render(fmt.Sprintf("%d skill(s)", len(selSkills))),
		strings.Join(selSkills, ", ")))
	s.WriteString(fmt.Sprintf("  Destination: %s\n", m.destination))

	s.WriteString("\n  " + helpStyle.Render("Proceed? [Y/n] · ← back"))
	return s.String()
}

func (m SetupModel) viewInstalling() string {
	return "Installing... please wait.\n"
}

func (m SetupModel) viewDone() string {
	var s strings.Builder

	if m.installResult != nil {
		s.WriteString(successStyle.Render(fmt.Sprintf(
			"✓ Done: %d succeeded, %d failed\n",
			m.installResult.Succeeded, m.installResult.Failed)))

		if len(m.installResult.Errors) > 0 {
			s.WriteString("\nErrors:\n")
			for _, err := range m.installResult.Errors {
				s.WriteString(errorStyle.Render("  ✗ "+err.Error()) + "\n")
			}
		}
	}

	s.WriteString("\nPress any key to exit.")
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

func (m SetupModel) runInstall() tea.Cmd {
	return func() tea.Msg {
		result := m.orchestrator.InstallItems(
			m.resultManifest.Plugins,
			m.resultManifest.Skills,
		)
		return installDoneMsg{result: result}
	}
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
			strings.Contains(strings.ToLower(item.description), lower) {
			indices = append(indices, i)
		}
	}
	return indices
}
