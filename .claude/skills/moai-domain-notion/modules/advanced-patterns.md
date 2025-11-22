# Advanced Notion Patterns

## Multi-Database Synchronization

### Pattern: Keep Multiple Databases in Sync

**Use Case**: Maintain view databases that reflect source data changes.

```python
class NotionSyncManager:
    """Manage synchronization between related databases."""

    def __init__(self, client: AsyncClient):
        self.client = client
        self.sync_history = []

    async def sync_databases(
        self,
        source_db_id: str,
        target_db_id: str,
        property_mapping: dict
    ):
        """Sync source database to target database."""

        # Query source database
        source_items = await self.client.databases.query(
            database_id=source_db_id
        )

        # Get existing target items
        target_items = await self.client.databases.query(
            database_id=target_db_id
        )

        existing_ids = {
            item["properties"]["Source ID"]["rich_text"][0]["text"]["content"]
            for item in target_items["results"]
            if "Source ID" in item["properties"]
        }

        # Sync new and updated items
        for source_item in source_items["results"]:
            source_id = source_item["id"]
            properties = {}

            # Map properties
            for source_prop, target_prop in property_mapping.items():
                value = source_item["properties"].get(source_prop)
                if value:
                    properties[target_prop] = value

            # Add source reference
            properties["Source ID"] = {
                "rich_text": [{"text": {"content": source_id}}]
            }

            if source_id not in existing_ids:
                # Create new item
                await self.client.pages.create(
                    parent={"database_id": target_db_id},
                    properties=properties
                )
            else:
                # Update existing item
                target_page = next(
                    item for item in target_items["results"]
                    if source_id in [
                        rt["text"]["content"]
                        for rt in item["properties"].get("Source ID", {}).get("rich_text", [])
                    ]
                )
                await self.client.pages.update(
                    page_id=target_page["id"],
                    properties=properties
                )

            await asyncio.sleep(0.1)  # Rate limiting
```

---

## Batch Operations with Error Handling

### Pattern: Process Large Datasets Safely

```python
class BatchOperationManager:
    """Manage batch operations with error recovery."""

    async def batch_update_pages(
        self,
        client: AsyncClient,
        page_updates: list,
        batch_size: int = 10,
        retry_count: int = 3
    ) -> dict:
        """Update multiple pages with retry logic."""

        results = {
            "success": 0,
            "failed": 0,
            "errors": []
        }

        for i in range(0, len(page_updates), batch_size):
            batch = page_updates[i:i + batch_size]
            tasks = []

            for update in batch:
                task = self._update_with_retry(
                    client,
                    update["page_id"],
                    update["properties"],
                    retry_count
                )
                tasks.append(task)

            # Execute batch
            batch_results = await asyncio.gather(*tasks, return_exceptions=True)

            for result in batch_results:
                if isinstance(result, Exception):
                    results["failed"] += 1
                    results["errors"].append(str(result))
                else:
                    results["success"] += 1

            # Rate limiting between batches
            await asyncio.sleep(0.5)

        return results

    async def _update_with_retry(
        self,
        client: AsyncClient,
        page_id: str,
        properties: dict,
        max_retries: int
    ):
        """Update page with exponential backoff."""

        for attempt in range(max_retries):
            try:
                return await client.pages.update(
                    page_id=page_id,
                    properties=properties
                )
            except Exception as e:
                if attempt == max_retries - 1:
                    raise
                wait_time = 2 ** attempt
                await asyncio.sleep(wait_time)
```

---

## Notion Database Migrations

### Pattern: Migrate Data Between Schemas

```python
class NotionMigrationManager:
    """Manage complex database migrations."""

    async def migrate_database_schema(
        self,
        client: AsyncClient,
        source_db_id: str,
        target_db_id: str,
        transformer: callable
    ):
        """Migrate database to new schema."""

        # Step 1: Get all source data
        all_pages = []
        cursor = None

        while True:
            response = await client.databases.query(
                database_id=source_db_id,
                start_cursor=cursor,
                page_size=100
            )

            all_pages.extend(response["results"])

            if not response["has_more"]:
                break

            cursor = response["next_cursor"]

        # Step 2: Transform and migrate
        migrated_count = 0
        for page in all_pages:
            # Transform properties using provided function
            new_properties = transformer(page["properties"])

            # Create in target database
            await client.pages.create(
                parent={"database_id": target_db_id},
                properties=new_properties
            )

            # Copy content blocks
            blocks = await client.blocks.children.list(page["id"])
            if blocks["results"]:
                await client.blocks.children.append(
                    block_id=migrated_count,  # Get new page ID
                    children=blocks["results"]
                )

            migrated_count += 1
            await asyncio.sleep(0.1)

        return migrated_count
```

