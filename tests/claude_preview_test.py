import json, os, pathlib, subprocess, tempfile, unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
INSTALL = ROOT / "adapters/claude/install.py"
RENDER = ROOT / "adapters/claude/statusline.py"

class ClaudePreviewTest(unittest.TestCase):
    def render(self, data, env=None):
        return subprocess.run(["python3", str(RENDER)], input=data, text=True, capture_output=True, env=env, check=True).stdout

    def install(self, project):
        return subprocess.run(["python3", str(INSTALL), str(project)], text=True, capture_output=True)

    def test_payloads(self):
        self.assertIn("Sonnet", self.render('{"model":{"display_name":"Sonnet"},"context_window":{"used_percentage":42.5}}'))
        self.assertIn("42.5%", self.render('{"context_window":{"used_percentage":42.5}}'))
        for payload in ("", "{", "null", "[]", '{"context_window":{"used_percentage":null}}', '{"context_window":{"used_percentage":true}}', '{"context_window":{"used_percentage":101}}'):
            self.assertIn("context: --", self.render(payload))
        self.assertIn("0%", self.render('{"context_window":{"used_percentage":0}}'))
        self.assertIn("context: --", self.render(json.dumps({"context_window": {"used_percentage": 10 ** 400}})))
        self.assertNotIn("\x1b", self.render('{"model":{"display_name":"x\\u001b[31m"}}'))

    def test_secret_is_not_used(self):
        with tempfile.TemporaryDirectory() as temp:
            fake = pathlib.Path(temp)
            (fake / ".moai-cli").mkdir()
            (fake / ".moai-cli/.env.glm").write_text("MARKER=secret_should_not_appear\n")
            env = dict(os.environ, HOME=temp, MARKER="secret_should_not_appear")
            self.assertNotIn("secret_should_not_appear", self.render("{}", env))

    def test_reinstall_quoted_path_and_conflicts(self):
        with tempfile.TemporaryDirectory() as temp:
            project = pathlib.Path(temp) / "project space ' quote"
            project.mkdir()
            self.assertEqual(self.install(project).returncode, 0)
            settings = json.loads((project / ".claude/settings.json").read_text())
            before = {str(p.relative_to(project)): p.read_bytes() for p in project.rglob("*") if p.is_file()}
            self.assertEqual(self.install(project).returncode, 0)
            after = {str(p.relative_to(project)): p.read_bytes() for p in project.rglob("*") if p.is_file()}
            self.assertEqual(before, after)
            env = dict(os.environ, CLAUDE_PROJECT_DIR=str(project))
            result = subprocess.run(settings["statusLine"]["command"], shell=True, input="{}", text=True, capture_output=True, env=env)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("MoAI", result.stdout)
            style = project / ".claude/output-styles/moai.md"
            style.write_text("user edit")
            self.assertNotEqual(self.install(project).returncode, 0)
            self.assertEqual(style.read_text(), "user edit")

    def test_existing_settings_and_symlink_are_preserved(self):
        with tempfile.TemporaryDirectory() as temp:
            project = pathlib.Path(temp) / "project"
            project.mkdir()
            (project / ".claude").mkdir()
            settings = project / ".claude/settings.json"
            settings.write_text('{"unknown":true,"hooks":{"user":[]}}')
            self.assertEqual(self.install(project).returncode, 0)
            merged = json.loads(settings.read_text())
            self.assertEqual(merged["unknown"], True)
            self.assertEqual(merged["hooks"], {"user": []})
            self.assertEqual(merged["outputStyle"], "MoAI-Easy")
            before = settings.read_bytes()
            self.assertEqual(self.install(project).returncode, 0)
            self.assertEqual(settings.read_bytes(), before)
        with tempfile.TemporaryDirectory() as temp:
            project = pathlib.Path(temp)
            (project / ".claude").mkdir()
            settings = project / ".claude/settings.json"
            user_data = '{"statusLine":{"type":"command","command":"orca"},"unknown":true}'
            settings.write_text(user_data)
            self.assertNotEqual(self.install(project).returncode, 0)
            self.assertEqual(settings.read_text(), user_data)
            self.assertFalse((project / ".moai-cli").exists())
        with tempfile.TemporaryDirectory() as temp:
            project = pathlib.Path(temp)
            (project / ".claude").mkdir()
            settings = project / ".claude/settings.json"
            user_data = '{"unknown":1,"unknown":2}'
            settings.write_text(user_data)
            self.assertNotEqual(self.install(project).returncode, 0)
            self.assertEqual(settings.read_text(), user_data)
            self.assertFalse((project / ".moai-cli").exists())
        with tempfile.TemporaryDirectory() as temp:
            project = pathlib.Path(temp) / "project"
            project.mkdir()
            outside = pathlib.Path(temp) / "outside"
            outside.mkdir()
            (project / ".claude").symlink_to(outside, target_is_directory=True)
            self.assertNotEqual(self.install(project).returncode, 0)
            self.assertEqual(list(outside.iterdir()), [])

    def test_style_metadata(self):
        for name, flag in [("moai.md", "true"), ("moai-easy.md", "true"), ("moai-learn.md", "false")]:
            text = (ROOT / "adapters/claude/output-styles" / name).read_text()
            self.assertTrue(text.startswith("---\n"))
            self.assertIn("keep-coding-instructions: " + flag, text)
        self.assertIn("Do not write or modify code", (ROOT / "adapters/claude/output-styles/moai-learn.md").read_text())

if __name__ == "__main__":
    unittest.main()
