# Reducing Hallucinations in Claude

## Basic Hallucination Minimization Strategies

### 1. Allow Claude to Say "I Don't Know"

Explicitly permit Claude to acknowledge uncertainty. This simple technique can significantly reduce incorrect information.

**Example Prompt for M&A Report Analysis:**

- Instruct Claude to state "I do not have sufficient information to evaluate this" when uncertain
- Focus on financial prospects, integration risks, and regulatory barriers
- Encourage transparent communication about knowledge gaps

### 2. Use Direct Quotations for Fact Basis

For long documents (>20K tokens), request Claude to:

- Extract verbatim quotes before performing analysis
- Ground responses directly in source text
- Minimize potential hallucinations by anchoring to original content

**Example: Data Privacy Policy Audit**

- Review GDPR and CCPA compliance
- Extract precise policy quotations
- Analyze compliance using only extracted text
- Explicitly note when no relevant quotations exist

### 3. Verify with Citations

Require Claude to:

- Provide citations for each claim
- Validate claims against source documents
- Remove statements without supporting evidence

**Example: Product Launch Press Release**

- Use only provided product brief and market reports
- Review each claim
- Find direct supporting quotations
- Remove unsupported statements

## Advanced Techniques

### Thought Chain Verification

- Request Claude to explain reasoning step-by-step
- Identify potential logical errors or assumptions

### Multiple Iteration Validation

- Run the same prompt multiple times
- Compare outputs for inconsistencies

### Iterative Improvement

- Use Claude's previous output as input for verification
- Detect and correct discrepancies

### Limit External Knowledge

- Explicitly instruct Claude to use only provided documents
- Prevent drawing from general knowledge

## Important Caveat

> These techniques significantly reduce hallucinations but cannot completely eliminate them.

Always verify critical information, especially for important decisions.
