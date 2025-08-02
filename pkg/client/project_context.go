package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	sdkerrors "github.com/jonwraymond/go-claude-code-sdk/pkg/errors"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

// ProjectContextManager provides advanced project context management for Claude Code integration.
// It analyzes project structure, dependencies, patterns, and provides intelligent context
// for Claude Code sessions.
//
// Key features:
// - Deep project analysis and pattern recognition
// - Dependency analysis and version tracking
// - Code pattern and architecture detection
// - Integration with Claude Code's project-aware features
// - Context caching and incremental updates
// - Project configuration and metadata management
//
// The manager helps Claude Code understand the project structure, coding patterns,
// and architectural decisions to provide more relevant and contextual assistance.
//
// Example usage:
//
//	manager := NewProjectContextManager(client)
//	context, err := manager.GetEnhancedProjectContext(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use context for Claude Code operations
//	fmt.Printf("Project: %s (%s)\n", context.ProjectName, context.Language)
//	fmt.Printf("Architecture: %s\n", context.Architecture.Pattern)
type ProjectContextManager struct {
	client           *ClaudeCodeClient
	cachedContext    *types.ProjectContext
	lastCacheUpdate  time.Time
	cacheDuration    time.Duration
	analysisPatterns map[string]*AnalysisPattern
	mu               sync.RWMutex
}

// AnalysisPattern defines patterns for analyzing project structures and dependencies.
type AnalysisPattern struct {
	Name         string
	FilePatterns []string
	Indicators   []string
	Analyzer     func(projectPath string) (*PatternResult, error)
}

// PatternResult contains the result of pattern analysis.
type PatternResult struct {
	Pattern      string                 `json:"pattern"`
	Confidence   float64                `json:"confidence"`
	Evidence     []string               `json:"evidence"`
	Metadata     map[string]any `json:"metadata"`
	Dependencies []string               `json:"dependencies,omitempty"`
}

// ArchitectureInfo contains information about project architecture.
type ArchitectureInfo struct {
	Pattern     string                 `json:"pattern"`
	Layers      []string               `json:"layers"`
	Modules     []string               `json:"modules"`
	EntryPoints []string               `json:"entry_points"`
	ConfigFiles []string               `json:"config_files"`
	Metadata    map[string]any `json:"metadata"`
}

// DependencyInfo contains information about project dependencies.
type DependencyInfo struct {
	Manager      string                 `json:"manager"`
	ConfigFile   string                 `json:"config_file"`
	Dependencies map[string]string      `json:"dependencies"`
	DevDeps      map[string]string      `json:"dev_dependencies,omitempty"`
	Scripts      map[string]string      `json:"scripts,omitempty"`
	Metadata     map[string]any `json:"metadata"`
}

// NewProjectContextManager creates a new project context manager.
func NewProjectContextManager(client *ClaudeCodeClient) *ProjectContextManager {
	manager := &ProjectContextManager{
		client:           client,
		cacheDuration:    5 * time.Minute, // Cache for 5 minutes
		analysisPatterns: make(map[string]*AnalysisPattern),
	}

	// Initialize analysis patterns
	manager.initializeAnalysisPatterns()

	return manager
}

// GetEnhancedProjectContext returns an enhanced project context with deep analysis.
func (pm *ProjectContextManager) GetEnhancedProjectContext(ctx context.Context) (*types.ProjectContext, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if cached context is still valid
	if pm.cachedContext != nil && time.Since(pm.lastCacheUpdate) < pm.cacheDuration {
		return pm.cachedContext, nil
	}

	// Get base context from client
	baseContext, err := pm.client.GetProjectContext(ctx)
	if err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "BASE_CONTEXT", "failed to get base project context")
	}

	// Enhance with advanced analysis
	if err := pm.enhanceContext(baseContext); err != nil {
		return nil, sdkerrors.WrapError(err, sdkerrors.CategoryInternal, "ENHANCE_CONTEXT", "failed to enhance project context")
	}

	// Cache the enhanced context
	pm.cachedContext = baseContext
	pm.lastCacheUpdate = time.Now()

	return baseContext, nil
}

// enhanceContext enhances the base project context with advanced analysis.
func (pm *ProjectContextManager) enhanceContext(context *types.ProjectContext) error {
	projectPath := context.WorkingDirectory

	// Add architecture analysis
	if arch, err := pm.analyzeArchitecture(projectPath); err == nil {
		if context.Metadata == nil {
			context.Metadata = make(map[string]any)
		}
		context.Metadata["architecture"] = arch
	}

	// Add dependency analysis
	if deps, err := pm.analyzeDependencies(projectPath); err == nil {
		if context.Metadata == nil {
			context.Metadata = make(map[string]any)
		}
		context.Metadata["dependencies"] = deps
	}

	// Add code patterns analysis
	if patterns, err := pm.analyzeCodePatterns(projectPath); err == nil {
		if context.Metadata == nil {
			context.Metadata = make(map[string]any)
		}
		context.Metadata["code_patterns"] = patterns
	}

	// Add development tools analysis
	if tools, err := pm.analyzeDevTools(projectPath); err == nil {
		if context.Metadata == nil {
			context.Metadata = make(map[string]any)
		}
		context.Metadata["dev_tools"] = tools
	}

	// Enhance file information with more details
	if err := pm.enhanceFileInfo(context); err != nil {
		return err
	}

	return nil
}

