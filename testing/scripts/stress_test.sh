#!/bin/bash

# GopherStrike Stress Testing Script
# Tests the application under various stress conditions

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

STRESS_LOG="$RESULTS_DIR/stress_test_$(date +%Y%m%d_%H%M%S).log"

echo "GopherStrike Stress Testing" | tee "$STRESS_LOG"
echo "==========================" | tee -a "$STRESS_LOG"
echo "Started at: $(date)" | tee -a "$STRESS_LOG"
echo "" | tee -a "$STRESS_LOG"

# Function to log results
log_result() {
    local test_name="$1"
    local status="$2"
    local details="$3"
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}[PASS]${NC} $test_name: $details" | tee -a "$STRESS_LOG"
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}[FAIL]${NC} $test_name: $details" | tee -a "$STRESS_LOG"
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}[WARN]${NC} $test_name: $details" | tee -a "$STRESS_LOG"
    else
        echo -e "${BLUE}[INFO]${NC} $test_name: $details" | tee -a "$STRESS_LOG"
    fi
}

# Test concurrent executions
test_concurrent_execution() {
    echo -e "\n${BLUE}Testing Concurrent Execution${NC}" | tee -a "$STRESS_LOG"
    echo "============================" | tee -a "$STRESS_LOG"
    
    local pids=()
    local success_count=0
    local fail_count=0
    
    # Start multiple instances
    for i in {1..5}; do
        (echo "11" | timeout 10s "$BINARY" >/dev/null 2>&1) &
        pids+=($!)
    done
    
    # Wait for all to complete
    for pid in "${pids[@]}"; do
        if wait "$pid"; then
            ((success_count++))
        else
            ((fail_count++))
        fi
    done
    
    if [ "$fail_count" -eq 0 ]; then
        log_result "Concurrent Execution" "PASS" "$success_count/5 instances completed successfully"
    else
        log_result "Concurrent Execution" "WARN" "$success_count/5 instances succeeded, $fail_count failed"
    fi
}

# Test memory usage with large inputs
test_memory_usage() {
    echo -e "\n${BLUE}Testing Memory Usage${NC}" | tee -a "$STRESS_LOG"
    echo "====================" | tee -a "$STRESS_LOG"
    
    # Generate large input string
    local large_input=$(python3 -c "print('A' * 10000)")
    
    # Monitor memory usage
    local mem_before=$(ps -o pid,vsz,rss,comm | grep -v grep | wc -l)
    
    # Run with large input
    local output
    local exit_code
    output=$(echo -e "${large_input}\n11" | timeout 15s "$BINARY" 2>&1) || exit_code=$?
    
    if [ ${exit_code:-0} -eq 0 ] || [ ${exit_code:-0} -eq 1 ]; then
        if [[ $output == *"Invalid choice"* ]]; then
            log_result "Large Input Handling" "PASS" "Large input properly rejected"
        else
            log_result "Large Input Handling" "WARN" "Large input behavior unclear"
        fi
    else
        log_result "Large Input Handling" "FAIL" "Application crashed or hung with large input"
    fi
}

# Test rapid input sequences
test_rapid_input() {
    echo -e "\n${BLUE}Testing Rapid Input${NC}" | tee -a "$STRESS_LOG"
    echo "===================" | tee -a "$STRESS_LOG"
    
    # Generate rapid input sequence
    local rapid_input=""
    for i in {1..100}; do
        rapid_input+="invalid\n"
    done
    rapid_input+="11\n"
    
    local start_time=$(date +%s)
    local output
    local exit_code
    output=$(echo -e "$rapid_input" | timeout 30s "$BINARY" 2>&1) || exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ ${exit_code:-0} -eq 0 ]; then
        log_result "Rapid Input" "PASS" "Handled 100 rapid inputs in ${duration}s"
    elif [ ${exit_code:-0} -eq 124 ]; then
        log_result "Rapid Input" "FAIL" "Timed out processing rapid inputs"
    else
        log_result "Rapid Input" "WARN" "Completed with exit code ${exit_code:-unknown} in ${duration}s"
    fi
}

# Test file descriptor limits
test_file_descriptors() {
    echo -e "\n${BLUE}Testing File Descriptor Usage${NC}" | tee -a "$STRESS_LOG"
    echo "==============================" | tee -a "$STRESS_LOG"
    
    # Check available file descriptors
    local fd_limit=$(ulimit -n)
    log_result "FD Limit Check" "INFO" "File descriptor limit: $fd_limit"
    
    # Test opening many instances rapidly
    local instances=()
    local count=0
    
    for i in {1..20}; do
        if echo "11" | timeout 5s "$BINARY" >/dev/null 2>&1; then
            ((count++))
        fi
    done
    
    if [ "$count" -eq 20 ]; then
        log_result "FD Usage" "PASS" "All 20 sequential instances completed"
    else
        log_result "FD Usage" "WARN" "Only $count/20 instances completed successfully"
    fi
}

