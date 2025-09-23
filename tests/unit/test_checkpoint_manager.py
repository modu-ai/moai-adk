import importlib.util
from datetime import datetime
from pathlib import Path

import pytest

SPEC = importlib.util.spec_from_file_location(
    "moai_checkpoint_manager",
    Path('.moai/scripts/checkpoint_manager.py'),
)
checkpoint_module = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(checkpoint_module)
CheckpointManager = checkpoint_module.CheckpointManager


def test_generate_checkpoint_id_uniqueness(monkeypatch):
    manager = CheckpointManager()

    # 고정된 타임스탬프와 중복 태그 존재 시퀀스를 주입한다.
    monkeypatch.setattr(manager, '_current_time', lambda: datetime(2024, 1, 1, 12, 0, 0))
    responses = iter([True, True, False])
    monkeypatch.setattr(manager, '_tag_exists', lambda tag: next(responses))

    tag_name = manager.generate_checkpoint_id()

    assert tag_name.startswith('moai_cp/20240101_120000')
    assert tag_name.endswith('_02')


def test_record_checkpoint_removes_old_tag(monkeypatch):
    manager = CheckpointManager()

    metadata = {
        'checkpoints': [
            {
                'id': 'moai_cp/old',
                'timestamp': '2024-01-01T00:00:00',
                'kind': 'tag',
                'tag': 'moai_cp/old',
            }
        ]
    }

    saved = {}

    monkeypatch.setattr(manager, '_load_metadata', lambda: metadata)
    monkeypatch.setattr(manager, '_save_metadata', lambda data: saved.setdefault('data', data))
    monkeypatch.setattr(
        manager,
        'load_config',
        lambda: {'git_strategy': {'personal': {'max_checkpoints': 1}}},
    )

    deleted_tags: list[str] = []
    monkeypatch.setattr(manager, '_delete_tag', lambda tag: deleted_tags.append(tag))
    monkeypatch.setattr(manager, '_drop_stash_entry', lambda commit: None)

    new_entry = {
        'id': 'moai_cp/new',
        'timestamp': '2024-02-01T00:00:00',
        'kind': 'tag',
        'tag': 'moai_cp/new',
    }

    manager._record_checkpoint(new_entry)

    assert deleted_tags == ['moai_cp/old']
    assert saved['data']['checkpoints'][0]['id'] == 'moai_cp/new'
