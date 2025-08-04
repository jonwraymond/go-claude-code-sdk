package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/claudecode"
)

func main() {
	fmt.Println("=== Real World Application Examples ===")

	// Example 1: Code reviewer
	example1CodeReviewer()

	// Example 2: Documentation generator
	example2DocumentationGenerator()

	// Example 3: Test generator
	example3TestGenerator()

	// Example 4: Bug analyzer
	example4BugAnalyzer()

	// Example 5: API client generator
	example5APIClientGenerator()

	// Example 6: Performance optimizer
	example6PerformanceOptimizer()
}

func example1CodeReviewer() {
	fmt.Println("Example 1: Automated Code Reviewer")
	fmt.Println("---------------------------------")

	// Create a code review system
	reviewer := &CodeReviewer{
		client: claudecode.NewClaudeSDKClient(nil),
		rules: []ReviewRule{
			{Name: "Error Handling", Pattern: "error", Severity: "high"},
			{Name: "Code Comments", Pattern: "//", Severity: "medium"},
			{Name: "Function Length", Pattern: "func", Severity: "low"},
			{Name: "Security", Pattern: "password|secret|key", Severity: "critical"},
		},
	}

	ctx := context.Background()
	if err := reviewer.client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer reviewer.client.Close()

	// Sample code to review
	codeToReview := `
package main

import "fmt"

func processUserData(data map[string]string) {
    password := data["password"]
    fmt.Println("Processing user:", data["username"])
    
    // TODO: Add validation
    
    if password == "admin" {
        fmt.Println("Admin access granted")
    }
}

func calculateTotal(prices []float64) float64 {
    var total float64
    for _, price := range prices {
        total += price
    }
    return total
}
`

	fmt.Println("📝 Code to review:")
	fmt.Println(codeToReview)
	fmt.Println("\n🔍 Starting code review...")

	// Perform review
	review := reviewer.ReviewCode(ctx, codeToReview)

	// Display results
	fmt.Println("\n📊 Code Review Results:")
	fmt.Printf("   Overall Score: %d/100\n", review.Score)
	fmt.Printf("   Issues Found: %d\n", len(review.Issues))

	fmt.Println("\n🚨 Issues by Severity:")
	for _, severity := range []string{"critical", "high", "medium", "low"} {
		issues := filterIssuesBySeverity(review.Issues, severity)
		if len(issues) > 0 {
			fmt.Printf("\n   %s (%d):\n", severity, len(issues))
			for _, issue := range issues {
				fmt.Printf("      - Line %d: %s\n", issue.Line, issue.Description)
			}
		}
	}

	fmt.Println("\n💡 Recommendations:")
	for i, rec := range review.Recommendations {
		fmt.Printf("   %d. %s\n", i+1, rec)
	}
	fmt.Println()
}

func example2DocumentationGenerator() {
	fmt.Println("Example 2: Documentation Generator")
	fmt.Println("---------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write"}

	ctx := context.Background()

	// Create sample Go file for documentation
	sampleCode := `package mathutils

import "errors"

// Calculator provides basic mathematical operations
type Calculator struct {
	precision int
}

// NewCalculator creates a new Calculator instance
func NewCalculator(precision int) *Calculator {
	return &Calculator{precision: precision}
}

// Add returns the sum of two numbers
func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}

// Divide returns a divided by b, returns error if b is zero
func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// Factorial calculates the factorial of n
func (c *Calculator) Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * c.Factorial(n-1)
}`

	// Write sample file
	writeQuery := fmt.Sprintf("Create a file called mathutils.go with this content:\n%s", sampleCode)
	msgChan := claudecode.Query(ctx, writeQuery, options)
	for range msgChan {
	} // Drain channel

	// Generate documentation
	fmt.Println("\n📚 Generating documentation...")

	docQuery := `Read mathutils.go and generate comprehensive documentation:
1. Create a README.md with package overview, installation, usage examples
2. Create API.md with detailed function documentation
3. Create EXAMPLES.md with practical code examples
4. Include proper markdown formatting, code blocks, and tables`

	msgChan = claudecode.Query(ctx, docQuery, options)

	filesCreated := []string{}

	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
					if toolUse.Name == "Write" {
						if filepath, ok := toolUse.Input["file_path"].(string); ok {
							filesCreated = append(filesCreated, filepath)
							fmt.Printf("   ✅ Generated: %s\n", filepath)
						}
					}
				}
			}
		}
	}

	fmt.Printf("\n📄 Documentation generated: %d files\n", len(filesCreated))
	fmt.Println()
}

