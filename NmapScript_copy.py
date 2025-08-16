from datetime import datetime
import json
import socket
import os
import sys
import ipaddress
import re
import concurrent
import math
import logging
import logging.handlers
from concurrent.futures import ThreadPoolExecutor
import asyncio
import csv
import xml.etree.ElementTree as ET
from typing import List, Dict, Tuple, Optional, Union
import time
import threading
import subprocess
import venv

# Global variable to store detailed scan results
scan_detailed_results = {}


def setup_virtual_environment():
    """Setup and activate virtual environment"""
    script_dir = os.path.dirname(os.path.abspath(__file__))
    venv_dir = os.path.join(script_dir, "venv")
    
    # Check if already in virtual environment
    if hasattr(sys, 'real_prefix') or (hasattr(sys, 'base_prefix') and sys.base_prefix != sys.prefix):
        print("[+] Already running in virtual environment")
        return True
    
    # Check if venv directory exists
    if not os.path.exists(venv_dir):
        print("[+] Creating virtual environment...")
        try:
            venv.create(venv_dir, with_pip=True)
            print(f"[+] Virtual environment created at: {venv_dir}")
        except Exception as e:
            print(f"[-] Failed to create virtual environment: {e}")
            return False
    
    # Activate virtual environment by restarting with venv python
    if os.name == 'nt':  # Windows
        venv_python = os.path.join(venv_dir, "Scripts", "python.exe")
        venv_pip = os.path.join(venv_dir, "Scripts", "pip.exe")
    else:  # Unix/Linux
        venv_python = os.path.join(venv_dir, "bin", "python")
        venv_pip = os.path.join(venv_dir, "bin", "pip")
    
    if not os.path.exists(venv_python):
        print("[-] Virtual environment python not found")
        return False
    
    # Install required packages if not already installed
    required_packages = ["python-nmap", "scapy"]
    for package in required_packages:
        try:
            __import__(package.replace("-", "_"))
        except ImportError:
            print(f"[+] Installing {package} in virtual environment...")
            try:
                subprocess.check_call([venv_pip, "install", package], 
                                    stdout=subprocess.DEVNULL, 
                                    stderr=subprocess.DEVNULL)
                print(f"[+] {package} installed successfully")
            except subprocess.CalledProcessError as e:
                print(f"[-] Failed to install {package}: {e}")
    
    # If we're not in the venv python, restart with it
    if sys.executable != venv_python:
        print(f"[+] Activating virtual environment and restarting...")
        try:
            os.execv(venv_python, [venv_python] + sys.argv)
        except Exception as e:
            print(f"[-] Failed to restart with virtual environment: {e}")
            return False
    
    return True


def print_banner():
    """Display ASCII art banner for PythMap"""
    banner = """

 ██████╗ ██╗   ██╗████████╗██╗  ██╗███╗   ███╗ █████╗ ██████╗ 
 ██╔══██╗╚██╗ ██╔╝╚══██╔══╝██║  ██║████╗ ████║██╔══██╗██╔══██╗
 ██████╔╝ ╚████╔╝    ██║   ███████║██╔████╔██║███████║██████╔╝
 ██╔═══╝   ╚██╔╝     ██║   ██╔══██║██║╚██╔╝██║██╔══██║██╔═══╝ 
 ██║        ██║      ██║   ██║  ██║██║ ╚═╝ ██║██║  ██║██║     
 ╚═╝        ╚═╝      ╚═╝   ╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝     
                                                               
    An Advanced Network Port Scanner & Security Assessment Tool    
                     Created By TheBitty               

"""
    print(banner)


