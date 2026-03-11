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
	standaloneCount := 0
	for _, g := range standaloneGroups {
		standaloneCount += len(g.skills)
	}
	totalSkills := standaloneCount + pluginSkillCount - len(duplicates)

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

	// Build set of duplicate skill names for quick lookup
	duplicateSet := map[string]string{}
	for _, d := range duplicates {
		duplicateSet[d.skillName] = d.pluginName
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
			desc := s.desc
			if pname, isDup := duplicateSet[s.name]; isDup {
				desc += fmt.Sprintf(" *(redundant — provided by **%s** plugin)*", pname)
			}
			b.WriteString(fmt.Sprintf("| `%s` | %s | `npx skills add %s -g -s %s` |\n", s.name, desc, src, s.name))
		}
		b.WriteString("\n")
	}

	// Emit redundant skills warning blockquote
	if len(duplicates) > 0 {
		b.WriteString("> **Redundant standalone skills:** These are already provided by plugins and can be removed:\n")
		for _, d := range duplicates {
			b.WriteString(fmt.Sprintf("> - `%s` (provided by **%s** plugin) — `npx skills remove %s`\n", d.skillName, d.pluginName, d.skillName))
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
