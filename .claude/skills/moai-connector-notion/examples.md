# Notion Integration Examples

## Example 1: Basic Page Creation

```python
import asyncio
from notion_client import AsyncClient

async def create_simple_page():
    """Create a basic Notion page."""
    client = AsyncClient(auth=NOTION_API_KEY)

    response = await client.pages.create(
        parent={"database_id": "database_id_here"},
        properties={
            "Name": {
                "title": [
                    {
                        "text": {
                            "content": "My First Page"
                        }
                    }
                ]
            },
            "Tags": {
                "multi_select": [
                    {"name": "Important"},
                    {"name": "Review"}
                ]
            }
        }
    )

    return response["id"]

# Usage
page_id = asyncio.run(create_simple_page())
print(f"Created page: {page_id}")
```

---

## Example 2: Database Query with Filtering

```python
async def query_active_tasks():
    """Query tasks with 'Active' status."""
    client = AsyncClient(auth=NOTION_API_KEY)

    response = await client.databases.query(
        database_id="database_id_here",
        filter={
            "property": "Status",
            "select": {
                "equals": "Active"
            }
        },
        sorts=[
            {
                "property": "Created",
                "direction": "descending"
            }
        ]
    )

    tasks = []
    for result in response["results"]:
        task_name = result["properties"]["Name"]["title"][0]["text"]["content"]
        status = result["properties"]["Status"]["select"]["name"]
        tasks.append({"id": result["id"], "name": task_name, "status": status})

    return tasks

# Usage
tasks = asyncio.run(query_active_tasks())
for task in tasks:
    print(f"- {task['name']} ({task['status']})")
```

---

## Example 3: Update Page Properties

```python
async def update_task_status(page_id: str, new_status: str):
    """Update task status to completed."""
    client = AsyncClient(auth=NOTION_API_KEY)

    response = await client.pages.update(
        page_id=page_id,
        properties={
            "Status": {
                "select": {
                    "name": new_status
                }
            },
            "Completed": {
                "checkbox": new_status == "Done"
            },
            "Last Updated": {
                "date": {
                    "start": datetime.now().isoformat()
                }
            }
        }
    )

    return response

# Usage
asyncio.run(update_task_status("page_id_123", "Done"))
```

---

## Example 4: Add Block Content to Page

```python
async def add_page_content(page_id: str):
    """Add rich content blocks to a page."""
    client = AsyncClient(auth=NOTION_API_KEY)

    await client.blocks.children.append(
        block_id=page_id,
        children=[
            {
                "object": "block",
                "type": "heading_2",
                "heading_2": {
                    "rich_text": [{"type": "text", "text": {"content": "Overview"}}],
                    "color": "default"
                }
            },
            {
                "object": "block",
                "type": "paragraph",
                "paragraph": {
                    "rich_text": [
                        {
                            "type": "text",
                            "text": {"content": "This is the main content with "},
                            "annotations": {"bold": False}
                        },
                        {
                            "type": "text",
                            "text": {"content": "bold"},
                            "annotations": {"bold": True}
                        },
                        {
                            "type": "text",
                            "text": {"content": " formatting."}
                        }
                    ]
                }
            },
            {
                "object": "block",
                "type": "bulleted_list_item",
                "bulleted_list_item": {
                    "rich_text": [{"type": "text", "text": {"content": "First item"}}]
                }
            },
            {
                "object": "block",
                "type": "bulleted_list_item",
                "bulleted_list_item": {
                    "rich_text": [{"type": "text", "text": {"content": "Second item"}}]
                }
            }
        ]
    )

# Usage
asyncio.run(add_page_content("page_id_123"))
```

---

## Example 5: Bulk Create Pages from List

```python
async def bulk_create_pages(items: list, database_id: str):
    """Create multiple pages in batch."""
    client = AsyncClient(auth=NOTION_API_KEY)
    created_pages = []

    for item in items:
        response = await client.pages.create(
            parent={"database_id": database_id},
            properties={
                "Title": {
                    "title": [{"text": {"content": item["title"]}}]
                },
                "Description": {
                    "rich_text": [{"text": {"content": item.get("description", "")}}]
                },
                "Priority": {
                    "select": {"name": item.get("priority", "Medium")}
                },
                "Assigned To": {
                    "people": [{"id": item.get("assignee_id")}] if item.get("assignee_id") else []
                }
            }
        )
        created_pages.append(response["id"])
        # Add delay to respect rate limits
        await asyncio.sleep(0.1)

    return created_pages

# Usage
items = [
    {"title": "Task 1", "description": "First task", "priority": "High"},
    {"title": "Task 2", "description": "Second task", "priority": "Medium"},
    {"title": "Task 3", "description": "Third task", "priority": "Low"}
]
page_ids = asyncio.run(bulk_create_pages(items, "database_id_here"))
print(f"Created {len(page_ids)} pages")
```

---

## Example 6: Search Notion Workspace