def setup_logging():
    """Configure logging for the application"""
    # Create logger
    logger = logging.getLogger('portscan')
    logger.setLevel(logging.INFO)

    # Create console handler (this will always work)
    console = logging.StreamHandler()
    console.setLevel(logging.INFO)
    console_format = logging.Formatter('%(message)s')
    console.setFormatter(console_format)
    logger.addHandler(console)

    # Try to create a file handler, but gracefully handle permission errors
    try:
        # Try current directory first (more likely to have permissions)
        script_dir = os.path.dirname(os.path.abspath(__file__))
        log_folder = os.path.join(script_dir, "logs")

        # Try to create the directory
        os.makedirs(log_folder, exist_ok=True)

        # Create file handler for detailed logs
        log_file = os.path.join(log_folder, f"scan_log_{datetime.now().strftime('%Y-%m-%d')}.log")
        file_handler = logging.handlers.RotatingFileHandler(
            log_file, maxBytes=10485760, backupCount=5)
        file_handler.setLevel(logging.DEBUG)
        file_format = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
        file_handler.setFormatter(file_format)
        logger.addHandler(file_handler)

        print(f"Log file created at: {log_file}")
    except PermissionError:
        print("Warning: Could not create log file due to permission error.")
        print("Continuing with console logging only.")
    except Exception as e:
        print(f"Warning: Could not set up file logging: {e}")
        print("Continuing with console logging only.")

    return logger


def check_root():
    """Check and obtain root privileges"""
    # First check if we're on a platform where we can check for root
    if hasattr(os, 'geteuid'):
        if os.geteuid() != 0:
            print("\nPlease run as root")
            print("\nAttempting to run as root...")
            try:
                # Get the full path to the current script
                script_path = os.path.abspath(sys.argv[0])
                # Use sys.executable to get the correct Python interpreter path
                interpreter = sys.executable
                # Execute: sudo [current-python] [full-script-path] [all-arguments]
                os.execvp('sudo', ['sudo', interpreter, script_path] + sys.argv[1:])
            except PermissionError:
                print("\n[-] Failed to obtain root privileges: Permission denied")
                print("\nExiting...please use the sudo command to run the script as root")
                sys.exit(1)
            except FileNotFoundError:
                print("\n[-] Failed to obtain root privileges: sudo command not found")
                print("\nExiting...please install sudo or run the script as root")
                sys.exit(1)
            except Exception as e:
                print(f"\n[-] Failed to obtain root privileges: {e}")
                print("\nExiting...please use the sudo command to run the script as root")
                sys.exit(1)  # scans are needed for root
    else:
        # On Windows or other platforms where geteuid isn't available
        print("\nWarning: Cannot check for root/admin privileges on this platform.")
        print("Some scanning features may not work without admin privileges.")
        print("Please make sure you're running this script with administrator rights.")

        # On Windows, we could try to check for admin, but for now just warn the user
        if os.name == 'nt':
            print("On Windows, right-click the command prompt and select 'Run as administrator'")

        # Continue anyway
        return


def validate_ip(ip):
    """Validate and clean IP address input"""
    try:
        ip = ip.strip()
        ipaddress.ip_address(ip)
        return ip, True
    except ValueError:
        return ip, False


def get_target_ip():
    """Get and validate target IP with user feedback"""
    while True:
        target = input("Enter target IP: ")
        ip, is_valid = validate_ip(target)

        if is_valid:
            return ip
        else:
            logger.warning(f"[-] Invalid IP address: {target}")
            logger.info("[!] Please enter a valid IP (e.g., 192.168.1.1)")
            continue


def get_port_range():
    """Get custom port range from user"""
    while True:
        try:
            logger.info("\nSelect port range to scan:")
            logger.info("1. Common ports (1-1024)")
            logger.info("2. Extended range (1-5000)")
            logger.info("3. Full range (1-65535)")
            logger.info("4. Custom range")

            choice = input("\nEnter choice (1-4): ").strip()

            if choice == '1':
                return 1, 1024
            elif choice == '2':
                return 1, 5000
            elif choice == '3':
                return 1, 65535
            elif choice == '4':
                start = int(input("Enter start port: "))
                end = int(input("Enter end port: "))
                if 0 < start < end <= 65535:
                    return start, end
                else:
                    logger.warning("Invalid port range!")
            else:
                logger.warning("Invalid choice!")
        except ValueError:
            logger.warning("Please enter valid numbers!")


