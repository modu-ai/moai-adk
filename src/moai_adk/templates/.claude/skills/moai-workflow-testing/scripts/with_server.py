#!/usr/bin/env python3
"""
Start one or more servers, wait for them to be ready, run a command, then clean up.

Usage:
    # Single server
    python scripts/with_server.py --server "npm run dev" --port 5173 -- python automation.py
    python scripts/with_server.py --server "npm start" --port 3000 -- python test.py

    # Multiple servers with working directories
    python scripts/with_server.py \
      --server "python server.py" --port 3000 --cwd backend \
      --server "npm run dev" --port 5173 --cwd frontend \
      -- python test.py

Security Note:
    This script uses shell=False to prevent command injection attacks (CWE-78).
    Use --cwd to specify working directories instead of shell operators like 'cd'.
"""

import argparse
import shlex
import socket
import subprocess
import sys
import time
from pathlib import Path


def is_server_ready(port, timeout=30):
    """Wait for server to be ready by polling the port."""
    start_time = time.time()
    while time.time() - start_time < timeout:
        try:
            with socket.create_connection(("localhost", port), timeout=1):
                return True
        except (socket.error, ConnectionRefusedError):
            time.sleep(0.5)
    return False


def parse_command_safely(cmd_string):
    """
    Parse a command string into a list for subprocess execution.

    Uses shlex.split() for safe command parsing to prevent shell injection.
    This handles quoted arguments properly but does NOT support shell operators
    like &&, ||, |, ;, or redirections.

    Args:
        cmd_string: Command string to parse

    Returns:
        List of command arguments suitable for subprocess with shell=False

    Raises:
        ValueError: If command contains unsupported shell operators
    """
    # Check for shell operators that are not supported
    shell_operators = ["&&", "||", "|", ";", ">", "<", ">>", "<<"]
    for op in shell_operators:
        if op in cmd_string:
            raise ValueError(
                f"Shell operator '{op}' is not supported for security reasons. "
                f"Use --cwd to change working directory instead of 'cd &&'."
            )

    try:
        return shlex.split(cmd_string)
    except ValueError as e:
        raise ValueError(f"Failed to parse command '{cmd_string}': {e}") from e


def main():
    parser = argparse.ArgumentParser(
        description="Run command with one or more servers",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Security:
  This script uses shell=False to prevent command injection attacks (CWE-78).
  Shell operators (&&, ||, |, ;) are not supported for security reasons.
  Use --cwd to specify working directories instead of 'cd &&' patterns.

Examples:
  # Single server
  python with_server.py --server "npm run dev" --port 3000 -- pytest

  # Multiple servers with different working directories
  python with_server.py \\
    --server "python server.py" --port 3000 --cwd backend \\
    --server "npm run dev" --port 5173 --cwd frontend \\
    -- pytest tests/
""",
    )
    parser.add_argument(
        "--server",
        action="append",
        dest="servers",
        required=True,
        help="Server command (can be repeated)",
    )
    parser.add_argument(
        "--port",
        action="append",
        dest="ports",
        type=int,
        required=True,
        help="Port for each server (must match --server count)",
    )
    parser.add_argument(
        "--cwd",
        action="append",
        dest="cwds",
        default=None,
        help="Working directory for each server (optional, defaults to current directory)",
    )
    parser.add_argument(
        "--timeout",
        type=int,
        default=30,
        help="Timeout in seconds per server (default: 30)",
    )
    parser.add_argument("command", nargs=argparse.REMAINDER, help="Command to run after server(s) ready")

    args = parser.parse_args()

    # Remove the '--' separator if present
    if args.command and args.command[0] == "--":
        args.command = args.command[1:]

    if not args.command:
        print("Error: No command specified to run")
        sys.exit(1)

    # Parse server configurations
    if len(args.servers) != len(args.ports):
        print("Error: Number of --server and --port arguments must match")
        sys.exit(1)

    # Handle working directories - default to None (current dir) for each server
    cwds = args.cwds or []
    while len(cwds) < len(args.servers):
        cwds.append(None)

    if len(cwds) > len(args.servers):
        print("Error: Too many --cwd arguments (must not exceed number of --server arguments)")
        sys.exit(1)

    servers = []
    for cmd, port, cwd in zip(args.servers, args.ports, cwds):
        # Validate and resolve working directory
        resolved_cwd = None
        if cwd:
            cwd_path = Path(cwd)
            if not cwd_path.exists():
                print(f"Error: Working directory '{cwd}' does not exist")
                sys.exit(1)
            if not cwd_path.is_dir():
                print(f"Error: '{cwd}' is not a directory")
                sys.exit(1)
            resolved_cwd = str(cwd_path.resolve())

        # Parse command safely
        try:
            cmd_list = parse_command_safely(cmd)
        except ValueError as e:
            print(f"Error: {e}")
            sys.exit(1)

        servers.append({"cmd": cmd, "cmd_list": cmd_list, "port": port, "cwd": resolved_cwd})

    server_processes = []

    try:
        # Start all servers
        for i, server in enumerate(servers):
            cwd_info = f" (cwd: {server['cwd']})" if server["cwd"] else ""
            print(f"Starting server {i + 1}/{len(servers)}: {server['cmd']}{cwd_info}")

            # Use shell=False with parsed command list to prevent command injection (CWE-78)
            process = subprocess.Popen(
                server["cmd_list"],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                cwd=server["cwd"],
            )
            server_processes.append(process)

            # Wait for this server to be ready
            print(f"Waiting for server on port {server['port']}...")
            if not is_server_ready(server["port"], timeout=args.timeout):
                raise RuntimeError(f"Server failed to start on port {server['port']} within {args.timeout}s")

            print(f"Server ready on port {server['port']}")

        print(f"\nAll {len(servers)} server(s) ready")

        # Run the command
        print(f"Running: {' '.join(args.command)}\n")
        result = subprocess.run(args.command)
        sys.exit(result.returncode)

    finally:
        # Clean up all servers
        print(f"\nStopping {len(server_processes)} server(s)...")
        for i, process in enumerate(server_processes):
            try:
                process.terminate()
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
            print(f"Server {i + 1} stopped")
        print("All servers stopped")


if __name__ == "__main__":
    main()