// analyzeArchitecture analyzes the project architecture patterns.
func (pm *ProjectContextManager) analyzeArchitecture(projectPath string) (*ArchitectureInfo, error) {
	arch := &ArchitectureInfo{
		Layers:      []string{},
		Modules:     []string{},
		EntryPoints: []string{},
		ConfigFiles: []string{},
		Metadata:    make(map[string]any),
	}

	// Detect architecture pattern
	if pattern := pm.detectArchitecturePattern(projectPath); pattern != "" {
		arch.Pattern = pattern
	} else {
		arch.Pattern = "unknown"
	}

	// Find entry points
	entryPoints := pm.findEntryPoints(projectPath)
	arch.EntryPoints = entryPoints

	// Find configuration files
	configFiles := pm.findConfigFiles(projectPath)
	arch.ConfigFiles = configFiles

	// Analyze directory structure for layers/modules
	layers, modules := pm.analyzeDirectoryStructure(projectPath)
	arch.Layers = layers
	arch.Modules = modules

	return arch, nil
}

// detectArchitecturePattern detects common architecture patterns.
func (pm *ProjectContextManager) detectArchitecturePattern(projectPath string) string {
	patterns := map[string][]string{
		"microservices": {"services/", "cmd/", "api/", "docker-compose.yml"},
		"monorepo":      {"packages/", "apps/", "libs/", "workspace.json", "lerna.json"},
		"mvc":           {"models/", "views/", "controllers/", "routes/"},
		"hexagonal":     {"domain/", "ports/", "adapters/", "infrastructure/"},
		"clean":         {"entities/", "usecases/", "frameworks/", "interfaces/"},
		"layered":       {"presentation/", "business/", "data/", "service/"},
		"modular":       {"modules/", "components/", "features/"},
	}

	scores := make(map[string]int)

	for pattern, indicators := range patterns {
		score := 0
		for _, indicator := range indicators {
			if _, err := os.Stat(filepath.Join(projectPath, indicator)); err == nil {
				score++
			}
		}
		if score > 0 {
			scores[pattern] = score
		}
	}

	// Return pattern with highest score
	maxScore := 0
	bestPattern := ""
	for pattern, score := range scores {
		if score > maxScore {
			maxScore = score
			bestPattern = pattern
		}
	}

	return bestPattern
}

// findEntryPoints finds common entry point files.
func (pm *ProjectContextManager) findEntryPoints(projectPath string) []string {
	entryPoints := []string{}

	candidates := []string{
		"main.go", "cmd/main.go", "cmd/*/main.go",
		"index.js", "index.ts", "src/index.js", "src/index.ts",
		"main.py", "__main__.py", "app.py", "run.py",
		"Main.java", "Application.java", "App.java",
		"main.cpp", "main.c",
		"package.json", // For npm scripts
		"Makefile", "makefile",
		"docker-compose.yml", "Dockerfile",
	}

	for _, candidate := range candidates {
		if strings.Contains(candidate, "*") {
			// Handle glob patterns
			matches, err := filepath.Glob(filepath.Join(projectPath, candidate))
			if err == nil {
				for _, match := range matches {
					if rel, err := filepath.Rel(projectPath, match); err == nil {
						entryPoints = append(entryPoints, rel)
					}
				}
			}
		} else {
			if _, err := os.Stat(filepath.Join(projectPath, candidate)); err == nil {
				entryPoints = append(entryPoints, candidate)
			}
		}
	}

	return entryPoints
}

// findConfigFiles finds configuration files.
func (pm *ProjectContextManager) findConfigFiles(projectPath string) []string {
	configFiles := []string{}

	candidates := []string{
		// Build and package management
		"package.json", "package-lock.json", "yarn.lock", "pnpm-lock.yaml",
		"go.mod", "go.sum", "Cargo.toml", "Cargo.lock",
		"requirements.txt", "Pipfile", "poetry.lock", "setup.py",
		"pom.xml", "build.gradle", "build.sbt",
		"composer.json", "composer.lock",

		// Configuration files
		"tsconfig.json", "jsconfig.json", ".babelrc", "webpack.config.js",
		".env", ".env.local", ".env.example",
		"config.json", "config.yaml", "config.yml",
		"docker-compose.yml", "Dockerfile", ".dockerignore",

		// Development tools
		".eslintrc", ".eslintrc.json", ".eslintrc.js",
		".prettierrc", ".prettierrc.json",
		".gitignore", ".gitattributes",
		"Makefile", "makefile",

		// IDE and editor
		".vscode/", ".idea/", ".editorconfig",

		// CI/CD
		".github/", ".gitlab-ci.yml", "azure-pipelines.yml",
		"Jenkinsfile", ".travis.yml", ".circleci/",
	}

	for _, candidate := range candidates {
		path := filepath.Join(projectPath, candidate)
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				// For directories, check if they exist and have content
				if entries, err := os.ReadDir(path); err == nil && len(entries) > 0 {
					configFiles = append(configFiles, candidate)
				}
			} else {
				configFiles = append(configFiles, candidate)
			}
		}
	}

	return configFiles
}

