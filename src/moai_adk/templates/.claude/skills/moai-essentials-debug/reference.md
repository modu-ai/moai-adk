# MoAI Essentials Debug - Language Reference Guide

## Table of Contents
1. [Systems Programming](#systems-programming)
2. [JVM Ecosystem](#jvm-ecosystem)
3. [Scripting Languages](#scripting-languages)
4. [Web & Mobile](#web--mobile)
5. [Functional & Concurrency](#functional--concurrency)
6. [Enterprise & Data](#enterprise--data)
7. [Container & Distributed Systems](#container--distributed-systems)

---

## Systems Programming

### C

#### Installation
```bash
# Linux/macOS (usually pre-installed)
which gdb

# macOS with Homebrew
brew install gdb

# Ubuntu/Debian
sudo apt-get install gdb

# LLDB alternative
brew install llvm  # macOS
```

#### CLI Usage
```bash
# Basic debugging
gdb ./program
(gdb) break main
(gdb) run arg1 arg2
(gdb) step         # Step into
(gdb) next         # Step over
(gdb) continue     # Continue execution
(gdb) print var    # Inspect variable
(gdb) backtrace    # Stack trace

# Debug with core dump
gdb ./program core

# Attach to running process
gdb -p <pid>
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "C: Debug",
      "type": "cppdbg",
      "request": "launch",
      "program": "${workspaceFolder}/build/myapp",
      "args": [],
      "stopAtEntry": false,
      "cwd": "${workspaceFolder}",
      "environment": [],
      "externalConsole": false,
      "MIMode": "gdb",
      "setupCommands": [
        {
          "description": "Enable pretty-printing",
          "text": "-enable-pretty-printing",
          "ignoreFailures": true
        }
      ]
    }
  ]
}
```

#### Common Commands
```bash
# Watchpoints
watch variable_name    # Break when variable changes

# Conditional breakpoints
break line_number if condition

# Memory inspection
x/10x address         # Display 10 hex values at address
```

---

### C++

#### Installation
Same as C (gdb/lldb), plus sanitizers:
```bash
# Compile with AddressSanitizer
g++ -fsanitize=address -g program.cpp -o program

# Valgrind for memory analysis
sudo apt-get install valgrind
brew install valgrind  # macOS
```

#### CLI Usage
```bash
# GDB with C++ pretty-printers
gdb ./program
(gdb) set print pretty on
(gdb) set print object on

# LLDB
lldb ./program
(lldb) breakpoint set -n main
(lldb) process launch -- arg1 arg2
(lldb) thread backtrace
```

#### Valgrind Usage
```bash
# Memory leak detection
valgrind --leak-check=full --show-leak-kinds=all ./program

# Callgrind profiling
valgrind --tool=callgrind ./program
callgrind_annotate callgrind.out.<pid>
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "C++: Debug",
      "type": "cppdbg",
      "request": "launch",
      "program": "${workspaceFolder}/build/myapp",
      "args": [],
      "stopAtEntry": false,
      "cwd": "${workspaceFolder}",
      "environment": [],
      "externalConsole": false,
      "MIMode": "gdb",
      "miDebuggerPath": "/usr/bin/gdb"
    }
  ]
}
```

---

### Rust

#### Installation
```bash
# Installed with rustup (automatic)
rustup component add rust-src
rustup component add rust-analyzer

# VSCode extension
code --install-extension vadimcn.vscode-lldb
```

#### CLI Usage
```bash
# Run with backtrace
RUST_BACKTRACE=1 cargo run
RUST_BACKTRACE=full cargo run

# Debug build
cargo build
rust-lldb target/debug/myapp

# In LLDB
(lldb) b main
(lldb) r
(lldb) bt  # Backtrace
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "lldb",
      "request": "launch",
      "name": "Debug executable",
      "cargo": {
        "args": [
          "build",
          "--bin=myapp",
          "--package=myapp"
        ],
        "filter": {
          "name": "myapp",
          "kind": "bin"
        }
      },
      "args": [],
      "cwd": "${workspaceFolder}"
    },
    {
      "type": "lldb",
      "request": "launch",
      "name": "Debug unit tests",
      "cargo": {
        "args": [
          "test",
          "--no-run",
          "--lib",
          "--package=myapp"
        ],
        "filter": {
          "name": "myapp",
          "kind": "lib"
        }
      },
      "args": [],
      "cwd": "${workspaceFolder}"
    }
  ]
}
```

---

### Go

#### Installation
```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Verify
dlv version

# VSCode extension
code --install-extension golang.go
```

#### CLI Usage
```bash
# Debug main package
dlv debug

# Debug with arguments
dlv debug -- --config=dev.json

# Debug tests
dlv test

# Attach to running process
dlv attach <pid>

# Remote debugging
dlv debug --headless --listen=:2345 --api-version=2

# In Delve
(dlv) break main.main
(dlv) continue
(dlv) print myVar
(dlv) goroutines           # List all goroutines
(dlv) goroutine 5 bt       # Stack trace for goroutine 5
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Package",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}",
      "env": {},
      "args": []
    },
    {
      "name": "Launch File",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${file}"
    },
    {
      "name": "Attach to Process",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": "${command:pickProcess}"
    },
    {
      "name": "Connect to Remote",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "remotePath": "/app",
      "port": 2345,
      "host": "localhost"
    }
  ]
}
```

---

## JVM Ecosystem

### Java

#### Installation
```bash
# JDB comes with JDK
java -version
jdb -version

# IntelliJ IDEA (recommended for GUI debugging)
# Download from jetbrains.com
```

#### CLI Usage
```bash
# Start application with debug port
java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005 -jar app.jar

# Connect with JDB
jdb -attach localhost:5005

# JDB commands
> stop at ClassName:lineNumber
> run
> step
> next
> print variable
> locals
> where    # Stack trace
```

#### Remote Debugging Setup
```bash
# Server side
java -Xdebug -Xrunjdwp:transport=dt_socket,server=y,suspend=n,address=*:5005 -jar app.jar

# Or with modern syntax
java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005 -jar app.jar
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "java",
      "name": "Debug (Launch)",
      "request": "launch",
      "mainClass": "com.example.Main",
      "projectName": "myapp"
    },
    {
      "type": "java",
      "name": "Debug (Attach)",
      "request": "attach",
      "hostName": "localhost",
      "port": 5005
    }
  ]
}
```

#### IntelliJ IDEA Setup
1. Run → Edit Configurations
2. Add New Configuration → Remote JVM Debug
3. Set host and port
4. Set breakpoints and click Debug

---

### Kotlin

#### Installation
Same as Java (uses JVM debug infrastructure)

#### CLI Usage
```bash
# Gradle with debug
./gradlew run --debug-jvm

# Maven with debug
mvn spring-boot:run -Dspring-boot.run.jvmArguments="-Xdebug -Xrunjdwp:transport=dt_socket,server=y,suspend=n,address=*:5005"
```

#### Coroutine Debugging
```kotlin
// Enable coroutine debugging
System.setProperty("kotlinx.coroutines.debug", "on")

// In IntelliJ IDEA:
// Run → Edit Configurations → VM Options:
-Dkotlinx.coroutines.debug
```

#### VSCode Configuration
Same as Java, works with Kotlin bytecode

---

### Scala

#### Installation
```bash
# sbt with debug support
sbt

# Metals language server (VSCode)
code --install-extension scalameta.metals
```

#### CLI Usage
```bash
# sbt with debug port
sbt -jvm-debug 5005

# In sbt shell
sbt:myapp> ~run  # Continuous compilation + run
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "scala",
      "request": "launch",
      "name": "Launch Scala App",
      "mainClass": "com.example.Main",
      "args": [],
      "jvmOptions": []
    }
  ]
}
```

---

### Clojure

#### Installation
```bash
# Leiningen
brew install leiningen  # macOS
# Ubuntu: see leiningen.org

# CIDER (Emacs)
# Add to ~/.emacs: (require 'cider)

# Cursive (IntelliJ)
# Install from JetBrains Marketplace
```

#### REPL-Based Debugging
```clojure
; Load file in REPL
(load-file "src/myapp/core.clj")

; Add breakpoints
(require '[clojure.tools.trace :as trace])
(trace/trace-ns 'myapp.core)

; Call function to see trace
(myapp.core/my-function arg1 arg2)

; Inspect vars
(def result (my-function arg1))
(println result)
```

#### Leiningen Debug Plugin
```bash
# project.clj
:plugins [[cider/cider-nrepl "0.30.0"]]

# Start REPL with debug
lein repl

# Connect to running process
lein repl :connect 5005
```

---

## Scripting Languages

### Python

#### Installation
```bash
# pdb (built-in, no installation needed)
python3 -m pdb

# Enhanced debuggers
pip install pudb ipdb debugpy

# VSCode extension
code --install-extension ms-python.python
```

#### CLI Usage
```bash
# pdb
python -m pdb script.py
(Pdb) break script.py:42
(Pdb) continue
(Pdb) step
(Pdb) next
(Pdb) print variable
(Pdb) list      # Show code context
(Pdb) where     # Stack trace

# pudb (TUI)
pudb script.py

# ipdb (IPython-enhanced)
ipdb script.py
```

#### Code-Level Debugging
```python
# Insert breakpoint
import pdb; pdb.set_trace()

# Python 3.7+
breakpoint()

# Remote debugging with debugpy
import debugpy
debugpy.listen(5678)
print("Waiting for debugger attach...")
debugpy.wait_for_client()
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Python: Current File",
      "type": "debugpy",
      "request": "launch",
      "program": "${file}",
      "console": "integratedTerminal",
      "justMyCode": false
    },
    {
      "name": "Python: Remote Attach",
      "type": "debugpy",
      "request": "attach",
      "connect": {
        "host": "localhost",
        "port": 5678
      },
      "pathMappings": [
        {
          "localRoot": "${workspaceFolder}",
          "remoteRoot": "/app"
        }
      ]
    },
    {
      "name": "Python: Django",
      "type": "debugpy",
      "request": "launch",
      "program": "${workspaceFolder}/manage.py",
      "args": ["runserver", "--noreload"],
      "django": true
    },
    {
      "name": "Python: Flask",
      "type": "debugpy",
      "request": "launch",
      "module": "flask",
      "env": {
        "FLASK_APP": "app.py",
        "FLASK_DEBUG": "1"
      },
      "args": ["run", "--no-debugger"],
      "jinja": true
    }
  ]
}
```

#### Async Debugging
```python
import asyncio
import pdb

async def debug_async():
    await asyncio.sleep(1)
    breakpoint()  # Works in async context
    await asyncio.sleep(1)

asyncio.run(debug_async())
```

---

### Ruby

#### Installation
```bash
# Ruby 3.2+ (built-in debug gem)
ruby -v

# For older versions
gem install debug

# Alternative debuggers
gem install byebug
gem install pry-byebug

# VSCode extension
code --install-extension Shopify.ruby-lsp
```

#### CLI Usage
```bash
# debug gem
ruby -r debug script.rb

# In code
require 'debug'
binding.break

# byebug
byebug script.rb

# Commands
(rdbg) break script.rb:42
(rdbg) continue
(rdbg) step
(rdbg) next
(rdbg) display variable
(rdbg) backtrace
```

#### Rails Debugging
```ruby
# In controller/view
binding.break  # Ruby 3.2+
byebug         # Older versions

# Start Rails with debugger
rails server
# Trigger breakpoint, debugger opens in terminal
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Ruby",
      "type": "rdbg",
      "request": "launch",
      "script": "${file}",
      "args": [],
      "useBundler": true
    },
    {
      "name": "Debug Rails",
      "type": "rdbg",
      "request": "launch",
      "command": "rails",
      "script": "server",
      "args": [],
      "useBundler": true
    }
  ]
}
```

---

### PHP

#### Installation
```bash
# Xdebug (most common)
# Ubuntu/Debian
sudo apt-get install php-xdebug

# macOS with Homebrew
brew install php
pecl install xdebug

# Configure php.ini
zend_extension=xdebug.so
xdebug.mode=debug
xdebug.start_with_request=yes
xdebug.client_port=9003

# VSCode extension
code --install-extension xdebug.php-debug
```

#### CLI Usage
```bash
# phpdbg (built-in CLI debugger)
phpdbg -qrr script.php

# phpdbg commands
phpdbg> break file.php:42
phpdbg> run
phpdbg> step
phpdbg> print $variable

# Xdebug from command line
php -dxdebug.mode=debug -dxdebug.start_with_request=yes script.php
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Listen for Xdebug",
      "type": "php",
      "request": "launch",
      "port": 9003,
      "pathMappings": {
        "/var/www/html": "${workspaceFolder}"
      }
    },
    {
      "name": "Launch current script",
      "type": "php",
      "request": "launch",
      "program": "${file}",
      "cwd": "${workspaceFolder}",
      "port": 9003
    }
  ]
}
```

#### Laravel Debugging
```php
// Use Laravel Debugbar
composer require barryvdh/laravel-debugbar --dev

// Or Telescope for advanced debugging
composer require laravel/telescope --dev
php artisan telescope:install
```

---

### Lua

#### Installation
```bash
# ZeroBrane Studio (GUI)
# Download from studio.zerobrane.com

# MobDebug
luarocks install mobdebug

# VSCode extension
code --install-extension actboy168.lua-debug
```

#### CLI Usage
```lua
-- MobDebug
local mobdebug = require('mobdebug')
mobdebug.start()  -- Start debugging session

-- In code
mobdebug.pause()  -- Breakpoint
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Lua",
      "type": "lua",
      "request": "launch",
      "program": "${file}"
    }
  ]
}
```

---

### Shell (Bash)

#### Built-in Debugging
```bash
# Trace mode
bash -x script.sh

# Or in script
#!/bin/bash
set -x  # Enable tracing
# ... code ...
set +x  # Disable tracing

# Custom trace format
PS4='+(${BASH_SOURCE}:${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
set -x
```

#### Debugging Options
```bash
# Exit on error
set -e

# Exit on undefined variable
set -u

# Exit on pipe failure
set -o pipefail

# Combine (common)
set -euo pipefail
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Bash-Debug",
      "type": "bashdb",
      "request": "launch",
      "program": "${file}",
      "cwd": "${workspaceFolder}",
      "args": [],
      "internalConsoleOptions": "openOnSessionStart"
    }
  ]
}
```

---

## Web & Mobile

### JavaScript (Node.js)

#### Installation
```bash
# Built-in with Node.js
node --version

# VSCode (built-in debugger)
```

#### CLI Usage
```bash
# Start with inspector
node --inspect script.js

# Start with inspector and break at first line
node --inspect-brk script.js

# Connect with Chrome DevTools
# Navigate to chrome://inspect

# Built-in REPL debugging
node inspect script.js
debug> cont
debug> next
debug> step
debug> repl  # Evaluate expressions
```

#### Code-Level Debugging
```javascript
// Trigger debugger breakpoint
debugger;

// When Chrome DevTools is open, execution will pause here
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "node",
      "request": "launch",
      "name": "Launch Program",
      "skipFiles": ["<node_internals>/**"],
      "program": "${workspaceFolder}/app.js"
    },
    {
      "type": "node",
      "request": "attach",
      "name": "Attach",
      "port": 9229,
      "skipFiles": ["<node_internals>/**"]
    },
    {
      "name": "Jest Tests",
      "type": "node",
      "request": "launch",
      "runtimeArgs": [
        "--inspect-brk",
        "${workspaceFolder}/node_modules/.bin/jest",
        "--runInBand"
      ],
      "console": "integratedTerminal"
    }
  ]
}
```

---

### TypeScript

#### Installation
```bash
# TypeScript + ts-node
npm install -D typescript ts-node @types/node

# VSCode (built-in support)
```

#### CLI Usage
```bash
# Compile with source maps
npx tsc --sourceMap

# Debug with Node.js
node --inspect -r ts-node/register script.ts

# Or use ts-node directly
ts-node --inspect script.ts
```

#### tsconfig.json
```json
{
  "compilerOptions": {
    "sourceMap": true,
    "outDir": "./dist",
    "rootDir": "./src"
  }
}
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "node",
      "request": "launch",
      "name": "TypeScript: Current File",
      "runtimeArgs": ["-r", "ts-node/register"],
      "args": ["${file}"],
      "cwd": "${workspaceFolder}",
      "protocol": "inspector"
    },
    {
      "type": "node",
      "request": "launch",
      "name": "TypeScript: Jest",
      "program": "${workspaceFolder}/node_modules/.bin/jest",
      "args": ["--runInBand"],
      "console": "integratedTerminal"
    }
  ]
}
```

---

### Dart/Flutter

#### Installation
```bash
# Install Flutter SDK
# flutter.dev/docs/get-started/install

# VSCode extension
code --install-extension Dart-Code.flutter
```

#### CLI Usage
```bash
# Run with debugger enabled
flutter run --observe

# Attach to running app
flutter attach

# Dart VM Observatory
# Opens in browser automatically

# Hot reload
r  # Reload
R  # Hot restart
```

#### VSCode Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Flutter: Launch",
      "request": "launch",
      "type": "dart",
      "program": "lib/main.dart"
    },
    {
      "name": "Flutter: Attach",
      "request": "attach",
      "type": "dart"
    },
    {
      "name": "Dart: Tests",
      "request": "launch",
      "type": "dart",
      "program": "test/"
    }
  ]
}
```

#### Flutter DevTools
```bash
# Open DevTools
flutter pub global activate devtools
flutter pub global run devtools

# Or from running app
# Click DevTools link in terminal
```

---

### Swift

#### Installation
```bash
# Xcode (macOS only)
xcode-select --install

# LLDB comes with Xcode
lldb --version

# VSCode extension
code --install-extension vknabel.vscode-swift-development-environment
```

#### CLI Usage
```bash
# Build and debug
swift build
lldb .build/debug/MyApp

# In LLDB
(lldb) breakpoint set -n main
(lldb) run
(lldb) thread backtrace
(lldb) frame variable
```

#### Xcode Debugging
1. Set breakpoints in gutter
2. Run (⌘+R) or Debug (⌘+Y)
3. Use LLDB console in bottom pane
4. View hierarchy debugger: Debug → View Debugging → Capture View Hierarchy

#### LLDB Configuration for Swift
```bash
# ~/.lldbinit
command script import lldb.macosx.heap
settings set target.process.follow-fork-mode child
settings set target.x86-disassembly-flavor intel
```

---

## Functional & Concurrency

### Haskell

#### Installation
```bash
# GHC with GHCi debugger
ghcup install ghc
ghci --version

# VSCode extension
code --install-extension haskell.haskell
```

#### GHCi Debugger
```bash
# Load module with debugging
ghci -fbreak-on-exception Main.hs

# Set breakpoint
:break main
:break MyModule.myFunction

# Run
:main

# Step through
:step
:steplocal  # Don't step into libraries
:trace myFunction args

# Inspect
:show breaks
:show bindings
:print variable
```

#### Profiling
```bash
# Compile with profiling
ghc -prof -fprof-auto Main.hs

# Run with profiling
./Main +RTS -p

# View report
cat Main.prof

# Heap profiling
./Main +RTS -hy
hp2ps Main.hp
open Main.ps
```

---

### Elixir

#### Installation
```bash
# Elixir includes debugging tools
elixir --version

# VSCode extension
code --install-extension JakeBecker.elixir-ls
```

#### IEx Debugger
```elixir
# Start IEx
iex -S mix

# Load debugger
:debugger.start()

# Set breakpoint
:int.break(MyModule, :my_function, 2)  # arity 2

# Call function, GUI debugger opens
MyModule.my_function(arg1, arg2)
```

#### Observer
```elixir
# Start observer (live system inspection)
:observer.start()

# Trace calls
:sys.trace(pid, true)

# Get state
:sys.get_status(pid)
```

#### Phoenix LiveDashboard
```elixir
# Add to mix.exs
{:phoenix_live_dashboard, "~> 0.8"}

# Add to router
live_dashboard "/dashboard"

# Navigate to localhost:4000/dashboard
```

---

### Julia

#### Installation
```bash
# Install Julia
# julialang.org/downloads

# Install debugger
julia> using Pkg
julia> Pkg.add("Debugger")
julia> Pkg.add("Infiltrator")

# VSCode extension
code --install-extension julialang.language-julia
```

#### CLI Usage
```julia
# Debugger.jl
using Debugger

@enter my_function(args)

# Commands
n  # next
s  # step into
finish  # finish current function
bt  # backtrace
```

#### Infiltrator.jl (Lightweight)
```julia
using Infiltrator

function my_function(x)
    y = x * 2
    @infiltrate  # Breakpoint
    z = y + 1
    return z
end

# Run function, drops into REPL at breakpoint
```

#### VSCode Debugging
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "julia",
      "request": "launch",
      "name": "Run Julia",
      "program": "${file}",
      "stopOnEntry": false,
      "cwd": "${workspaceFolder}"
    }
  ]
}
```

---

### R

#### Installation
```bash
# R comes with debugging tools
R --version

# RStudio (recommended)
# Download from rstudio.com

# VSCode extension
code --install-extension REditorSupport.r
```

#### CLI Usage
```R
# Set breakpoint
debug(myFunction)
myFunction(args)  # Enters debugger

# Or use browser()
myFunction <- function(x) {
  browser()  # Breakpoint
  result <- x * 2
  return(result)
}

# Commands
n  # next
s  # step into
c  # continue
Q  # quit
```

#### RStudio Debugging
1. Set breakpoints by clicking line numbers
2. Source file with Debug button
3. Use Debug menu: Debug → On Error → Break in Code

#### Traceback
```R
# After error
traceback()

# Recover mode (interactive debugging)
options(error = recover)
myFunction()  # On error, choose frame to debug
```

---

## Enterprise & Data

### C#

#### Installation
```bash
# .NET SDK
# dotnet.microsoft.com/download

# VSCode extension
code --install-extension ms-dotnettools.csharp

# Visual Studio (full IDE, Windows)
# visualstudio.microsoft.com
```

#### CLI Usage
```bash
# Run with debugging
dotnet run

# In VSCode, use integrated debugger
```

#### Visual Studio Debugging
1. Set breakpoints (F9)
2. Start debugging (F5)
3. Step over (F10), Step into (F11)
4. Immediate Window for expression evaluation

#### Remote Debugging (vsdbg)
```bash
# Install vsdbg on remote machine
curl -sSL https://aka.ms/getvsdbgsh | bash /dev/stdin -v latest -l ~/vsdbg

# VSCode configuration
{
  "name": ".NET Core Attach (Remote)",
  "type": "coreclr",
  "request": "attach",
  "processId": "${command:pickRemoteProcess}",
  "pipeTransport": {
    "pipeCwd": "${workspaceFolder}",
    "pipeProgram": "ssh",
    "pipeArgs": ["user@remote-host"],
    "debuggerPath": "~/vsdbg/vsdbg"
  }
}
```

---

### SQL

#### PostgreSQL Debugging
```sql
-- Enable verbose error messages
\set VERBOSITY verbose

-- Explain query execution
EXPLAIN ANALYZE SELECT * FROM users WHERE id = 1;

-- Check slow queries
SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;

-- Enable query logging in postgresql.conf
log_statement = 'all'
log_duration = on
log_min_duration_statement = 100  -- ms
```

#### MySQL Debugging
```sql
-- Show warnings
SHOW WARNINGS;

-- Explain query
EXPLAIN SELECT * FROM users WHERE id = 1;

-- Enable slow query log (in my.cnf)
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 1
```

#### pgAdmin Debugger
1. Install pldbgapi extension
2. Right-click function in pgAdmin
3. Select "Debugging" → "Set breakpoint"
4. Execute function, debugger opens

---

## Container & Distributed Systems

### Docker Debugging

#### Interactive Shell
```bash
# Exec into running container
docker exec -it <container-name> /bin/sh

# Run with interactive shell
docker run -it myimage /bin/bash
```

#### Remote Debugging
```dockerfile
# Python example
FROM python:3.11
RUN pip install debugpy
EXPOSE 5678
CMD ["python", "-m", "debugpy", "--listen", "0.0.0.0:5678", "--wait-for-client", "app.py"]
```

```bash
# Run container
docker run -p 5678:5678 myapp

# Attach from VSCode (see Python Remote Attach config above)
```

#### Log Analysis
```bash
# Stream logs
docker logs -f <container>

# Logs with timestamps
docker logs --timestamps <container>

# Last N lines
docker logs --tail 100 <container>
```

---

### Kubernetes Debugging

#### Port Forwarding
```bash
# Forward debugger port
kubectl port-forward pod/myapp-pod 5005:5005

# Forward service
kubectl port-forward svc/myapp-service 5005:5005
```

#### Exec into Pod
```bash
# Get shell
kubectl exec -it myapp-pod -- /bin/bash

# Specific container in multi-container pod
kubectl exec -it myapp-pod -c sidecar -- /bin/sh
```

#### Ephemeral Containers (K8s 1.23+)
```bash
# Debug with temporary container
kubectl debug -it myapp-pod --image=busybox --target=myapp

# Copy pod and add debug tools
kubectl debug myapp-pod -it --copy-to=myapp-debug --container=debug-tools --image=ubuntu
```

#### Log Streaming
```bash
# Stream logs
kubectl logs -f deployment/myapp

# All containers
kubectl logs -f deployment/myapp --all-containers=true

# Previous container instance (after crash)
kubectl logs myapp-pod --previous
```

---

### OpenTelemetry Distributed Tracing

See SKILL.md for detailed code examples.

#### Setup Checklist
- [ ] Install OpenTelemetry SDK for your language
- [ ] Configure OTLP exporter (gRPC or HTTP)
- [ ] Set up trace collector (Jaeger, Zipkin, or vendor)
- [ ] Instrument key operations with spans
- [ ] Add span attributes for context
- [ ] Propagate trace context across services

#### Common Endpoints
```bash
# OTLP gRPC (default)
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# OTLP HTTP
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# Jaeger UI
http://localhost:16686
```

---

This reference guide provides installation, CLI usage, and VSCode configuration for all 23 supported languages plus container and distributed system debugging.
