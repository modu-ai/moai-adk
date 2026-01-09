"""TAG-001: Test Terminal Models

RED Phase: Tests for terminal session and message models.
These tests verify the Pydantic models for terminal data structures.
"""

from datetime import UTC, datetime


class TestTerminalSessionModel:
    """Test cases for TerminalSession model"""

    def test_terminal_session_model_exists(self):
        """Test that TerminalSession model can be imported"""
        from moai_adk.web.models.terminal import TerminalSession

        assert TerminalSession is not None

    def test_terminal_session_has_id_field(self):
        """Test that TerminalSession has id field"""
        from moai_adk.web.models.terminal import TerminalSession

        session = TerminalSession(
            id="term-001",
            status="pending",
            created_at=datetime.now(UTC),
        )
        assert hasattr(session, "id")
        assert session.id == "term-001"

    def test_terminal_session_has_spec_id_field(self):
        """Test that TerminalSession has optional spec_id field"""
        from moai_adk.web.models.terminal import TerminalSession

        session = TerminalSession(
            id="term-001",
            spec_id="SPEC-001",
            status="pending",
            created_at=datetime.now(UTC),
        )
        assert hasattr(session, "spec_id")
        assert session.spec_id == "SPEC-001"

    def test_terminal_session_spec_id_is_optional(self):
        """Test that spec_id can be None"""
        from moai_adk.web.models.terminal import TerminalSession

        session = TerminalSession(
            id="term-001",
            status="pending",
            created_at=datetime.now(UTC),
        )
        assert session.spec_id is None

    def test_terminal_session_has_worktree_path_field(self):
        """Test that TerminalSession has optional worktree_path field"""
        from moai_adk.web.models.terminal import TerminalSession

        session = TerminalSession(
            id="term-001",
            worktree_path="/path/to/worktree",
            status="pending",
            created_at=datetime.now(UTC),
        )
        assert hasattr(session, "worktree_path")
        assert session.worktree_path == "/path/to/worktree"

    def test_terminal_session_worktree_path_is_optional(self):
        """Test that worktree_path can be None"""
        from moai_adk.web.models.terminal import TerminalSession

        session = TerminalSession(
            id="term-001",
            status="pending",
            created_at=datetime.now(UTC),
        )
        assert session.worktree_path is None

    def test_terminal_session_has_pid_field(self):
        """Test that TerminalSession has optional pid field"""
        from moai_adk.web.models.terminal import TerminalSession

        session = TerminalSession(
            id="term-001",
            pid=12345,
            status="running",
            created_at=datetime.now(UTC),
        )
        assert hasattr(session, "pid")
        assert session.pid == 12345

    def test_terminal_session_pid_is_optional(self):
        """Test that pid can be None"""
        from moai_adk.web.models.terminal import TerminalSession

        session = TerminalSession(
            id="term-001",
            status="pending",
            created_at=datetime.now(UTC),
        )
        assert session.pid is None

    def test_terminal_session_has_status_field(self):
        """Test that TerminalSession has status field with Literal type"""
        from moai_adk.web.models.terminal import TerminalSession

        session = TerminalSession(
            id="term-001",
            status="pending",
            created_at=datetime.now(UTC),
        )
        assert hasattr(session, "status")
        assert session.status == "pending"

    def test_terminal_session_status_values(self):
        """Test that status accepts all valid values"""
        from moai_adk.web.models.terminal import TerminalSession

        valid_statuses = ["pending", "running", "completed", "error"]
        for status in valid_statuses:
            session = TerminalSession(
                id=f"term-{status}",
                status=status,
                created_at=datetime.now(UTC),
            )
            assert session.status == status

    def test_terminal_session_has_created_at_field(self):
        """Test that TerminalSession has created_at field"""
        from moai_adk.web.models.terminal import TerminalSession

        now = datetime.now(UTC)
        session = TerminalSession(
            id="term-001",
            status="pending",
            created_at=now,
        )
        assert hasattr(session, "created_at")
        assert session.created_at == now

    def test_terminal_session_has_completed_at_field(self):
        """Test that TerminalSession has optional completed_at field"""
        from moai_adk.web.models.terminal import TerminalSession

        now = datetime.now(UTC)
        session = TerminalSession(
            id="term-001",
            status="completed",
            created_at=now,
            completed_at=now,
        )
        assert hasattr(session, "completed_at")
        assert session.completed_at == now

    def test_terminal_session_completed_at_is_optional(self):
        """Test that completed_at can be None"""
        from moai_adk.web.models.terminal import TerminalSession

        session = TerminalSession(
            id="term-001",
            status="pending",
            created_at=datetime.now(UTC),
        )
        assert session.completed_at is None

    def test_terminal_session_serialization(self):
        """Test that TerminalSession can be serialized to JSON"""
        from moai_adk.web.models.terminal import TerminalSession

        now = datetime.now(UTC)
        session = TerminalSession(
            id="term-001",
            spec_id="SPEC-001",
            worktree_path="/path/to/worktree",
            pid=12345,
            status="running",
            created_at=now,
        )

        json_data = session.model_dump(mode="json")

        assert json_data["id"] == "term-001"
        assert json_data["spec_id"] == "SPEC-001"
        assert json_data["worktree_path"] == "/path/to/worktree"
        assert json_data["pid"] == 12345
        assert json_data["status"] == "running"
        assert "created_at" in json_data