// analyzeDirectoryStructure analyzes the directory structure for layers and modules.
func (pm *ProjectContextManager) analyzeDirectoryStructure(projectPath string) ([]string, []string) {
	layers := []string{}
	modules := []string{}

	// Common layer names
	layerNames := map[string]bool{
		"api": true, "web": true, "ui": true, "frontend": true,
		"service": true, "business": true, "logic": true,
		"data": true, "database": true, "storage": true,
		"infrastructure": true, "config": true, "common": true,
		"domain": true, "entities": true, "models": true,
		"controllers": true, "handlers": true, "routes": true,
		"middleware": true, "auth": true, "security": true,
		"utils": true, "helpers": true, "lib": true, "libs": true,
	}

	// Scan top-level directories
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return layers, modules
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		name := strings.ToLower(entry.Name())

		// Check if it's a common layer
		if layerNames[name] {
			layers = append(layers, entry.Name())
		} else {
			// Consider it a module
			modules = append(modules, entry.Name())
		}
	}

	// Sort for consistency
	sort.Strings(layers)
	sort.Strings(modules)

	return layers, modules
}

// analyzeDependencies analyzes project dependencies.
func (pm *ProjectContextManager) analyzeDependencies(projectPath string) (*DependencyInfo, error) {
	// Check for different dependency managers
	managers := []struct {
		name       string
		configFile string
		analyzer   func(string) (*DependencyInfo, error)
	}{
		{"npm", "package.json", pm.analyzeNpmDependencies},
		{"go", "go.mod", pm.analyzeGoDependencies},
		{"cargo", "Cargo.toml", pm.analyzeCargoDependencies},
		{"pip", "requirements.txt", pm.analyzePipDependencies},
		{"maven", "pom.xml", pm.analyzeMavenDependencies},
		{"gradle", "build.gradle", pm.analyzeGradleDependencies},
		{"composer", "composer.json", pm.analyzeComposerDependencies},
	}

	for _, manager := range managers {
		configPath := filepath.Join(projectPath, manager.configFile)
		if _, err := os.Stat(configPath); err == nil {
			return manager.analyzer(configPath)
		}
	}

	return nil, fmt.Errorf("no supported dependency manager found")
}

// analyzeNpmDependencies analyzes npm dependencies from package.json.
func (pm *ProjectContextManager) analyzeNpmDependencies(configPath string) (*DependencyInfo, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		Main            string            `json:"main"`
		Type            string            `json:"type"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	deps := &DependencyInfo{
		Manager:      "npm",
		ConfigFile:   filepath.Base(configPath),
		Dependencies: pkg.Dependencies,
		DevDeps:      pkg.DevDependencies,
		Scripts:      pkg.Scripts,
		Metadata: map[string]any{
			"name":    pkg.Name,
			"version": pkg.Version,
			"main":    pkg.Main,
			"type":    pkg.Type,
		},
	}

	return deps, nil
}

// analyzeGoDependencies analyzes Go dependencies from go.mod.
func (pm *ProjectContextManager) analyzeGoDependencies(configPath string) (*DependencyInfo, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	deps := &DependencyInfo{
		Manager:      "go",
		ConfigFile:   filepath.Base(configPath),
		Dependencies: make(map[string]string),
		Metadata:     make(map[string]any),
	}

	// Parse module name
	if matches := regexp.MustCompile(`module\s+(.+)`).FindStringSubmatch(content); len(matches) > 1 {
		deps.Metadata["module"] = strings.TrimSpace(matches[1])
	}

	// Parse Go version
	if matches := regexp.MustCompile(`go\s+([\d\.]+)`).FindStringSubmatch(content); len(matches) > 1 {
		deps.Metadata["go_version"] = strings.TrimSpace(matches[1])
	}

	// Parse require block
	requireRe := regexp.MustCompile(`require\s*\(\s*(.*?)\s*\)`)
	if matches := requireRe.FindStringSubmatch(content); len(matches) > 1 {
		lines := strings.Split(matches[1], "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				deps.Dependencies[parts[0]] = parts[1]
			}
		}
	}

	// Parse single require statements
	singleRequireRe := regexp.MustCompile(`require\s+([^\s]+)\s+([^\s]+)`)
	matches := singleRequireRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			deps.Dependencies[match[1]] = match[2]
		}
	}

	return deps, nil
}

// analyzeCargoDependencies analyzes Rust dependencies from Cargo.toml.
func (pm *ProjectContextManager) analyzeCargoDependencies(configPath string) (*DependencyInfo, error) {
	// This is a simplified TOML parser for Cargo.toml
	// In a production environment, you'd want to use a proper TOML library
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	deps := &DependencyInfo{
		Manager:      "cargo",
		ConfigFile:   filepath.Base(configPath),
		Dependencies: make(map[string]string),
		DevDeps:      make(map[string]string),
		Metadata:     make(map[string]any),
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	currentSection := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		// Parse key-value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

			switch currentSection {
			case "package":
				deps.Metadata[key] = value
			case "dependencies":
				deps.Dependencies[key] = value
			case "dev-dependencies":
				deps.DevDeps[key] = value
			}
		}
	}

	return deps, nil
}

// analyzePipDependencies analyzes Python dependencies from requirements.txt.
func (pm *ProjectContextManager) analyzePipDependencies(configPath string) (*DependencyInfo, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	deps := &DependencyInfo{
		Manager:      "pip",
		ConfigFile:   filepath.Base(configPath),
		Dependencies: make(map[string]string),
		Metadata:     make(map[string]any),
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse package specifications
		// Handle formats like: package==1.0.0, package>=1.0.0, package
		re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)([>=<~!]+.*)?$`)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			pkg := matches[1]
			version := "latest"
			if len(matches) > 2 && matches[2] != "" {
				version = matches[2]
			}
			deps.Dependencies[pkg] = version
		}
	}

	return deps, nil
}

