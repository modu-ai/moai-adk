# Ruby valid fixture — no rule violations
# This file should produce 0 findings when scanned with the ruby rule set.

x = "hello"
result = x.to_s
safe_len = result.length
puts safe_len

def implemented_method
  "actual implementation"
end
