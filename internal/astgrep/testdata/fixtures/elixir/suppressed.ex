# Elixir suppressed fixture — correctly paired ast-grep-ignore + @MX:REASON.
# checkSuppressionPairing should return 0 violations for this file.

defmodule Example do
  def code do
    # ast-grep-ignore
    # @MX:REASON test fixture for suppression policy; _unused assignment intentional for testing
    _unused = :test_value
    :ok
  end
end