def fast_port_discovery(target, start_port, end_port):
    """Perform fast port discovery scan to identify open ports"""
    logger.info(f"\n[STAGE 1] Fast port discovery on {target}...")
    
    try:
        import nmap
    except ImportError:
        logger.error("python-nmap package not found. Please run the script again to auto-install dependencies.")
        return []
    
    nm = nmap.PortScanner()
    
    try:
        # Fast port discovery scan
        if start_port == 1 and end_port == 65535:
            # Full port range scan
            logger.info("Scanning all ports (1-65535) with fast discovery...")
            nm.scan(target, arguments='-Pn -p- --min-rate=1000 -T4')
        else:
            # Custom port range
            port_range = f"{start_port}-{end_port}"
            logger.info(f"Scanning ports {port_range} with fast discovery...")
            nm.scan(target, ports=port_range, arguments='-Pn --min-rate=1000 -T4')
        
        open_ports = []
        if target in nm.all_hosts():
            for port in nm[target]['tcp'].keys():
                if nm[target]['tcp'][port]['state'] == 'open':
                    open_ports.append(port)
                    logger.info(f"[+] Open port found: {port}/tcp")
        
        logger.info(f"[STAGE 1] Fast discovery completed. Found {len(open_ports)} open ports")
        return sorted(open_ports)
        
    except Exception as e:
        logger.error(f"Fast port discovery error: {e}")
        return []


def comprehensive_service_scan(target, open_ports):
    """Perform comprehensive service detection and script scanning"""
    if not open_ports:
        logger.info("[STAGE 2] No open ports to scan")
        return {}
    
    logger.info(f"\n[STAGE 2] Comprehensive service detection on {len(open_ports)} ports...")
    
    try:
        import nmap
    except ImportError:
        logger.error("python-nmap package not found.")
        return {}
    
    nm = nmap.PortScanner()
    
    try:
        # Convert ports list to comma-separated string
        port_list = ','.join(map(str, open_ports))
        logger.info(f"Running detailed scan on ports: {port_list}")
        
        # Comprehensive service detection scan
        nm.scan(target, ports=port_list, arguments='-Pn -sC -sV')
        
        detailed_results = {}
        if target in nm.all_hosts():
            # Extract hostname/domain from scan results
            hostnames = nm[target].get('hostnames', [])
            
            for port in open_ports:
                if nm[target].has_tcp(port):
                    port_info = nm[target]['tcp'][port]
                    
                    detailed_results[port] = {
                        'state': port_info.get('state', 'unknown'),
                        'service': port_info.get('name', 'unknown'),
                        'version': port_info.get('version', ''),
                        'product': port_info.get('product', ''),
                        'extrainfo': port_info.get('extrainfo', ''),
                        'script_results': port_info.get('script', {}),
                        'cpe': port_info.get('cpe', '')
                    }
                    
                    # Log detailed service information
                    service_name = port_info.get('name', 'unknown')
                    product = port_info.get('product', '')
                    version = port_info.get('version', '')
                    extrainfo = port_info.get('extrainfo', '')
                    
                    service_line = f"{port}/tcp open {service_name}"
                    if product:
                        service_line += f" {product}"
                    if version:
                        service_line += f" {version}"
                    if extrainfo:
                        service_line += f" ({extrainfo})"
                    
                    logger.info(f"[+] {service_line}")
                    
                    # Display script results
                    scripts = port_info.get('script', {})
                    if scripts:
                        for script_name, script_output in scripts.items():
                            logger.info(f"| {script_name}:")
                            # Format multi-line script output
                            for line in str(script_output).split('\n'):
                                if line.strip():
                                    logger.info(f"|   {line}")
            
            # Display hostname information
            if hostnames:
                logger.info(f"\n[+] Hostnames discovered:")
                for hostname in hostnames:
                    logger.info(f"    {hostname.get('name', 'unknown')} ({hostname.get('type', 'unknown')})")
        
        logger.info(f"[STAGE 2] Comprehensive scan completed")
        return detailed_results
        
    except Exception as e:
        logger.error(f"Comprehensive service scan error: {e}")
        return {}


