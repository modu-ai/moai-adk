#!/usr/bin/env node
/**
 * @FEATURE:TAG-REPORT-001 Report generation logic for validate_tags module (TypeScript version)
 * @tags @FEATURE:TAG-REPORT-001 @QUALITY:TAG-001 @DATA:TAGS-001
 * Modern stack: Bun compatible, TypeScript 5.9.2+, 16-Core TAG system integration
 */

export interface TagReference {
    tag: string;
    file: string;
    line: number;
    context: string;
}

export interface TagHealthReport {
    total_tags: number;
    valid_tags: number;
    invalid_tags: number;
    orphan_tags: number;
    broken_links: number;
    quality_score: number;
    issues: string[];
    recommendations: string[];
}

export function calculateQualityScore(
    total_tags: number,
    valid_tags: number,
    orphan_count: number,
    broken_count: number
): number {
    if (total_tags === 0) return 100;

    const validity_ratio = valid_tags / total_tags;
    const orphan_penalty = Math.min(orphan_count / total_tags, 0.3);
    const broken_penalty = Math.min(broken_count / total_tags, 0.3);

    const score = Math.max(0, (validity_ratio - orphan_penalty - broken_penalty) * 100);
    return Math.round(score * 100) / 100;
}

export function generateHealthReport(
    all_tags: TagReference[],
    tag_index: Record<string, TagReference[]>,
    format_violations: string[],
    orphan_tags: string[],
    broken_links: Array<[string, string]>,
    chain_violations: string[]
): TagHealthReport {
    const total_tags = all_tags.length;
    const invalid_tags = format_violations.length;
    const valid_tags = total_tags - invalid_tags;

    const quality_score = calculateQualityScore(
        total_tags,
        valid_tags,
        orphan_tags.length,
        broken_links.length
    );

    const issues: string[] = [];
    const recommendations: string[] = [];

    // Add format violations
    issues.push(...format_violations);

    // Add orphan tag issues
    if (orphan_tags.length > 0) {
        issues.push(`Found ${orphan_tags.length} orphan tags`);
        recommendations.push('Link orphan tags to parent requirements');
    }

    // Add broken link issues
    if (broken_links.length > 0) {
        issues.push(`Found ${broken_links.length} broken links`);
        recommendations.push('Fix broken tag references');
    }

    // Add chain violations
    issues.push(...chain_violations);

    return {
        total_tags,
        valid_tags,
        invalid_tags,
        orphan_tags: orphan_tags.length,
        broken_links: broken_links.length,
        quality_score,
        issues,
        recommendations
    };
}

// CLI entry point for standalone execution
async function main(): Promise<void> {
    // Example usage for template
    const sampleTags: TagReference[] = [
        {
            tag: '@REQ:SAMPLE-001',
            file: 'sample.md',
            line: 10,
            context: 'Sample requirement'
        }
    ];

    const report = generateHealthReport(
        sampleTags,
        { '@REQ:SAMPLE-001': sampleTags },
        [],
        [],
        [],
        []
    );

    console.log('Sample tag health report generated');
    console.log(JSON.stringify(report, null, 2));
}

// Execute if run directly
if (require.main === module) {
    main().catch((error) => {
        console.error('Tag reporter failed:', error);
        process.exit(1);
    });
}