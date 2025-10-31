"""
Frontend Plugin Commands - /init-react, /setup-state, /setup-testing implementations

@CODE:FRONTEND-INIT-CMD-001:COMMANDS
"""

from dataclasses import dataclass
from pathlib import Path
from typing import Optional, List
import json
from datetime import datetime


# @CODE:FRONTEND-COMMAND-RESULT-001:RESULT
@dataclass
class CommandResult:
    """Result object for command execution"""
    success: bool
    project_dir: Optional[Path]
    files_created: List[str]
    message: str
    error: Optional[str] = None


class InitReactCommand:
    """
    /init-react command implementation

    Initializes React project with standard structure
    """

    VALID_FRAMEWORKS = ["react", "vite", "nextjs"]
    MIN_PROJECT_NAME_LENGTH = 3
    MAX_PROJECT_NAME_LENGTH = 50

    def validate_project_name(self, project_name: str) -> bool:
        """
        Validate project name format

        @CODE:FRONTEND-VALIDATE-NAME-001:VALIDATION
        """
        if not project_name:
            raise ValueError("Project name cannot be empty")

        if len(project_name) < self.MIN_PROJECT_NAME_LENGTH:
            raise ValueError(
                f"Project name must be at least {self.MIN_PROJECT_NAME_LENGTH} characters"
            )

        if len(project_name) > self.MAX_PROJECT_NAME_LENGTH:
            raise ValueError(
                f"Project name cannot exceed {self.MAX_PROJECT_NAME_LENGTH} characters"
            )

        if not all(c.islower() or c.isdigit() or c == "-" for c in project_name):
            raise ValueError(
                "Project name must contain only lowercase letters, numbers, and hyphens"
            )

        if project_name.startswith("-") or project_name.endswith("-"):
            raise ValueError("Project name cannot start or end with a hyphen")

        if "--" in project_name:
            raise ValueError("Project name cannot contain consecutive hyphens")

        return True

    def validate_framework(self, framework: str) -> bool:
        """
        Validate framework name

        @CODE:FRONTEND-VALIDATE-FRAMEWORK-001:VALIDATION
        """
        if framework not in self.VALID_FRAMEWORKS:
            raise ValueError(
                f"Invalid framework: {framework}\nSupported: {', '.join(self.VALID_FRAMEWORKS)}"
            )
        return True

    def create_project_directory(self, output_dir: Path, project_name: str) -> Path:
        """
        Create React project directory

        @CODE:FRONTEND-CREATE-DIR-001:DIRECTORY
        """
        project_dir = output_dir / project_name

        if project_dir.exists():
            raise FileExistsError(f"Project already exists: {project_dir}")

        project_dir.mkdir(parents=True, exist_ok=False)
        return project_dir

    def create_react_structure(
        self,
        project_dir: Path,
        framework: str = "react",
        use_typescript: bool = False,
        include_tailwind: bool = False,
        include_eslint: bool = False
    ) -> List[str]:
        """
        Create React project structure

        @CODE:FRONTEND-REACT-STRUCTURE-001:STRUCTURE
        """
        files_created = []

        # Create src directory
        src_dir = project_dir / "src"
        src_dir.mkdir(exist_ok=False)

        # Create App component
        app_extension = ".tsx" if use_typescript else ".jsx"
        app_content = """import React from 'react'
import './App.css'

function App() {
  return (
    <div className="app">
      <h1>Welcome to React</h1>
      <p>Start editing to see some magic happen!</p>
    </div>
  )
}

export default App
"""
        app_file = src_dir / f"App{app_extension}"
        app_file.write_text(app_content)
        files_created.append(str(app_file.relative_to(project_dir.parent)))

        # Create index file
        index_extension = ".tsx" if use_typescript else ".jsx"
        index_content = """import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App"""
        index_content += app_extension + """'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
"""
        index_file = src_dir / f"main{index_extension}"
        index_file.write_text(index_content)
        files_created.append(str(index_file.relative_to(project_dir.parent)))

        # Create CSS files
        css_content = """* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.app {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
}
"""
        (src_dir / "index.css").write_text(css_content)
        files_created.append(str((src_dir / "index.css").relative_to(project_dir.parent)))

        (src_dir / "App.css").write_text("/* App styles */")
        files_created.append(str((src_dir / "App.css").relative_to(project_dir.parent)))

        # Create public directory
        public_dir = project_dir / "public"
        public_dir.mkdir(exist_ok=False)

        index_html = """<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>React App</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.jsx"></script>
  </body>
</html>
"""
        (public_dir / "index.html").write_text(index_html)
        files_created.append(str((public_dir / "index.html").relative_to(project_dir.parent)))

        # Create package.json
        deps = {
            "react": "^18.2.0",
            "react-dom": "^18.2.0"
        }
        dev_deps = {
            "vite": "^5.0.0",
            "@vitejs/plugin-react": "^4.2.0"
        }

        if use_typescript:
            dev_deps["typescript"] = "^5.3.0"
            dev_deps["@types/react"] = "^18.2.0"
            dev_deps["@types/react-dom"] = "^18.2.0"

        if include_tailwind:
            deps["tailwindcss"] = "^3.4.0"
            dev_deps["postcss"] = "^8.4.0"

        if include_eslint:
            dev_deps["eslint"] = "^8.54.0"
            dev_deps["eslint-plugin-react"] = "^7.33.0"

        package_json = {
            "name": project_dir.name,
            "private": True,
            "version": "0.0.1",
            "type": "module",
            "scripts": {
                "dev": "vite",
                "build": "vite build",
                "lint": "eslint . --ext js,jsx,ts,tsx" if include_eslint else "echo 'no linter'",
                "preview": "vite preview"
            },
            "dependencies": deps,
            "devDependencies": dev_deps
        }

        (project_dir / "package.json").write_text(json.dumps(package_json, indent=2))
        files_created.append(str((project_dir / "package.json").relative_to(project_dir.parent)))

        # Create vite.config.js
        vite_config = """import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    open: true,
  },
})
"""
        (project_dir / "vite.config.js").write_text(vite_config)
        files_created.append(str((project_dir / "vite.config.js").relative_to(project_dir.parent)))

        # Create tsconfig if needed
        if use_typescript:
            tsconfig = {
                "compilerOptions": {
                    "target": "ES2020",
                    "useDefineForClassFields": True,
                    "lib": ["ES2020", "DOM", "DOM.Iterable"],
                    "module": "ESNext",
                    "skipLibCheck": True,
                    "esModuleInterop": True,
                    "allowSyntheticDefaultImports": True,
                    "strict": True,
                    "resolveJsonModule": True,
                    "jsx": "react-jsx"
                },
                "include": ["src"],
                "references": [{"path": "./tsconfig.node.json"}]
            }
            (project_dir / "tsconfig.json").write_text(json.dumps(tsconfig, indent=2))
            files_created.append(str((project_dir / "tsconfig.json").relative_to(project_dir.parent)))

        # Create tailwind.config if needed
        if include_tailwind:
            tailwind_config = """export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
"""
            (project_dir / "tailwind.config.js").write_text(tailwind_config)
            files_created.append(str((project_dir / "tailwind.config.js").relative_to(project_dir.parent)))

            postcss_config = """export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
"""
            (project_dir / "postcss.config.js").write_text(postcss_config)
            files_created.append(str((project_dir / "postcss.config.js").relative_to(project_dir.parent)))

        # Create .gitignore
        gitignore = """node_modules
dist
.env
.env.local
.env.*.local
.DS_Store
"""
        (project_dir / ".gitignore").write_text(gitignore)
        files_created.append(str((project_dir / ".gitignore").relative_to(project_dir.parent)))

        return files_created

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        framework: str = "react",
        use_typescript: bool = False,
        include_tailwind: bool = False,
        include_eslint: bool = False
    ) -> CommandResult:
        """
        Execute /init-react command

        @CODE:FRONTEND-REACT-EXECUTE-001:MAIN
        """
        # Validation
        self.validate_project_name(project_name)
        self.validate_framework(framework)

        try:
            # Create project directory
            project_dir = self.create_project_directory(Path(output_dir), project_name)

            # Create structure
            files_created = self.create_react_structure(
                project_dir,
                framework=framework,
                use_typescript=use_typescript,
                include_tailwind=include_tailwind,
                include_eslint=include_eslint
            )

            message = f"✅ React project '{project_name}' initialized successfully\n"
            message += f"📁 Location: {project_dir}\n"
            message += f"⚛️  Framework: {framework}\n"
            message += f"📝 Files created: {len(files_created)}\n"
            message += f"🚀 Next: cd {project_name} && npm install && npm run dev"

            return CommandResult(
                success=True,
                project_dir=project_dir,
                files_created=files_created,
                message=message
            )

        except FileExistsError:
            raise
        except Exception as e:
            return CommandResult(
                success=False,
                project_dir=None,
                files_created=[],
                message=f"❌ Error initializing React project",
                error=str(e)
            )