def scan_ports(target, start_port, end_port):
    """Perform two-stage scanning: fast discovery + comprehensive analysis"""
    logger.info(f"\nStarting two-stage scan on {target}...")
    
    # Stage 1: Fast port discovery
    open_ports = fast_port_discovery(target, start_port, end_port)
    
    if not open_ports:
        logger.info("No open ports found during fast discovery")
        return []
    
    # Stage 2: Comprehensive service detection
    detailed_results = comprehensive_service_scan(target, open_ports)
    
    # Store detailed results globally for later use
    global scan_detailed_results
    scan_detailed_results = detailed_results
    
    return open_ports


def get_service_name(port):
    """Identify common services by port number - expanded for security assessments"""
    common_ports = {
        21: "FTP",
        22: "SSH", 
        23: "Telnet",
        25: "SMTP",
        53: "DNS",
        67: "DHCP",
        68: "DHCP-Client",
        69: "TFTP",
        80: "HTTP",
        88: "Kerberos",
        110: "POP3",
        123: "NTP",
        135: "MS-RPC",
        139: "NetBIOS-SSN",
        143: "IMAP",
        161: "SNMP",
        162: "SNMP-Trap",
        389: "LDAP",
        443: "HTTPS",
        445: "Microsoft-DS",
        464: "Kpasswd5",
        465: "SMTPS",
        587: "SMTP Submission",
        593: "RPC-over-HTTP",
        636: "LDAPS",
        993: "IMAPS",
        995: "POP3S",
        1433: "MS-SQL-S",
        1521: "Oracle",
        3268: "GlobalCatalog",
        3269: "GlobalCatalog-SSL",
        3306: "MySQL",
        3389: "MS-WBT-Server",
        5357: "WSDAPI",
        5432: "PostgreSQL",
        5900: "VNC",
        5985: "WinRM-HTTP",
        5986: "WinRM-HTTPS",
        6379: "Redis",
        8080: "HTTP-Proxy",
        8443: "HTTPS-Alt",
        9389: "MC-NMF",
        27017: "MongoDB",
        47001: "WinRM",
        49152: "Dynamic-RPC",
        49153: "Dynamic-RPC",
        49154: "Dynamic-RPC",
        49155: "Dynamic-RPC",
        49156: "Dynamic-RPC",
        49157: "Dynamic-RPC",
        49158: "Dynamic-RPC"
    }
    
    # Handle dynamic RPC port ranges
    if 49152 <= port <= 65535:
        return "Dynamic-RPC"
    
    return common_ports.get(port, "Unknown")


def threaded_banner_grab(target, port):
    """Perform banner grabbing for a single port with improved protocol handling"""
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(2)
        s.connect((target, port))

        # Define port-specific probes
        port_probes = {
            21: b"USER anonymous\r\n",
            22: b"SSH-2.0-OpenSSH_8.2p1\r\n",
            25: b"EHLO scan.local\r\n",
            80: b"GET / HTTP/1.1\r\nHost: " + target.encode() + b"\r\n\r\n",
            110: b"USER test\r\n",
            143: b"A1 CAPABILITY\r\n",
            443: None,  # HTTPS requires SSL/TLS - handle specially
            3306: b"\x00\x00\x00\x00\x00",  # MySQL probe
            5432: b"\x00\x00\x00\x08\x04\xd2\x16\x2f",  # PostgreSQL probe
            8080: b"GET / HTTP/1.1\r\nHost: " + target.encode() + b"\r\n\r\n",
            8443: None  # HTTPS requires SSL/TLS - handle specially
        }

        # SSL/TLS ports that should be handled specially
        ssl_ports = {443, 465, 636, 993, 995, 8443}

        if port in ssl_ports:
            return port, f"SSL/TLS Service (port {port})"

        # Send appropriate probe or default to empty string
        probe = port_probes.get(port, b"")
        if probe:
            s.send(probe)

        # Safer banner receiving with graceful handling of connection issues
        try:
            banner_data = s.recv(1024)
            banner = banner_data.decode('utf-8', errors='ignore').strip()
        except socket.timeout:
            banner = "Connection timed out while receiving data"
        except ConnectionResetError:
            banner = "Connection reset by peer"
        except Exception as e:
            banner = f"Error receiving banner: {str(e)}"

        s.close()
        return port, banner
    except ConnectionRefusedError:
        return port, "Error: Connection refused"
    except socket.timeout:
        return port, "Error: Connection timeout"
    except OSError as e:
        return port, f"Error: Network error - {str(e)}"
    except Exception as e:
        return port, f"Error: {str(e)}"


