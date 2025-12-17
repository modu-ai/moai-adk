# Notion MCP Advanced Patterns

## 고급 데이터베이스 쿼리

### 복잡한 필터 조건

```python
async def advanced_database_query(database_id: str):
 """고급 필터 및 정렬이 적용된 데이터베이스 쿼리."""

 query_params = {
 "database_id": database_id,
 "filter": {
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
 "date": {"on_or_before": "2025-12-31"}
 }
 ]
 },
 "sorts": [
 {"property": "Priority", "direction": "ascending"},
 {"property": "Due Date", "direction": "ascending"}
 ]
 }

 result = await notion_server.invoke_tool("query_notion_database", query_params)

 return {
 "results": result["results"],
 "has_more": result["has_more"],
 "next_cursor": result.get("next_cursor")
 }
```

### 페이지네이션 처리

```python
async def fetch_all_pages(database_id: str):
 """페이지네이션을 통한 전체 데이터베이스 레코드 추출."""

 all_results = []
 has_more = True
 start_cursor = None

 while has_more:
 query_params = {
 "database_id": database_id,
 "start_cursor": start_cursor
 }

 response = await notion_server.invoke_tool("query_notion_database", query_params)

 all_results.extend(response["results"])
 has_more = response["has_more"]
 start_cursor = response.get("next_cursor")

 return {
 "total_count": len(all_results),
 "results": all_results
 }
```

---

## 페이지 및 블록 관리

### 복잡한 페이지 구조 생성

```python
async def create_structured_page(database_id: str, page_data: dict):
 """구조화된 콘텐츠를 가진 페이지 생성."""

 # 페이지 생성
 page = await notion_server.invoke_tool("create_notion_page", {
 "database_id": database_id,
 "properties": page_data["properties"]
 })

 # 블록 콘텐츠 추가
 blocks = [
 {
 "type": "heading_1",
 "heading_1": {
 "rich_text": [{"text": {"content": page_data["title"]}}]
 }
 },
 {
 "type": "paragraph",
 "paragraph": {
 "rich_text": [{"text": {"content": page_data["introduction"]}}]
 }
 },
 {
 "type": "heading_2",
 "heading_2": {
 "rich_text": [{"text": {"content": "Requirements"}}]
 }
 },
 {
 "type": "bulleted_list_item",
 "bulleted_list_item": {
 "rich_text": [{"text": {"content": req}}]
 }
 } for req in page_data["requirements"]
 ]

 await notion_server.invoke_tool("append_blocks", {
 "page_id": page["id"],
 "blocks": blocks
 })

 return page
```

### 중첩 블록 구조

```python
async def create_nested_structure(page_id: str, structure: dict):
 """중첩된 블록 구조 생성 (예: 토글 리스트 내부에 콘텐츠)."""

 toggle_block = {
 "type": "toggle",
 "toggle": {
 "rich_text": [{"text": {"content": structure["toggle_title"]}}],
 "children": [
 {
 "type": "paragraph",
 "paragraph": {
 "rich_text": [{"text": {"content": content}}]
 }
 } for content in structure["nested_content"]
 ]
 }
 }

 result = await notion_server.invoke_tool("append_blocks", {
 "page_id": page_id,
 "blocks": [toggle_block]
 })

 return result
```

---

## 지식 베이스 워크플로우

### 자동화된 지식 추출

```python
async def knowledge_extraction_workflow(database_id: str, analysis_goals: list):
 """Notion 데이터베이스에서 지식 추출 및 구조화."""

 # 1단계: 데이터베이스 쿼리
 content = await notion_server.invoke_tool("query_notion_database", {
 "database_id": database_id,
 "filter": {
 "property": "Status",
 "select": {"equals": "Published"}
 }
 })

 # 2단계: 페이지 콘텐츠 추출
 pages_content = []
 for page in content["results"]:
 blocks = await notion_server.invoke_tool("get_page_blocks", {
 "page_id": page["id"]
 })
 pages_content.append({
 "page_id": page["id"],
 "properties": page["properties"],
 "content": blocks["results"]
 })

 # 3단계: AI 분석 (MCP 통합)
 analyses = {}
 for goal in analysis_goals:
 analysis = await notion_server.invoke_tool("analyze_with_ai", {
 "content": json.dumps(pages_content),
 "analysis_type": goal,
 "max_tokens": 3000
 })
 analyses[goal] = analysis

 # 4단계: 구조화된 지식 베이스 생성
 structured_kb = await notion_server.invoke_tool("generate_structured_knowledge", {
 "analyses": analyses,
 "source_count": len(pages_content)
 })

 return {
 "raw_content": pages_content,
 "analyses": analyses,
 "structured_knowledge": structured_kb,
 "metadata": {
 "source_count": len(pages_content),
 "analysis_goals": analysis_goals
 }
 }
```