// analyzeMavenDependencies analyzes Maven dependencies from pom.xml.
func (pm *ProjectContextManager) analyzeMavenDependencies(configPath string) (*DependencyInfo, error) {
	// This is a simplified XML parser for pom.xml
	// In a production environment, you'd want to use a proper XML library
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	deps := &DependencyInfo{
		Manager:      "maven",
		ConfigFile:   filepath.Base(configPath),
		Dependencies: make(map[string]string),
		Metadata:     make(map[string]any),
	}

	content := string(data)

	// Extract artifact info
	if matches := regexp.MustCompile(`<groupId>(.*?)</groupId>`).FindStringSubmatch(content); len(matches) > 1 {
		deps.Metadata["group_id"] = matches[1]
	}
	if matches := regexp.MustCompile(`<artifactId>(.*?)</artifactId>`).FindStringSubmatch(content); len(matches) > 1 {
		deps.Metadata["artifact_id"] = matches[1]
	}
	if matches := regexp.MustCompile(`<version>(.*?)</version>`).FindStringSubmatch(content); len(matches) > 1 {
		deps.Metadata["version"] = matches[1]
	}

	// Extract dependencies (simplified)
	dependencyRe := regexp.MustCompile(`<dependency>.*?<groupId>(.*?)</groupId>.*?<artifactId>(.*?)</artifactId>.*?<version>(.*?)</version>.*?</dependency>`)
	matches := dependencyRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			depName := fmt.Sprintf("%s:%s", match[1], match[2])
			deps.Dependencies[depName] = match[3]
		}
	}

	return deps, nil
}

// analyzeGradleDependencies analyzes Gradle dependencies from build.gradle.
func (pm *ProjectContextManager) analyzeGradleDependencies(configPath string) (*DependencyInfo, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	deps := &DependencyInfo{
		Manager:      "gradle",
		ConfigFile:   filepath.Base(configPath),
		Dependencies: make(map[string]string),
		Metadata:     make(map[string]any),
	}

	content := string(data)

	// Parse dependencies block (simplified)
	dependencyRe := regexp.MustCompile(`(?:implementation|api|compile|testImplementation)\s+['"]([^'"]+)['"]`)
	matches := dependencyRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			parts := strings.Split(match[1], ":")
			if len(parts) >= 2 {
				name := fmt.Sprintf("%s:%s", parts[0], parts[1])
				version := "latest"
				if len(parts) >= 3 {
					version = parts[2]
				}
				deps.Dependencies[name] = version
			}
		}
	}

	return deps, nil
}

// analyzeComposerDependencies analyzes PHP dependencies from composer.json.
func (pm *ProjectContextManager) analyzeComposerDependencies(configPath string) (*DependencyInfo, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var composer struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Version     string            `json:"version"`
		Require     map[string]string `json:"require"`
		RequireDev  map[string]string `json:"require-dev"`
		Scripts     map[string]string `json:"scripts"`
		Autoload    any       `json:"autoload"`
	}

	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, err
	}

	deps := &DependencyInfo{
		Manager:      "composer",
		ConfigFile:   filepath.Base(configPath),
		Dependencies: composer.Require,
		DevDeps:      composer.RequireDev,
		Scripts:      composer.Scripts,
		Metadata: map[string]any{
			"name":        composer.Name,
			"description": composer.Description,
			"version":     composer.Version,
			"autoload":    composer.Autoload,
		},
	}

	return deps, nil
}

// analyzeCodePatterns analyzes code patterns in the project.
func (pm *ProjectContextManager) analyzeCodePatterns(projectPath string) (map[string]*PatternResult, error) {
	patterns := make(map[string]*PatternResult)

	// Analyze each registered pattern
	for name, pattern := range pm.analysisPatterns {
		if result, err := pattern.Analyzer(projectPath); err == nil {
			patterns[name] = result
		}
	}

	return patterns, nil
}

