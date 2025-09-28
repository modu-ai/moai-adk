#!/usr/bin/env node
/**
 * @FEATURE:LICENSE-REPORT-001 License reporting for check_licenses module (TypeScript version)
 * @tags @FEATURE:LICENSE-REPORT-001 @QUALITY:SEC-001
 * Modern stack: Bun compatible, TypeScript 5.9.2+, ESM modules
 */

import { writeFileSync, mkdirSync } from 'node:fs';
import { resolve, join } from 'node:path';

export interface PackageLicense {
    package: string;
    version: string;
    license: string;
    status: string;
    license_info?: {
        risk_level: 'low' | 'medium' | 'high' | 'critical';
    };
}

export interface LicenseReport {
    timestamp: string;
    status: 'PASS' | 'FAIL';
    summary: {
        total_packages: number;
        violations: number;
        warnings: number;
        license_counts: Record<string, number>;
        risk_distribution: Record<string, number>;
    };
    violations: string[];
    warnings: string[];
    packages: Array<{
        name: string;
        version: string;
        license: string;
        status: string;
        risk_level: string;
    }>;
}

export function generateLicenseReport(
    packages: PackageLicense[],
    violations: string[],
    warnings: string[]
): LicenseReport {
    const status = violations.length === 0 ? 'PASS' : 'FAIL';

    // Count by license type
    const license_counts: Record<string, number> = {};
    const risk_counts = { low: 0, medium: 0, high: 0, critical: 0 };

    for (const pkg of packages) {
        license_counts[pkg.license] = (license_counts[pkg.license] || 0) + 1;
        if (pkg.license_info) {
            const risk = pkg.license_info.risk_level;
            risk_counts[risk] = (risk_counts[risk] || 0) + 1;
        }
    }

    const report: LicenseReport = {
        timestamp: new Date().toISOString(),
        status,
        summary: {
            total_packages: packages.length,
            violations: violations.length,
            warnings: warnings.length,
            license_counts,
            risk_distribution: risk_counts
        },
        violations,
        warnings,
        packages: packages.map(pkg => ({
            name: pkg.package,
            version: pkg.version,
            license: pkg.license,
            status: pkg.status,
            risk_level: pkg.license_info?.risk_level || 'unknown'
        }))
    };

    return report;
}

export function saveLicenseReport(report: LicenseReport, project_root: string): void {
    try {
        const reports_dir = join(project_root, '.moai', 'reports');
        mkdirSync(reports_dir, { recursive: true });

        const report_file = join(reports_dir, 'license_report.json');
        writeFileSync(report_file, JSON.stringify(report, null, 2));
        console.log(`\n📄 Report saved: ${report_file}`);

    } catch (error) {
        console.error(`⚠️  Failed to save report: ${error}`);
    }
}

// CLI entry point for standalone execution
async function main(): Promise<void> {
    // Example usage for template
    const samplePackages: PackageLicense[] = [
        {
            package: 'example-package',
            version: '1.0.0',
            license: 'MIT',
            status: 'approved',
            license_info: { risk_level: 'low' }
        }
    ];

    const report = generateLicenseReport(samplePackages, [], []);
    console.log('Sample license report generated');
    console.log(JSON.stringify(report, null, 2));
}

// Execute if run directly
if (require.main === module) {
    main().catch((error) => {
        console.error('License reporter failed:', error);
        process.exit(1);
    });
}