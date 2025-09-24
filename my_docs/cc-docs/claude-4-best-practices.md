# Claude 4 Prompt Engineering Best Practices

## General Principles

### Provide Clear Instructions

Claude 4 models respond well to explicit and clear instructions. Being specific about desired output can help improve results. For previous Claude models that required "above and beyond" behavior, you may now need to explicitly request those behaviors.

Example:

- Less Effective: `Create an analytics dashboard`
- More Effective: `Create an analytics dashboard. Include as many relevant features and interactions as possible. Go beyond the basics to create a fully functional implementation.`

### Add Context for Performance Improvement

Providing context or motivation behind instructions can help Claude 4 better understand goals and deliver more targeted responses.

Example:

- Less Effective: `Never use ellipses`
- More Effective: `Avoid ellipses because your response will be read aloud by a text-to-speech engine, which may not know how to pronounce them.`

### Pay Attention to Examples and Details

Claude 4 pays close attention to details and examples as part of instruction following. Ensure examples align with desired behaviors and minimize undesired actions.

## Specific Situation Guidelines

### Controlling Response Format

Effective methods for adjusting Claude 4's output format:

1. Tell Claude what to do, not what to avoid
   - Instead of "Don't use markdown"
   - Try "Responses should be composed of smoothly flowing prose paragraphs"

2. Use XML format indicators
   - "Write the prose section within <smoothly_flowing_prose_paragraphs> tags"

3. Match prompt style to desired output
   - Adjust prompt formatting to influence response style

### Utilizing Reasoning Capabilities

Claude 4 excels at multi-step reasoning and tool use reflection. Guide initial or cross-reasoning for better results:

```
After receiving tool results, carefully review their quality and determine the optimal next steps before proceeding. Use reasoning to plan and iterate based on new information, then take the best next action.
```

### Optimizing Parallel Tool Calls

Claude 4 performs excellently with parallel tool execution. A recommended prompt:

```
For maximum efficiency, always call all relevant tools simultaneously when performing multiple independent tasks, rather than sequentially.
```