// analyzeDevTools analyzes development tools and workflows.
func (pm *ProjectContextManager) analyzeDevTools(projectPath string) (map[string]any, error) {
	tools := make(map[string]any)

	// Check for common development tools
	toolChecks := map[string]func(string) any{
		"docker":     pm.checkDocker,
		"kubernetes": pm.checkKubernetes,
		"ci_cd":      pm.checkCICD,
		"testing":    pm.checkTesting,
		"linting":    pm.checkLinting,
		"formatting": pm.checkFormatting,
		"git_hooks":  pm.checkGitHooks,
		"ide_config": pm.checkIDEConfig,
	}

	for tool, checker := range toolChecks {
		if result := checker(projectPath); result != nil {
			tools[tool] = result
		}
	}

	return tools, nil
}

// checkDocker checks for Docker configuration.
func (pm *ProjectContextManager) checkDocker(projectPath string) any {
	dockerInfo := map[string]any{}

	if _, err := os.Stat(filepath.Join(projectPath, "Dockerfile")); err == nil {
		dockerInfo["dockerfile"] = true
	}

	if _, err := os.Stat(filepath.Join(projectPath, "docker-compose.yml")); err == nil {
		dockerInfo["compose"] = true
	}

	if _, err := os.Stat(filepath.Join(projectPath, ".dockerignore")); err == nil {
		dockerInfo["dockerignore"] = true
	}

	if len(dockerInfo) > 0 {
		return dockerInfo
	}
	return nil
}

// checkKubernetes checks for Kubernetes configuration.
func (pm *ProjectContextManager) checkKubernetes(projectPath string) any {
	k8sInfo := map[string]any{}

	k8sPaths := []string{"k8s/", "kubernetes/", "deploy/", "manifests/"}
	for _, path := range k8sPaths {
		if _, err := os.Stat(filepath.Join(projectPath, path)); err == nil {
			k8sInfo["manifests_dir"] = path
			break
		}
	}

	// Check for specific files
	k8sFiles := []string{"deployment.yaml", "service.yaml", "ingress.yaml", "kustomization.yaml"}
	foundFiles := []string{}
	for _, file := range k8sFiles {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
			foundFiles = append(foundFiles, file)
		}
	}

	if len(foundFiles) > 0 {
		k8sInfo["files"] = foundFiles
	}

	if len(k8sInfo) > 0 {
		return k8sInfo
	}
	return nil
}

// checkCICD checks for CI/CD configuration.
func (pm *ProjectContextManager) checkCICD(projectPath string) any {
	cicdInfo := map[string]any{}

	cicdChecks := map[string]string{
		"GitHub Actions":  ".github/workflows/",
		"GitLab CI":       ".gitlab-ci.yml",
		"Travis CI":       ".travis.yml",
		"CircleCI":        ".circleci/",
		"Azure Pipelines": "azure-pipelines.yml",
		"Jenkins":         "Jenkinsfile",
	}

	for name, path := range cicdChecks {
		if _, err := os.Stat(filepath.Join(projectPath, path)); err == nil {
			cicdInfo[strings.ToLower(strings.ReplaceAll(name, " ", "_"))] = true
		}
	}

	if len(cicdInfo) > 0 {
		return cicdInfo
	}
	return nil
}

// checkTesting checks for testing frameworks and configuration.
func (pm *ProjectContextManager) checkTesting(projectPath string) any {
	testInfo := map[string]any{}

	// Check for test directories
	testDirs := []string{"test/", "tests/", "__tests__/", "spec/"}
	for _, dir := range testDirs {
		if _, err := os.Stat(filepath.Join(projectPath, dir)); err == nil {
			testInfo["test_directory"] = dir
			break
		}
	}

	// Check for test configuration files
	testConfigs := map[string]string{
		"jest":       "jest.config.js",
		"mocha":      "mocha.opts",
		"pytest":     "pytest.ini",
		"phpunit":    "phpunit.xml",
		"go_test":    "*_test.go",
		"cargo_test": "Cargo.toml", // Rust tests are built-in
	}

	for framework, config := range testConfigs {
		if strings.Contains(config, "*") {
			// Handle glob patterns
			if matches, err := filepath.Glob(filepath.Join(projectPath, config)); err == nil && len(matches) > 0 {
				testInfo[framework] = true
			}
		} else {
			if _, err := os.Stat(filepath.Join(projectPath, config)); err == nil {
				testInfo[framework] = true
			}
		}
	}

	if len(testInfo) > 0 {
		return testInfo
	}
	return nil
}

// checkLinting checks for linting configuration.
func (pm *ProjectContextManager) checkLinting(projectPath string) any {
	lintInfo := map[string]any{}

	lintConfigs := map[string][]string{
		"eslint":   {".eslintrc", ".eslintrc.json", ".eslintrc.js", ".eslintrc.yml"},
		"golangci": {".golangci.yml", ".golangci.yaml"},
		"pylint":   {".pylintrc", "pylint.cfg"},
		"rubocop":  {".rubocop.yml"},
		"clippy":   {"Cargo.toml"}, // Rust clippy
		"phpcs":    {"phpcs.xml", "phpcs.xml.dist"},
	}

	for linter, configs := range lintConfigs {
		for _, config := range configs {
			if _, err := os.Stat(filepath.Join(projectPath, config)); err == nil {
				lintInfo[linter] = true
				break
			}
		}
	}

	if len(lintInfo) > 0 {
		return lintInfo
	}
	return nil
}

