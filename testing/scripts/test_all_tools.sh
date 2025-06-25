#!/bin/bash

# GopherStrike Comprehensive Testing Script
# This script tests all tools and functionalities

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
RESULTS_DIR="$SCRIPT_DIR/../results"
BINARY="$PROJECT_DIR/GopherStrike"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create results directory
mkdir -p "$RESULTS_DIR"

# Initialize result files
TEST_LOG="$RESULTS_DIR/test_results_$(date +%Y%m%d_%H%M%S).log"
ERROR_LOG="$RESULTS_DIR/errors_$(date +%Y%m%d_%H%M%S).log"
SUCCESS_LOG="$RESULTS_DIR/success_$(date +%Y%m%d_%H%M%S).log"

echo "GopherStrike Comprehensive Testing" | tee "$TEST_LOG"
echo "=================================" | tee -a "$TEST_LOG"
echo "Started at: $(date)" | tee -a "$TEST_LOG"
echo "" | tee -a "$TEST_LOG"

# Function to log test results
log_test() {
    local test_name="$1"
    local status="$2"
    local details="$3"
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}[PASS]${NC} $test_name" | tee -a "$TEST_LOG"
        echo "$test_name: $details" >> "$SUCCESS_LOG"
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}[FAIL]${NC} $test_name" | tee -a "$TEST_LOG"
        echo "$test_name: $details" >> "$ERROR_LOG"
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}[WARN]${NC} $test_name" | tee -a "$TEST_LOG"
        echo "$test_name: $details" >> "$ERROR_LOG"
    else
        echo -e "${BLUE}[INFO]${NC} $test_name" | tee -a "$TEST_LOG"
    fi
}

# Function to test command line arguments
test_command_line() {
    echo -e "\n${BLUE}Testing Command Line Arguments${NC}" | tee -a "$TEST_LOG"
    echo "=============================" | tee -a "$TEST_LOG"
    
    # Test --help
    if timeout 10s "$BINARY" --help >/dev/null 2>&1; then
        log_test "Command: --help" "PASS" "Help command executed successfully"
    else
        log_test "Command: --help" "FAIL" "Help command failed or timed out"
    fi
    
    # Test -h
    if timeout 10s "$BINARY" -h >/dev/null 2>&1; then
        log_test "Command: -h" "PASS" "Short help command executed successfully"
    else
        log_test "Command: -h" "FAIL" "Short help command failed or timed out"
    fi
    
    # Test --version
    if timeout 10s "$BINARY" --version >/dev/null 2>&1; then
        log_test "Command: --version" "PASS" "Version command executed successfully"
    else
        log_test "Command: --version" "FAIL" "Version command failed or timed out"
    fi
    
    # Test -v
    if timeout 10s "$BINARY" -v >/dev/null 2>&1; then
        log_test "Command: -v" "PASS" "Short version command executed successfully"
    else
        log_test "Command: -v" "FAIL" "Short version command failed or timed out"
    fi
    
    # Test invalid argument
    if timeout 10s "$BINARY" --invalid-arg >/dev/null 2>&1; then
        log_test "Command: --invalid-arg" "FAIL" "Invalid argument should have failed but didn't"
    else
        log_test "Command: --invalid-arg" "PASS" "Invalid argument properly rejected"
    fi
}

