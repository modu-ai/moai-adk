"""Show full-line context of two suspicious words in a tracked doc."""
path = ".moai/docs/git-workflow-doctrine.md"
lines = open(path, encoding="utf-8").read().split("\n")
print("L237:", lines[236])
print()
print("L274:", lines[273])