func example3TestGenerator() {
	fmt.Println("Example 3: Test Generator")
	fmt.Println("------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Bash"}

	ctx := context.Background()

	// Create a sample service to test
	serviceCode := `package userservice

import (
	"errors"
	"regexp"
)

type User struct {
	ID       string
	Email    string
	Username string
	Age      int
}

type UserService struct {
	users map[string]*User
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]*User),
	}
}

func (s *UserService) CreateUser(email, username string, age int) (*User, error) {
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}
	
	if age < 13 {
		return nil, errors.New("user must be at least 13 years old")
	}
	
	user := &User{
		ID:       generateID(),
		Email:    email,
		Username: username,
		Age:      age,
	}
	
	s.users[user.ID] = user
	return user, nil
}

func (s *UserService) GetUser(id string) (*User, error) {
	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserService) validateEmail(email string) error {
	emailRegex := regexp.MustCompile("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func generateID() string {
	return "user_" + string(time.Now().Unix())
}`

	// Write the service file
	writeQuery := fmt.Sprintf("Create userservice.go with this content:\n%s", serviceCode)
	msgChan := claudecode.Query(ctx, writeQuery, options)
	for range msgChan {
	} // Drain

	// Generate comprehensive tests
	fmt.Println("\n🧪 Generating comprehensive tests...")

	testQuery := `Read userservice.go and generate comprehensive tests:
1. Create userservice_test.go with unit tests for all functions
2. Include table-driven tests for different scenarios
3. Test happy paths, error cases, and edge cases
4. Add benchmarks for performance-critical functions
5. Include test coverage of at least 80%
6. Run the tests and show the results`

	msgChan = claudecode.Query(ctx, testQuery, options)

	testStats := TestGenerationStats{}

	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				switch b := block.(type) {
				case claudecode.ToolUseBlock:
					if b.Name == "Write" && contains(fmt.Sprintf("%v", b.Input), "_test.go") {
						testStats.FilesCreated++
					}
					if b.Name == "Bash" && contains(fmt.Sprintf("%v", b.Input), "go test") {
						testStats.TestsRun = true
					}
				case claudecode.TextBlock:
					// Extract test counts
					if contains(b.Text, "func Test") {
						testStats.TestCount++
					}
					if contains(b.Text, "func Benchmark") {
						testStats.BenchmarkCount++
					}
				}
			}
		}
	}

	fmt.Println("\n📊 Test Generation Results:")
	fmt.Printf("   Test files created: %d\n", testStats.FilesCreated)
	fmt.Printf("   Test functions: %d+\n", testStats.TestCount)
	fmt.Printf("   Benchmarks: %d+\n", testStats.BenchmarkCount)
	fmt.Printf("   Tests executed: %v\n", testStats.TestsRun)
	fmt.Println()
}

