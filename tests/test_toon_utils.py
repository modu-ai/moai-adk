"""Tests for TOON utilities."""

import json
import tempfile
from pathlib import Path

import pytest

from moai_adk.utils import (
    compare_formats,
    migrate_json_to_toon,
    toon_decode,
    toon_encode,
    toon_load,
    toon_save,
    validate_roundtrip,
)


class TestToonEncode:
    """Tests for toon_encode function."""

    def test_encode_simple_dict(self):
        """Test encoding simple dictionary."""
        data = {"name": "Alice", "age": 30}
        result = toon_encode(data)
        assert isinstance(result, str)
        assert "Alice" in result
        assert "30" in result

    def test_encode_list_of_objects(self):
        """Test encoding list of uniform objects."""
        data = {
            "users": [
                {"id": 1, "name": "Alice"},
                {"id": 2, "name": "Bob"},
            ]
        }
        result = toon_encode(data)
        assert isinstance(result, str)
        assert "users" in result

    def test_encode_nested_structure(self):
        """Test encoding nested data structures."""
        data = {
            "config": {
                "database": {
                    "host": "localhost",
                    "port": 5432,
                }
            }
        }
        result = toon_encode(data)
        assert isinstance(result, str)
        assert "localhost" in result

    def test_encode_invalid_data(self):
        """Test encoding with invalid data raises error."""

        # Non-serializable object
        class CustomObject:
            pass

        data = {"obj": CustomObject()}
        with pytest.raises(ValueError):
            toon_encode(data)


class TestToonDecode:
    """Tests for toon_decode function."""

    def test_decode_simple_toon(self):
        """Test decoding simple TOON/JSON string."""
        toon = '{"name": "Alice", "age": 30}'
        result = toon_decode(toon)
        assert result["name"] == "Alice"
        assert result["age"] == 30

    def test_decode_table_format(self):
        """Test decoding TOON/JSON table format."""
        toon = '{"users": [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]}'
        result = toon_decode(toon)
        assert isinstance(result["users"], list)
        assert len(result["users"]) == 2
        assert result["users"][0]["name"] == "Alice"

    def test_decode_invalid_toon(self):
        """Test decoding invalid TOON raises error."""
        invalid_toon = "invalid [[[["
        with pytest.raises(ValueError):
            toon_decode(invalid_toon)


class TestRoundtrip:
    """Tests for roundtrip encode/decode."""

    def test_roundtrip_simple_dict(self):
        """Test roundtrip for simple dictionary."""
        original = {"name": "Alice", "age": 30, "active": True}
        encoded = toon_encode(original)
        decoded = toon_decode(encoded)
        assert original == decoded

    def test_roundtrip_list_of_objects(self):
        """Test roundtrip for list of objects."""
        original = {
            "users": [
                {"id": 1, "name": "Alice", "email": "alice@example.com"},
                {"id": 2, "name": "Bob", "email": "bob@example.com"},
            ]
        }
        encoded = toon_encode(original)
        decoded = toon_decode(encoded)
        assert original == decoded

    def test_roundtrip_nested_structure(self):
        """Test roundtrip for nested structures."""
        original = {"config": {"database": {"host": "localhost", "port": 5432, "pool": {"min": 2, "max": 10}}}}
        encoded = toon_encode(original)
        decoded = toon_decode(encoded)
        assert original == decoded

    def test_validate_roundtrip_function(self):
        """Test validate_roundtrip function."""
        valid_data = {"users": [{"id": 1, "name": "Alice"}]}
        assert validate_roundtrip(valid_data) is True

    def test_validate_roundtrip_with_invalid_data(self):
        """Test validate_roundtrip with data that cannot encode."""

        class NonSerializable:
            pass

        invalid_data = {"obj": NonSerializable()}
        assert validate_roundtrip(invalid_data) is False


class TestFileOperations:
    """Tests for file save/load operations."""

    def test_save_and_load(self):
        """Test saving and loading TOON files."""
        data = {"config": {"debug": True, "port": 8080}}

        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"

            # Save
            toon_save(data, path)
            assert path.exists()

            # Load
            loaded = toon_load(path)
            assert loaded == data

    def test_save_creates_directories(self):
        """Test that save creates parent directories."""
        data = {"key": "value"}

        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "a" / "b" / "c" / "test.toon"

            toon_save(data, path)
            assert path.exists()
            assert path.parent.exists()

    def test_load_nonexistent_file(self):
        """Test loading nonexistent file raises error."""
        with pytest.raises(IOError):
            toon_load(Path("/nonexistent/file.toon"))

    def test_save_invalid_data(self):
        """Test saving invalid data raises error."""

        class NonSerializable:
            pass

        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "test.toon"
            with pytest.raises(ValueError):
                toon_save({"obj": NonSerializable()}, path)


class TestCompareFormats:
    """Tests for format comparison."""

    def test_compare_formats_returns_metrics(self):
        """Test compare_formats returns correct metrics."""
        data = {
            "items": [
                {"id": 1, "name": "Item1"},
                {"id": 2, "name": "Item2"},
            ]
        }

        metrics = compare_formats(data)

        assert "json" in metrics
        assert "toon" in metrics
        assert "reduction" in metrics
        assert metrics["json"]["tokens"] > 0
        assert metrics["toon"]["tokens"] > 0
        # Note: Current implementation uses JSON, so reduction may be minimal

    def test_compare_formats_shows_metrics(self):
        """Test that compare_formats provides consistent metrics."""
        # Uniform array data benefits most from TOON
        data = {"records": [{"id": i, "value": f"val{i}", "status": "active"} for i in range(10)]}

        metrics = compare_formats(data)
        # Metrics should be available for comparison
        assert "size_reduction_percent" in metrics
        # Both formats should be encodable
        assert metrics["json"]["size_bytes"] > 0
        assert metrics["toon"]["size_bytes"] > 0

    def test_compare_formats_invalid_data(self):
        """Test compare_formats with non-serializable data."""

        class NonSerializable:
            pass

        data = {"obj": NonSerializable()}
        with pytest.raises(ValueError):
            compare_formats(data)


class TestMigration:
    """Tests for JSON to TOON migration."""

    def test_migrate_json_to_toon(self):
        """Test migrating JSON file to TOON."""
        json_data = {
            "config": {"debug": True, "port": 8080},
            "services": [
                {"name": "api", "port": 8000},
                {"name": "worker", "port": 8001},
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create JSON file
            json_path = Path(tmpdir) / "config.json"
            json_path.write_text(json.dumps(json_data))

            # Migrate
            toon_path = migrate_json_to_toon(json_path)

            # Verify
            assert toon_path.exists()
            assert toon_path.name == "config.toon"
            loaded = toon_load(toon_path)
            assert loaded == json_data

    def test_migrate_with_custom_output_path(self):
        """Test migration with custom output path."""
        json_data = {"key": "value"}

        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "input.json"
            json_path.write_text(json.dumps(json_data))

            custom_path = Path(tmpdir) / "output.toon"
            result_path = migrate_json_to_toon(json_path, custom_path)

            assert result_path == custom_path
            assert custom_path.exists()

    def test_migrate_invalid_json(self):
        """Test migration of invalid JSON raises error."""
        with tempfile.TemporaryDirectory() as tmpdir:
            json_path = Path(tmpdir) / "invalid.json"
            json_path.write_text("{ invalid json }")

            with pytest.raises(ValueError):
                migrate_json_to_toon(json_path)
