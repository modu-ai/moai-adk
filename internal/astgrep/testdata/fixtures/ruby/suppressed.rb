# Ruby suppressed fixture — correctly paired ast-grep-ignore + @MX:REASON.
# checkSuppressionPairing should return 0 violations for this file.

# ast-grep-ignore
# @MX:REASON test fixture for suppression policy; nil assignment intentional for reset logic
x = nil

result = "hello"
puts result