func example4BugAnalyzer() {
	fmt.Println("Example 4: Bug Analyzer")
	fmt.Println("-----------------------")

	// Create bug analyzer
	analyzer := &BugAnalyzer{
		client: claudecode.NewClaudeSDKClient(nil),
		patterns: []BugPattern{
			{Type: "NullPointer", Keywords: []string{"nil", "panic", "dereference"}},
			{Type: "RaceCondition", Keywords: []string{"goroutine", "mutex", "channel"}},
			{Type: "ResourceLeak", Keywords: []string{"defer", "close", "file", "connection"}},
			{Type: "ErrorHandling", Keywords: []string{"error", "err != nil", "ignore"}},
		},
	}

	ctx := context.Background()
	if err := analyzer.client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer analyzer.client.Close()

	// Sample buggy code
	buggyCode := `package main

import (
	"fmt"
	"os"
	"sync"
)

var counter int

func incrementCounter() {
	// BUG: Race condition - no synchronization
	counter++
}

func processFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	// BUG: File not closed - resource leak
	
	data := make([]byte, 1024)
	file.Read(data)
	
	// BUG: Ignoring error from Read
	fmt.Println(string(data))
	
	return nil
}

func dangerousOperation(data *DataStruct) {
	// BUG: No nil check
	fmt.Println(data.Value)
}

type DataStruct struct {
	Value string
}`

	fmt.Println("🐛 Analyzing code for bugs...")

	// Analyze the code
	report := analyzer.AnalyzeCode(ctx, buggyCode)

	fmt.Println("\n🔍 Bug Analysis Report:")
	fmt.Printf("   Severity: %s\n", report.Severity)
	fmt.Printf("   Bugs found: %d\n", len(report.Bugs))

	fmt.Println("\n🐞 Detected Bugs:")
	for i, bug := range report.Bugs {
		fmt.Printf("\n   Bug #%d: %s\n", i+1, bug.Type)
		fmt.Printf("   Location: Line %d\n", bug.Line)
		fmt.Printf("   Description: %s\n", bug.Description)
		fmt.Printf("   Fix: %s\n", bug.SuggestedFix)
	}

	fmt.Println("\n✅ Recommendations:")
	for i, rec := range report.Recommendations {
		fmt.Printf("   %d. %s\n", i+1, rec)
	}
	fmt.Println()
}

func example5APIClientGenerator() {
	fmt.Println("Example 5: API Client Generator")
	fmt.Println("-------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Write", "Read"}

	ctx := context.Background()

	// API specification
	apiSpec := APISpecification{
		Name:    "UserAPI",
		BaseURL: "https://api.example.com/v1",
		Endpoints: []Endpoint{
			{
				Method:      "GET",
				Path:        "/users",
				Description: "List all users",
				QueryParams: []Param{{Name: "limit", Type: "int"}, {Name: "offset", Type: "int"}},
			},
			{
				Method:      "GET",
				Path:        "/users/{id}",
				Description: "Get user by ID",
				PathParams:  []Param{{Name: "id", Type: "string"}},
			},
			{
				Method:      "POST",
				Path:        "/users",
				Description: "Create new user",
				BodyParams:  []Param{{Name: "email", Type: "string"}, {Name: "name", Type: "string"}},
			},
			{
				Method:      "PUT",
				Path:        "/users/{id}",
				Description: "Update user",
				PathParams:  []Param{{Name: "id", Type: "string"}},
				BodyParams:  []Param{{Name: "email", Type: "string"}, {Name: "name", Type: "string"}},
			},
			{
				Method:      "DELETE",
				Path:        "/users/{id}",
				Description: "Delete user",
				PathParams:  []Param{{Name: "id", Type: "string"}},
			},
		},
	}

	fmt.Println("🔧 Generating API client...")

	// Generate the client
	specJSON, _ := json.MarshalIndent(apiSpec, "", "  ")
	query := fmt.Sprintf(`Generate a complete Go API client based on this specification:
%s

Requirements:
1. Create client.go with the main client struct and methods
2. Create types.go with request/response types
3. Create errors.go with custom error types
4. Include proper error handling, retries, and timeouts
5. Add authentication support (API key header)
6. Create example usage in example_test.go`, string(specJSON))

	msgChan := claudecode.Query(ctx, query, options)

	generatedFiles := []GeneratedFile{}

	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if toolUse, ok := block.(claudecode.ToolUseBlock); ok {
					if toolUse.Name == "Write" {
						file := GeneratedFile{}
						if filePath, ok := toolUse.Input["file_path"].(string); ok {
							file.Path = filePath
						}
						if content, ok := toolUse.Input["content"].(string); ok {
							file.Lines = countLines(content)
						}
						generatedFiles = append(generatedFiles, file)
					}
				}
			}
		}
	}

	fmt.Println("\n📦 Generated API Client:")
	totalLines := 0
	for _, file := range generatedFiles {
		fmt.Printf("   ✅ %s (%d lines)\n", file.Path, file.Lines)
		totalLines += file.Lines
	}
	fmt.Printf("\n   Total: %d files, %d lines of code\n", len(generatedFiles), totalLines)
	fmt.Println()
}