class SetupStateCommand:
    """
    /setup-state command implementation

    Configures state management for React project
    """

    VALID_STATE_TYPES = ["react-context", "zustand", "redux", "recoil"]

    def validate_state_type(self, state_type: str) -> bool:
        """
        Validate state management type

        @CODE:FRONTEND-STATE-VALIDATE-001:VALIDATION
        """
        if state_type not in self.VALID_STATE_TYPES:
            raise ValueError(
                f"Invalid state type: {state_type}\nSupported: {', '.join(self.VALID_STATE_TYPES)}"
            )
        return True

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        state_type: str = "zustand"
    ) -> CommandResult:
        """
        Execute /setup-state command

        @CODE:FRONTEND-STATE-EXECUTE-001:MAIN
        """
        self.validate_state_type(state_type)

        try:
            project_dir = output_dir / project_name
            src_dir = project_dir / "src"
            store_dir = src_dir / "store"
            store_dir.mkdir(parents=True, exist_ok=True)

            files_created = []

            # Create state management setup based on type
            if state_type == "zustand":
                store_content = """import { create } from 'zustand'

export const useStore = create((set) => ({
  count: 0,
  increment: () => set((state) => ({ count: state.count + 1 })),
  decrement: () => set((state) => ({ count: state.count - 1 })),
}))
"""
                store_file = store_dir / "useStore.js"
                store_file.write_text(store_content)
                files_created.append(str(store_file.relative_to(output_dir)))

            elif state_type == "redux":
                store_content = """import { createSlice, configureStore } from '@reduxjs/toolkit'

const counterSlice = createSlice({
  name: 'counter',
  initialState: { value: 0 },
  reducers: {
    increment: (state) => { state.value += 1 },
    decrement: (state) => { state.value -= 1 },
  },
})

export const store = configureStore({
  reducer: {
    counter: counterSlice.reducer,
  },
})

export const { increment, decrement } = counterSlice.actions
"""
                store_file = store_dir / "store.js"
                store_file.write_text(store_content)
                files_created.append(str(store_file.relative_to(output_dir)))

            else:  # react-context or recoil
                context_content = """import React, { createContext, useState } from 'react'

export const AppContext = createContext()

export function AppProvider({ children }) {
  const [state, setState] = useState({
    count: 0,
  })

  return (
    <AppContext.Provider value={{ state, setState }}>
      {children}
    </AppContext.Provider>
  )
}
"""
                context_file = store_dir / "AppContext.jsx"
                context_file.write_text(context_content)
                files_created.append(str(context_file.relative_to(output_dir)))

            message = f"✅ State management setup with {state_type} completed\n"
            message += f"📁 Location: {store_dir}\n"
            message += f"📝 Files created: {len(files_created)}"

            return CommandResult(
                success=True,
                project_dir=store_dir,
                files_created=files_created,
                message=message
            )

        except Exception as e:
            return CommandResult(
                success=False,
                project_dir=None,
                files_created=[],
                message=f"❌ Error setting up state management",
                error=str(e)
            )


