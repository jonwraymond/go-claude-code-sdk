# Update Go Import Statements - Module Path Migration

## Task Overview
Update all Go import statements from old module path to new GitHub repository path:
- **Old**: `github.com/jraymond/claude-code-go-sdk`
- **New**: `github.com/jonwraymond/go-claude-code-sdk`

## Todo List

### Phase 1: Examples Directory (5 files)
- [ ] Update examples/basic_usage/main.go
- [ ] Update examples/streaming/main.go
- [ ] Update examples/anthropic_api/main.go
- [ ] Update examples/file_operations/main.go
- [ ] Update examples/batch_operations/main.go

### Phase 2: Integration Tests Directory (6 files)
- [ ] Update tests/integration/client_test.go
- [ ] Update tests/integration/auth_test.go
- [ ] Update tests/integration/streaming_test.go
- [ ] Update tests/integration/file_operations_test.go
- [ ] Update tests/integration/batch_operations_test.go
- [ ] Update tests/integration/anthropic_api_test.go

### Phase 3: Client Package Directory (13 files)
- [ ] Update pkg/client/client.go
- [ ] Update pkg/client/streaming.go
- [ ] Update pkg/client/file_operations.go
- [ ] Update pkg/client/batch_operations.go
- [ ] Update pkg/client/anthropic_api.go
- [ ] Update pkg/client/types.go
- [ ] Update pkg/client/client_test.go
- [ ] Update pkg/client/streaming_test.go
- [ ] Update pkg/client/file_operations_test.go
- [ ] Update pkg/client/batch_operations_test.go
- [ ] Update pkg/client/anthropic_api_test.go
- [ ] Update pkg/client/types_test.go
- [ ] Update pkg/client/utils.go

### Phase 4: Auth Package Directory (4 files)
- [ ] Update pkg/auth/auth.go
- [ ] Update pkg/auth/types.go
- [ ] Update pkg/auth/auth_test.go
- [ ] Update pkg/auth/utils.go

## Execution Strategy
- Use Edit tool with replace_all=true for efficient updates
- Update all occurrences of the old module path in each file
- Verify each file has been properly updated

## Success Criteria
- All 28 Go files updated with new module path
- No references to old module path remain
- Code compiles successfully with new imports