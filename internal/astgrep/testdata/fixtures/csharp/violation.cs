// C# violation fixture — demonstrates patterns matched by csharp rules.
// This file should produce >= 1 finding when scanned with the csharp rule set.

using System;

public class Example
{
    public void BadCode()
    {
        // Matches csharp-unused-var: string variable assigned null
        string x = null;

        // Matches csharp-null-deref: Convert.ToString with null argument
        var s = Convert.ToString(null);

        Console.WriteLine(x);
        Console.WriteLine(s);
    }

    // Matches csharp-todo-marker: NotImplementedException stub
    public string NotImplemented()
    {
        throw new NotImplementedException();
    }
}