class SetupTestingCommand:
    """
    /setup-testing command implementation

    Configures testing setup for React project
    """

    VALID_FRAMEWORKS = ["vitest", "jest"]

    def validate_framework(self, framework: str) -> bool:
        """
        Validate test framework

        @CODE:FRONTEND-TESTING-VALIDATE-001:VALIDATION
        """
        if framework not in self.VALID_FRAMEWORKS:
            raise ValueError(
                f"Invalid framework: {framework}\nSupported: {', '.join(self.VALID_FRAMEWORKS)}"
            )
        return True

    def execute(
        self,
        project_name: str,
        output_dir: Path,
        test_framework: str = "vitest",
        include_testing_library: bool = False
    ) -> CommandResult:
        """
        Execute /setup-testing command

        @CODE:FRONTEND-TESTING-EXECUTE-001:MAIN
        """
        self.validate_framework(test_framework)

        try:
            project_dir = output_dir / project_name
            tests_dir = project_dir / "src" / "__tests__"
            tests_dir.mkdir(parents=True, exist_ok=True)

            files_created = []

            # Create test example file
            test_content = """import { describe, it, expect } from 'vitest'

describe('Math operations', () => {
  it('should add numbers correctly', () => {
    expect(1 + 1).toBe(2)
  })

  it('should subtract numbers correctly', () => {
    expect(5 - 3).toBe(2)
  })
})
"""
            test_file = tests_dir / "example.test.js"
            test_file.write_text(test_content)
            files_created.append(str(test_file.relative_to(output_dir)))

            # Create vitest/jest config
            if test_framework == "vitest":
                config_content = """import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: [],
  },
})
"""
                (project_dir / "vitest.config.js").write_text(config_content)
                files_created.append(str((project_dir / "vitest.config.js").relative_to(output_dir)))

            message = f"✅ Testing setup with {test_framework} completed\n"
            message += f"📁 Location: {tests_dir}\n"
            message += f"📝 Files created: {len(files_created)}"

            return CommandResult(
                success=True,
                project_dir=tests_dir,
                files_created=files_created,
                message=message
            )

        except Exception as e:
            return CommandResult(
                success=False,
                project_dir=None,
                files_created=[],
                message=f"❌ Error setting up testing",
                error=str(e)
            )


# Create module-level command instances
init_react = InitReactCommand()
setup_state = SetupStateCommand()
setup_testing = SetupTestingCommand()