class TestTerminalMessageModel:
    """Test cases for TerminalMessage model"""

    def test_terminal_message_model_exists(self):
        """Test that TerminalMessage model can be imported"""
        from moai_adk.web.models.terminal import TerminalMessage

        assert TerminalMessage is not None

    def test_terminal_message_has_type_field(self):
        """Test that TerminalMessage has type field"""
        from moai_adk.web.models.terminal import TerminalMessage

        msg = TerminalMessage(type="input", data="ls -la")
        assert hasattr(msg, "type")
        assert msg.type == "input"

    def test_terminal_message_type_values(self):
        """Test that type accepts all valid values"""
        from moai_adk.web.models.terminal import TerminalMessage

        valid_types = ["input", "output", "resize", "heartbeat", "close"]
        for msg_type in valid_types:
            msg = TerminalMessage(type=msg_type)
            assert msg.type == msg_type

    def test_terminal_message_has_data_field(self):
        """Test that TerminalMessage has optional data field"""
        from moai_adk.web.models.terminal import TerminalMessage

        msg = TerminalMessage(type="input", data="echo hello")
        assert hasattr(msg, "data")
        assert msg.data == "echo hello"

    def test_terminal_message_data_is_optional(self):
        """Test that data can be None"""
        from moai_adk.web.models.terminal import TerminalMessage

        msg = TerminalMessage(type="heartbeat")
        assert msg.data is None

    def test_terminal_message_has_cols_field(self):
        """Test that TerminalMessage has optional cols field"""
        from moai_adk.web.models.terminal import TerminalMessage

        msg = TerminalMessage(type="resize", cols=120)
        assert hasattr(msg, "cols")
        assert msg.cols == 120

    def test_terminal_message_cols_is_optional(self):
        """Test that cols can be None"""
        from moai_adk.web.models.terminal import TerminalMessage

        msg = TerminalMessage(type="input", data="test")
        assert msg.cols is None

    def test_terminal_message_has_rows_field(self):
        """Test that TerminalMessage has optional rows field"""
        from moai_adk.web.models.terminal import TerminalMessage

        msg = TerminalMessage(type="resize", rows=40)
        assert hasattr(msg, "rows")
        assert msg.rows == 40

    def test_terminal_message_rows_is_optional(self):
        """Test that rows can be None"""
        from moai_adk.web.models.terminal import TerminalMessage

        msg = TerminalMessage(type="input", data="test")
        assert msg.rows is None

    def test_terminal_message_resize_with_dimensions(self):
        """Test resize message with both cols and rows"""
        from moai_adk.web.models.terminal import TerminalMessage

        msg = TerminalMessage(type="resize", cols=80, rows=24)
        assert msg.type == "resize"
        assert msg.cols == 80
        assert msg.rows == 24

    def test_terminal_message_serialization(self):
        """Test that TerminalMessage can be serialized to JSON"""
        from moai_adk.web.models.terminal import TerminalMessage

        msg = TerminalMessage(type="resize", cols=120, rows=40)
        json_data = msg.model_dump(mode="json")

        assert json_data["type"] == "resize"
        assert json_data["cols"] == 120
        assert json_data["rows"] == 40