def banner_grabbing(target, ports):
    """Perform parallel banner grabbing with improved error logging"""
    logger.info("\nPerforming banner grabbing...")
    banners = {}

    # Create a separate error counter to reduce console clutter
    error_count = 0
    error_types = {}

    with ThreadPoolExecutor(max_workers=10) as executor:
        future_to_port = {
            executor.submit(threaded_banner_grab, target, port): port
            for port in ports
        }

        for future in concurrent.futures.as_completed(future_to_port):
            try:
                port, banner = future.result()
                banners[port] = banner

                # Check if this is an error banner and handle accordingly
                if banner.startswith("Error:"):
                    error_count += 1
                    error_type = banner.split(":", 1)[1].strip()
                    # Count error types for summary
                    error_types[error_type] = error_types.get(error_type, 0) + 1
                    # Log error to debug log but don't show in main output
                    logger.debug(f"[-] Banner error for port {port}: {banner}")
                else:
                    # Only show successful banner grabs in main output
                    logger.info(f"[+] Banner for port {port}: {banner}")
            except Exception as e:
                error_count += 1
                error_type = str(e)
                error_types[error_type] = error_types.get(error_type, 0) + 1
                logger.debug(f"[-] Exception in banner grabbing for port {future_to_port[future]}: {e}")

        # Provide a summary of errors if any occurred
        if error_count > 0:
            logger.info(f"\n[!] Banner grabbing completed with {error_count} errors")
            for error_type, count in error_types.items():
                logger.debug(f"    - {count} x {error_type}")
            logger.info("[!] Run with debug logging enabled to see detailed error information")

    return banners


def detect_service_version(banner):
    """Extract version information from banner with improved multi-line handling"""
    version_patterns = {
        'ssh': [r'SSH-\d+\.\d+-([\w._-]+)', r'OpenSSH[_-]([\d.]+)'],
        'http': [r'Server:\s+([\w._/-]+)', r'Apache/([\d.]+)', r'nginx/([\d.]+)', r'Microsoft-IIS/([\d.]+)'],
        'ftp': [r'([\w._-]+) FTP', r'FTP server \(Version ([\w._-]+)\)'],
        'smtp': [r'([\w._-]+) ESMTP', r'([\w._-]+) Mail Service'],
        'mysql': [r'([\d.]+)-MariaDB', r'MySQL\s+([\d.]+)', r'mysql_native_password'],
        'telnet': [r'([\w._-]+) telnetd'],
        'pop3': [r'POP3 Server ([\w._-]+)'],
        'imap': [r'IMAP4rev1 ([\w._-]+)'],
        'generic': [r'version[\s:]+([\w._-]+)', r'([\d.]+\d)']  # Generic patterns as fallback
    }

    try:
        # Split banner into lines to handle multi-line responses
        banner_lines = banner.splitlines()

        # Try to match service-specific patterns first, line by line
        for line in banner_lines:
            for service, patterns in version_patterns.items():
                for pattern in patterns:
                    match = re.search(pattern, line, re.IGNORECASE)
                    if match:
                        return match.group(1)

        # If no match in line-by-line search, try the whole banner
        for service, patterns in version_patterns.items():
            for pattern in patterns:
                match = re.search(pattern, banner, re.IGNORECASE)
                if match:
                    return match.group(1)

        # If still no match, check for common version patterns in the whole banner
        for pattern in version_patterns['generic']:
            match = re.search(pattern, banner, re.IGNORECASE)
            if match:
                return match.group(1)

    except re.error as e:
        logger.debug(f"Regex error in service detection: {e}")
    except Exception as e:
        logger.debug(f"Error detecting version: {e}")

    return "Unknown Version"


