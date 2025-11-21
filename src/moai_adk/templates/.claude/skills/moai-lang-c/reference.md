# moai-lang-c - CLI Reference & Tools

_Last updated: 2025-11-21_

## Compiler Setup

### GCC Installation & Usage

```bash
# Install GCC (macOS)
brew install gcc

# Install GCC (Ubuntu/Debian)
sudo apt-get install gcc

# Compile with warnings enabled
gcc -Wall -Wextra -std=c23 -o program program.c

# Compile with optimization
gcc -O2 -Wall -o program program.c

# Compile with debugging symbols
gcc -g -Wall -o program program.c

# Position-independent code (PIE)
gcc -fPIE -o program program.c

# Static linking
gcc -static -o program program.c
```

### Clang Usage

```bash
# Clang typically offers better error messages than GCC
clang -Wall -Wextra -std=c23 -o program program.c

# Address Sanitizer (detect memory errors)
clang -fsanitize=address -g -o program program.c
./program  # Will report memory issues

# Memory Sanitizer
clang -fsanitize=memory -g -o program program.c

# Undefined Behavior Sanitizer
clang -fsanitize=undefined -o program program.c
```

## Debugging with GDB

```bash
# Compile with debugging symbols first
gcc -g -o program program.c

# Start debugger
gdb ./program

# Common GDB commands
gdb> break main              # Set breakpoint at main
gdb> break file.c:42         # Set breakpoint at line 42
gdb> run                     # Start execution
gdb> next                    # Step over function
gdb> step                    # Step into function
gdb> print variable_name     # Print variable value
gdb> continue                # Continue execution
gdb> backtrace              # Show call stack
gdb> quit                    # Exit debugger

# Run with command-line arguments
gdb --args ./program arg1 arg2
```

## Memory Analysis with Valgrind

```bash
# Install Valgrind
brew install valgrind         # macOS
sudo apt-get install valgrind # Ubuntu

# Basic memory leak check
valgrind ./program

# Detailed leak report
valgrind --leak-check=full --show-leak-kinds=all ./program

# Track memory allocations
valgrind --track-origins=yes ./program

# Generate HTML report
valgrind --html=yes ./program
```

## Build Systems

### Make & Makefile

```makefile
# Basic Makefile for C project
CC = gcc
CFLAGS = -Wall -Wextra -std=c23 -g
LDFLAGS =

SOURCES = main.c utils.c
OBJECTS = $(SOURCES:.c=.o)
TARGET = program

all: $(TARGET)

$(TARGET): $(OBJECTS)
	$(CC) $(LDFLAGS) -o $@ $^

%.o: %.c
	$(CC) $(CFLAGS) -c $<

clean:
	rm -f $(OBJECTS) $(TARGET)

.PHONY: all clean
```

### CMake

```cmake
# CMakeLists.txt
cmake_minimum_required(VERSION 3.10)
project(MyProject C)

set(CMAKE_C_STANDARD 23)
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -Wall -Wextra")

add_executable(program main.c utils.c)

# Link against pthread if needed
find_package(Threads REQUIRED)
target_link_libraries(program Threads::Threads)
```

```bash
# Build with CMake
mkdir build
cd build
cmake ..
make

# Or with specific compiler
cmake -DCMAKE_C_COMPILER=clang ..
make
```

## Code Quality Tools

### CPPCheck (Static Analysis)

```bash
# Install
brew install cppcheck

# Check all C files
cppcheck .

# Generate HTML report
cppcheck --html --html-report=report .

# Specific severity
cppcheck --enable=all src/
```

### Formatting with clang-format

```bash
# Install
brew install clang-format

# Format file in-place
clang-format -i program.c

# Check formatting without modifying
clang-format -i --dry-run program.c

# Format entire directory
find . -name "*.c" -o -name "*.h" | xargs clang-format -i
```

## C23 Standard Features

### Key Improvements in C23

```c
// 1. Auto keyword for type inference
auto x = 5;  // int

// 2. _Bool renamed to bool
bool flag = true;

// 3. Undefined behavior sanitizers
// Compile with: clang -fsanitize=undefined

// 4. Bit-field improvements
struct Flags {
    bool flag1 : 1;
    bool flag2 : 1;
    unsigned padding : 6;
};

// 5. typeof operator
typeof(variable) copy = variable;
```

## Performance Profiling

### Using perf (Linux)

```bash
# Record performance data
perf record ./program

# Report statistics
perf report

# CPU flame graph
perf record -g ./program
perf script | stackcollapse-perf.pl | flamegraph.pl
```

## Tool Versions (2025-11-21)

- **GCC**: 14.2.0
- **Clang**: 19.1.7
- **GDB**: 14.1
- **Valgrind**: 3.23.0
- **cppcheck**: 2.16.0
- **CMake**: 3.31.0
- **Make**: 4.4.1

---

_For examples and advanced patterns, see SKILL.md_