---

## Rich Content Management

### Pattern: Create Complex Pages with Markdown

```python
class NotionPageBuilder:
    """Build complex Notion pages with rich content."""

    async def create_markdown_page(
        self,
        client: AsyncClient,
        parent_id: str,
        title: str,
        markdown_content: str
    ):
        """Convert markdown to Notion blocks."""

        # Create page
        page = await client.pages.create(
            parent={"database_id": parent_id},
            properties={
                "Title": {
                    "title": [{"text": {"content": title}}]
                }
            }
        )

        # Parse markdown and create blocks
        blocks = self._parse_markdown(markdown_content)

        # Add blocks to page
        for block in blocks:
            await client.blocks.children.append(
                block_id=page["id"],
                children=[block]
            )
            await asyncio.sleep(0.05)

        return page["id"]

    def _parse_markdown(self, markdown: str) -> list:
        """Convert markdown to Notion block format."""
        blocks = []
        lines = markdown.split('\n')

        for line in lines:
            if line.startswith('# '):
                blocks.append(self._create_heading(line[2:], level=1))
            elif line.startswith('## '):
                blocks.append(self._create_heading(line[3:], level=2))
            elif line.startswith('- '):
                blocks.append(self._create_bullet(line[2:]))
            elif line.startswith('> '):
                blocks.append(self._create_quote(line[2:]))
            elif line.strip():
                blocks.append(self._create_paragraph(line))

        return blocks

    def _create_heading(self, text: str, level: int) -> dict:
        """Create heading block."""
        heading_type = f"heading_{level}"
        return {
            "type": heading_type,
            heading_type: {
                "rich_text": [{"type": "text", "text": {"content": text}}]
            }
        }

    def _create_bullet(self, text: str) -> dict:
        """Create bullet point."""
        return {
            "type": "bulleted_list_item",
            "bulleted_list_item": {
                "rich_text": [{"type": "text", "text": {"content": text}}]
            }
        }
```

---

## Advanced Filtering and Queries

### Pattern: Complex Multi-Condition Queries

```python
class NotionQueryBuilder:
    """Build complex database queries."""

    def build_advanced_filter(self) -> dict:
        """Create complex AND/OR filter."""
        return {
            "and": [
                {
                    "property": "Status",
                    "select": {"equals": "Active"}
                },
                {
                    "or": [
                        {
                            "property": "Priority",
                            "select": {"equals": "High"}
                        },
                        {
                            "property": "Deadline",
                            "date": {
                                "before": datetime.now().isoformat()
                            }
                        }
                    ]
                },
                {
                    "property": "Assigned To",
                    "people": {
                        "contains": "user_id_123"
                    }
                }
            ]
        }

    async def execute_complex_query(
        self,
        client: AsyncClient,
        database_id: str
    ) -> list:
        """Execute complex query and aggregate results."""

        response = await client.databases.query(
            database_id=database_id,
            filter=self.build_advanced_filter(),
            sorts=[
                {"property": "Priority", "direction": "descending"},
                {"property": "Deadline", "direction": "ascending"}
            ]
        )

        return self._aggregate_results(response["results"])

    def _aggregate_results(self, pages: list) -> list:
        """Aggregate and process query results."""
        results = []

        for page in pages:
            result = {
                "id": page["id"],
                "title": self._extract_title(page["properties"]),
                "status": self._extract_select(page["properties"]["Status"]),
                "priority": self._extract_select(page["properties"]["Priority"]),
                "deadline": self._extract_date(page["properties"]["Deadline"])
            }
            results.append(result)

        return results
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
