# Ruby violation fixture — demonstrates patterns matched by ruby rules.
# This file should produce >= 1 finding when scanned with the ruby rule set.

# Matches ruby-unused-var: variable assigned nil
x = nil

# Matches ruby-nil-method-call: calling length on nil
bad_len = nil.length

# Matches ruby-todo-marker: stub implementation
def not_implemented
  raise "TODO"
end
