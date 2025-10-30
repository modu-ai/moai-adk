"""
DevOps Plugin - Docker/CI/K8s Infrastructure Setup for Alfred Framework

@CODE:DEVOPS-PLUGIN-INIT-001:INIT
"""

__version__ = "1.0.0-dev"
__author__ = "GOOS🪿"
__license__ = "MIT"

from .commands import setup_docker, setup_ci, setup_k8s

__all__ = ["setup_docker", "setup_ci", "setup_k8s"]