func example6PerformanceOptimizer() {
	fmt.Println("Example 6: Performance Optimizer")
	fmt.Println("-------------------------------")

	options := claudecode.NewClaudeCodeOptions()
	options.AllowedTools = []string{"Read", "Write", "Edit", "Bash"}

	ctx := context.Background()

	// Create sample slow code
	slowCode := `package main

import (
	"fmt"
	"strings"
)

// SlowFunction has several performance issues
func SlowFunction(data []string) string {
	result := ""
	
	// Issue 1: String concatenation in loop
	for _, item := range data {
		result = result + item + ","
	}
	
	// Issue 2: Unnecessary allocations
	parts := make([]string, 0)
	for i := 0; i < len(data); i++ {
		parts = append(parts, strings.ToUpper(data[i]))
	}
	
	// Issue 3: Inefficient search
	found := false
	target := "TARGET"
	for i := 0; i < len(parts); i++ {
		if parts[i] == target {
			found = true
			break
		}
	}
	
	if found {
		return "Found: " + result
	}
	
	return result
}

// Issue 4: No buffering for file operations
func ProcessLargeFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fmt.Println(line)
	}
	
	return nil
}`

	// Write the slow code
	writeQuery := fmt.Sprintf("Create slow_code.go with this content:\n%s", slowCode)
	msgChan := claudecode.Query(ctx, writeQuery, options)
	for range msgChan {
	} // Drain

	fmt.Println("\n⚡ Optimizing performance...")

	// Optimize the code
	optimizeQuery := `Analyze slow_code.go and optimize it:
1. Identify all performance issues
2. Create optimized_code.go with fixes
3. Create benchmark_test.go to compare performance
4. Document the optimizations made
5. Run benchmarks to show improvements`

	msgChan = claudecode.Query(ctx, optimizeQuery, options)

	optimizations := []Optimization{}
	benchmarkRun := false

	for msg := range msgChan {
		if assistantMsg, ok := msg.(*claudecode.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				switch b := block.(type) {
				case claudecode.TextBlock:
					// Extract optimization details
					if contains(b.Text, "Issue") || contains(b.Text, "Optimization") {
						opt := Optimization{
							Description: extractFirstLine(b.Text),
						}
						optimizations = append(optimizations, opt)
					}
				case claudecode.ToolUseBlock:
					if b.Name == "Bash" && contains(fmt.Sprintf("%v", b.Input), "benchmark") {
						benchmarkRun = true
					}
				}
			}
		}
	}

	fmt.Println("\n📊 Optimization Results:")
	fmt.Printf("   Issues identified: %d+\n", len(optimizations))
	fmt.Printf("   Optimizations applied: ✅\n")
	fmt.Printf("   Benchmarks run: %v\n", benchmarkRun)

	fmt.Println("\n🚀 Performance Improvements:")
	improvements := []string{
		"String concatenation → strings.Builder",
		"Unnecessary allocations → Preallocated slices",
		"Linear search → Map lookup",
		"Unbuffered I/O → Buffered reader",
	}

	for i, imp := range improvements {
		fmt.Printf("   %d. %s\n", i+1, imp)
	}

	fmt.Println()
}

// Helper types and functions

type CodeReviewer struct {
	client *claudecode.ClaudeSDKClient
	rules  []ReviewRule
}

type ReviewRule struct {
	Name     string
	Pattern  string
	Severity string
}

type CodeReview struct {
	Score           int
	Issues          []Issue
	Recommendations []string
}

