#!/bin/bash

# Detailed Tool Testing Script for GopherStrike
# Tests each tool individually and captures detailed error information

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
DETAILED_LOG="$RESULTS_DIR/detailed_tool_test_$(date +%Y%m%d_%H%M%S).log"
ERROR_SUMMARY="$RESULTS_DIR/error_summary_$(date +%Y%m%d_%H%M%S).md"

echo "GopherStrike Detailed Tool Testing" | tee "$DETAILED_LOG"
echo "==================================" | tee -a "$DETAILED_LOG"
echo "Started at: $(date)" | tee -a "$DETAILED_LOG"
echo "" | tee -a "$DETAILED_LOG"

# Initialize error summary
cat > "$ERROR_SUMMARY" << 'EOF'
# GopherStrike Error Analysis Report

Generated on: $(date)

## Summary

This report contains detailed analysis of errors found in GopherStrike tools.

## Issues Found

EOF

# Function to test individual tool
test_tool() {
    local tool_number="$1"
    local tool_name="$2"
    local additional_input="$3"
    
    echo -e "\n${BLUE}Testing Tool $tool_number: $tool_name${NC}" | tee -a "$DETAILED_LOG"
    echo "$(printf '=%.0s' {1..50})" | tee -a "$DETAILED_LOG"
    
    local test_input="$tool_number\n"
    if [ -n "$additional_input" ]; then
        test_input="${test_input}${additional_input}\n"
    fi
    test_input="${test_input}11\n"  # Always end with exit
    
    local output_file="$RESULTS_DIR/tool_${tool_number}_output.txt"
    local error_file="$RESULTS_DIR/tool_${tool_number}_error.txt"
    
    # Run the tool and capture output
    local exit_code=0
    timeout 30s bash -c "echo -e '$test_input' | '$BINARY'" > "$output_file" 2> "$error_file" || exit_code=$?
    
    # Analyze output
    local has_error=false
    local error_description=""
    
    if [ $exit_code -eq 124 ]; then
        error_description="Tool timed out after 30 seconds"
        has_error=true
    elif [ $exit_code -ne 0 ]; then
        error_description="Tool exited with code $exit_code"
        has_error=true
    fi
    
    # Check for specific error patterns
    if grep -q -i "error\|failed\|exception\|panic\|fatal" "$output_file" "$error_file" 2>/dev/null; then
        if [ "$has_error" = false ]; then
            error_description="Tool reported errors in output"
        else
            error_description="$error_description and reported errors in output"
        fi
        has_error=true
    fi
    
    # Check for EOF issues
    if grep -q "EOF" "$output_file" "$error_file" 2>/dev/null; then
        if [ "$has_error" = false ]; then
            error_description="Tool has EOF handling issues"
        else
            error_description="$error_description and has EOF handling issues"
        fi
        has_error=true
    fi
    
    # Check for sudo/permission issues
    if grep -q -i "sudo\|password\|permission" "$output_file" "$error_file" 2>/dev/null; then
        if [ "$has_error" = false ]; then
            error_description="Tool has permission/sudo issues"
        else
            error_description="$error_description and has permission issues"
        fi
        has_error=true
    fi
    
    # Report results
    if [ "$has_error" = true ]; then
        echo -e "${RED}[FAIL]${NC} $tool_name: $error_description" | tee -a "$DETAILED_LOG"
        
        # Add to error summary
        cat >> "$ERROR_SUMMARY" << EOF

### Tool $tool_number: $tool_name

**Status:** FAILED  
**Issue:** $error_description

**Output:**
\`\`\`
$(head -20 "$output_file" 2>/dev/null || echo "No output")
\`\`\`

**Errors:**
\`\`\`
$(head -20 "$error_file" 2>/dev/null || echo "No errors")
\`\`\`

EOF
    else
        echo -e "${GREEN}[PASS]${NC} $tool_name: Tool executed successfully" | tee -a "$DETAILED_LOG"
    fi
    
    # Show some output for debugging
    echo "Exit code: $exit_code" | tee -a "$DETAILED_LOG"
    echo "Output preview:" | tee -a "$DETAILED_LOG"
    head -5 "$output_file" 2>/dev/null | sed 's/^/  /' | tee -a "$DETAILED_LOG"
    if [ -s "$error_file" ]; then
        echo "Error preview:" | tee -a "$DETAILED_LOG"
        head -5 "$error_file" 2>/dev/null | sed 's/^/  /' | tee -a "$DETAILED_LOG"
    fi
}

# Function to test command line arguments
test_command_args() {
    echo -e "\n${BLUE}Testing Command Line Arguments${NC}" | tee -a "$DETAILED_LOG"
    echo "=============================" | tee -a "$DETAILED_LOG"
    
    # Test invalid argument handling
    local output
    local exit_code
    output=$(timeout 5s "$BINARY" --invalid-arg 2>&1) || exit_code=$?
    
    if [ ${exit_code:-0} -eq 0 ]; then
        echo -e "${RED}[FAIL]${NC} Invalid argument acceptance: --invalid-arg was accepted when it should be rejected" | tee -a "$DETAILED_LOG"
        
        cat >> "$ERROR_SUMMARY" << EOF

### Command Line Argument Handling

**Status:** FAILED  
**Issue:** Invalid arguments are being accepted instead of rejected

**Details:**
- Command: \`--invalid-arg\`
- Expected: Error message and non-zero exit code
- Actual: Exit code 0 (success)

EOF
    else
        echo -e "${GREEN}[PASS]${NC} Invalid argument rejection: --invalid-arg properly rejected" | tee -a "$DETAILED_LOG"
    fi
}

# Main testing function
main() {
    echo -e "${GREEN}Starting Detailed Tool Testing${NC}"
    echo "==============================="
    
    # Change to project directory
    cd "$PROJECT_DIR"
    
    # Test command line arguments
    test_command_args
    
    # Test each tool individually
    test_tool "1" "Subdomain Scanner" "example.com"
    test_tool "2" "OSINT & Vulnerability Tool" "example.com"
    test_tool "3" "Web Application Security Scanner" "http://example.com"
    test_tool "4" "S3 Bucket Scanner" "example-bucket"
    test_tool "5" "Email Harvester" "example.com"
    test_tool "6" "Directory Bruteforcer" "http://example.com"
    test_tool "7" "Report Generator" ""
    test_tool "8" "Host & Subdomain Resolver" "example.com"
    test_tool "9" "Check Dependencies" "n"  # Don't install dependencies
    
    # Finalize error summary
    cat >> "$ERROR_SUMMARY" << EOF

## Recommendations

Based on the errors found, the following fixes are recommended:

1. **EOF Handling**: Tools that prompt for user input need to handle EOF gracefully
2. **Permission Issues**: Tools requiring elevated privileges need better error handling
3. **Input Validation**: Better validation of user inputs before processing
4. **Dependency Checking**: Improved dependency validation before tool execution
5. **Error Recovery**: Better error recovery and user feedback

## Files to Check

The following source files likely need fixes:

EOF
    
    # Count errors
    local error_count=$(grep -c "FAILED" "$ERROR_SUMMARY" 2>/dev/null || echo "0")
    
    echo -e "\n${BLUE}Testing Complete${NC}" | tee -a "$DETAILED_LOG"
    echo "================" | tee -a "$DETAILED_LOG"
    echo "" | tee -a "$DETAILED_LOG"
    echo "Results saved to:" | tee -a "$DETAILED_LOG"
    echo "  Detailed log: $DETAILED_LOG" | tee -a "$DETAILED_LOG"
    echo "  Error summary: $ERROR_SUMMARY" | tee -a "$DETAILED_LOG"
    echo "" | tee -a "$DETAILED_LOG"
    
    if [ "$error_count" -gt 0 ]; then
        echo -e "${RED}Found $error_count tools with errors${NC}" | tee -a "$DETAILED_LOG"
        echo "" | tee -a "$DETAILED_LOG"
        echo "Next steps:" | tee -a "$DETAILED_LOG"
        echo "1. Review $ERROR_SUMMARY for detailed error analysis" | tee -a "$DETAILED_LOG"
        echo "2. Fix the identified issues in the source code" | tee -a "$DETAILED_LOG"
        echo "3. Re-run tests to verify fixes" | tee -a "$DETAILED_LOG"
        exit 1
    else
        echo -e "${GREEN}All tools passed testing!${NC}" | tee -a "$DETAILED_LOG"
        exit 0
    fi
}

# Run main function
main "$@"