class TestTerminalCreateModel:
    """Test cases for TerminalCreate request model"""

    def test_terminal_create_model_exists(self):
        """Test that TerminalCreate model can be imported"""
        from moai_adk.web.models.terminal import TerminalCreate

        assert TerminalCreate is not None

    def test_terminal_create_has_spec_id_field(self):
        """Test that TerminalCreate has optional spec_id field"""
        from moai_adk.web.models.terminal import TerminalCreate

        create = TerminalCreate(spec_id="SPEC-001")
        assert hasattr(create, "spec_id")
        assert create.spec_id == "SPEC-001"

    def test_terminal_create_spec_id_is_optional(self):
        """Test that spec_id can be None"""
        from moai_adk.web.models.terminal import TerminalCreate

        create = TerminalCreate()
        assert create.spec_id is None

    def test_terminal_create_has_worktree_path_field(self):
        """Test that TerminalCreate has optional worktree_path field"""
        from moai_adk.web.models.terminal import TerminalCreate

        create = TerminalCreate(worktree_path="/path/to/worktree")
        assert hasattr(create, "worktree_path")
        assert create.worktree_path == "/path/to/worktree"

    def test_terminal_create_has_initial_command_field(self):
        """Test that TerminalCreate has optional initial_command field"""
        from moai_adk.web.models.terminal import TerminalCreate

        create = TerminalCreate(initial_command="claude /moai:2-run SPEC-001")
        assert hasattr(create, "initial_command")
        assert create.initial_command == "claude /moai:2-run SPEC-001"

    def test_terminal_create_has_cols_field(self):
        """Test that TerminalCreate has cols field with default"""
        from moai_adk.web.models.terminal import TerminalCreate

        create = TerminalCreate()
        assert hasattr(create, "cols")
        assert create.cols == 80  # default value

    def test_terminal_create_has_rows_field(self):
        """Test that TerminalCreate has rows field with default"""
        from moai_adk.web.models.terminal import TerminalCreate

        create = TerminalCreate()
        assert hasattr(create, "rows")
        assert create.rows == 24  # default value

    def test_terminal_create_custom_dimensions(self):
        """Test TerminalCreate with custom dimensions"""
        from moai_adk.web.models.terminal import TerminalCreate

        create = TerminalCreate(cols=120, rows=40)
        assert create.cols == 120
        assert create.rows == 40


class TestTerminalResizeModel:
    """Test cases for TerminalResize request model"""

    def test_terminal_resize_model_exists(self):
        """Test that TerminalResize model can be imported"""
        from moai_adk.web.models.terminal import TerminalResize

        assert TerminalResize is not None

    def test_terminal_resize_has_cols_field(self):
        """Test that TerminalResize has cols field"""
        from moai_adk.web.models.terminal import TerminalResize

        resize = TerminalResize(cols=120, rows=40)
        assert hasattr(resize, "cols")
        assert resize.cols == 120

    def test_terminal_resize_has_rows_field(self):
        """Test that TerminalResize has rows field"""
        from moai_adk.web.models.terminal import TerminalResize

        resize = TerminalResize(cols=120, rows=40)
        assert hasattr(resize, "rows")
        assert resize.rows == 40


class TestTerminalListModel:
    """Test cases for TerminalList response model"""

    def test_terminal_list_model_exists(self):
        """Test that TerminalList model can be imported"""
        from moai_adk.web.models.terminal import TerminalList

        assert TerminalList is not None

    def test_terminal_list_has_terminals_field(self):
        """Test that TerminalList has terminals list field"""
        from moai_adk.web.models.terminal import TerminalList

        terminal_list = TerminalList()
        assert hasattr(terminal_list, "terminals")
        assert terminal_list.terminals == []

    def test_terminal_list_has_total_field(self):
        """Test that TerminalList has total count field"""
        from moai_adk.web.models.terminal import TerminalList

        terminal_list = TerminalList()
        assert hasattr(terminal_list, "total")
        assert terminal_list.total == 0

    def test_terminal_list_with_sessions(self):
        """Test TerminalList with actual sessions"""
        from moai_adk.web.models.terminal import TerminalList, TerminalSession

        sessions = [
            TerminalSession(
                id="term-001",
                status="running",
                created_at=datetime.now(UTC),
            ),
            TerminalSession(
                id="term-002",
                status="pending",
                created_at=datetime.now(UTC),
            ),
        ]
        terminal_list = TerminalList(terminals=sessions, total=2)
        assert len(terminal_list.terminals) == 2
        assert terminal_list.total == 2