def vuln_scan(target, ports):
    """Perform vulnerability scan using nmap NSE scripts"""
    logger.info("\nPerforming vulnerability scan...")
    
    try:
        import nmap
    except ImportError:
        logger.error("python-nmap package not found. Skipping vulnerability scan.")
        return {}
    
    nm = nmap.PortScanner()

    if not ports:
        logger.warning("No open ports to scan for vulnerabilities")
        return {}

    port_list = ','.join(map(str, ports))

    vuln_results = {}
    try:
        logger.info("Running vulnerability scripts (this may take a while)...")
        nm.scan(
            target,
            ports=port_list,
            arguments='--script vuln,exploit,auth,default,version -sV'
        )

        if target in nm.all_hosts():
            for port in ports:
                try:
                    if nm[target].has_tcp(port):
                        port_info = nm[target]['tcp'][port]
                        scripts_results = port_info.get('script', {})

                        if scripts_results:
                            vuln_results[port] = {
                                'service': port_info.get('name', 'unknown'),
                                'version': port_info.get('version', 'unknown'),
                                'vulnerabilities': scripts_results
                            }
                            logger.info(f"\n[+] Found potential vulnerabilities on port {port}:")
                            for script_name, result in scripts_results.items():
                                logger.info(f"  - {script_name}: {result}")
                except KeyError:
                    logger.debug(f"Port {port} not found in scan results")
                except Exception as e:
                    logger.error(f"Error processing vulnerability results for port {port}: {e}")

    except Exception as e:
        logger.error(f"\nNmap vulnerability scanning error: {e}")
    except Exception as e:
        logger.error(f"\nError during vulnerability scan: {e}")

    return vuln_results


def nmap_logger(ports, target, start_port, end_port, scan_start_time):
    """Log comprehensive scan results to JSON file"""
    timestamp = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
    log_folder = "logs"
    os.makedirs(log_folder, exist_ok=True)
    log_filename = os.path.join(log_folder, f"scan_{target}_{timestamp}.json")

    scan_duration = (datetime.now() - scan_start_time).total_seconds()

    try:
        hostname = socket.gethostbyaddr(target)[0]
    except socket.herror:
        hostname = "Unable to resolve"
    except Exception as e:
        hostname = f"Error resolving: {str(e)}"

    # Use detailed scan results from comprehensive scan
    global scan_detailed_results
    
    scan_data = {
        "metadata": {
            "scan_time": timestamp,
            "scan_duration_seconds": scan_duration,
            "target_ip": target,
            "target_hostname": hostname,
            "ports_scanned": {
                "start": start_port,
                "end": end_port,
                "total": end_port - start_port + 1
            },
            "open_ports_count": len(ports),
            "scan_method": "Two-stage: Fast discovery + Comprehensive service detection"
        },
        "open_ports": []
    }

    for port in ports:
        # Get detailed information from comprehensive scan
        detailed_info = scan_detailed_results.get(port, {})
        
        # Fallback to basic service identification if detailed scan failed
        service_name = detailed_info.get('service', get_service_name(port))
        product = detailed_info.get('product', '')
        version = detailed_info.get('version', '')
        extrainfo = detailed_info.get('extrainfo', '')
        script_results = detailed_info.get('script_results', {})
        
        # Build comprehensive service string
        service_description = service_name
        if product:
            service_description += f" {product}"
        if version:
            service_description += f" {version}"
        if extrainfo:
            service_description += f" ({extrainfo})"

        port_data = {
            "port_number": port,
            "service": service_name,
            "product": product,
            "version": version,
            "extrainfo": extrainfo,
            "service_description": service_description,
            "script_results": script_results,
            "cpe": detailed_info.get('cpe', ''),
            "state": detailed_info.get('state', 'open'),
            "scan_time": datetime.now().strftime("%H:%M:%S")
        }
        scan_data["open_ports"].append(port_data)

    try:
        with open(log_filename, 'w') as f:
            json.dump(scan_data, f, indent=4)
        logger.info(f"\n[+] Comprehensive scan results saved to {log_filename}")
        os.chmod(log_filename, 0o644)
    except PermissionError:
        logger.error(f"\n[-] Failed to save scan results: Permission denied for {log_filename}")
    except Exception as e:
        logger.error(f"\n[-] Failed to save scan results: {e}")