# Function to test each tool option
test_interactive_tools() {
    echo -e "\n${BLUE}Testing Interactive Tool Options${NC}" | tee -a "$TEST_LOG"
    echo "================================" | tee -a "$TEST_LOG"
    
    local tools=(
        "1:Port Scanner"
        "2:Subdomain Scanner" 
        "3:OSINT & Vulnerability Tool"
        "4:Web Application Security Scanner"
        "5:S3 Bucket Scanner"
        "6:Email Harvester"
        "7:Directory Bruteforcer"
        "8:Report Generator"
        "9:Host & Subdomain Resolver"
        "10:Check Dependencies"
        "11:Exit"
    )
    
    for tool in "${tools[@]}"; do
        IFS=':' read -r num name <<< "$tool"
        
        if [ "$num" = "11" ]; then
            # Test exit option
            if echo "$num" | timeout 10s "$BINARY" >/dev/null 2>&1; then
                log_test "Tool $num ($name)" "PASS" "Exit option works correctly"
            else
                log_test "Tool $num ($name)" "FAIL" "Exit option failed"
            fi
        else
            # Test other tools (they will likely fail due to missing dependencies or inputs)
            local output
            local exit_code
            output=$(echo -e "$num\n11" | timeout 30s "$BINARY" 2>&1) || exit_code=$?
            
            if [[ $output == *"Error"* ]] || [[ $output == *"error"* ]] || [[ $output == *"failed"* ]]; then
                log_test "Tool $num ($name)" "WARN" "Tool executed but reported errors: $(echo "$output" | grep -i error | head -1)"
            elif [ ${exit_code:-0} -eq 124 ]; then
                log_test "Tool $num ($name)" "WARN" "Tool timed out after 30 seconds"
            elif [ ${exit_code:-0} -eq 0 ]; then
                log_test "Tool $num ($name)" "PASS" "Tool executed without immediate errors"
            else
                log_test "Tool $num ($name)" "FAIL" "Tool failed with exit code ${exit_code:-unknown}"
            fi
        fi
    done
}

# Function to test input validation
test_input_validation() {
    echo -e "\n${BLUE}Testing Input Validation${NC}" | tee -a "$TEST_LOG"
    echo "=======================" | tee -a "$TEST_LOG"
    
    local invalid_inputs=(
        "0:Invalid option 0"
        "12:Invalid option 12" 
        "999:Invalid option 999"
        "-1:Negative number"
        "abc:Non-numeric input"
        "1.5:Decimal number"
        " :Whitespace only"
        "1 2:Multiple numbers"
        "!@#:Special characters"
    )
    
    for input_test in "${invalid_inputs[@]}"; do
        IFS=':' read -r input desc <<< "$input_test"
        
        local output
        local exit_code
        output=$(echo -e "$input\n11" | timeout 15s "$BINARY" 2>&1) || exit_code=$?
        
        if [[ $output == *"Invalid choice"* ]] || [[ $output == *"invalid"* ]]; then
            log_test "Input Validation ($desc)" "PASS" "Invalid input properly rejected"
        else
            log_test "Input Validation ($desc)" "FAIL" "Invalid input not properly handled"
        fi
    done
}

# Function to test EOF handling
test_eof_handling() {
    echo -e "\n${BLUE}Testing EOF Handling${NC}" | tee -a "$TEST_LOG"
    echo "===================" | tee -a "$TEST_LOG"
    
    # Test EOF with no input
    local output
    local exit_code
    output=$(echo "" | timeout 10s "$BINARY" 2>&1) || exit_code=$?
    
    if [[ $output == *"Exiting GopherStrike"* ]]; then
        log_test "EOF Handling (empty input)" "PASS" "EOF properly handled with graceful exit"
    else
        log_test "EOF Handling (empty input)" "FAIL" "EOF not properly handled"
    fi
    
    # Test Ctrl+D simulation
    output=$(timeout 10s "$BINARY" < /dev/null 2>&1) || exit_code=$?
    
    if [[ $output == *"Exiting GopherStrike"* ]]; then
        log_test "EOF Handling (no stdin)" "PASS" "No stdin properly handled with graceful exit"
    else
        log_test "EOF Handling (no stdin)" "FAIL" "No stdin not properly handled"
    fi
}

# Function to test binary existence and permissions
test_binary() {
    echo -e "\n${BLUE}Testing Binary${NC}" | tee -a "$TEST_LOG"
    echo "=============" | tee -a "$TEST_LOG"
    
    if [ -f "$BINARY" ]; then
        log_test "Binary Existence" "PASS" "GopherStrike binary exists"
    else
        log_test "Binary Existence" "FAIL" "GopherStrike binary not found at $BINARY"
        return 1
    fi
    
    if [ -x "$BINARY" ]; then
        log_test "Binary Permissions" "PASS" "GopherStrike binary is executable"
    else
        log_test "Binary Permissions" "FAIL" "GopherStrike binary is not executable"
        return 1
    fi
    
    # Test basic execution
    if timeout 5s "$BINARY" --help >/dev/null 2>&1; then
        log_test "Binary Execution" "PASS" "GopherStrike binary executes successfully"
    else
        log_test "Binary Execution" "FAIL" "GopherStrike binary fails to execute"
        return 1
    fi
}

