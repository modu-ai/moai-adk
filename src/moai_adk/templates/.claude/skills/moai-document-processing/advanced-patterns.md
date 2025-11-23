---
name: moai-document-processing/advanced-patterns
description: Advanced document processing patterns, OCR integration, format transformations, and content extraction
---

# Advanced Document Processing Patterns (v5.0.0)

## Intelligent Document Classification

### 1. Multi-Modal Document Classification

```typescript
interface DocumentMetadata {
    format: 'pdf' | 'docx' | 'xlsx' | 'pptx' | 'text' | 'markdown';
    mimeType: string;
    size: number;
    pages?: number;
    languages: string[];
    documentType: DocumentType;
    confidence: number;
}

type DocumentType = 'invoice' | 'contract' | 'report' | 'presentation' | 'spreadsheet' | 'article' | 'other';

class IntelligentDocumentClassifier {
    async classifyDocument(
        file: Buffer,
        filename: string
    ): Promise<DocumentMetadata> {
        // Step 1: Detect file format
        const format = this.detectFormat(file, filename);

        // Step 2: Extract text content
        const content = await this.extractText(file, format);

        // Step 3: Analyze content structure
        const structure = await this.analyzeStructure(content, format);

        // Step 4: Use ML to classify
        const classification = await this.mlClassifier.classify({
            content,
            structure,
            format,
            filesize: file.length
        });

        // Step 5: Detect languages
        const languages = await this.detectLanguages(content);

        return {
            format,
            mimeType: this.getMimeType(format),
            size: file.length,
            pages: structure.pageCount,
            languages,
            documentType: classification.type,
            confidence: classification.confidence
        };
    }

    private async analyzeStructure(content: string, format: string) {
        const structure = {
            pageCount: content.split('\f').length,
            headings: (content.match(/^#+\s/gm) || []).length,
            tables: (content.match(/\|.*\|/gm) || []).length,
            lists: (content.match(/^[-*•]\s/gm) || []).length,
            codeBlocks: (content.match(/```/g) || []).length / 2,
            links: (content.match(/https?:\/\/\S+/g) || []).length
        };

        return structure;
    }
}
```

## OCR and Handwriting Recognition

### 1. Advanced OCR with Context

```typescript
import Tesseract from 'tesseract.js';
import { Client } = require('@microsoft/cognitiveservices-vision-computervision');

class AdvancedOCRProcessor {
    async extractTextWithContext(
        imagePath: string,
        language: string = 'eng'
    ): Promise<OCRResult> {
        // Use Tesseract for local OCR
        const tesseractResult = await this.runTesseract(imagePath, language);

        // Use Azure Vision API for enhanced accuracy
        const azureResult = await this.runAzureVision(imagePath);

        // Ensemble results for better accuracy
        const ensembleResult = this.ensembleOCRResults(
            tesseractResult,
            azureResult
        );

        // Post-process with language model
        const corrected = await this.correctWithLanguageModel(
            ensembleResult.text
        );

        return {
            text: corrected.text,
            confidence: Math.max(tesseractResult.confidence, azureResult.confidence),
            tables: ensembleResult.tables,
            handwriting: azureResult.handwritingConfidence > 0.7,
            layouts: ensembleResult.layouts,
            metadata: {
                processed_at: new Date(),
                processor: 'ensemble_ocr',
                source: imagePath
            }
        };
    }

    private async runTesseract(imagePath: string, language: string) {
        const { data } = await Tesseract.recognize(
            imagePath,
            language,
            { logger: m => console.log(m) }
        );

        return {
            text: data.text,
            confidence: data.confidence / 100,
            words: data.words.map(w => ({
                text: w.text,
                confidence: w.confidence / 100,
                bbox: w.bbox
            }))
        };
    }

    private async runAzureVision(imagePath: string) {
        const client = new Client(this.azureConfig);

        const result = await client.recognizeTextInStream(
            fs.createReadStream(imagePath)
        );

        // Extract handwriting and print text
        const handwritingRegions = result.regions
            .filter(r => r.metadata?.estimatedSkewAngle || false);

        return {
            text: result.regions.map(r =>
                r.lines.map(l => l.words.map(w => w.text).join(' ')).join('\n')
            ).join('\n\n'),
            confidence: 0.95,
            handwritingConfidence: handwritingRegions.length / result.regions.length
        };
    }