```python
async def search_workspace(query: str):
    """Search entire Notion workspace."""
    client = AsyncClient(auth=NOTION_API_KEY)

    response = await client.search(
        query=query,
        filter={"value": "page", "property": "object"},
        sort={"direction": "ascending", "timestamp": "last_edited_time"}
    )

    results = []
    for item in response["results"]:
        title = "Untitled"
        if item["object"] == "page" and "title" in item:
            title = item.get("properties", {}).get("title", "Untitled")

        results.append({
            "id": item["id"],
            "title": title,
            "type": item["object"],
            "last_edited": item["last_edited_time"]
        })

    return results

# Usage
results = asyncio.run(search_workspace("project"))
for result in results:
    print(f"- {result['title']} ({result['type']})")
```

---

## Example 7: Create Database with Relations

```python
async def create_linked_database():
    """Create database with relationship properties."""
    client = AsyncClient(auth=NOTION_API_KEY)

    response = await client.databases.create(
        parent={"page_id": "page_id_here"},
        title="Projects with Tasks",
        properties={
            "Name": {"title": {}},
            "Status": {
                "select": {
                    "options": [
                        {"name": "Not Started", "color": "gray"},
                        {"name": "In Progress", "color": "yellow"},
                        {"name": "Completed", "color": "green"}
                    ]
                }
            },
            "Related Tasks": {
                "relation": {
                    "database_id": "tasks_database_id",
                    "synced_property_name": "Related Project"
                }
            },
            "Owner": {
                "people": {}
            }
        }
    )

    return response["id"]

# Usage
db_id = asyncio.run(create_linked_database())
print(f"Created database: {db_id}")
```

---

## Example 8: Sync External Data to Notion

```python
import httpx

async def sync_github_issues_to_notion():
    """Sync GitHub issues to Notion database."""
    client = AsyncClient(auth=NOTION_API_KEY)
    notion_db_id = "notion_database_id"

    # Fetch GitHub issues
    async with httpx.AsyncClient() as http_client:
        response = await http_client.get(
            "https://api.github.com/repos/owner/repo/issues",
            headers={"Authorization": f"token {GITHUB_TOKEN}"}
        )
        issues = response.json()

    # Create Notion pages for each issue
    for issue in issues:
        await client.pages.create(
            parent={"database_id": notion_db_id},
            properties={
                "Title": {
                    "title": [{"text": {"content": issue["title"]}}]
                },
                "URL": {
                    "url": issue["html_url"]
                },
                "Status": {
                    "select": {"name": "Open" if issue["state"] == "open" else "Closed"}
                },
                "Labels": {
                    "multi_select": [
                        {"name": label["name"]} for label in issue["labels"]
                    ]
                },
                "Created": {
                    "date": {"start": issue["created_at"]}
                }
            }
        )
        await asyncio.sleep(0.1)  # Rate limiting

# Usage
asyncio.run(sync_github_issues_to_notion())
```

---

## Example 9: Handle Errors and Retries

```python
import asyncio
from notion_client import APIResponseError

async def create_page_with_retry(max_retries: int = 3):
    """Create page with exponential backoff retry."""
    client = AsyncClient(auth=NOTION_API_KEY)

    for attempt in range(max_retries):
        try:
            response = await client.pages.create(
                parent={"database_id": "database_id"},
                properties={
                    "Name": {
                        "title": [{"text": {"content": "Test Page"}}]
                    }
                }
            )
            return response["id"]

        except APIResponseError as e:
            if e.status_code == 429:  # Rate limited
                wait_time = 2 ** attempt  # Exponential backoff
                print(f"Rate limited. Waiting {wait_time}s...")
                await asyncio.sleep(wait_time)
            else:
                print(f"API Error: {e}")
                raise

        except Exception as e:
            print(f"Unexpected error: {e}")
            if attempt == max_retries - 1:
                raise
            await asyncio.sleep(0.5)

    raise Exception("Failed to create page after retries")

# Usage
try:
    page_id = asyncio.run(create_page_with_retry())
    print(f"Created page: {page_id}")
except Exception as e:
    print(f"Failed: {e}")
```

---

## Example 10: Complex Database Query

```python
async def advanced_query():
    """Query with multiple filters and conditions."""
    client = AsyncClient(auth=NOTION_API_KEY)

    response = await client.databases.query(
        database_id="database_id",
        filter={
            "and": [
                {
                    "property": "Status",
                    "select": {"equals": "In Progress"}
                },
                {
                    "property": "Priority",
                    "select": {"equals": "High"}
                },
                {
                    "property": "Due Date",
                    "date": {
                        "before": (datetime.now() + timedelta(days=7)).isoformat()
                    }
                },
                {
                    "or": [
                        {
                            "property": "Assigned To",
                            "people": {
                                "contains": "user_id_123"
                            }
                        },
                        {
                            "property": "Assigned To",
                            "people": {
                                "is_empty": True
                            }
                        }
                    ]
                }
            ]
        },
        sorts=[
            {"property": "Priority", "direction": "descending"},
            {"property": "Due Date", "direction": "ascending"}
        ],
        page_size=100
    )

    results = []
    for page in response["results"]:
        results.append({
            "id": page["id"],
            "title": page["properties"]["Title"]["title"][0]["text"]["content"],
            "status": page["properties"]["Status"]["select"]["name"],
            "priority": page["properties"]["Priority"]["select"]["name"]
        })

    return results

# Usage
results = asyncio.run(advanced_query())
print(f"Found {len(results)} high-priority in-progress items")
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