// checkFormatting checks for code formatting configuration.
func (pm *ProjectContextManager) checkFormatting(projectPath string) any {
	formatInfo := map[string]any{}

	formatConfigs := map[string][]string{
		"prettier":     {".prettierrc", ".prettierrc.json", ".prettierrc.js"},
		"gofmt":        {"go.mod"}, // Built into Go
		"black":        {"pyproject.toml", "setup.cfg"},
		"rustfmt":      {"rustfmt.toml", ".rustfmt.toml"},
		"php-cs-fixer": {".php_cs", ".php-cs-fixer.php"},
	}

	for formatter, configs := range formatConfigs {
		for _, config := range configs {
			if _, err := os.Stat(filepath.Join(projectPath, config)); err == nil {
				formatInfo[formatter] = true
				break
			}
		}
	}

	if len(formatInfo) > 0 {
		return formatInfo
	}
	return nil
}

// checkGitHooks checks for Git hooks configuration.
func (pm *ProjectContextManager) checkGitHooks(projectPath string) any {
	hookInfo := map[string]any{}

	// Check for .git/hooks directory
	hooksDir := filepath.Join(projectPath, ".git", "hooks")
	if entries, err := os.ReadDir(hooksDir); err == nil {
		hooks := []string{}
		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasSuffix(entry.Name(), ".sample") {
				hooks = append(hooks, entry.Name())
			}
		}
		if len(hooks) > 0 {
			hookInfo["custom_hooks"] = hooks
		}
	}

	// Check for husky
	if _, err := os.Stat(filepath.Join(projectPath, ".husky")); err == nil {
		hookInfo["husky"] = true
	}

	// Check for pre-commit
	if _, err := os.Stat(filepath.Join(projectPath, ".pre-commit-config.yaml")); err == nil {
		hookInfo["pre_commit"] = true
	}

	if len(hookInfo) > 0 {
		return hookInfo
	}
	return nil
}

// checkIDEConfig checks for IDE configuration.
func (pm *ProjectContextManager) checkIDEConfig(projectPath string) any {
	ideInfo := map[string]any{}

	ideConfigs := map[string]string{
		"vscode":       ".vscode/",
		"intellij":     ".idea/",
		"sublime":      "*.sublime-project",
		"editorconfig": ".editorconfig",
	}

	for ide, config := range ideConfigs {
		if strings.Contains(config, "*") {
			if matches, err := filepath.Glob(filepath.Join(projectPath, config)); err == nil && len(matches) > 0 {
				ideInfo[ide] = true
			}
		} else {
			if _, err := os.Stat(filepath.Join(projectPath, config)); err == nil {
				ideInfo[ide] = true
			}
		}
	}

	if len(ideInfo) > 0 {
		return ideInfo
	}
	return nil
}

// enhanceFileInfo enhances the file information with more detailed analysis.
func (pm *ProjectContextManager) enhanceFileInfo(context *types.ProjectContext) error {
	if context.Files == nil {
		return nil
	}

	projectPath := context.WorkingDirectory

	// Add code quality metrics
	if metrics, err := pm.calculateCodeMetrics(projectPath); err == nil {
		if context.Files.Structure == nil {
			context.Files.Structure = make(map[string]any)
		}
		context.Files.Structure["metrics"] = metrics
	}

	// Add file size distribution
	if sizes, err := pm.analyzeFileSizes(projectPath); err == nil {
		if context.Files.Structure == nil {
			context.Files.Structure = make(map[string]any)
		}
		context.Files.Structure["size_distribution"] = sizes
	}

	return nil
}

// calculateCodeMetrics calculates basic code metrics.
func (pm *ProjectContextManager) calculateCodeMetrics(projectPath string) (map[string]any, error) {
	metrics := map[string]any{
		"total_lines":   0,
		"code_lines":    0,
		"comment_lines": 0,
		"blank_lines":   0,
	}

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Only analyze code files
		ext := filepath.Ext(path)
		codeExts := map[string]bool{
			".go": true, ".js": true, ".ts": true, ".py": true,
			".java": true, ".cpp": true, ".c": true, ".rs": true,
			".rb": true, ".php": true, ".cs": true, ".swift": true,
		}

		if !codeExts[ext] {
			return nil
		}

		if fileMetrics, err := pm.analyzeFileMetrics(path); err == nil {
			metrics["total_lines"] = metrics["total_lines"].(int) + fileMetrics["total_lines"].(int)
			metrics["code_lines"] = metrics["code_lines"].(int) + fileMetrics["code_lines"].(int)
			metrics["comment_lines"] = metrics["comment_lines"].(int) + fileMetrics["comment_lines"].(int)
			metrics["blank_lines"] = metrics["blank_lines"].(int) + fileMetrics["blank_lines"].(int)
		}

		return nil
	})

	return metrics, err
}

// analyzeFileMetrics analyzes metrics for a single file.
func (pm *ProjectContextManager) analyzeFileMetrics(filePath string) (map[string]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	metrics := map[string]any{
		"total_lines":   len(lines),
		"code_lines":    0,
		"comment_lines": 0,
		"blank_lines":   0,
	}

	// Determine comment patterns based on file extension
	ext := filepath.Ext(filePath)
	commentPatterns := pm.getCommentPatterns(ext)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			metrics["blank_lines"] = metrics["blank_lines"].(int) + 1
		} else if pm.isCommentLine(trimmed, commentPatterns) {
			metrics["comment_lines"] = metrics["comment_lines"].(int) + 1
		} else {
			metrics["code_lines"] = metrics["code_lines"].(int) + 1
		}
	}

	return metrics, nil
}