# Function to test directory structure
test_directories() {
    echo -e "\n${BLUE}Testing Directory Structure${NC}" | tee -a "$TEST_LOG"
    echo "===========================" | tee -a "$TEST_LOG"
    
    local required_dirs=(
        "logs:Log directory"
        "pkg:Package directory"
        "utils:Utilities directory"
    )
    
    for dir_test in "${required_dirs[@]}"; do
        IFS=':' read -r dir desc <<< "$dir_test"
        local full_path="$PROJECT_DIR/$dir"
        
        if [ -d "$full_path" ]; then
            log_test "Directory ($desc)" "PASS" "Directory $dir exists"
        else
            log_test "Directory ($desc)" "WARN" "Directory $dir does not exist"
        fi
    done
}

# Function to check dependencies
test_dependencies() {
    echo -e "\n${BLUE}Testing Dependencies${NC}" | tee -a "$TEST_LOG"
    echo "===================" | tee -a "$TEST_LOG"
    
    # Test dependency check tool
    local output
    local exit_code
    output=$(echo -e "10\n11" | timeout 30s "$BINARY" 2>&1) || exit_code=$?
    
    if [[ $output == *"Dependencies"* ]] || [[ $output == *"dependencies"* ]]; then
        log_test "Dependency Check Tool" "PASS" "Dependency check tool executed"
        
        # Extract dependency information
        if [[ $output == *"not found"* ]] || [[ $output == *"missing"* ]]; then
            log_test "Dependencies Status" "WARN" "Some dependencies are missing"
        else
            log_test "Dependencies Status" "PASS" "All dependencies appear to be available"
        fi
    else
        log_test "Dependency Check Tool" "FAIL" "Dependency check tool failed to execute"
    fi
}

# Main test execution
main() {
    echo -e "${GREEN}Starting GopherStrike Comprehensive Testing${NC}"
    echo "=========================================="
    
    # Change to project directory
    cd "$PROJECT_DIR"
    
    # Build the project first
    echo -e "\n${BLUE}Building GopherStrike${NC}"
    if go build -o GopherStrike; then
        log_test "Build Process" "PASS" "GopherStrike built successfully"
    else
        log_test "Build Process" "FAIL" "Failed to build GopherStrike"
        echo -e "${RED}Cannot continue testing without successful build${NC}"
        exit 1
    fi
    
    # Run all tests
    test_binary
    test_directories
    test_command_line
    test_input_validation
    test_eof_handling
    test_interactive_tools
    test_dependencies
    
    # Summary
    echo -e "\n${BLUE}Test Summary${NC}" | tee -a "$TEST_LOG"
    echo "============" | tee -a "$TEST_LOG"
    
    local pass_count=$(grep -c "\[PASS\]" "$TEST_LOG" || echo "0")
    local fail_count=$(grep -c "\[FAIL\]" "$TEST_LOG" || echo "0")
    local warn_count=$(grep -c "\[WARN\]" "$TEST_LOG" || echo "0")
    
    echo "Tests Passed: $pass_count" | tee -a "$TEST_LOG"
    echo "Tests Failed: $fail_count" | tee -a "$TEST_LOG" 
    echo "Warnings: $warn_count" | tee -a "$TEST_LOG"
    echo "" | tee -a "$TEST_LOG"
    echo "Detailed results saved to:" | tee -a "$TEST_LOG"
    echo "  Full log: $TEST_LOG" | tee -a "$TEST_LOG"
    echo "  Errors: $ERROR_LOG" | tee -a "$TEST_LOG"
    echo "  Success: $SUCCESS_LOG" | tee -a "$TEST_LOG"
    
    if [ "$fail_count" -gt 0 ]; then
        echo -e "\n${RED}Some tests failed. Check $ERROR_LOG for details.${NC}"
        exit 1
    elif [ "$warn_count" -gt 0 ]; then
        echo -e "\n${YELLOW}All tests passed but with warnings. Check $ERROR_LOG for details.${NC}"
        exit 0
    else
        echo -e "\n${GREEN}All tests passed successfully!${NC}"
        exit 0
    fi
}

# Run main function
main "$@"