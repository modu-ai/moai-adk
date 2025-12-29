# Advanced Implementation Patterns

Advanced patterns for custom template development, performance optimization, and integration workflows.

---

## Custom Template Development

### Documentation Templates

To create a custom template for a specific project type, define the following structure:

Template Configuration Fields:
- Project type: Specify the target project type such as mobile_application, web_application, or cli_tool
- Language: Set the primary language code
- Sections: Define the documentation sections

For Product Section Configuration:
- Mission: Define the product mission statement
- Metrics: List the key performance indicators
- Success criteria: Define measurable success conditions

For Technical Section Configuration:
- Frameworks: List the technology frameworks used
- Performance targets: Define performance requirements

After configuring the template, generate documentation by specifying the project type and language parameters.

### Language-Specific Customization

To add support for a custom language, configure the following fields:
- Code: The language code such as "de" for German
- Name: The display name in English such as "German"
- Native name: The display name in the native language such as "Deutsch"
- Locale: The system locale string such as "de_DE.UTF-8"
- Date format: The date formatting pattern
- RTL: Whether the language uses right-to-left text direction

After defining the custom language configuration, register it with the language initialization system.

---

## Performance Optimization Strategies

### Template Caching

To enable template caching for improved performance:

Step 1: Initialize an empty cache storage for optimization results

Step 2: Implement cached optimization by following this process:
- Generate a cache key using the template path and current date
- Check if the cache key exists in the optimization cache
- If the key does not exist, execute the template optimization and store the result
- If the key exists, retrieve the cached result

Step 3: Use the cached optimization function for all template operations to avoid redundant processing

### Batch Processing

To process multiple templates efficiently:

Step 1: Collect the list of template file paths to process

Step 2: For each template in the list:
- Attempt to optimize the template file
- If successful: Add the result to the results collection
- If an error occurs: Record the file path, mark success as false, and include the error message

Step 3: Return the complete results collection for review

---

## Integration Workflows

### Complete Project Lifecycle

The full project lifecycle workflow consists of five phases:

Phase 1 - Project Initialization:
- Create a new project instance with the target directory
- Execute complete initialization with language, domains, and optimization settings

Phase 2 - Feature Development with SPEC:
- Prepare SPEC data with identifier, title, requirements, and API endpoint definitions
- Generate documentation from the SPEC data

Phase 3 - Performance Optimization:
- Execute template optimization to improve performance and reduce file sizes

Phase 4 - Documentation Export:
- Export project documentation to the desired format such as HTML, Markdown, or PDF

Phase 5 - Backup Creation:
- Create a complete project backup for safety and version control

Review the results from each phase to verify successful completion.

### Multilingual Project Management

To manage a multilingual project with cost optimization:

Step 1: Initialize the project with the primary language
- Set the language to the primary user language such as Korean
- Configure the user name for personalization
- Specify the project domains

Step 2: Optimize agent prompts for cost efficiency
- Update the agent prompt language setting to English
- This reduces token usage while maintaining user-facing localization

Step 3: Generate documentation in multiple languages
- For each target language such as Korean, English, and Japanese:
  - Export documentation in Markdown format for the specified language
  - Verify successful export completion

Step 4: Review the initialization and export results to confirm multilingual setup

---

For detailed implementation patterns, refer to the main SKILL.md Implementation Guide section.
