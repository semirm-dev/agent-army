package plugindoc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/semir/agent-army/internal/frontmatter"
	"github.com/semir/agent-army/internal/termcolor"
)

// --- JSON types ---

type installedPluginsFile struct {
	Plugins map[string][]pluginInstance `json:"plugins"`
}

type pluginInstance struct {
	InstallPath string `json:"installPath"`
	Version     string `json:"version"`
}

type pluginMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repository  string `json:"repository"`
	Version     string `json:"version"`
}

type skillLockFile struct {
	Skills map[string]skillEntry `json:"skills"`
}

type skillEntry struct {
	Source    string `json:"source"`
	SourceURL string `json:"sourceUrl"`
}

type marketplaceEntry struct {
	Source struct {
		Repo string `json:"repo"`
	} `json:"source"`
}

type mcpServerEntry struct {
	URL     string   `json:"url"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// --- Loading functions (return zero values for missing files) ---

func loadInstalledPlugins(home string) installedPluginsFile {
	var result installedPluginsFile
	data, err := os.ReadFile(filepath.Join(home, ".claude", "plugins", "installed_plugins.json"))
	if err != nil {
		return installedPluginsFile{Plugins: map[string][]pluginInstance{}}
	}
	if err := json.Unmarshal(data, &result); err != nil || result.Plugins == nil {
		return installedPluginsFile{Plugins: map[string][]pluginInstance{}}
	}
	return result
}

func loadPluginMeta(installPath string) pluginMeta {
	var result pluginMeta
	data, err := os.ReadFile(filepath.Join(installPath, ".claude-plugin", "plugin.json"))
	if err != nil {
		return result
	}
	json.Unmarshal(data, &result)
	return result
}

func loadSkillLock(home string) skillLockFile {
	var result skillLockFile
	data, err := os.ReadFile(filepath.Join(home, ".agents", ".skill-lock.json"))
	if err != nil {
		return skillLockFile{Skills: map[string]skillEntry{}}
	}
	if err := json.Unmarshal(data, &result); err != nil || result.Skills == nil {
		return skillLockFile{Skills: map[string]skillEntry{}}
	}
	return result
}

func loadMarketplaces(home string) map[string]marketplaceEntry {
	result := map[string]marketplaceEntry{}
	data, err := os.ReadFile(filepath.Join(home, ".claude", "plugins", "known_marketplaces.json"))
	if err != nil {
		return result
	}
	json.Unmarshal(data, &result)
	return result
}

func loadMCPServers(installPath string) map[string]mcpServerEntry {
	result := map[string]mcpServerEntry{}
	data, err := os.ReadFile(filepath.Join(installPath, ".mcp.json"))
	if err != nil {
		return result
	}
	json.Unmarshal(data, &result)
	return result
}

// --- Helpers ---

var githubSlugRe = regexp.MustCompile(`github\.com/([^/]+/[^/.\s]+)`)

func githubSlug(url string) string {
	m := githubSlugRe.FindStringSubmatch(url)
	if m == nil {
		return ""
	}
	return strings.TrimSuffix(m[1], ".git")
}

func extractDescription(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	content := string(data)
	fm := frontmatter.ParseFrontmatter(content)
	desc := fm.StringVal("description", "")

	// Handle multiline block scalar: ParseFrontmatter returns literal "|"
	if desc == "|" || desc == ">" {
		lines := strings.Split(content, "\n")
		inFrontmatter := false
		for i, line := range lines {
			trimmed := strings.TrimRight(line, " \t\r")
			if trimmed == "---" {
				inFrontmatter = !inFrontmatter
				continue
			}
			if inFrontmatter && strings.HasPrefix(line, "description:") {
				// Next line should be the indented content
				if i+1 < len(lines) {
					desc = strings.TrimSpace(lines[i+1])
				} else {
					desc = ""
				}
				break
			}
		}
	}

	// Escape pipe characters for markdown table safety
	desc = strings.ReplaceAll(desc, "|", "\u2014")
	// Strip surrounding quotes
	if len(desc) >= 2 {
		if (desc[0] == '"' && desc[len(desc)-1] == '"') || (desc[0] == '\'' && desc[len(desc)-1] == '\'') {
			desc = desc[1 : len(desc)-1]
		}
	}
	return desc
}

var xmlTagRe = regexp.MustCompile(`<[^>]*>`)

func shortDescription(desc string) string {
	// Strip XML tags
	desc = xmlTagRe.ReplaceAllString(desc, "")
	desc = strings.Join(strings.Fields(desc), " ")

	// First sentence (up to first period followed by space or end)
	if idx := strings.Index(desc, ". "); idx >= 0 {
		desc = desc[:idx+1]
	} else if idx := strings.LastIndex(desc, "."); idx == len(desc)-1 {
		// ends with period, keep as-is
	}

	if len(desc) > 200 {
		desc = desc[:200]
	}
	return desc
}

func buildPluginRepoMap(plugins installedPluginsFile) map[string]string {
	result := map[string]string{}
	for _, instances := range plugins.Plugins {
		if len(instances) == 0 {
			continue
		}
		meta := loadPluginMeta(instances[0].InstallPath)
		if meta.Repository != "" && meta.Name != "" {
			slug := githubSlug(meta.Repository)
			if slug != "" {
				result[slug] = meta.Name
			}
		}
	}
	return result
}

// buildPluginSkillNames collects bare skill names from all installed plugins
// (skills/*/  subdirs and non-deprecated commands/*.md) → map[skillName]pluginName.
func buildPluginSkillNames(plugins installedPluginsFile) map[string]string {
	result := map[string]string{}
	for _, instances := range plugins.Plugins {
		if len(instances) == 0 {
			continue
		}
		installPath := instances[0].InstallPath
		meta := loadPluginMeta(installPath)
		pname := meta.Name
		if pname == "" {
			continue
		}

		// Skills from skills/ directory
		skillsDir := filepath.Join(installPath, "skills")
		if entries, err := os.ReadDir(skillsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					result[entry.Name()] = pname
				}
			}
		}

		// Commands from commands/ directory (non-deprecated)
		cmdsDir := filepath.Join(installPath, "commands")
		if entries, err := os.ReadDir(cmdsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				cmdFile := filepath.Join(cmdsDir, entry.Name())
				desc := extractDescription(cmdFile)
				if strings.Contains(strings.ToLower(desc), "deprecated") {
					continue
				}
				cmdName := strings.TrimSuffix(entry.Name(), ".md")
				result[cmdName] = pname
			}
		}
	}
	return result
}

// --- Markdown generation ---

func Generate() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}

	plugins := loadInstalledPlugins(home)
	skillLock := loadSkillLock(home)
	marketplaces := loadMarketplaces(home)
	pluginRepoMap := buildPluginRepoMap(plugins)
	pluginSkillNames := buildPluginSkillNames(plugins)

	var b strings.Builder

	b.WriteString("# Claude Code \u2014 Installed Plugins & Skills\n\n")
	b.WriteString(fmt.Sprintf("> Generated: %s\n\n", time.Now().Format("2006-01-02")))

	generatePluginsSection(&b, plugins)
	generateSkillsSection(&b, plugins, skillLock, pluginRepoMap, pluginSkillNames)
	generateAgentsSection(&b, plugins)
	generateMarketplacesSection(&b, marketplaces)
	generateMCPSection(&b, plugins)

	return b.String(), nil
}

func generatePluginsSection(b *strings.Builder, plugins installedPluginsFile) {
	keys := sortedKeys(plugins.Plugins)

	b.WriteString(fmt.Sprintf("## Plugins (%d)\n\n", len(keys)))
	b.WriteString("| # | Name | Description | Source | Install |\n")
	b.WriteString("|---|------|-------------|--------|--------|\n")

	for i, key := range keys {
		instances := plugins.Plugins[key]
		if len(instances) == 0 {
			continue
		}
		inst := instances[0]

		pluginName := strings.SplitN(key, "@", 2)[0]
		marketplace := ""
		if parts := strings.SplitN(key, "@", 2); len(parts) == 2 {
			marketplace = parts[1]
		}

		meta := loadPluginMeta(inst.InstallPath)

		name := meta.Name
		if name == "" {
			name = pluginName
		}
		ver := meta.Version
		if ver == "" {
			ver = inst.Version
		}

		// Format name with optional version (skip commit-hash versions)
		nameDisplay := fmt.Sprintf("**%s**", name)
		if ver != "" && isSemanticVersion(ver) {
			nameDisplay = fmt.Sprintf("**%s** (v%s)", name, ver)
		}

		// Format source
		sourceDisplay := ""
		if meta.Repository != "" {
			slug := githubSlug(meta.Repository)
			if slug != "" {
				sourceDisplay = fmt.Sprintf("[%s](%s) via `%s`", slug, meta.Repository, marketplace)
			} else {
				sourceDisplay = fmt.Sprintf("[%s](%s) via `%s`", meta.Repository, meta.Repository, marketplace)
			}
		} else if marketplace != "" {
			sourceDisplay = fmt.Sprintf("`%s`", marketplace)
		}

		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | `/plugin install %s@%s` |\n",
			i+1, nameDisplay, meta.Description, sourceDisplay, pluginName, marketplace))
	}

	b.WriteString("\n---\n\n")
}

func isSemanticVersion(ver string) bool {
	return len(ver) > 0 && ver[0] >= '0' && ver[0] <= '9' && strings.Contains(ver, ".")
}

type skillGroup struct {
	source    string
	sourceURL string
	skills    []skillInfo
}

type skillInfo struct {
	name string
	desc string
}

func generateSkillsSection(b *strings.Builder, plugins installedPluginsFile, skillLock skillLockFile, pluginRepoMap map[string]string, pluginSkillNames map[string]string) {
	home, _ := os.UserHomeDir()

	// Separate skills into plugin-provided (superpowers) and standalone
	var spSkills []string
	var spSource, spSourceURL string
	standaloneGroups := map[string]*skillGroup{}

	skillNames := sortedKeys(skillLock.Skills)
	for _, skillName := range skillNames {
		entry := skillLock.Skills[skillName]
		if _, isPlugin := pluginRepoMap[entry.Source]; isPlugin {
			spSkills = append(spSkills, skillName)
			spSource = entry.Source
			spSourceURL = entry.SourceURL
		} else {
			skillMD := filepath.Join(home, ".agents", "skills", skillName, "SKILL.md")
			desc := extractDescription(skillMD)

			if _, ok := standaloneGroups[entry.Source]; !ok {
				standaloneGroups[entry.Source] = &skillGroup{
					source:    entry.Source,
					sourceURL: entry.SourceURL,
				}
			}
			standaloneGroups[entry.Source].skills = append(standaloneGroups[entry.Source].skills, skillInfo{name: skillName, desc: desc})
		}
	}

	// Track duplicate standalone skills that are also provided by plugins
	type duplicateInfo struct {
		skillName  string
		pluginName string
	}
	var duplicates []duplicateInfo
	for _, g := range standaloneGroups {
		for _, s := range g.skills {
			if pname, ok := pluginSkillNames[s.name]; ok {
				duplicates = append(duplicates, duplicateInfo{skillName: s.name, pluginName: pname})
			}
		}
	}
	sort.Slice(duplicates, func(i, j int) bool { return duplicates[i].skillName < duplicates[j].skillName })

	// Count plugin-provided skills (from plugin dirs)
	pluginSkillCount := countPluginProvidedSkills(plugins)

	// Remove duplicate standalone skills (already provided by plugins)
	duplicateSet := map[string]bool{}
	for _, d := range duplicates {
		duplicateSet[d.skillName] = true
	}
	for src, g := range standaloneGroups {
		filtered := g.skills[:0]
		for _, s := range g.skills {
			if !duplicateSet[s.name] {
				filtered = append(filtered, s)
			}
		}
		g.skills = filtered
		if len(filtered) == 0 {
			delete(standaloneGroups, src)
		}
	}

	// Count after filtering duplicates
	standaloneCount := 0
	for _, g := range standaloneGroups {
		standaloneCount += len(g.skills)
	}
	totalSkills := standaloneCount + pluginSkillCount

	b.WriteString(fmt.Sprintf("## Skills (%d)\n\n", totalSkills))
	b.WriteString("Install skills globally with `npx skills add <repo> -g -s <skill-name>`. Add `-l` to list available skills before installing.\n\n")

	// Superpowers note
	if len(spSkills) > 0 {
		sort.Strings(spSkills)
		cleanURL := strings.TrimSuffix(spSourceURL, ".git")
		spPluginName := pluginRepoMap[spSource]

		b.WriteString(fmt.Sprintf("> **Note:** The %d [%s](%s) skills (%s) are provided by the **%s plugin** and invoked via the `%s:` prefix (e.g., `%s:brainstorming`). They are not installed as standalone skills.\n",
			len(spSkills), spSource, cleanURL, strings.Join(spSkills, ", "), spPluginName, spPluginName, spPluginName))

		// Detect deprecated aliases
		deprecatedAliases := findDeprecatedAliases(plugins, spPluginName)
		if len(deprecatedAliases) > 0 {
			b.WriteString(">\n")
			b.WriteString("> **Deprecated aliases:** The following superpowers skill names are deprecated but still functional:\n")
			for _, alias := range deprecatedAliases {
				b.WriteString(fmt.Sprintf("> - `%s` \u2192 use `%s`\n", alias.old, alias.new))
			}
		}
		b.WriteString("\n")
	}

	// Standalone skill groups (sorted by source)
	groupSources := make([]string, 0, len(standaloneGroups))
	for src := range standaloneGroups {
		groupSources = append(groupSources, src)
	}
	sort.Strings(groupSources)

	for _, src := range groupSources {
		g := standaloneGroups[src]
		cleanURL := strings.TrimSuffix(g.sourceURL, ".git")
		skillWord := "skills"
		if len(g.skills) == 1 {
			skillWord = "skill"
		}

		b.WriteString(fmt.Sprintf("### From [%s](%s) (%d %s)\n\n", src, cleanURL, len(g.skills), skillWord))
		b.WriteString("| Skill | Description | Install |\n")
		b.WriteString("|-------|-------------|--------|\n")

		sort.Slice(g.skills, func(i, j int) bool { return g.skills[i].name < g.skills[j].name })
		for _, s := range g.skills {
			b.WriteString(fmt.Sprintf("| `%s` | %s | `npx skills add %s -g -s %s` |\n", s.name, s.desc, src, s.name))
		}
		b.WriteString("\n")
	}

	// Plugin-Provided Skills
	b.WriteString("### Plugin-Provided Skills\n\n")
	b.WriteString("Skills exposed by installed plugins, invoked via the `Skill` tool or `/skill-name` shorthand. These do not require separate installation.\n\n")
	b.WriteString("| Skill | Description | Plugin Source |\n")
	b.WriteString("|-------|-------------|---------------|\n")

	keys := sortedKeys(plugins.Plugins)
	for _, key := range keys {
		instances := plugins.Plugins[key]
		if len(instances) == 0 {
			continue
		}
		installPath := instances[0].InstallPath
		meta := loadPluginMeta(installPath)
		pname := meta.Name
		if pname == "" {
			pname = strings.SplitN(key, "@", 2)[0]
		}

		// Skills from skills/ directory
		skillsDir := filepath.Join(installPath, "skills")
		if entries, err := os.ReadDir(skillsDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				skillMD := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
				desc := extractDescription(skillMD)
				b.WriteString(fmt.Sprintf("| `%s:%s` | %s | %s |\n", pname, entry.Name(), desc, pname))
			}
		}

		// Commands from commands/ directory
		cmdsDir := filepath.Join(installPath, "commands")
		if entries, err := os.ReadDir(cmdsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				cmdFile := filepath.Join(cmdsDir, entry.Name())
				cmdDesc := extractDescription(cmdFile)
				if strings.Contains(strings.ToLower(cmdDesc), "deprecated") {
					continue
				}
				cmdName := strings.TrimSuffix(entry.Name(), ".md")
				b.WriteString(fmt.Sprintf("| `%s:%s` | %s | %s |\n", pname, cmdName, cmdDesc, pname))
			}
		}
	}

	// Emit redundant standalone skills blockquote for sync to pick up
	if len(duplicates) > 0 {
		b.WriteString("> **Redundant standalone skills:** These are already provided by plugins and can be removed:\n")
		for _, d := range duplicates {
			b.WriteString(fmt.Sprintf("> - `npx skills remove %s`\n", d.skillName))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n---\n\n")
}

func countPluginProvidedSkills(plugins installedPluginsFile) int {
	count := 0
	for _, instances := range plugins.Plugins {
		if len(instances) == 0 {
			continue
		}
		installPath := instances[0].InstallPath

		skillsDir := filepath.Join(installPath, "skills")
		if entries, err := os.ReadDir(skillsDir); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					count++
				}
			}
		}

		cmdsDir := filepath.Join(installPath, "commands")
		if entries, err := os.ReadDir(cmdsDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					cmdFile := filepath.Join(cmdsDir, e.Name())
					desc := extractDescription(cmdFile)
					if !strings.Contains(strings.ToLower(desc), "deprecated") {
						count++
					}
				}
			}
		}
	}
	return count
}

type deprecatedAlias struct {
	old string
	new string
}

func findDeprecatedAliases(plugins installedPluginsFile, targetPluginName string) []deprecatedAlias {
	var aliases []deprecatedAlias

	for _, instances := range plugins.Plugins {
		if len(instances) == 0 {
			continue
		}
		installPath := instances[0].InstallPath
		meta := loadPluginMeta(installPath)
		if meta.Name != targetPluginName {
			continue
		}

		cmdsDir := filepath.Join(installPath, "commands")
		entries, err := os.ReadDir(cmdsDir)
		if err != nil {
			continue
		}

		replacementRe := regexp.MustCompile(regexp.QuoteMeta(targetPluginName) + ` ([a-z-]+)`)
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			cmdFile := filepath.Join(cmdsDir, entry.Name())
			desc := extractDescription(cmdFile)
			if !strings.Contains(strings.ToLower(desc), "deprecated") {
				continue
			}

			cmdName := strings.TrimSuffix(entry.Name(), ".md")
			oldName := targetPluginName + ":" + cmdName

			// Try to find replacement from file content
			data, err := os.ReadFile(cmdFile)
			if err != nil {
				continue
			}
			m := replacementRe.FindStringSubmatch(string(data))
			newName := ""
			if m != nil {
				newName = targetPluginName + ":" + m[1]
			}
			aliases = append(aliases, deprecatedAlias{old: oldName, new: newName})
		}
	}

	sort.Slice(aliases, func(i, j int) bool { return aliases[i].old < aliases[j].old })
	return aliases
}

func generateAgentsSection(b *strings.Builder, plugins installedPluginsFile) {
	type agentInfo struct {
		id   string
		desc string
		src  string
	}
	var agents []agentInfo

	keys := sortedKeys(plugins.Plugins)
	for _, key := range keys {
		instances := plugins.Plugins[key]
		if len(instances) == 0 {
			continue
		}
		installPath := instances[0].InstallPath
		meta := loadPluginMeta(installPath)
		pname := meta.Name
		if pname == "" {
			pname = strings.SplitN(key, "@", 2)[0]
		}

		agentsDir := filepath.Join(installPath, "agents")
		entries, err := os.ReadDir(agentsDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			agentFile := filepath.Join(agentsDir, entry.Name())
			desc := shortDescription(extractDescription(agentFile))
			agentName := strings.TrimSuffix(entry.Name(), ".md")
			agents = append(agents, agentInfo{
				id:   fmt.Sprintf("%s:%s", pname, agentName),
				desc: desc,
				src:  pname + " plugin",
			})
		}
	}

	b.WriteString(fmt.Sprintf("## Custom Agents (%d)\n\n", len(agents)))
	b.WriteString("| Agent | Description | Provided By |\n")
	b.WriteString("|-------|-------------|-------------|\n")

	for _, a := range agents {
		b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", a.id, a.desc, a.src))
	}

	b.WriteString("\n---\n\n")
}

func generateMarketplacesSection(b *strings.Builder, marketplaces map[string]marketplaceEntry) {
	keys := make([]string, 0, len(marketplaces))
	for k := range marketplaces {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b.WriteString(fmt.Sprintf("## Plugin Marketplaces (%d)\n\n", len(keys)))
	b.WriteString("| Marketplace | Source | Browse |\n")
	b.WriteString("|-------------|--------|--------|\n")

	for _, name := range keys {
		entry := marketplaces[name]
		if entry.Source.Repo != "" {
			b.WriteString(fmt.Sprintf("| `%s` | [%s](https://github.com/%s) | `/plugins` |\n",
				name, entry.Source.Repo, entry.Source.Repo))
		} else {
			b.WriteString(fmt.Sprintf("| `%s` | \u2014 | `/plugins` |\n", name))
		}
	}

	b.WriteString("\n---\n\n")
}

func generateMCPSection(b *strings.Builder, plugins installedPluginsFile) {
	type mcpInfo struct {
		name      string
		transport string
		endpoint  string
	}
	var servers []mcpInfo

	keys := sortedKeys(plugins.Plugins)
	for _, key := range keys {
		instances := plugins.Plugins[key]
		if len(instances) == 0 {
			continue
		}
		installPath := instances[0].InstallPath
		mcpServers := loadMCPServers(installPath)

		serverNames := make([]string, 0, len(mcpServers))
		for name := range mcpServers {
			serverNames = append(serverNames, name)
		}
		sort.Strings(serverNames)

		for _, name := range serverNames {
			entry := mcpServers[name]
			transport := ""
			endpoint := ""
			if entry.URL != "" {
				transport = "HTTP"
				endpoint = entry.URL
			} else if entry.Command != "" {
				transport = "stdio"
				args := strings.Join(entry.Args, " ")
				if args != "" {
					endpoint = fmt.Sprintf("`%s %s`", entry.Command, args)
				} else {
					endpoint = fmt.Sprintf("`%s`", entry.Command)
				}
			}
			servers = append(servers, mcpInfo{name: name, transport: transport, endpoint: endpoint})
		}
	}

	b.WriteString(fmt.Sprintf("## MCP Servers (%d)\n\n", len(servers)))
	b.WriteString("| Server | Transport | Endpoint |\n")
	b.WriteString("|--------|-----------|----------|\n")

	for _, s := range servers {
		b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", s.name, s.transport, s.endpoint))
	}
}

// DriftEntry represents a skill listed in the lock file whose directory no longer exists on disk.
type DriftEntry struct {
	Name   string
	Source string
}

// DetectDrift returns skills listed in .skill-lock.json whose SKILL.md no longer exists on disk.
func DetectDrift() ([]DriftEntry, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}

	skillLock := loadSkillLock(home)
	skillNames := sortedKeys(skillLock.Skills)

	var entries []DriftEntry
	for _, skillName := range skillNames {
		skillMD := filepath.Join(home, ".agents", "skills", skillName, "SKILL.md")
		if _, err := os.Stat(skillMD); os.IsNotExist(err) {
			entries = append(entries, DriftEntry{
				Name:   skillName,
				Source: skillLock.Skills[skillName].Source,
			})
		}
	}
	return entries, nil
}

// OrphanEntry represents a skill directory that exists on disk but has no entry in skill-lock.json.
type OrphanEntry struct {
	Name string
}

// DetectOrphans returns skills in ~/.agents/skills/ that have no entry in .skill-lock.json.
// Plugin-provided skills (those whose name matches a known plugin skill) are excluded.
func DetectOrphans() ([]OrphanEntry, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}

	skillsDir := filepath.Join(home, ".agents", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading skills dir: %w", err)
	}

	skillLock := loadSkillLock(home)
	plugins := loadInstalledPlugins(home)
	pluginSkillNames := buildPluginSkillNames(plugins)

	var orphans []OrphanEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, inLock := skillLock.Skills[name]; inLock {
			continue
		}
		if _, isPluginSkill := pluginSkillNames[name]; isPluginSkill {
			continue
		}
		orphans = append(orphans, OrphanEntry{Name: name})
	}
	return orphans, nil
}

// RemoveDriftEntries removes the specified skill entries from ~/.agents/.skill-lock.json.
func RemoveDriftEntries(entries []DriftEntry) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home dir: %w", err)
	}

	lockPath := filepath.Join(home, ".agents", ".skill-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return fmt.Errorf("reading skill-lock: %w", err)
	}

	var lock map[string]interface{}
	if err := json.Unmarshal(data, &lock); err != nil {
		return fmt.Errorf("parsing skill-lock: %w", err)
	}

	skills, ok := lock["skills"].(map[string]interface{})
	if !ok {
		return nil
	}

	for _, e := range entries {
		delete(skills, e.Name)
	}

	out, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling skill-lock: %w", err)
	}
	out = append(out, '\n')
	return os.WriteFile(lockPath, out, 0644)
}

// Analyze produces a terminal-friendly report of installed plugins, skills, and duplicates.
func Analyze() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}

	plugins := loadInstalledPlugins(home)
	skillLock := loadSkillLock(home)
	pluginRepoMap := buildPluginRepoMap(plugins)
	pluginSkillNames := buildPluginSkillNames(plugins)

	var b strings.Builder

	// --- Installed Plugins ---
	pluginKeys := sortedKeys(plugins.Plugins)
	b.WriteString(termcolor.Header("Installed Plugins", len(pluginKeys)))
	for i, key := range pluginKeys {
		instances := plugins.Plugins[key]
		if len(instances) == 0 {
			continue
		}
		meta := loadPluginMeta(instances[0].InstallPath)
		name := meta.Name
		if name == "" {
			name = strings.SplitN(key, "@", 2)[0]
		}
		ver := meta.Version
		if ver == "" {
			ver = instances[0].Version
		}
		verDisplay := ""
		if ver != "" && isSemanticVersion(ver) {
			verDisplay = fmt.Sprintf(" (v%s)", ver)
		}
		b.WriteString(termcolor.Numbered(i+1, name, verDisplay) + "\n")
	}
	b.WriteString("\n")

	// --- Plugin-Provided Skills (grouped by plugin) ---
	type pluginSkillInfo struct {
		name string
		desc string
	}
	pluginSkillsByPlugin := map[string][]pluginSkillInfo{}
	for _, key := range pluginKeys {
		instances := plugins.Plugins[key]
		if len(instances) == 0 {
			continue
		}
		installPath := instances[0].InstallPath
		meta := loadPluginMeta(installPath)
		pname := meta.Name
		if pname == "" {
			pname = strings.SplitN(key, "@", 2)[0]
		}

		skillsDir := filepath.Join(installPath, "skills")
		if entries, err := os.ReadDir(skillsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					skillMD := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
					desc := shortDescription(extractDescription(skillMD))
					pluginSkillsByPlugin[pname] = append(pluginSkillsByPlugin[pname], pluginSkillInfo{name: entry.Name(), desc: desc})
				}
			}
		}

		cmdsDir := filepath.Join(installPath, "commands")
		if entries, err := os.ReadDir(cmdsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				cmdFile := filepath.Join(cmdsDir, entry.Name())
				desc := extractDescription(cmdFile)
				if strings.Contains(strings.ToLower(desc), "deprecated") {
					continue
				}
				cmdName := strings.TrimSuffix(entry.Name(), ".md")
				pluginSkillsByPlugin[pname] = append(pluginSkillsByPlugin[pname], pluginSkillInfo{name: cmdName, desc: shortDescription(desc)})
			}
		}
	}

	totalPluginSkills := 0
	for _, skills := range pluginSkillsByPlugin {
		totalPluginSkills += len(skills)
	}
	b.WriteString(termcolor.Header("Plugin-Provided Skills", totalPluginSkills))
	pnames := sortedKeys(pluginSkillsByPlugin)
	for _, pname := range pnames {
		skills := pluginSkillsByPlugin[pname]
		b.WriteString(termcolor.Section(pname) + "\n")
		for _, s := range skills {
			b.WriteString(termcolor.Item(s.name) + "\n")
		}
	}
	b.WriteString("\n")

	// --- Standalone Skills (grouped by source) ---
	type standaloneSkillInfo struct {
		name string
		desc string
	}
	standaloneBySource := map[string][]standaloneSkillInfo{}
	skillNames := sortedKeys(skillLock.Skills)
	for _, skillName := range skillNames {
		entry := skillLock.Skills[skillName]
		if _, isPlugin := pluginRepoMap[entry.Source]; isPlugin {
			continue
		}
		skillMD := filepath.Join(home, ".agents", "skills", skillName, "SKILL.md")
		desc := shortDescription(extractDescription(skillMD))
		standaloneBySource[entry.Source] = append(standaloneBySource[entry.Source], standaloneSkillInfo{name: skillName, desc: desc})
	}

	totalStandalone := 0
	for _, skills := range standaloneBySource {
		totalStandalone += len(skills)
	}
	b.WriteString(termcolor.Header("Standalone Skills", totalStandalone))
	sources := sortedKeys(standaloneBySource)
	for _, src := range sources {
		skills := standaloneBySource[src]
		b.WriteString(termcolor.Section(src) + "\n")
		for _, s := range skills {
			b.WriteString(termcolor.Item(s.name) + "\n")
		}
	}
	b.WriteString("\n")

	// --- Duplicates ---
	type dupInfo struct {
		skillName  string
		pluginName string
	}
	var duplicates []dupInfo
	for _, src := range sources {
		for _, s := range standaloneBySource[src] {
			if pname, ok := pluginSkillNames[s.name]; ok {
				duplicates = append(duplicates, dupInfo{skillName: s.name, pluginName: pname})
			}
		}
	}
	sort.Slice(duplicates, func(i, j int) bool { return duplicates[i].skillName < duplicates[j].skillName })

	b.WriteString(termcolor.Header("Duplicates", len(duplicates)))
	if len(duplicates) == 0 {
		b.WriteString("  " + termcolor.Success("No duplicates found.") + "\n")
	} else {
		for _, d := range duplicates {
			b.WriteString("  " + termcolor.Warn(fmt.Sprintf("\"%s\" installed standalone AND provided by plugin \"%s\"", d.skillName, d.pluginName)) + "\n")
			b.WriteString(fmt.Sprintf("    %s→%s Remove standalone: %snpx skills remove %s%s\n", termcolor.Dim, termcolor.Reset, termcolor.Bold, d.skillName, termcolor.Reset))
		}
	}

	b.WriteString("\n")

	// --- Skill Lock Drift ---
	driftEntries, _ := DetectDrift()

	b.WriteString(termcolor.Header("Skill Lock Drift", len(driftEntries)))
	if len(driftEntries) == 0 {
		b.WriteString("  " + termcolor.Success("No drift detected.") + "\n")
	} else {
		for _, entry := range driftEntries {
			b.WriteString("  " + termcolor.Err(fmt.Sprintf("\"%s\" in lock file but missing from filesystem (source: %s)", entry.Name, entry.Source)) + "\n")
		}
	}

	b.WriteString("\n")

	// --- Orphaned Skills ---
	orphanEntries, _ := DetectOrphans()

	b.WriteString(termcolor.Header("Orphaned Skills (on disk, not in lock)", len(orphanEntries)))
	if len(orphanEntries) == 0 {
		b.WriteString("  " + termcolor.Success("No orphaned skills found.") + "\n")
	} else {
		for _, entry := range orphanEntries {
			b.WriteString("  " + termcolor.Warn(fmt.Sprintf("\"%s\" exists on disk but missing from .skill-lock.json", entry.Name)) + "\n")
		}
	}

	return b.String(), nil
}

// WritePluginsAndSkills generates the doc and writes it atomically.
func WritePluginsAndSkills(outputPath string) error {
	content, err := Generate()
	if err != nil {
		return err
	}
	tmp := outputPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, outputPath)
}

// sortedKeys returns the sorted keys of a map.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