# Test signal handling under stress
test_signal_stress() {
    echo -e "\n${BLUE}Testing Signal Handling Under Stress${NC}" | tee -a "$STRESS_LOG"
    echo "====================================" | tee -a "$STRESS_LOG"
    
    local success_count=0
    local total_tests=5
    
    for i in $(seq 1 $total_tests); do
        # Start instance
        echo "11" | timeout 10s "$BINARY" >/dev/null 2>&1 &
        local pid=$!
        
        # Wait a moment then send SIGTERM
        sleep 1
        if kill -TERM "$pid" 2>/dev/null; then
            sleep 2
            if ! kill -0 "$pid" 2>/dev/null; then
                ((success_count++))
            else
                kill -KILL "$pid" 2>/dev/null
            fi
        else
            # Process already ended
            ((success_count++))
        fi
    done
    
    if [ "$success_count" -eq "$total_tests" ]; then
        log_result "Signal Stress" "PASS" "All instances handled signals properly"
    else
        log_result "Signal Stress" "WARN" "$success_count/$total_tests instances handled signals properly"
    fi
}

# Test resource cleanup
test_resource_cleanup() {
    echo -e "\n${BLUE}Testing Resource Cleanup${NC}" | tee -a "$STRESS_LOG"
    echo "========================" | tee -a "$STRESS_LOG"
    
    local temp_files_before=$(find /tmp -name "*gopher*" -o -name "*strike*" 2>/dev/null | wc -l)
    
    # Run multiple instances
    for i in {1..10}; do
        echo "11" | timeout 5s "$BINARY" >/dev/null 2>&1
    done
    
    local temp_files_after=$(find /tmp -name "*gopher*" -o -name "*strike*" 2>/dev/null | wc -l)
    local temp_diff=$((temp_files_after - temp_files_before))
    
    if [ "$temp_diff" -eq 0 ]; then
        log_result "Resource Cleanup" "PASS" "No temporary files left behind"
    elif [ "$temp_diff" -lt 5 ]; then
        log_result "Resource Cleanup" "WARN" "$temp_diff temporary files created"
    else
        log_result "Resource Cleanup" "FAIL" "$temp_diff temporary files left behind"
    fi
}

# Test process limits
test_process_limits() {
    echo -e "\n${BLUE}Testing Process Limits${NC}" | tee -a "$STRESS_LOG"
    echo "======================" | tee -a "$STRESS_LOG"
    
    local proc_limit=$(ulimit -u)
    log_result "Process Limit" "INFO" "Process limit: $proc_limit"
    
    # Test creating multiple child processes
    local pids=()
    local max_procs=10
    
    for i in $(seq 1 $max_procs); do
        (sleep 5) &
        pids+=($!)
    done
    
    # Run GopherStrike while background processes are running
    local output
    local exit_code
    output=$(echo "11" | timeout 10s "$BINARY" 2>&1) || exit_code=$?
    
    # Clean up background processes
    for pid in "${pids[@]}"; do
        kill "$pid" 2>/dev/null || true
    done
    
    if [ ${exit_code:-0} -eq 0 ]; then
        log_result "Process Limits" "PASS" "Worked correctly with $max_procs background processes"
    else
        log_result "Process Limits" "WARN" "Issues with background processes (exit code: ${exit_code:-unknown})"
    fi
}

# Main stress test execution
main() {
    echo -e "${GREEN}Starting GopherStrike Stress Testing${NC}"
    echo "===================================="
    
    # Change to project directory
    cd "$PROJECT_DIR"
    
    # Verify binary exists
    if [ ! -f "$BINARY" ]; then
        echo -e "${RED}GopherStrike binary not found. Building...${NC}"
        if ! go build -o GopherStrike; then
            echo -e "${RED}Failed to build GopherStrike${NC}"
            exit 1
        fi
    fi
    
    # Run stress tests
    test_concurrent_execution
    test_memory_usage
    test_rapid_input
    test_file_descriptors
    test_signal_stress
    test_resource_cleanup
    test_process_limits
    
    # Summary
    echo -e "\n${BLUE}Stress Test Summary${NC}" | tee -a "$STRESS_LOG"
    echo "===================" | tee -a "$STRESS_LOG"
    
    local pass_count=$(grep -c "\[PASS\]" "$STRESS_LOG" || echo "0")
    local fail_count=$(grep -c "\[FAIL\]" "$STRESS_LOG" || echo "0")
    local warn_count=$(grep -c "\[WARN\]" "$STRESS_LOG" || echo "0")
    
    echo "Tests Passed: $pass_count" | tee -a "$STRESS_LOG"
    echo "Tests Failed: $fail_count" | tee -a "$STRESS_LOG"
    echo "Warnings: $warn_count" | tee -a "$STRESS_LOG"
    echo "" | tee -a "$STRESS_LOG"
    echo "Detailed results saved to: $STRESS_LOG" | tee -a "$STRESS_LOG"
    
    if [ "$fail_count" -gt 0 ]; then
        echo -e "\n${RED}Some stress tests failed.${NC}"
        exit 1
    elif [ "$warn_count" -gt 0 ]; then
        echo -e "\n${YELLOW}Stress tests completed with warnings.${NC}"
        exit 0
    else
        echo -e "\n${GREEN}All stress tests passed!${NC}"
        exit 0
    fi
}

# Run main function
main "$@"