    private ensembleOCRResults(tesseract: any, azure: any) {
        // Merge results from multiple OCR engines
        const tesseractWords = new Map(
            tesseract.words.map((w: any) => [w.text, w])
        );

        // Use majority voting for better accuracy
        const consensusText = this.buildConsensusText(
            tesseract.text,
            azure.text,
            tesseractWords
        );

        return {
            text: consensusText,
            tables: this.extractTables(tesseract, azure),
            layouts: this.analyzeLayouts(tesseract.words)
        };
    }
}
```

## Content Extraction and Transformation

### 1. Structured Data Extraction

```typescript
interface StructuredContent {
    title: string;
    abstract?: string;
    sections: Section[];
    metadata: Metadata;
    citations: Citation[];
}

interface Section {
    heading: string;
    level: number;
    content: string;
    subsections: Section[];
}

class ContentExtractor {
    async extractStructuredContent(
        text: string,
        documentType: DocumentType
    ): Promise<StructuredContent> {
        // Parse document structure
        const structure = this.parseStructure(text);

        // Extract metadata
        const metadata = await this.extractMetadata(text, structure);

        // Identify sections
        const sections = this.extractSections(text, structure);

        // Extract citations/references
        const citations = this.extractCitations(text);

        // Generate abstract if not present
        const abstract = metadata.abstract || await this.generateAbstract(text);

        return {
            title: metadata.title,
            abstract,
            sections,
            metadata,
            citations
        };
    }

    private parseStructure(text: string) {
        // Identify hierarchical structure
        const lines = text.split('\n');
        const structure: Array<{ level: number; text: string; index: number }> = [];

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];

