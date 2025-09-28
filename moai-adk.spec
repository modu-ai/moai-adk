# -*- mode: python ; coding: utf-8 -*-
# @TASK:PYINSTALLER-SPEC-001
"""
MoAI-ADK PyInstaller Build Specification

This spec file builds a single executable for Windows distribution.
Optimized for minimal size and cross-platform compatibility.

@REQ:WINDOWS-EXE-001 → @DESIGN:PYINSTALLER-BUILD-001 → @TASK:EXE-PACKAGING-001 → @TEST:EXE-VALIDATION-001
"""

import os
import sys
from pathlib import Path

# Get the current directory (project root)
WORKPATH = os.getcwd()
SRCPATH = os.path.join(WORKPATH, 'src')

# Define paths
block_cipher = None

# Add src to Python path for analysis
sys.path.insert(0, SRCPATH)

a = Analysis(
    # Entry point: unified CLI module
    ['src/moai_adk/cli.py'],

    pathex=[SRCPATH],

    binaries=[
        # No external binaries needed for core functionality
    ],

    datas=[
        # Include all template and resource files
        ('src/moai_adk/resources', 'moai_adk/resources'),

        # Include version file
        ('src/moai_adk/utils/_version.py', 'moai_adk/utils'),

        # Include any other necessary data files
        ('README.md', '.'),
        ('LICENSE', '.'),
    ],

    hiddenimports=[
        # Core dependencies that might not be auto-detected
        'sqlite3',
        'colorama',
        'click',
        'toml',
        'yaml',
        'jinja2',
        'watchdog',
        'gitpython',

        # MoAI-ADK modules that might be imported dynamically
        'moai_adk.cli.commands',
        'moai_adk.cli.wizard',
        'moai_adk.cli.helpers',
        'moai_adk.core.config_manager',
        'moai_adk.core.directory_manager',
        'moai_adk.core.file_manager',
        'moai_adk.core.git_manager',
        'moai_adk.install.installer',
        'moai_adk.install.post_install',
        'moai_adk.utils.logger',
        'moai_adk.utils.version',

        # Template rendering dependencies
        'jinja2.ext',

        # YAML dependencies for template processing
        'yaml.loader',
        'yaml.dumper',

        # Git dependencies
        'git.repo',
        'git.config',

        # Watchdog platform-specific modules
        'watchdog.observers',
        'watchdog.events',
    ],

    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],

    excludes=[
        # Exclude heavy dependencies not needed for core functionality
        'tkinter',           # GUI toolkit
        'matplotlib',        # Plotting library
        'PIL',              # Image processing
        'numpy',            # Scientific computing
        'pandas',           # Data analysis
        'scipy',            # Scientific computing
        'IPython',          # Interactive Python
        'jupyter',          # Jupyter notebooks
        'notebook',         # Jupyter notebook server
        'pytest',           # Testing framework
        'mypy',             # Type checking
        'black',            # Code formatting
        'isort',            # Import sorting
        'flake8',           # Linting
        'bandit',           # Security linting
        'ruff',             # Fast linting

        # Development tools
        'setuptools',
        'pip',
        'wheel',
        'twine',

        # Documentation tools
        'sphinx',
        'mkdocs',

        # Test directories
        'tests',
        'test',

        # Legacy modules
        'legacy',

        # Platform-specific modules we don't need
        '_tkinter',
        'tcl',
        'tk',
    ],

    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=block_cipher,
    noarchive=False,
)

# Filter out test files and development dependencies
a.datas = [x for x in a.datas if not any(
    exclude in x[0].lower() for exclude in [
        'test', 'pytest', 'mypy', 'black', 'isort', 'flake8', 'bandit', 'ruff',
        'sphinx', 'mkdocs', '__pycache__', '.pyc', '.git', 'legacy'
    ]
)]

pyz = PYZ(a.pure, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.zipfiles,
    a.datas,
    [],
    name='moai-adk',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,  # Enable UPX compression to reduce file size
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,  # Console application
    disable_windowed_traceback=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,

    # Windows-specific options
    version=None,  # Version info (can be added later)
    uac_admin=False,  # Don't require admin privileges

    # Icon (if available)
    icon=None,  # Can be set to 'assets/moai.ico' if icon file exists
)

# Optional: Create a distribution folder
# Uncomment the following if you want a folder distribution instead of single file
# coll = COLLECT(
#     exe,
#     a.binaries,
#     a.zipfiles,
#     a.datas,
#     strip=False,
#     upx=True,
#     upx_exclude=[],
#     name='moai-adk'
# )