def print_summary(target, open_ports, scan_start_time):
    """Print a comprehensive summary of the scan results"""
    scan_duration = (datetime.now() - scan_start_time).total_seconds()
    global scan_detailed_results

    logger.info("\n" + "=" * 80)
    logger.info(f"COMPREHENSIVE SCAN SUMMARY FOR {target}")
    logger.info("=" * 80)
    logger.info(f"Scan started at: {scan_start_time.strftime('%Y-%m-%d %H:%M:%S')}")
    logger.info(f"Scan duration: {scan_duration:.2f} seconds")
    logger.info(f"Open ports found: {len(open_ports)}")
    logger.info(f"Scan method: Two-stage (Fast discovery + Service detection)")

    if open_ports:
        logger.info("\nDETAILED PORT INFORMATION:")
        logger.info("-" * 80)
        
        for port in sorted(open_ports):
            detailed_info = scan_detailed_results.get(port, {})
            
            # Build service description
            service_name = detailed_info.get('service', get_service_name(port))
            product = detailed_info.get('product', '')
            version = detailed_info.get('version', '')
            extrainfo = detailed_info.get('extrainfo', '')
            
            service_line = f"{port}/tcp open {service_name}"
            if product:
                service_line += f" {product}"
            if version:
                service_line += f" {version}"
            if extrainfo:
                service_line += f" ({extrainfo})"
            
            logger.info(f"[+] {service_line}")
            
            # Show important script results in summary
            scripts = detailed_info.get('script_results', {})
            important_scripts = ['ftp-anon', 'http-title', 'http-server-header', 'ssh-hostkey', 
                               'ssl-cert', 'smb-os-discovery', 'ms-sql-info', 'mysql-info']
            
            for script_name in important_scripts:
                if script_name in scripts:
                    script_output = str(scripts[script_name])
                    # Show first line of script output in summary
                    first_line = script_output.split('\n')[0].strip()
                    if first_line:
                        logger.info(f"    | {script_name}: {first_line}")

        # Show any discovered hostnames
        logger.info("\nADDITIONAL INFORMATION:")
        logger.info("-" * 80)
        
        # Check for domain indicators in script results
        domain_indicators = []
        for port, details in scan_detailed_results.items():
            scripts = details.get('script_results', {})
            
            # Extract domain from various script results
            for script_name, script_output in scripts.items():
                script_str = str(script_output).lower()
                if 'domain:' in script_str or 'commonname=' in script_str:
                    domain_indicators.append(f"Port {port}: {script_name}")
        
        if domain_indicators:
            logger.info("[+] Domain/Certificate information detected:")
            for indicator in domain_indicators:
                logger.info(f"    - {indicator}")
        
        # Service analysis
        unique_services = set()
        for port in open_ports:
            service = scan_detailed_results.get(port, {}).get('service', get_service_name(port))
            unique_services.add(service)
        
        logger.info(f"[+] Unique services detected: {', '.join(sorted(unique_services))}")
        
        # Security analysis insights
        security_insights = analyze_security_posture(open_ports, scan_detailed_results)
        if security_insights:
            logger.info("\nSECURITY ANALYSIS:")
            logger.info("-" * 80)
            for insight in security_insights:
                logger.info(f"[!] {insight}")

    logger.info("=" * 80)