### 문서 자동 생성

```python
async def generate_documentation_page(database_id: str, spec_data: dict):
 """SPEC 데이터를 기반으로 Notion 문서 페이지 생성."""

 # 페이지 속성 구성
 properties = {
 "Title": {
 "title": [{"text": {"content": spec_data["title"]}}]
 },
 "Category": {
 "select": {"name": "Technical Documentation"}
 },
 "Status": {
 "select": {"name": "Draft"}
 },
 "SPEC ID": {
 "rich_text": [{"text": {"content": spec_data["id"]}}]
 }
 }

 # 페이지 생성
 page = await notion_server.invoke_tool("create_notion_page", {
 "database_id": database_id,
 "properties": properties
 })

 # 문서 콘텐츠 블록 구성
 blocks = [
 {
 "type": "heading_1",
 "heading_1": {
 "rich_text": [{"text": {"content": spec_data["title"]}}]
 }
 },
 {
 "type": "paragraph",
 "paragraph": {
 "rich_text": [{"text": {"content": spec_data["description"]}}]
 }
 },
 {
 "type": "heading_2",
 "heading_2": {
 "rich_text": [{"text": {"content": "Requirements"}}]
 }
 }
 ]

 # 요구사항 항목 추가
 for req in spec_data.get("requirements", []):
 blocks.append({
 "type": "bulleted_list_item",
 "bulleted_list_item": {
 "rich_text": [{"text": {"content": req}}]
 }
 })

 # API 엔드포인트 섹션
 if spec_data.get("api_endpoints"):
 blocks.append({
 "type": "heading_2",
 "heading_2": {
 "rich_text": [{"text": {"content": "API Endpoints"}}]
 }
 })

 for endpoint in spec_data["api_endpoints"]:
 blocks.append({
 "type": "code",
 "code": {
 "rich_text": [{
 "text": {
 "content": f"{endpoint['method']} {endpoint['path']}\n{endpoint['description']}"
 }
 }],
 "language": "plain text"
 }
 })

 # 블록 추가
 await notion_server.invoke_tool("append_blocks", {
 "page_id": page["id"],
 "blocks": blocks
 })

 return {
 "page_id": page["id"],
 "page_url": page["url"],
 "spec_id": spec_data["id"]
 }
```

---

## 템플릿 자동화

### 재사용 가능한 페이지 템플릿

```python
class NotionTemplateManager:
 """Notion 페이지 템플릿 관리자."""

 def __init__(self, notion_server):
 self.server = notion_server
 self.templates = {}

 def register_template(self, template_name: str, template_config: dict):
 """템플릿 등록."""
 self.templates[template_name] = template_config

 async def create_from_template(
 self,
 template_name: str,
 database_id: str,
 variables: dict
 ):
 """템플릿을 사용하여 페이지 생성."""

 template = self.templates.get(template_name)
 if not template:
 raise ValueError(f"Template {template_name} not found")

 # 변수를 템플릿에 적용
 properties = self._apply_variables(template["properties"], variables)
 blocks = self._apply_variables(template["blocks"], variables)

 # 페이지 생성
 page = await self.server.invoke_tool("create_notion_page", {
 "database_id": database_id,
 "properties": properties
 })

 # 블록 추가
 await self.server.invoke_tool("append_blocks", {
 "page_id": page["id"],
 "blocks": blocks
 })

 return page

 def _apply_variables(self, template_data: dict, variables: dict):
 """템플릿에 변수 적용."""
 import json
 template_json = json.dumps(template_data)

 for key, value in variables.items():
 template_json = template_json.replace(f"{{{{ {key} }}}}", str(value))

 return json.loads(template_json)

# 사용 예시
template_manager = NotionTemplateManager(notion_server)

# 템플릿 등록
template_manager.register_template("project_spec", {
 "properties": {
 "Title": {"title": [{"text": {"content": "{{ title }}"}}]},
 "Category": {"select": {"name": "{{ category }}"}},
 "Status": {"select": {"name": "Planning"}}
 },
 "blocks": [
 {
 "type": "heading_1",
 "heading_1": {"rich_text": [{"text": {"content": "{{ title }}"}}]}
 },
 {
 "type": "paragraph",
 "paragraph": {"rich_text": [{"text": {"content": "{{ description }}"}}]}
 }
 ]
})

# 템플릿으로 페이지 생성
page = await template_manager.create_from_template(
 "project_spec",
 database_id="abc123",
 variables={
 "title": "User Authentication System",
 "category": "Backend",
 "description": "Implement JWT-based authentication"
 }
)
```

