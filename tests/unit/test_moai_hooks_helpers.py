import importlib
import importlib.util
import json
import sys
import time
from pathlib import Path
from types import SimpleNamespace
from typing import Dict, List

import pytest


def _load_tags_module(module_name: str = "tags_module"):
    """Dynamically load core/tags.py as a fresh module."""
    repo_root = Path(__file__).resolve().parents[2]

    hooks_dir = repo_root / "src" / "moai_adk" / "templates" / ".claude" / "hooks" / "alfred"
    if str(hooks_dir) not in sys.path:
        sys.path.insert(0, str(hooks_dir))

    module_path = hooks_dir / "core" / "tags.py"

    spec = importlib.util.spec_from_file_location(module_name, module_path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = module
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


@pytest.fixture
def tags_module():
    """Provide a clean tags module instance per test."""
    module_name = f"tags_module_{time.time_ns()}"
    module = _load_tags_module(module_name=module_name)
    module._tag_cache.clear()
    module._lib_version_cache.clear()
    yield module
    module._tag_cache.clear()
    module._lib_version_cache.clear()
    sys.modules.pop(module_name, None)


def test_search_tags_uses_cache(monkeypatch: pytest.MonkeyPatch, tmp_path: Path, tags_module):
    pattern = "@SPEC:AUTH-001"
    scope_dir = tmp_path / "workspace"
    scope_dir.mkdir()
    target_file = scope_dir / "spec.md"
    target_file.write_text("@SPEC:AUTH-001 Example spec\n", encoding="utf-8")

    call_count = {"run": 0}

    def fake_run(cmd, capture_output, text, timeout, check):
        call_count["run"] += 1
        output = json.dumps(
            {
                "type": "match",
                "data": {
                    "path": {"text": str(target_file)},
                    "lines": {"text": "@SPEC:AUTH-001 Example spec"},
                    "line_number": 1,
                },
            }
        )
        return SimpleNamespace(stdout=f"{output}\n", returncode=0)

    monkeypatch.setattr(tags_module.subprocess, "run", fake_run)

    results_first = tags_module.search_tags(pattern, [str(scope_dir)], cache_ttl=120)
    results_second = tags_module.search_tags(pattern, [str(scope_dir)], cache_ttl=120)

    assert call_count["run"] == 1, "subprocess.run should be called only once when cache hits."
    assert results_first == results_second
    assert results_first[0]["file"] == str(target_file)


def test_search_tags_cache_expires(monkeypatch: pytest.MonkeyPatch, tmp_path: Path, tags_module):
    pattern = "@SPEC:AUTH-002"
    scope_dir = tmp_path / "workspace"
    scope_dir.mkdir()
    target_file = scope_dir / "spec.md"
    target_file.write_text("@SPEC:AUTH-002 Another spec\n", encoding="utf-8")

    call_count = {"run": 0}

    def fake_run(cmd, capture_output, text, timeout, check):
        call_count["run"] += 1
        output = json.dumps(
            {
                "type": "match",
                "data": {
                    "path": {"text": str(target_file)},
                    "lines": {"text": "@SPEC:AUTH-002 Another spec"},
                    "line_number": 1,
                },
            }
        )
        return SimpleNamespace(stdout=f"{output}\n", returncode=0)

    monkeypatch.setattr(tags_module.subprocess, "run", fake_run)

    tags_module.search_tags(pattern, [str(scope_dir)], cache_ttl=1)
    cache_key = f"{pattern}:{str(scope_dir)}"
    matches, mtime, cached_at = tags_module._tag_cache[cache_key]

    # Force a TTL expiration by backdating the cached timestamp.
    tags_module._tag_cache[cache_key] = (matches, mtime, cached_at - 120)

    tags_module.search_tags(pattern, [str(scope_dir)], cache_ttl=1)

    assert call_count["run"] == 2, "Cache must be refreshed once the TTL has expired."


def test_verify_tag_chain_complete(monkeypatch: pytest.MonkeyPatch, tags_module):
    responses: Dict[str, List[dict]] = {
        "@SPEC:AUTH-100": [{"tag": "@SPEC:AUTH-100"}],
        "@TEST:AUTH-100": [{"tag": "@TEST:AUTH-100"}],
        "@CODE:AUTH-100": [{"tag": "@CODE:AUTH-100"}],
    }

    def fake_search(pattern: str, scope=None, cache_ttl=60):
        return responses.get(pattern, [])

    monkeypatch.setattr(tags_module, "search_tags", fake_search)

    result = tags_module.verify_tag_chain("AUTH-100")

    assert result["complete"] is True
    assert result["orphans"] == []


def test_verify_tag_chain_orphans(monkeypatch: pytest.MonkeyPatch, tags_module):
    responses = {
        "@SPEC:AUTH-200": [],
        "@TEST:AUTH-200": [{"tag": "@TEST:AUTH-200"}],
        "@CODE:AUTH-200": [{"tag": "@CODE:AUTH-200"}],
    }

    def fake_search(pattern: str, scope=None, cache_ttl=60):
        return responses.get(pattern, [])

    monkeypatch.setattr(tags_module, "search_tags", fake_search)

    result = tags_module.verify_tag_chain("AUTH-200")

    assert result["complete"] is False
    assert len(result["orphans"]) == 2


def test_find_all_tags_by_type(monkeypatch: pytest.MonkeyPatch, tags_module):
    sample_matches = [
        {"tag": "@SPEC:AUTH-001"},
        {"tag": "@SPEC:AUTH-002"},
        {"tag": "@SPEC:INSTALLER-001"},
        {"tag": "@SPEC:AUTH-002"},  # Duplicate entry should be removed.
    ]

    def fake_search(pattern: str, scope=None, cache_ttl=60):
        return sample_matches

    monkeypatch.setattr(tags_module, "search_tags", fake_search)

    grouped = tags_module.find_all_tags_by_type("SPEC")
    assert grouped == {"AUTH": ["AUTH-001", "AUTH-002"], "INSTALLER": ["INSTALLER-001"]}


def test_suggest_tag_reuse(monkeypatch: pytest.MonkeyPatch, tags_module):
    sample_index = {
        "AUTH": [f"AUTH-{i:03d}" for i in range(1, 7)],
        "INSTALLER": ["INSTALLER-001"],
    }

    monkeypatch.setattr(tags_module, "find_all_tags_by_type", lambda _: sample_index)

    suggestions = tags_module.suggest_tag_reuse("auth")

    assert len(suggestions) == 5  # Limit suggestions to five entries.
    assert suggestions[0] == "AUTH-001"


def test_library_version_cache(tags_module):
    tags_module._lib_version_cache.clear()
    tags_module.set_library_version("fastapi", "0.115.0")

    assert tags_module.get_library_version("fastapi") == "0.115.0"

    version, cached_at = tags_module._lib_version_cache["fastapi"]
    tags_module._lib_version_cache["fastapi"] = (version, cached_at - 90000)

    assert tags_module.get_library_version("fastapi", cache_ttl=10) is None


def test_workflow_context_helpers(tags_module):
    context_module = importlib.import_module("core.context")

    context_module.save_phase_context("analysis", {"spec": "data"})

    assert context_module.load_phase_context("analysis") == {"spec": "data"}

    entry = context_module._workflow_context["analysis"]
    context_module._workflow_context["analysis"] = {"data": entry["data"], "timestamp": entry["timestamp"] - 1000}

    assert context_module.load_phase_context("analysis") is None

    context_module.save_phase_context("implementation", {"state": "ok"})
    context_module.clear_workflow_context()

    assert context_module.load_phase_context("implementation") is None
