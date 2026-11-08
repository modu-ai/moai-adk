"""Read-only selected TODO promotion preview; contains no publishing client."""
def preview(tasks, selected, published):
    by_id = {}
    for task in tasks:
        key = task["id"]
        if key in by_id:
            raise ValueError("duplicate task identity")
        by_id[key] = task
    if len(selected) != len(set(selected)) or not set(selected).issubset(by_id):
        raise ValueError("invalid selection")
    rows = []
    for key in selected:
        task = by_id[key]
        if key in published:
            rows.append({"id": key, "state": "already_published"})
        elif task.get("disclosure_review") != "approved":
            rows.append({"id": key, "state": "blocked_review"})
        else:
            rows.append({"id": key, "state": "preview_only", "title": task["title"]})
    return rows
