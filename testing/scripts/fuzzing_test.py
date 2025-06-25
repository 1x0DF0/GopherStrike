#!/usr/bin/env python3

"""
GopherStrike Fuzzing Test Script
This script performs fuzzing tests on GopherStrike to find edge cases and potential crashes.
"""

import subprocess
import sys
import time
import random
import string
import json
import os
from datetime import datetime

class GopherStrikeFuzzer:
    def __init__(self, binary_path, results_dir):
        self.binary_path = binary_path
        self.results_dir = results_dir
        self.results = []
        self.start_time = datetime.now()
        
        # Create results directory
        os.makedirs(results_dir, exist_ok=True)
        
    def log_result(self, test_type, input_data, expected, actual, status):
        """Log test result"""
        result = {
            'timestamp': datetime.now().isoformat(),
            'test_type': test_type,
            'input': input_data,
            'expected': expected,
            'actual': actual,
            'status': status
        }
        self.results.append(result)
        
        # Print result
        status_color = {
            'PASS': '\033[92m',
            'FAIL': '\033[91m',
            'WARN': '\033[93m',
            'INFO': '\033[94m'
        }
        reset_color = '\033[0m'
        
        print(f"{status_color.get(status, '')}{status:4}{reset_color} {test_type}: {input_data[:50]}...")
        
    def run_command(self, input_data, timeout=10):
        """Run GopherStrike with given input"""
        try:
            process = subprocess.Popen(
                [self.binary_path],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            
            stdout, stderr = process.communicate(input=input_data, timeout=timeout)
            return process.returncode, stdout, stderr
            
        except subprocess.TimeoutExpired:
            process.kill()
            return -1, "", "TIMEOUT"
        except Exception as e:
            return -2, "", str(e)
    
    def test_boundary_values(self):
        """Test boundary values for menu options"""
        print("\n=== Testing Boundary Values ===")
        
        boundary_tests = [
            ("0", "Below minimum"),
            ("1", "Minimum valid"),
            ("11", "Maximum valid"),
            ("12", "Above maximum"),
            ("100", "Far above maximum"),
            ("-1", "Negative number"),
            ("-999", "Large negative"),
            ("2147483647", "Max 32-bit int"),
            ("2147483648", "Max 32-bit int + 1"),
            ("-2147483648", "Min 32-bit int"),
        ]
        
        for input_val, description in boundary_tests:
            test_input = f"{input_val}\n11\n"  # Try option then exit
            
            returncode, stdout, stderr = self.run_command(test_input)
            
            if returncode == 0:
                if "Invalid choice" in stdout or "invalid" in stdout.lower():
                    self.log_result("Boundary", f"{input_val} ({description})", 
                                  "Invalid choice rejection", "Correctly rejected", "PASS")
                elif input_val in ["1", "11"]:
                    self.log_result("Boundary", f"{input_val} ({description})", 
                                  "Valid option acceptance", "Correctly accepted", "PASS")
                else:
                    self.log_result("Boundary", f"{input_val} ({description})", 
                                  "Invalid choice rejection", "Unexpectedly accepted", "FAIL")
            else:
                self.log_result("Boundary", f"{input_val} ({description})", 
                              "Graceful handling", f"Exit code: {returncode}", "WARN")
    
    def test_random_strings(self):
        """Test random string inputs"""
        print("\n=== Testing Random Strings ===")
        
        for i in range(20):
            # Generate random string
            length = random.randint(1, 100)
            chars = string.ascii_letters + string.digits + string.punctuation
            random_string = ''.join(random.choice(chars) for _ in range(length))
            
            test_input = f"{random_string}\n11\n"
            
            returncode, stdout, stderr = self.run_command(test_input)
            
            if returncode == 0:
                if "Invalid choice" in stdout or "invalid" in stdout.lower():
                    self.log_result("Random String", random_string[:20], 
                                  "Invalid choice rejection", "Correctly rejected", "PASS")
                else:
                    self.log_result("Random String", random_string[:20], 
                                  "Invalid choice rejection", "Unexpectedly accepted", "FAIL")
            else:
                self.log_result("Random String", random_string[:20], 
                              "Graceful handling", f"Exit code: {returncode}", "WARN")
    
    def test_special_characters(self):
        """Test special characters and escape sequences"""
        print("\n=== Testing Special Characters ===")
        
        special_tests = [
            ("\\n", "Newline character"),
            ("\\t", "Tab character"),
            ("\\r", "Carriage return"),
            ("\\0", "Null character"),
            ("\\x01", "Control character"),
            ("\\x7f", "DEL character"),
            ("\\xff", "High byte"),
            ("$(echo hello)", "Command injection attempt"),
            ("`echo hello`", "Backtick command injection"),
            ("; echo hello", "Command separator"),
            ("\\\\ echo hello", "Backslash escape"),
            ("' OR 1=1 --", "SQL injection attempt"),
            ("<script>alert(1)</script>", "XSS attempt"),
            ("../../../etc/passwd", "Path traversal"),
            ("%n%n%n%n", "Format string"),
            ("A" * 1000, "Buffer overflow attempt"),
        ]
        
        for input_val, description in special_tests:
            test_input = f"{input_val}\n11\n"
            
            returncode, stdout, stderr = self.run_command(test_input)
            
            if returncode == 0:
                if "Invalid choice" in stdout or "invalid" in stdout.lower():
                    self.log_result("Special Chars", f"{description}", 
                                  "Invalid choice rejection", "Correctly rejected", "PASS")
                else:
                    self.log_result("Special Chars", f"{description}", 
                                  "Invalid choice rejection", "Unexpectedly accepted", "FAIL")
            elif returncode == -1:
                self.log_result("Special Chars", f"{description}", 
                              "Graceful handling", "TIMEOUT", "WARN")
            else:
                self.log_result("Special Chars", f"{description}", 
                              "Graceful handling", f"Exit code: {returncode}", "WARN")
    
    def test_empty_inputs(self):
        """Test various empty input scenarios"""
        print("\n=== Testing Empty Inputs ===")
        
        empty_tests = [
            ("", "Empty string"),
            (" ", "Single space"),
            ("  ", "Multiple spaces"),
            ("\n", "Just newline"),
            ("\t", "Just tab"),
            ("\r\n", "Windows newline"),
            ("   \n", "Spaces then newline"),
        ]
        
        for input_val, description in empty_tests:
            returncode, stdout, stderr = self.run_command(input_val, timeout=5)
            
            if returncode == 0:
                if "Exiting GopherStrike" in stdout:
                    self.log_result("Empty Input", description, 
                                  "Graceful exit", "Correctly exited", "PASS")
                else:
                    self.log_result("Empty Input", description, 
                                  "Graceful handling", "Unexpected behavior", "WARN")
            elif returncode == -1:
                self.log_result("Empty Input", description, 
                              "Graceful handling", "TIMEOUT", "WARN")
            else:
                self.log_result("Empty Input", description, 
                              "Graceful handling", f"Exit code: {returncode}", "PASS")
    
    def test_rapid_inputs(self):
        """Test rapid input sequences"""
        print("\n=== Testing Rapid Inputs ===")
        
        rapid_tests = [
            ("1\n2\n3\n4\n5\n11\n", "Sequential valid options"),
            ("1\n1\n1\n1\n11\n", "Repeated option"),
            ("invalid\ninvalid\ninvalid\n11\n", "Repeated invalid"),
            ("1\ninvalid\n2\ninvalid\n11\n", "Mixed valid/invalid"),
            ("\n\n\n\n11\n", "Multiple empty lines"),
        ]
        
        for input_val, description in rapid_tests:
            returncode, stdout, stderr = self.run_command(input_val, timeout=30)
            
            if returncode == 0:
                self.log_result("Rapid Input", description, 
                              "Graceful handling", "Completed successfully", "PASS")
            elif returncode == -1:
                self.log_result("Rapid Input", description, 
                              "Graceful handling", "TIMEOUT", "WARN")
            else:
                self.log_result("Rapid Input", description, 
                              "Graceful handling", f"Exit code: {returncode}", "WARN")
    
    def test_signal_handling(self):
        """Test signal handling"""
        print("\n=== Testing Signal Handling ===")
        
        # Test Ctrl+C simulation
        try:
            process = subprocess.Popen(
                [self.binary_path],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            
            time.sleep(2)  # Let it start
            process.send_signal(subprocess.signal.SIGINT)  # Ctrl+C
            
            try:
                stdout, stderr = process.communicate(timeout=5)
                self.log_result("Signal", "SIGINT (Ctrl+C)", 
                              "Graceful exit", f"Exit code: {process.returncode}", "PASS")
            except subprocess.TimeoutExpired:
                process.kill()
                self.log_result("Signal", "SIGINT (Ctrl+C)", 
                              "Graceful exit", "TIMEOUT after signal", "FAIL")
                
        except Exception as e:
            self.log_result("Signal", "SIGINT (Ctrl+C)", 
                          "Graceful exit", f"Exception: {e}", "WARN")
    
    def save_results(self):
        """Save test results to file"""
        timestamp = self.start_time.strftime("%Y%m%d_%H%M%S")
        
        # Save JSON results
        json_file = os.path.join(self.results_dir, f"fuzzing_results_{timestamp}.json")
        with open(json_file, 'w') as f:
            json.dump(self.results, f, indent=2)
        
        # Save summary report
        summary_file = os.path.join(self.results_dir, f"fuzzing_summary_{timestamp}.txt")
        with open(summary_file, 'w') as f:
            f.write("GopherStrike Fuzzing Test Summary\n")
            f.write("=" * 40 + "\n\n")
            f.write(f"Test started: {self.start_time}\n")
            f.write(f"Test completed: {datetime.now()}\n")
            f.write(f"Duration: {datetime.now() - self.start_time}\n\n")
            
            # Count results by status
            status_counts = {}
            for result in self.results:
                status = result['status']
                status_counts[status] = status_counts.get(status, 0) + 1
            
            f.write("Test Results:\n")
            for status, count in sorted(status_counts.items()):
                f.write(f"  {status}: {count}\n")
            
            f.write(f"\nTotal tests: {len(self.results)}\n\n")
            
            # List failures
            failures = [r for r in self.results if r['status'] == 'FAIL']
            if failures:
                f.write("FAILURES:\n")
                for failure in failures:
                    f.write(f"  {failure['test_type']}: {failure['input'][:50]}...\n")
                    f.write(f"    Expected: {failure['expected']}\n")
                    f.write(f"    Actual: {failure['actual']}\n\n")
            
            # List warnings
            warnings = [r for r in self.results if r['status'] == 'WARN']
            if warnings:
                f.write("WARNINGS:\n")
                for warning in warnings:
                    f.write(f"  {warning['test_type']}: {warning['input'][:50]}...\n")
                    f.write(f"    Expected: {warning['expected']}\n")
                    f.write(f"    Actual: {warning['actual']}\n\n")
        
        print(f"\nResults saved to:")
        print(f"  Detailed: {json_file}")
        print(f"  Summary: {summary_file}")
        
        return status_counts
    
    def run_all_tests(self):
        """Run all fuzzing tests"""
        print("Starting GopherStrike Fuzzing Tests")
        print("=" * 40)
        
        self.test_boundary_values()
        self.test_random_strings()
        self.test_special_characters()
        self.test_empty_inputs()
        self.test_rapid_inputs()
        self.test_signal_handling()
        
        print("\n" + "=" * 40)
        print("Fuzzing Tests Complete")
        
        status_counts = self.save_results()
        
        # Print summary
        print(f"\nTest Summary:")
        for status, count in sorted(status_counts.items()):
            print(f"  {status}: {count}")
        
        # Return exit code based on results
        if status_counts.get('FAIL', 0) > 0:
            return 1
        elif status_counts.get('WARN', 0) > 0:
            return 2
        else:
            return 0

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 fuzzing_test.py <path_to_gopherstrike_binary>")
        sys.exit(1)
    
    binary_path = sys.argv[1]
    results_dir = os.path.join(os.path.dirname(__file__), "../results")
    
    # Check if binary exists
    if not os.path.exists(binary_path):
        print(f"Error: Binary not found at {binary_path}")
        sys.exit(1)
    
    # Check if binary is executable
    if not os.access(binary_path, os.X_OK):
        print(f"Error: Binary is not executable: {binary_path}")
        sys.exit(1)
    
    fuzzer = GopherStrikeFuzzer(binary_path, results_dir)
    exit_code = fuzzer.run_all_tests()
    sys.exit(exit_code)

if __name__ == "__main__":
    main()