def analyze_security_posture(open_ports, detailed_results):
    """Analyze discovered services for security insights"""
    insights = []
    
    # Check for Active Directory services
    ad_ports = {88: "Kerberos", 389: "LDAP", 636: "LDAPS", 3268: "Global Catalog", 3269: "Global Catalog SSL"}
    ad_services_found = [port for port in open_ports if port in ad_ports]
    
    if len(ad_services_found) >= 2:
        insights.append("Domain Controller detected - Kerberos, LDAP services indicate AD environment")
    
    # Check for database services
    db_ports = {1433: "MS-SQL", 3306: "MySQL", 5432: "PostgreSQL", 1521: "Oracle"}
    db_services_found = [db_ports[port] for port in open_ports if port in db_ports]
    
    if db_services_found:
        insights.append(f"Database services detected: {', '.join(db_services_found)} - Potential data targets")
    
    # Check for remote access services
    remote_ports = {22: "SSH", 3389: "RDP", 5985: "WinRM", 5986: "WinRM-HTTPS", 23: "Telnet"}
    remote_services_found = [remote_ports[port] for port in open_ports if port in remote_ports]
    
    if remote_services_found:
        insights.append(f"Remote access services: {', '.join(remote_services_found)} - Authentication targets")
    
    # Check for file sharing services
    if 445 in open_ports:
        insights.append("SMB service detected - Check for anonymous access and share enumeration")
    
    if 21 in open_ports:
        ftp_info = detailed_results.get(21, {})
        scripts = ftp_info.get('script_results', {})
        if 'ftp-anon' in scripts:
            if 'Anonymous FTP login allowed' in str(scripts['ftp-anon']):
                insights.append("Anonymous FTP access enabled - Potential information disclosure")
    
    # Check for web services
    web_ports = [port for port in open_ports if port in [80, 443, 8080, 8443]]
    if web_ports:
        insights.append(f"Web services on ports {web_ports} - Web application attack surface")
    
    # Check for potentially risky services
    if 23 in open_ports:
        insights.append("Telnet service detected - Unencrypted protocol, consider SSH alternative")
    
    # Check for high port count (possible port scan evasion)
    if len(open_ports) > 20:
        insights.append(f"High number of open ports ({len(open_ports)}) - Review service necessity")
    
    return insights


if __name__ == "__main__":
    # Setup virtual environment first
    print("Initializing PythMap Scanner...")
    if not setup_virtual_environment():
        print("[-] Failed to setup virtual environment, continuing anyway...")
    
    # Import required packages after venv setup
    try:
        import nmap
        from scapy.all import *
        print("[+] Required packages loaded successfully")
    except ImportError as e:
        print(f"[-] Failed to import required packages: {e}")
        print("[!] Some functionality may be limited")
    
    # Set up logging
    try:
        logger = setup_logging()
    except Exception as e:
        print(f"Error setting up logging: {e}")
        print("Continuing with basic console output.")
        # Create a basic logger that just prints to console
        logger = logging.getLogger('portscan')
        logger.setLevel(logging.INFO)
        console = logging.StreamHandler()
        logger.addHandler(console)

    try:
        # Check for root privileges
        check_root()
        
        # Clear screen after privilege escalation
        os.system('clear' if os.name != 'nt' else 'cls')
        
        # Show banner after screen clear
        print_banner()
        logger.info("Starting advanced port scanner")

        # Get target information
        target = get_target_ip()
        logger.info(f"Target selected: {target}")

        start_port, end_port = get_port_range()
        logger.info(f"Port range selected: {start_port}-{end_port}")

        # Start scanning
        scan_start_time = datetime.now()
        logger.info(f"\nStarting scan at: {scan_start_time.strftime('%Y-%m-%d %H:%M:%S')}")

        open_ports = scan_ports(target, start_port, end_port)

        if open_ports:
            print_summary(target, open_ports, scan_start_time)
            nmap_logger(open_ports, target, start_port, end_port, scan_start_time)
        else:
            logger.info("\nNo open ports found.")
    except KeyboardInterrupt:
        logger.info("\n\nScan interrupted by user. Exiting...")
    except Exception as e:
        logger.error(f"\nUnexpected error: {e}")
        # Only try to log debug info if logger has handlers
        if logger.handlers:
            logger.debug("Exception details:", exc_info=True)
