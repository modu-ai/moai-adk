// C# suppressed fixture — correctly paired ast-grep-ignore + @MX:REASON.
// checkSuppressionPairing should return 0 violations for this file.

public class Example
{
    public void Code()
    {
        // ast-grep-ignore
        // @MX:REASON test fixture for suppression policy; null assignment intentional for reset
        string x = null;
        _ = x;
    }
}