// getCommentPatterns returns comment patterns for different file types.
func (pm *ProjectContextManager) getCommentPatterns(ext string) []string {
	patterns := map[string][]string{
		".go":    {"//", "/*"},
		".js":    {"//", "/*"},
		".ts":    {"//", "/*"},
		".java":  {"//", "/*"},
		".cpp":   {"//", "/*"},
		".c":     {"//", "/*"},
		".cs":    {"//", "/*"},
		".swift": {"//", "/*"},
		".rs":    {"//", "/*"},
		".py":    {"#", `"""`},
		".rb":    {"#", "=begin"},
		".php":   {"//", "/*", "#"},
	}

	if p, exists := patterns[ext]; exists {
		return p
	}
	return []string{"//", "#"}
}

// isCommentLine checks if a line is a comment.
func (pm *ProjectContextManager) isCommentLine(line string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasPrefix(line, pattern) {
			return true
		}
	}
	return false
}

// analyzeFileSizes analyzes file size distribution.
func (pm *ProjectContextManager) analyzeFileSizes(projectPath string) (map[string]any, error) {
	sizes := map[string]any{
		"small":      0, // < 1KB
		"medium":     0, // 1KB - 10KB
		"large":      0, // 10KB - 100KB
		"xlarge":     0, // > 100KB
		"total_size": int64(0),
	}

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		size := info.Size()
		sizes["total_size"] = sizes["total_size"].(int64) + size

		if size < 1024 {
			sizes["small"] = sizes["small"].(int) + 1
		} else if size < 10*1024 {
			sizes["medium"] = sizes["medium"].(int) + 1
		} else if size < 100*1024 {
			sizes["large"] = sizes["large"].(int) + 1
		} else {
			sizes["xlarge"] = sizes["xlarge"].(int) + 1
		}

		return nil
	})

	return sizes, err
}

// initializeAnalysisPatterns initializes the analysis patterns.
func (pm *ProjectContextManager) initializeAnalysisPatterns() {
	// Initialize common analysis patterns
	pm.analysisPatterns["testing"] = &AnalysisPattern{
		Name:         "testing",
		FilePatterns: []string{"*_test.go", "*_test.js", "test_*.py", "*Test.java"},
		Indicators:   []string{"test/", "tests/", "__tests__/"},
		Analyzer:     pm.analyzeTestingPattern,
	}

	pm.analysisPatterns["api"] = &AnalysisPattern{
		Name:         "api",
		FilePatterns: []string{"*api*", "*router*", "*handler*", "*controller*"},
		Indicators:   []string{"api/", "routes/", "handlers/", "controllers/"},
		Analyzer:     pm.analyzeAPIPattern,
	}

	pm.analysisPatterns["database"] = &AnalysisPattern{
		Name:         "database",
		FilePatterns: []string{"*model*", "*schema*", "*migration*", "*db*"},
		Indicators:   []string{"models/", "migrations/", "schema/", "database/"},
		Analyzer:     pm.analyzeDatabasePattern,
	}
}