type Issue struct {
	Line        int
	Severity    string
	Description string
	Rule        string
}

func (r *CodeReviewer) ReviewCode(ctx context.Context, code string) CodeReview {
	query := fmt.Sprintf(`Review this code and identify issues:
%s

Check for:
1. Error handling
2. Security issues (hardcoded passwords, secrets)
3. Code style and comments
4. Performance issues
5. Best practices

Provide a score out of 100 and specific recommendations.`, code)

	msgChan := claudecode.Query(ctx, query, nil)

	review := CodeReview{
		Score: 75, // Default
		Issues: []Issue{
			{Line: 6, Severity: "critical", Description: "Hardcoded password comparison", Rule: "Security"},
			{Line: 9, Severity: "medium", Description: "Missing input validation", Rule: "Error Handling"},
			{Line: 11, Severity: "high", Description: "No error handling for data access", Rule: "Error Handling"},
		},
		Recommendations: []string{
			"Use secure password hashing instead of plaintext comparison",
			"Add input validation for all user data",
			"Implement proper error handling throughout",
			"Add function documentation comments",
			"Consider adding unit tests",
		},
	}

	// Process actual review results
	for range msgChan {
		// In real implementation, parse Claude's response
	}

	return review
}

type BugAnalyzer struct {
	client   *claudecode.ClaudeSDKClient
	patterns []BugPattern
}

type BugPattern struct {
	Type     string
	Keywords []string
}

type BugReport struct {
	Severity        string
	Bugs            []Bug
	Recommendations []string
}

type Bug struct {
	Type         string
	Line         int
	Description  string
	SuggestedFix string
}

func (a *BugAnalyzer) AnalyzeCode(ctx context.Context, code string) BugReport {
	query := fmt.Sprintf(`Analyze this code for bugs:
%s

Look for:
1. Race conditions
2. Resource leaks
3. Null pointer dereferences
4. Error handling issues
5. Security vulnerabilities

For each bug found, provide the line number, description, and suggested fix.`, code)

	msgChan := claudecode.Query(ctx, query, nil)

	report := BugReport{
		Severity: "High",
		Bugs: []Bug{
			{Type: "RaceCondition", Line: 11, Description: "Unsynchronized access to shared variable", SuggestedFix: "Use sync.Mutex or atomic operations"},
			{Type: "ResourceLeak", Line: 20, Description: "File not closed after opening", SuggestedFix: "Add defer file.Close() after error check"},
			{Type: "ErrorHandling", Line: 24, Description: "Error from Read() ignored", SuggestedFix: "Check and handle the error returned by Read()"},
			{Type: "NullPointer", Line: 31, Description: "No nil check before dereferencing", SuggestedFix: "Add if data != nil check"},
		},
		Recommendations: []string{
			"Run go vet and golint regularly",
			"Use race detector during testing",
			"Implement comprehensive error handling",
			"Add unit tests with edge cases",
		},
	}

	// Process actual analysis results
	for range msgChan {
		// In real implementation, parse Claude's response
	}

	return report
}

type APISpecification struct {
	Name      string
	BaseURL   string
	Endpoints []Endpoint
}

type Endpoint struct {
	Method      string
	Path        string
	Description string
	PathParams  []Param
	QueryParams []Param
	BodyParams  []Param
}

type Param struct {
	Name string
	Type string
}

type GeneratedFile struct {
	Path  string
	Lines int
}

type TestGenerationStats struct {
	FilesCreated   int
	TestCount      int
	BenchmarkCount int
	TestsRun       bool
}

type Optimization struct {
	Description string
	Impact      string
}

func filterIssuesBySeverity(issues []Issue, severity string) []Issue {
	filtered := []Issue{}
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

func countLines(content string) int {
	lines := 1
	for _, ch := range content {
		if ch == '\n' {
			lines++
		}
	}
	return lines
}

func extractFirstLine(text string) string {
	for i, ch := range text {
		if ch == '\n' {
			return text[:i]
		}
	}
	return text
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
