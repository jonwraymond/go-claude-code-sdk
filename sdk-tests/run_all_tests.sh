#!/bin/bash

# Run all SDK tests
echo "Running Go Claude Code SDK Test Suite"
echo "===================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test counter
TOTAL=0
PASSED=0

# Function to run a test
run_test() {
    local test_name=$1
    local test_file=$2
    
    echo "Running $test_name..."
    echo "----------------------------------------"
    
    TOTAL=$((TOTAL + 1))
    
    if go run "$test_file" > "/tmp/${test_file}.log" 2>&1; then
        echo -e "${GREEN}✅ PASSED${NC}: $test_name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}❌ FAILED${NC}: $test_name"
        echo "Error output:"
        tail -n 20 "/tmp/${test_file}.log"
    fi
    
    echo ""
}

# Run all tests
run_test "Basic Initialization" "test_basic_init.go"
run_test "Query Functionality (Simple)" "test_query_simple.go"
run_test "Session Management" "test_sessions.go"
run_test "Command Execution" "test_commands.go"
run_test "MCP Integration" "test_mcp.go"
run_test "MCP Common Servers" "test_mcp_common.go"
run_test "Error Handling" "test_error_handling.go"

# Summary
echo "===================================="
echo "Test Summary"
echo "===================================="
echo "Total Tests: $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$((TOTAL - PASSED))${NC}"

if [ $PASSED -eq $TOTAL ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed!${NC}"
    exit 1
fi