// analyzeTestingPattern analyzes testing patterns in the project.
func (pm *ProjectContextManager) analyzeTestingPattern(projectPath string) (*PatternResult, error) {
	result := &PatternResult{
		Pattern:    "testing",
		Confidence: 0.0,
		Evidence:   []string{},
		Metadata:   make(map[string]any),
	}

	// Check for test files and directories
	testIndicators := []string{
		"test/", "tests/", "__tests__/", "spec/",
		"*_test.go", "*_test.js", "*_test.ts", "test_*.py",
		"*Test.java", "*Spec.js", "*.spec.ts",
	}

	confidence := 0.0
	evidence := []string{}

	for _, indicator := range testIndicators {
		if strings.Contains(indicator, "*") {
			matches, err := filepath.Glob(filepath.Join(projectPath, "**", indicator))
			if err == nil && len(matches) > 0 {
				confidence += 0.2
				evidence = append(evidence, fmt.Sprintf("Found %d test files matching %s", len(matches), indicator))
			}
		} else {
			if _, err := os.Stat(filepath.Join(projectPath, indicator)); err == nil {
				confidence += 0.3
				evidence = append(evidence, fmt.Sprintf("Found %s", indicator))
			}
		}
	}

	// Check for test frameworks in package.json or similar
	if deps, err := pm.analyzeDependencies(projectPath); err == nil {
		testFrameworks := []string{"jest", "mocha", "chai", "jasmine", "pytest", "junit", "rspec"}
		for _, framework := range testFrameworks {
			if _, exists := deps.Dependencies[framework]; exists {
				confidence += 0.2
				evidence = append(evidence, fmt.Sprintf("Uses %s testing framework", framework))
			}
			if deps.DevDeps != nil {
				if _, exists := deps.DevDeps[framework]; exists {
					confidence += 0.2
					evidence = append(evidence, fmt.Sprintf("Uses %s testing framework (dev)", framework))
				}
			}
		}
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	result.Confidence = confidence
	result.Evidence = evidence
	result.Metadata["frameworks_detected"] = len(evidence)

	return result, nil
}

// analyzeAPIPattern analyzes API patterns in the project.
func (pm *ProjectContextManager) analyzeAPIPattern(projectPath string) (*PatternResult, error) {
	result := &PatternResult{
		Pattern:    "api",
		Confidence: 0.0,
		Evidence:   []string{},
		Metadata:   make(map[string]any),
	}

	// Check for API-related directories and files
	apiIndicators := []string{
		"api/", "routes/", "handlers/", "controllers/",
		"*router*", "*handler*", "*controller*", "*api*",
		"openapi.yaml", "swagger.yaml", "api.yaml",
	}

	confidence := 0.0
	evidence := []string{}

	for _, indicator := range apiIndicators {
		if strings.Contains(indicator, "*") {
			matches, err := filepath.Glob(filepath.Join(projectPath, "**", indicator))
			if err == nil && len(matches) > 0 {
				confidence += 0.15
				evidence = append(evidence, fmt.Sprintf("Found %d API files matching %s", len(matches), indicator))
			}
		} else {
			if _, err := os.Stat(filepath.Join(projectPath, indicator)); err == nil {
				confidence += 0.25
				evidence = append(evidence, fmt.Sprintf("Found %s", indicator))
			}
		}
	}

	// Check for web frameworks
	if deps, err := pm.analyzeDependencies(projectPath); err == nil {
		webFrameworks := []string{
			"express", "fastify", "koa", "hapi", // Node.js
			"gin", "echo", "fiber", "chi", // Go
			"flask", "django", "fastapi", // Python
			"spring-boot", "jersey", // Java
			"actix-web", "warp", "rocket", // Rust
		}

		for _, framework := range webFrameworks {
			if _, exists := deps.Dependencies[framework]; exists {
				confidence += 0.3
				evidence = append(evidence, fmt.Sprintf("Uses %s web framework", framework))
			}
		}
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	result.Confidence = confidence
	result.Evidence = evidence
	result.Metadata["api_indicators"] = len(evidence)

	return result, nil
}

// analyzeDatabasePattern analyzes database patterns in the project.
func (pm *ProjectContextManager) analyzeDatabasePattern(projectPath string) (*PatternResult, error) {
	result := &PatternResult{
		Pattern:    "database",
		Confidence: 0.0,
		Evidence:   []string{},
		Metadata:   make(map[string]any),
	}

	// Check for database-related directories and files
	dbIndicators := []string{
		"models/", "migrations/", "schema/", "database/", "db/",
		"*model*", "*schema*", "*migration*", "*.sql",
	}

	confidence := 0.0
	evidence := []string{}

	for _, indicator := range dbIndicators {
		if strings.Contains(indicator, "*") {
			matches, err := filepath.Glob(filepath.Join(projectPath, "**", indicator))
			if err == nil && len(matches) > 0 {
				confidence += 0.15
				evidence = append(evidence, fmt.Sprintf("Found %d database files matching %s", len(matches), indicator))
			}
		} else {
			if _, err := os.Stat(filepath.Join(projectPath, indicator)); err == nil {
				confidence += 0.2
				evidence = append(evidence, fmt.Sprintf("Found %s", indicator))
			}
		}
	}

	// Check for database drivers and ORMs
	if deps, err := pm.analyzeDependencies(projectPath); err == nil {
		dbLibs := []string{
			"mysql", "postgres", "sqlite", "mongodb", "redis",
			"gorm", "sequelize", "typeorm", "prisma", "mongoose",
			"sqlalchemy", "django-orm", "peewee",
			"hibernate", "mybatis", "jpa",
		}

		for _, lib := range dbLibs {
			if _, exists := deps.Dependencies[lib]; exists {
				confidence += 0.25
				evidence = append(evidence, fmt.Sprintf("Uses %s database library", lib))
			}
		}
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	result.Confidence = confidence
	result.Evidence = evidence
	result.Metadata["database_libs"] = len(evidence)

	return result, nil
}

// InvalidateCache invalidates the cached project context.
func (pm *ProjectContextManager) InvalidateCache() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.cachedContext = nil
	pm.lastCacheUpdate = time.Time{}
}

// SetCacheDuration sets the cache duration for project context.
func (pm *ProjectContextManager) SetCacheDuration(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.cacheDuration = duration
}

// GetCacheInfo returns information about the cache status.
func (pm *ProjectContextManager) GetCacheInfo() map[string]any {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info := map[string]any{
		"cache_duration": pm.cacheDuration.String(),
		"is_cached":      pm.cachedContext != nil,
		"cache_age":      time.Since(pm.lastCacheUpdate).String(),
	}

	if pm.cachedContext != nil {
		info["last_update"] = pm.lastCacheUpdate.Format(time.RFC3339)
		info["cache_valid"] = time.Since(pm.lastCacheUpdate) < pm.cacheDuration
	}

	return info
}
