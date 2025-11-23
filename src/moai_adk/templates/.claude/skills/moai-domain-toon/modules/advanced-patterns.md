# Advanced TOON Patterns

## Streaming TOON Encoding

### Pattern: Encode Large Objects Incrementally

**Use Case**: Streaming data to LLMs without buffering entire objects.

```python
class StreamingTOONEncoder:
    """Stream TOON encoding for large datasets."""

    async def encode_stream(self, data_iterator, buffer_size: int = 1024):
        """Encode data stream with buffering."""
        buffer = []
        buffer_bytes = 0

        async for item in data_iterator:
            encoded = self.encode_value(item)
            buffer.append(encoded)
            buffer_bytes += len(encoded.encode())

            if buffer_bytes >= buffer_size:
                yield '\n'.join(buffer)
                buffer = []
                buffer_bytes = 0

        if buffer:
            yield '\n'.join(buffer)

    def encode_value(self, value):
        """Encode single value to TOON."""
        if isinstance(value, dict):
            return self._encode_dict(value)
        elif isinstance(value, list):
            return self._encode_list(value)
        elif isinstance(value, str):
            return f'"{value}"'
        elif value is None:
            return 'null'
        elif isinstance(value, bool):
            return 'true' if value else 'false'
        else:
            return str(value)
```

---

## TOON Compression for RAG

### Pattern: Compress Retrieved Context

**Use Case**: Reduce tokens in RAG document context.

```python
class TOONRAGCompressor:
    """Compress RAG documents using TOON."""

    def compress_documents(self, documents: list) -> str:
        """Convert RAG documents to TOON format."""
        compressed = []

        for doc in documents:
            # Extract essential fields only
            toon_doc = {
                "id": doc.get("id"),
                "title": doc.get("title"),
                "content": doc.get("content"),
                "score": doc.get("score")
            }
            compressed.append(toon_doc)

        return self._to_toon(compressed)

    def _to_toon(self, data):
        """Convert to minimal TOON representation."""
        lines = []
        self._format_value(data, lines, indent=0)
        return '\n'.join(lines)

    def _format_value(self, value, lines, indent):
        """Recursively format TOON."""
        if isinstance(value, dict):
            for key, val in value.items():
                if isinstance(val, (dict, list)):
                    lines.append("  " * indent + key + ":")
                    self._format_value(val, lines, indent + 1)
                else:
                    lines.append("  " * indent + key + ": " + str(val))
        elif isinstance(value, list):
            for item in value:
                lines.append("  " * indent + "-")
                self._format_value(item, lines, indent + 1)
```

---

## TOON Validation & Schema

### Pattern: Validate TOON Against Schema

```python
class TOONValidator:
    """Validate TOON against predefined schema."""

    def __init__(self, schema: dict):
        self.schema = schema

    def validate(self, toon_string: str) -> tuple[bool, list]:
        """Validate TOON and return errors."""
        try:
            data = self.parse(toon_string)
            errors = self.validate_data(data, self.schema)
            return len(errors) == 0, errors
        except Exception as e:
            return False, [str(e)]

    def validate_data(self, data, schema, path="root"):
        """Recursively validate data against schema."""
        errors = []

        # Type check
        if "type" in schema:
            if not self._check_type(data, schema["type"]):
                errors.append(f"{path}: Expected {schema['type']}, got {type(data).__name__}")

        # Required fields
        if "required" in schema and isinstance(data, dict):
            for field in schema["required"]:
                if field not in data:
                    errors.append(f"{path}: Missing required field '{field}'")

        # Field validation
        if "properties" in schema and isinstance(data, dict):
            for key, val in data.items():
                if key in schema["properties"]:
                    field_errors = self.validate_data(
                        val,
                        schema["properties"][key],
                        f"{path}.{key}"
                    )
                    errors.extend(field_errors)

        return errors

    def _check_type(self, data, expected_type):
        """Check if data matches expected type."""
        type_map = {
            "string": str,
            "number": (int, float),
            "boolean": bool,
            "array": list,
            "object": dict,
            "null": type(None)
        }
        return isinstance(data, type_map.get(expected_type, object))
```

---

## TOON for Prompt Engineering

### Pattern: Structured Prompts with TOON

```python
class StructuredPromptBuilder:
    """Build structured prompts using TOON."""

    def build_query_prompt(self, query: str, context: list, instructions: str):
        """Build prompt with TOON-encoded context."""
        prompt_data = {
            "task": "answer_question",
            "query": query,
            "context": context,
            "instructions": instructions,
            "format": "comprehensive"
        }

        toon_prompt = self._to_toon(prompt_data)
        return f"""You have the following information:

{toon_prompt}

Please provide a comprehensive answer based on the context provided."""

    def build_extraction_prompt(self, text: str, fields: list):
        """Build extraction prompt."""
        prompt_data = {
            "task": "extract_fields",
            "text": text,
            "fields": fields,
            "format": "structured"
        }

        toon_prompt = self._to_toon(prompt_data)
        return f"""Extract the following fields from the text:

{toon_prompt}

Respond with the extracted values in TOON format."""
```

---

## TOON to JSON Conversion

### Pattern: Bidirectional Conversion

```python
class TOONConverter:
    """Convert between TOON and JSON."""

    @staticmethod
    def toon_to_json(toon_string: str) -> dict:
        """Convert TOON string to JSON."""
        lines = toon_string.strip().split('\n')
        return TOONConverter._parse_toon(lines, 0)[0]

    @staticmethod
    def json_to_toon(data: dict) -> str:
        """Convert JSON to TOON."""
        lines = []
        TOONConverter._format_toon(data, lines, 0)
        return '\n'.join(lines)

    @staticmethod
    def _parse_toon(lines, start_indent, idx=0):
        """Recursively parse TOON."""
        result = {}
        i = idx

        while i < len(lines):
            line = lines[i]
            indent = len(line) - len(line.lstrip())

            if indent < start_indent:
                break

            if indent > start_indent:
                i += 1
                continue

            # Parse line
            if ':' in line:
                key, value = line.strip().split(':', 1)
                value = value.strip()

                if not value:
                    # Nested object
                    nested, i = TOONConverter._parse_toon(lines, indent + 2, i + 1)
                    result[key] = nested
                else:
                    result[key] = TOONConverter._parse_value(value)
                    i += 1
            else:
                i += 1

        return result, i

    @staticmethod
    def _parse_value(value):
        """Parse TOON value."""
        if value.startswith('"') and value.endswith('"'):
            return value[1:-1]
        elif value == 'null':
            return None
        elif value == 'true':
            return True
        elif value == 'false':
            return False
        elif value.replace('.', '', 1).isdigit():
            return float(value) if '.' in value else int(value)
        else:
            return value
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