            // Detect heading levels
            if (line.match(/^#+\s/)) {
                const level = (line.match(/^#+/)[0]).length;
                structure.push({ level, text: line, index: i });
            }
        }

        return structure;
    }

    private extractSections(text: string, structure: any[]): Section[] {
        const sections: Section[] = [];

        for (let i = 0; i < structure.length; i++) {
            const current = structure[i];
            const next = structure[i + 1];

            const start = current.index;
            const end = next ? next.index : text.length;
            const content = text.substring(start, end).trim();

            const section: Section = {
                heading: current.text.replace(/^#+\s/, ''),
                level: current.level,
                content,
                subsections: []
            };

            sections.push(section);
        }

        return sections;
    }

    private async generateAbstract(text: string): Promise<string> {
        // Use extractive summarization
        const sentences = text.match(/[^.!?]+[.!?]+/g) || [];

        // Score sentences by importance
        const scores = sentences.map(sent =>
            this.scoreSentence(sent, text)
        );

        // Select top 3-5 sentences
        const topSentences = scores
            .sort((a, b) => b.score - a.score)
            .slice(0, Math.min(5, Math.ceil(sentences.length * 0.3)))
            .sort((a, b) => a.index - b.index)
            .map(s => s.sentence)
            .join(' ');

        return topSentences;
    }

    private scoreSentence(sentence: string, document: string): { sentence: string; score: number; index: number } {
        // TF-IDF scoring
        const words = sentence.toLowerCase().split(/\s+/);
        let score = 0;

        for (const word of words) {
            const tf = (words.filter(w => w === word).length) / words.length;
            const idf = Math.log(document.length / (document.split(word).length - 1));
            score += tf * idf;
        }

        return {
            sentence,
            score,
            index: document.indexOf(sentence)
        };
    }
}
```

### 2. Format Transformation Pipeline

```typescript
class DocumentTransformer {
    async transformDocument(
        input: DocumentInput,
        targetFormat: string
    ): Promise<Buffer> {
        // Step 1: Normalize to intermediate format
        const intermediate = await this.normalizeToIntermediate(input);

        // Step 2: Validate structure
        await this.validateStructure(intermediate);

        // Step 3: Transform to target format
        const output = await this.transformFromIntermediate(
            intermediate,
            targetFormat
        );

        // Step 4: Post-process
        const final = await this.postProcess(output, targetFormat);

        return final;
    }

    private async normalizeToIntermediate(input: DocumentInput) {
        const normalized = {
            metadata: input.metadata,
            content: '',
            structure: [],
            assets: []
        };

        switch (input.format) {
            case 'pdf':
                normalized.content = await this.extractFromPDF(input.buffer);
                break;
            case 'docx':
                normalized.content = await this.extractFromDocx(input.buffer);
                break;
            case 'markdown':
                normalized.content = input.buffer.toString('utf-8');
                break;
            // ... other formats
        }

        return normalized;
    }

    private async transformFromIntermediate(
        intermediate: any,
        targetFormat: string
    ): Promise<Buffer> {
        switch (targetFormat) {
            case 'pdf':
                return await this.generatePDF(intermediate);
            case 'docx':
                return await this.generateDocx(intermediate);
            case 'html':
                return await this.generateHTML(intermediate);
            case 'markdown':
                return Buffer.from(intermediate.content);
            default:
                throw new Error(`Unsupported format: ${targetFormat}`);
        }
    }

    private async generateHTML(intermediate: any): Promise<Buffer> {
        const html = `
            <!DOCTYPE html>
            <html>
            <head>
                <meta charset="UTF-8">
                <title>${intermediate.metadata.title}</title>
                <style>
                    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
                    h1 { color: #333; }
                    code { background: #f4f4f4; padding: 2px 6px; }
                </style>
            </head>
            <body>
                <h1>${intermediate.metadata.title}</h1>
                <main>${intermediate.content}</main>
            </body>
            </html>
        `;

        return Buffer.from(html, 'utf-8');
    }
}
```

## Table and Data Extraction

### 1. Intelligent Table Detection

```typescript
class TableExtractor {
    async extractTables(documentBuffer: Buffer): Promise<Table[]> {
        // Detect tables in document
        const tables: Table[] = [];

        if (documentBuffer.toString().includes('table') ||
            documentBuffer.toString().includes('|')) {
            // Try multiple table detection methods
            tables.push(...await this.detectTablesRegex(documentBuffer));
            tables.push(...await this.detectTablesStructure(documentBuffer));

            // Deduplicate similar tables
            return this.deduplicateTables(tables);
        }

        return tables;
    }

    private async detectTablesStructure(buffer: Buffer): Promise<Table[]> {
        const text = buffer.toString('utf-8');
        const tables: Table[] = [];

        // Find lines with consistent column structure
        const lines = text.split('\n');
        let tableStart = -1;
        let headerRow = '';

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];

            // Detect table header (multiple pipes or columns)
            if (line.match(/\|.*\|/) && tableStart === -1) {
                tableStart = i;
                headerRow = line;
            } else if (tableStart !== -1 && !line.match(/\|.*\|/)) {
                // End of table
                const tableLines = lines.slice(tableStart, i);
                const table = this.parseTableLines(tableLines);
                if (table) tables.push(table);
                tableStart = -1;
            }
        }

        return tables;
    }

    private parseTableLines(lines: string[]): Table | null {
        const rows: string[][] = [];

        for (const line of lines) {
            if (line.match(/\|.*\|/)) {
                const cells = line.split('|')
                    .map(c => c.trim())
                    .filter(c => c.length > 0);
                rows.push(cells);
            }
        }

        if (rows.length < 2) return null;

        return {
            headers: rows[0],
            rows: rows.slice(1),
            data: this.rowsToObjects(rows[0], rows.slice(1)),
            metadata: {
                rows: rows.length,
                columns: rows[0].length
            }
        };
    }

    private rowsToObjects(headers: string[], rows: string[][]): Record<string, any>[] {
        return rows.map(row => {
            const obj: Record<string, any> = {};
            headers.forEach((header, i) => {
                obj[header] = row[i] || '';
            });
            return obj;
        });
    }
}
```

---

**Version**: 5.0.0 | **Last Updated**: 2025-11-22 | **Enterprise Ready**: ✓
