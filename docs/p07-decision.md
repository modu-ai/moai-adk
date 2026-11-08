# P07 local TODO promotion

When a forge is connected later, preview existing local TODOs and promote only user-selected items. Review disclosure before any transmission and reconcile stable task IDs with existing remote issues to prevent duplicates. No current repository is authorized for publication by this decision.

`experiments/todo_promotion.py` provides an in-memory selection preview. Nonselected tasks never enter its output; unreviewed text is omitted; mapped tasks are marked already published. An approval label is supplied data, not a secret scanner or remote permission. Remote reconciliation, concurrent publication and actual publishing are not implemented.
