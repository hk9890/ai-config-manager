#!/usr/bin/env python3
import platform
import sys


def detect_platform():
    """Detect OS and architecture for aimgr installation."""
    os_name = platform.system().lower()
    machine = platform.machine().lower()

    # Map to aimgr release naming convention
    if os_name == "linux":
        arch = "amd64" if machine in ["x86_64", "amd64"] else "arm64"
        return f"linux_{arch}"
    elif os_name == "darwin":
        arch = "arm64" if machine == "arm64" else "amd64"
        return f"darwin_{arch}"
    elif os_name == "windows":
        arch = "amd64" if machine in ["x86_64", "amd64"] else "arm64"
        return f"windows_{arch}"
    else:
        return "unknown"


if __name__ == "__main__":
    print(detect_platform())
