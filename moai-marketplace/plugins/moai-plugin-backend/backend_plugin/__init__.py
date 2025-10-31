"""
Backend Plugin - FastAPI/Database Backend Setup for Alfred Framework

@CODE:BACKEND-PLUGIN-INIT-001:INIT
"""

__version__ = "1.0.0-dev"
__author__ = "GOOS🪿"
__license__ = "MIT"

from .commands import init_fastapi, db_setup, resource_crud

__all__ = ["init_fastapi", "db_setup", "resource_crud"]