---

## 배치 작업 및 동기화

### 배치 페이지 업데이트

```python
async def batch_update_pages(page_ids: list, updates: dict):
 """여러 페이지를 배치로 업데이트."""

 results = []

 for page_id in page_ids:
 try:
 result = await notion_server.invoke_tool("update_notion_page", {
 "page_id": page_id,
 "properties": updates
 })
 results.append({
 "page_id": page_id,
 "status": "success",
 "result": result
 })
 except Exception as e:
 results.append({
 "page_id": page_id,
 "status": "error",
 "error": str(e)
 })

 return {
 "total": len(page_ids),
 "successful": sum(1 for r in results if r["status"] == "success"),
 "failed": sum(1 for r in results if r["status"] == "error"),
 "results": results
 }
```

### 데이터베이스 동기화

```python
async def sync_database_to_external(database_id: str, external_api: dict):
 """Notion 데이터베이스를 외부 시스템과 동기화."""

 # 1단계: Notion에서 전체 데이터 추출
 all_pages = await fetch_all_pages(database_id)

 # 2단계: 외부 API와 비교 및 동기화
 sync_results = {
 "created": [],
 "updated": [],
 "deleted": [],
 "errors": []
 }

 for page in all_pages["results"]:
 try:
 external_id = page["properties"].get("External_ID", {}).get("rich_text", [{}])[0].get("text", {}).get("content")

 if not external_id:
 # 새 레코드 생성
 external_record = await create_external_record(page, external_api)
 sync_results["created"].append(external_record)
 else:
 # 기존 레코드 업데이트
 external_record = await update_external_record(external_id, page, external_api)
 sync_results["updated"].append(external_record)

 except Exception as e:
 sync_results["errors"].append({
 "page_id": page["id"],
 "error": str(e)
 })

 return sync_results
```

---

## 실시간 이벤트 처리

### Webhook 통합

```python
from fastapi import FastAPI, Request
import hmac
import hashlib

app = FastAPI()

@app.post("/notion/webhook")
async def handle_notion_webhook(request: Request):
 """Notion Webhook 이벤트 처리."""

 # 서명 검증
 signature = request.headers.get("Notion-Signature")
 body = await request.body()

 if not verify_signature(signature, body):
 return {"error": "Invalid signature"}, 401

 # 이벤트 처리
 event = await request.json()

 event_type = event.get("type")

 if event_type == "page.created":
 await handle_page_created(event["data"])
 elif event_type == "page.updated":
 await handle_page_updated(event["data"])
 elif event_type == "database.updated":
 await handle_database_updated(event["data"])

 return {"status": "processed"}

def verify_signature(signature: str, body: bytes) -> bool:
 """Webhook 서명 검증."""
 secret = os.getenv("NOTION_WEBHOOK_SECRET")
 expected = hmac.new(
 secret.encode(),
 body,
 hashlib.sha256
 ).hexdigest()

 return hmac.compare_digest(signature, expected)

async def handle_page_created(page_data: dict):
 """페이지 생성 이벤트 처리."""
 # 자동화된 작업 실행
 page_id = page_data["id"]

 # 예: 새 프로젝트 페이지에 기본 템플릿 적용
 await notion_server.invoke_tool("append_blocks", {
 "page_id": page_id,
 "blocks": get_default_template_blocks()
 })
```

---

*더 자세한 패턴과 최적화 전략은 Notion API 공식 문서를 참조하세요.*
