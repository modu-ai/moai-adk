// C# valid fixture — no rule violations.
// This file should produce 0 findings when scanned with the csharp rule set.

using System;

public class Example
{
    public void GoodCode()
    {
        string x = "hello";
        var s = Convert.ToString(x);
        Console.WriteLine(s);
    }

    public string Implemented()
    {
        return "actual result";
    }
}
