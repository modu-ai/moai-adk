
// Node.js Implementation for claude-code/cli.js
// Replace the vulnerable code around line 89:708 with this robust version

class RobustJSONParser {
    constructor(debug = false) {
        this.debug = debug;
        this.errors = 0;
        this.processed = 0;
    }
    
    safeParseJSONChunk(chunk, defaultValue = null) {
        // Handle empty/null chunks
        if (!chunk || !chunk.trim()) {
            this.processed++;
            return defaultValue !== null ? defaultValue : { _empty: true };
        }
        
        // Check for obvious non-JSON content
        const trimmed = chunk.trim();
        if (!trimmed.startsWith('{') && !trimmed.startsWith('[') && !trimmed.startsWith('"')) {
            this.processed++;
            return { _text_content: trimmed };
        }
        
        try {
            // Try direct parsing first
            const result = JSON.parse(chunk);
            this.processed++;
            return result;
            
        } catch (error) {
            this.errors++;
            if (this.debug) {
                console.error(`JSON parse error: ${error.message}`);
            }
            
            // Attempt basic repairs
            try {
                const repaired = this.repairJSONChunk(trimmed);
                const result = JSON.parse(repaired);
                this.processed++;
                return { _repaired: true, _originalError: error.message, ...result };
                
            } catch (repairError) {
                // All parsing failed, return safe default
                this.processed++;
                return {
                    _parseError: true,
                    _error: error.message,
                    _originalChunk: chunk.length > 100 ? chunk.substring(0, 100) + '...' : chunk,
                    ...(defaultValue !== null ? defaultValue : {})
                };
            }
        }
    }
    
    repairJSONChunk(chunk) {
        // Basic repairs for common JSON issues
        if (chunk.endsWith('{')) return '{}';
        if (chunk.endsWith('[')) return '[]';
        
        // Remove trailing commas
        chunk = chunk.replace(/,}/g, '}').replace(/,]/g, ']');
        
        // Handle incomplete key-value pairs
        if (chunk.includes('{"incomplete":') && !chunk.endsWith('}')) {
            return chunk + '"_incomplete":true}';
        }
        
        return chunk;
    }
}

// Usage in claude-code/cli.js around line 89:708
// Replace vulnerable JSON.parse calls with:

const parser = new RobustJSONParser(process.env.DEBUG === 'true');

// Instead of: const data = JSON.parse(chunk);
// Use: const data = parser.safeParseJSONChunk(chunk, { fallback: true });

// This prevents: SyntaxError: Expected property name or '}' in JSON at position 1
