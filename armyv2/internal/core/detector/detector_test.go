package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

func TestDetect_FileExistence(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644)

	profiles := map[string]types.TechProfile{
		"go":     {Detect: []string{"go.mod"}},
		"python": {Detect: []string{"requirements.txt"}},
	}

	matched := Detect(dir, profiles)
	if len(matched) != 1 || matched[0] != "go" {
		t.Errorf("got %v, want [go]", matched)
	}
}

func TestDetect_GlobPattern(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.tsx"), []byte(""), 0o644)

	profiles := map[string]types.TechProfile{
		"react": {Detect: []string{"*.tsx"}},
	}

	matched := Detect(dir, profiles)
	if len(matched) != 1 || matched[0] != "react" {
		t.Errorf("got %v, want [react]", matched)
	}
}

func TestDetect_ContentMatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.txt"), []byte("some content with marker here"), 0o644)

	profiles := map[string]types.TechProfile{
		"custom": {Detect: []string{"config.txt:marker"}},
	}

	matched := Detect(dir, profiles)
	if len(matched) != 1 || matched[0] != "custom" {
		t.Errorf("got %v, want [custom]", matched)
	}
}

func TestDetect_ContentNoMatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.txt"), []byte("no match here"), 0o644)

	profiles := map[string]types.TechProfile{
		"custom": {Detect: []string{"config.txt:marker"}},
	}

	matched := Detect(dir, profiles)
	if len(matched) != 0 {
		t.Errorf("got %v, want empty", matched)
	}
}

func TestDetect_PackageJSONDependency(t *testing.T) {
	dir := t.TempDir()
	pkg := map[string]interface{}{
		"name":         "test",
		"dependencies": map[string]interface{}{"react": "^18.0.0"},
	}
	data, _ := json.Marshal(pkg)
	os.WriteFile(filepath.Join(dir, "package.json"), data, 0o644)

	profiles := map[string]types.TechProfile{
		"react":  {Detect: []string{"package.json:react"}},
		"vue":    {Detect: []string{"package.json:vue"}},
		"nestjs": {Detect: []string{"package.json:@nestjs/core"}},
	}

	matched := Detect(dir, profiles)
	if len(matched) != 1 || matched[0] != "react" {
		t.Errorf("got %v, want [react]", matched)
	}
}

func TestDetect_PackageJSONDevDependency(t *testing.T) {
	dir := t.TempDir()
	pkg := map[string]interface{}{
		"name":            "test",
		"devDependencies": map[string]interface{}{"jest": "^29.0.0"},
	}
	data, _ := json.Marshal(pkg)
	os.WriteFile(filepath.Join(dir, "package.json"), data, 0o644)

	profiles := map[string]types.TechProfile{
		"jest": {Detect: []string{"package.json:jest"}},
	}

	matched := Detect(dir, profiles)
	if len(matched) != 1 || matched[0] != "jest" {
		t.Errorf("got %v, want [jest]", matched)
	}
}

func TestDetect_ComposerJSONDependency(t *testing.T) {
	dir := t.TempDir()
	composer := map[string]interface{}{
		"require": map[string]interface{}{"laravel/framework": "^10.0"},
	}
	data, _ := json.Marshal(composer)
	os.WriteFile(filepath.Join(dir, "composer.json"), data, 0o644)

	profiles := map[string]types.TechProfile{
		"laravel": {Detect: []string{"composer.json:laravel/framework"}},
	}

	matched := Detect(dir, profiles)
	if len(matched) != 1 || matched[0] != "laravel" {
		t.Errorf("got %v, want [laravel]", matched)
	}
}

func TestDetect_NoMarkers(t *testing.T) {
	dir := t.TempDir()
	profiles := map[string]types.TechProfile{
		"go":     {Detect: []string{"go.mod"}},
		"python": {Detect: []string{"requirements.txt"}},
	}

	matched := Detect(dir, profiles)
	if len(matched) != 0 {
		t.Errorf("got %v, want empty", matched)
	}
}

func TestDetect_MultipleMatches(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644)
	os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0o644)

	profiles := map[string]types.TechProfile{
		"go":         {Detect: []string{"go.mod"}},
		"typescript": {Detect: []string{"tsconfig.json"}},
		"python":     {Detect: []string{"*.py"}},
	}

	matched := Detect(dir, profiles)
	sort.Strings(matched)
	if len(matched) != 2 {
		t.Fatalf("got %d matches, want 2: %v", len(matched), matched)
	}
	if matched[0] != "go" || matched[1] != "typescript" {
		t.Errorf("got %v, want [go, typescript]", matched)
	}
}

func TestDetect_FirstMarkerWins(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644)
	os.WriteFile(filepath.Join(dir, "go.sum"), []byte(""), 0o644)

	profiles := map[string]types.TechProfile{
		"go": {Detect: []string{"go.mod", "go.sum"}},
	}

	matched := Detect(dir, profiles)
	if len(matched) != 1 {
		t.Errorf("should match once even with multiple markers: %v", matched)
	}
}

func TestRecommendedItems_Basic(t *testing.T) {
	profiles := map[string]types.TechProfile{
		"go":   {Plugins: []string{"gopls"}, Skills: []string{"golang-pro"}},
		"react": {Plugins: []string{"frontend-design"}, Skills: []string{"react-expert", "js-pro"}},
	}

	plugins, skills := RecommendedItems([]string{"go", "react"}, profiles)

	if len(plugins) != 2 {
		t.Errorf("got %d plugins, want 2: %v", len(plugins), plugins)
	}
	if len(skills) != 3 {
		t.Errorf("got %d skills, want 3: %v", len(skills), skills)
	}
}

func TestRecommendedItems_Deduplication(t *testing.T) {
	profiles := map[string]types.TechProfile{
		"react":  {Plugins: []string{"frontend-design"}, Skills: []string{"js-pro"}},
		"nextjs": {Plugins: []string{"frontend-design"}, Skills: []string{"js-pro", "nextjs-dev"}},
	}

	plugins, skills := RecommendedItems([]string{"react", "nextjs"}, profiles)

	if len(plugins) != 1 {
		t.Errorf("plugins should be deduplicated: got %v", plugins)
	}
	if len(skills) != 2 {
		t.Errorf("skills should be deduplicated: got %v", skills)
	}
}

func TestRecommendedItems_UnknownProfile(t *testing.T) {
	profiles := map[string]types.TechProfile{
		"go": {Plugins: []string{"gopls"}, Skills: []string{"golang-pro"}},
	}

	plugins, skills := RecommendedItems([]string{"unknown"}, profiles)
	if len(plugins) != 0 || len(skills) != 0 {
		t.Errorf("unknown profile should return empty: plugins=%v skills=%v", plugins, skills)
	}
}

func TestRecommendedItems_EmptyInput(t *testing.T) {
	profiles := map[string]types.TechProfile{
		"go": {Plugins: []string{"gopls"}, Skills: []string{"golang-pro"}},
	}

	plugins, skills := RecommendedItems(nil, profiles)
	if len(plugins) != 0 || len(skills) != 0 {
		t.Error("nil input should return empty results")
	}
}
