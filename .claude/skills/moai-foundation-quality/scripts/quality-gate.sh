#!/bin/bash

# Enterprise Code Quality Gate Script
# Comprehensive quality validation with TRUST 5 framework
# Usage: ./quality-gate.sh [options] <project-path>

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_FILE="$PROJECT_ROOT/.moai/logs/quality-gate.log"
CONFIG_FILE="$PROJECT_ROOT/.moai/config/quality-config.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Quality thresholds
QUALITY_THRESHOLD=${QUALITY_THRESHOLD:-0.85}
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-0.80}
SECURITY_THRESHOLD=${SECURITY_THRESHOLD:-0.90}

# Default values
PROJECT_PATH="src/"
OUTPUT_DIR=".moai/reports/quality"
GENERATE_REPORTS=true
CI_MODE=false
VERBOSE=false
FAIL_ON_THRESHOLD=true

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS")
            echo -e "${GREEN}✓${NC} $message"
            ;;
        "FAIL")
            echo -e "${RED}✗${NC} $message"
            ;;
        "WARN")
            echo -e "${YELLOW}⚠${NC} $message"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ${NC} $message"
            ;;
        "HIGHLIGHT")
            echo -e "${PURPLE}★${NC} $message"
            ;;
        "DEBUG")
            [[ "$VERBOSE" == "true" ]] && echo -e "${CYAN}🐛${NC} $message"
            ;;
    esac
}

# Print section header
print_section() {
    local title=$1
    echo
    echo -e "${BLUE}=== $title ===${NC}"
    echo
}

# Print quality gauge
print_quality_gauge() {
    local score=$1
    local threshold=$2
    local label=$3

    # Calculate gauge
    local percentage=$(awk "BEGIN {printf \"%.0f\", $score * 100}")
    local filled=$(awk "BEGIN {printf \"%.0f\", $percentage / 5}")
    local empty=$((20 - filled))

    # Build gauge
    local gauge=""
    for ((i=1; i<=filled; i++)); do
        if (( i <= 16 )); then
            gauge="${gauge}$(echo -e "${GREEN}█${NC}")"
        elif (( i <= 18 )); then
            gauge="${gauge}$(echo -e "${YELLOW}█${NC}")"
        else
            gauge="${gauge}$(echo -e "${RED}█${NC}")"
        fi
    done
    for ((i=1; i<=empty; i++)); do
        gauge="${gauge}░"
    done

    # Determine status
    local status
    if (( $(awk "BEGIN {print ($score >= $threshold)}") )); then
        status="PASS"
    else
        status="FAIL"
    fi

    printf "%-25s %s %3d%% " "$label" "$gauge" "$percentage"
    print_status "$status" ""
}

# Check dependencies
check_dependencies() {
    local deps=("python3" "jq" "find")
    local missing=()

    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing+=("$dep")
        fi
    done

    if [ ${#missing[@]} -gt 0 ]; then
        print_status "FAIL" "Missing dependencies: ${missing[*]}"
        return 1
    fi

    # Check Python packages
    local python_deps=("moai-core-quality")
    for dep in "${python_deps[@]}"; do
        if ! python3 -c "import $dep" 2>/dev/null; then
            missing+=("python3-$dep")
        fi
    done

    if [ ${#missing[@]} -gt 0 ]; then
        print_status "FAIL" "Missing Python packages: ${missing[*]}"
        print_status "INFO" "Install with: pip install ${missing[*]}"
        return 1
    fi

    print_status "PASS" "All dependencies available"
    return 0
}

# Validate project structure
validate_project_structure() {
    local project_path=$1

    if [ ! -d "$project_path" ]; then
        print_status "FAIL" "Project path not found: $project_path"
        return 1
    fi

    # Check for code files
    local code_files
    code_files=$(find "$project_path" -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.go" | wc -l)

    if [ "$code_files" -eq 0 ]; then
        print_status "WARN" "No code files found in $project_path"
        return 1
    fi

    print_status "PASS" "Found $code_files code files"
    return 0
}

# Load configuration
load_configuration() {
    if [ -f "$CONFIG_FILE" ]; then
        print_status "INFO" "Loading configuration from $CONFIG_FILE"

        # Extract configuration values using jq
        local quality_threshold_config
        quality_threshold_config=$(jq -r '.quality_threshold // empty' "$CONFIG_FILE" 2>/dev/null || echo "")

        if [ -n "$quality_threshold_config" ] && [ "$quality_threshold_config" != "null" ]; then
            QUALITY_THRESHOLD=$quality_threshold_config
            print_status "INFO" "Quality threshold set to $QUALITY_THRESHOLD"
        fi

        local coverage_threshold_config
        coverage_threshold_config=$(jq -r '.test_coverage_threshold // empty' "$CONFIG_FILE" 2>/dev/null || echo "")

        if [ -n "$coverage_threshold_config" ] && [ "$coverage_threshold_config" != "null" ]; then
            COVERAGE_THRESHOLD=$coverage_threshold_config
            print_status "INFO" "Coverage threshold set to $COVERAGE_THRESHOLD"
        fi
    else
        print_status "INFO" "Using default configuration (no config file found)"
    fi
}

# Detect project languages
detect_languages() {
    local project_path=$1
    local languages=()

    # Detect Python
    if find "$project_path" -name "*.py" -type f | head -1 | grep -q .; then
        languages+=("python")
    fi

    # Detect JavaScript/TypeScript
    if find "$project_path" -name "*.js" -o -name "*.ts" -o -name "*.jsx" -o -name "*.tsx" | head -1 | grep -q .; then
        languages+=("javascript")
    fi

    # Detect TypeScript specifically
    if find "$project_path" -name "*.ts" -o -name "*.tsx" | head -1 | grep -q .; then
        languages+=("typescript")
    fi

    # Detect Go
    if find "$project_path" -name "*.go" | head -1 | grep -q .; then
        languages+=("go")
    fi

    # Detect Java
    if find "$project_path" -name "*.java" | head -1 | grep -q .; then
        languages+=("java")
    fi

    printf '%s\n' "${languages[@]}"
}

# Run quality analysis
run_quality_analysis() {
    local project_path=$1
    local languages=("${@:2}")

    print_section "Running Quality Analysis"

    # Create Python script for quality analysis
    local analysis_script=$(cat << 'EOF'
import asyncio
import json
import sys
from pathlib import Path
from moai_core_quality import QualityOrchestrator, QualityConfig

async def main():
    project_path = sys.argv[1]
    languages = sys.argv[2].split(',')
    quality_threshold = float(sys.argv[3])
    coverage_threshold = float(sys.argv[4])
    security_threshold = float(sys.argv[5])
    output_dir = sys.argv[6]
    ci_mode = sys.argv[7].lower() == 'true'

    # Configure quality analysis
    config = QualityConfig(
        quality_threshold=quality_threshold,
        test_coverage_threshold=coverage_threshold,
        security_threshold=security_threshold,
        languages=languages,
        generate_reports=True,
        output_directory=output_dir
    )

    # Initialize orchestrator
    orchestrator = QualityOrchestrator(config)

    print(f"Analyzing {project_path} with languages: {languages}")
    print(f"Quality threshold: {quality_threshold}")

    try:
        # Run comprehensive analysis
        result = await orchestrator.analyze_codebase(
            path=project_path,
            languages=languages,
            quality_threshold=quality_threshold
        )

        # Prepare analysis data
        analysis_data = {
            "overall_score": result.overall_score,
            "trust5_validation": {
                principle: validation.score
                for principle, validation in result.trust5_validation.principles.items()
            },
            "metrics": {
                "test_coverage": result.metrics.test_coverage,
                "security_score": result.metrics.security_score,
                "maintainability": result.metrics.maintainability_index,
                "technical_debt": result.metrics.technical_debt_hours,
                "complexity": result.metrics.cyclomatic_complexity
            },
            "passed": result.trust5_validation.passed,
            "recommendations": len(result.proactive_analysis.recommendations),
            "critical_issues": len([
                r for r in result.proactive_analysis.recommendations
                if r.severity == "critical"
            ])
        }

        # Print summary
        print(f"\nQuality Analysis Results:")
        print(f"Overall Score: {result.overall_score:.2f}")
        print(f"Quality Gate: {'PASSED' if result.trust5_validation.passed else 'FAILED'}")
        print(f"Test Coverage: {result.metrics.test_coverage:.1%}")
        print(f"Security Score: {result.metrics.security_score:.2f}")

        # Generate reports if not in CI mode
        if not ci_mode:
            await orchestrator.generate_report(
                result=result,
                output_path=f"{output_dir}/quality-report.html",
                format="html"
            )

            # Save detailed results
            with open(f"{output_dir}/quality-analysis.json", "w") as f:
                json.dump(analysis_data, f, indent=2)

        # Output JSON for CI integration
        print(json.dumps(analysis_data))

        # Exit code based on quality gate
        sys.exit(0 if result.trust5_validation.passed else 1)

    except Exception as e:
        print(f"Error during quality analysis: {e}", file=sys.stderr)
        sys.exit(2)

if __name__ == "__main__":
    asyncio.run(main())
EOF
)

    # Create output directory
    mkdir -p "$OUTPUT_DIR"

    # Run analysis
    local languages_str
    languages_str=$(IFS=,; echo "${languages[*]}")

    local analysis_result
    analysis_result=$(python3 -c "$analysis_script" \
        "$PROJECT_PATH" \
        "$languages_str" \
        "$QUALITY_THRESHOLD" \
        "$COVERAGE_THRESHOLD" \
        "$SECURITY_THRESHOLD" \
        "$OUTPUT_DIR" \
        "$CI_MODE" 2>&1)

    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        # Parse JSON results (last line should be JSON)
        local json_result
        json_result=$(echo "$analysis_result" | tail -1)

        # Display results
        echo "$analysis_result" | head -n -1

        # Store results for CI
        if [ "$CI_MODE" == "true" ]; then
            echo "$json_result" > "$OUTPUT_DIR/quality-results.json"

            # Set GitHub Actions outputs if available
            if command -v jq &> /dev/null; then
                local overall_score
                overall_score=$(echo "$json_result" | jq -r '.overall_score // 0')
                local quality_passed
                quality_passed=$(echo "$json_result" | jq -r '.passed // false')
                local test_coverage
                test_coverage=$(echo "$json_result" | jq -r '.metrics.test_coverage // 0')

                echo "quality_score=$overall_score" >> "$GITHUB_OUTPUT"
                echo "quality_passed=$quality_passed" >> "$GITHUB_OUTPUT"
                echo "test_coverage=$test_coverage" >> "$GITHUB_OUTPUT"
            fi
        fi

        return 0
    else
        print_status "FAIL" "Quality analysis failed"
        echo "$analysis_result"
        return $exit_code
    fi
}

# Display quality summary
display_quality_summary() {
    local output_dir=$1

    if [ ! -f "$output_dir/quality-results.json" ]; then
        print_status "WARN" "No quality results found"
        return 1
    fi

    print_section "Quality Summary"

    local overall_score
    overall_score=$(jq -r '.overall_score // 0' "$output_dir/quality-results.json")

    local quality_passed
    quality_passed=$(jq -r '.passed // false' "$output_dir/quality-results.json")

    # Display gauges
    print_quality_gauge "$overall_score" "$QUALITY_THRESHOLD" "Overall Quality"

    # TRUST 5 principles
    print_section "TRUST 5 Framework"

    local trust5_data
    trust5_data=$(jq -r '.trust5_validation // {}' "$output_dir/quality-results.json")

    for principle in testable readable unified secured trackable; do
        local score
        score=$(echo "$trust5_data" | jq -r ".${principle} // 0")
        print_quality_gauge "$score" "0.80" "$principle"
    done

    # Key metrics
    print_section "Key Metrics"

    local metrics
    metrics=$(jq -r '.metrics // {}' "$output_dir/quality-results.json")

    local test_coverage
    test_coverage=$(echo "$metrics" | jq -r '.test_coverage // 0')
    print_quality_gauge "$test_coverage" "$COVERAGE_THRESHOLD" "Test Coverage"

    local security_score
    security_score=$(echo "$metrics" | jq -r '.security_score // 0')
    print_quality_gauge "$security_score" "$SECURITY_THRESHOLD" "Security Score"

    local technical_debt
    technical_debt=$(echo "$metrics" | jq -r '.technical_debt // 0')
    printf "%-25s %s hours\n" "Technical Debt" "$technical_debt"

    local recommendations
    recommendations=$(jq -r '.recommendations // 0' "$output_dir/quality-results.json")
    printf "%-25s %s\n" "Recommendations" "$recommendations"

    local critical_issues
    critical_issues=$(jq -r '.critical_issues // 0' "$output_dir/quality-results.json")
    printf "%-25s %s\n" "Critical Issues" "$critical_issues"

    # Quality gate status
    echo
    if [ "$quality_passed" == "true" ]; then
        print_status "PASS" "Quality gate passed"
    else
        print_status "FAIL" "Quality gate failed"
        if [ "$FAIL_ON_THRESHOLD" == "true" ]; then
            print_status "INFO" "Use --no-fail to continue despite quality gate failure"
        fi
    fi
}

# Generate final report
generate_final_report() {
    local output_dir=$1

    if [ ! -f "$output_dir/quality-results.json" ]; then
        return 1
    fi

    print_section "Generating Reports"

    # Generate HTML report summary
    local report_html=$(cat << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Quality Gate Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .gauge { margin: 10px 0; }
        .passed { color: #28a745; font-weight: bold; }
        .failed { color: #dc3545; font-weight: bold; }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .metric { background: #f8f9fa; padding: 15px; border-radius: 5px; text-align: center; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Code Quality Gate Report</h1>
        <p>Generated on $(date)</p>
        <p>Project: $PROJECT_PATH</p>
    </div>

    <div class="metrics">
        <div class="metric">
            <h3>Overall Quality</h3>
            <div class="$([ "$(jq -r '.passed // false' "$output_dir/quality-results.json")" == "true" ] && echo "passed" || echo "failed")">
                $(jq -r '.overall_score // 0' "$output_dir/quality-results.json")
            </div>
        </div>

        <div class="metric">
            <h3>Test Coverage</h3>
            <div>$(jq -r '.metrics.test_coverage // 0' "$output_dir/quality-results.json" | awk '{printf "%.1f%%", $1 * 100}')</div>
        </div>

        <div class="metric">
            <h3>Security Score</h3>
            <div>$(jq -r '.metrics.security_score // 0' "$output_dir/quality-results.json")</div>
        </div>

        <div class="metric">
            <h3>Technical Debt</h3>
            <div>$(jq -r '.metrics.technical_debt // 0' "$output_dir/quality-results.json") hours</div>
        </div>
    </div>

    <p><a href="quality-report.html">View Detailed Report</a></p>
</body>
</html>
EOF
)

    echo "$report_html" > "$output_DIR/quality-summary.html"
    print_status "PASS" "Quality summary report generated: $OUTPUT_DIR/quality-summary.html"

    # Print report location
    print_status "INFO" "Detailed report available at: $OUTPUT_DIR/quality-report.html"
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --threshold)
                QUALITY_THRESHOLD="$2"
                shift 2
                ;;
            --coverage-threshold)
                COVERAGE_THRESHOLD="$2"
                shift 2
                ;;
            --security-threshold)
                SECURITY_THRESHOLD="$2"
                shift 2
                ;;
            --output-dir)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            --config)
                CONFIG_FILE="$2"
                shift 2
                ;;
            --ci)
                CI_MODE=true
                shift
                ;;
            --no-fail)
                FAIL_ON_THRESHOLD=false
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --help|-h)
                cat << EOF
Usage: $0 [options] [project-path]

Options:
  --threshold FLOAT          Quality threshold (default: 0.85)
  --coverage-threshold FLOAT Test coverage threshold (default: 0.80)
  --security-threshold FLOAT Security score threshold (default: 0.90)
  --output-dir DIR          Output directory for reports (default: .moai/reports/quality)
  --config FILE            Configuration file path
  --ci                     Enable CI mode (minimal output, JSON results)
  --no-fail                Don't fail on quality gate violation
  --verbose                Enable verbose logging
  --help, -h               Show this help message

Examples:
  $0                          # Analyze src/ with default settings
  $0 --threshold 0.90 src/    # Use higher quality threshold
  $0 --ci --no-fail src/      # CI mode without failing
  $0 --verbose --output-dir ./reports src/  # Verbose with custom output

Exit codes:
  0  - Quality gate passed
  1  - Quality gate failed
  2  - Analysis error occurred
EOF
                exit 0
                ;;
            -*)
                print_status "FAIL" "Unknown option: $1"
                echo "Use --help for usage information"
                exit 2
                ;;
            *)
                PROJECT_PATH="$1"
                shift
                ;;
        esac
    done
}

# Main execution
main() {
    local exit_code=0

    # Parse arguments
    parse_arguments "$@"

    print_section "Enterprise Code Quality Gate"

    log "Starting quality gate analysis for $PROJECT_PATH"

    # Check dependencies
    if ! check_dependencies; then
        exit 2
    fi

    # Load configuration
    load_configuration

    # Validate project structure
    if ! validate_project_structure "$PROJECT_PATH"; then
        exit 2
    fi

    # Detect languages
    print_status "INFO" "Detecting project languages..."
    local languages
    readarray -t languages < <(detect_languages "$PROJECT_PATH")

    if [ ${#languages[@]} -eq 0 ]; then
        print_status "FAIL" "No supported programming languages detected"
        exit 2
    fi

    print_status "INFO" "Detected languages: ${languages[*]}"

    # Run quality analysis
    if ! run_quality_analysis "$PROJECT_PATH" "${languages[@]}"; then
        exit_code=$?
        if [ $exit_code -eq 1 ]; then
            print_status "FAIL" "Quality gate failed"
        else
            print_status "FAIL" "Quality analysis encountered an error"
        fi
    fi

    # Display summary
    if [ "$CI_MODE" != "true" ]; then
        display_quality_summary "$OUTPUT_DIR"
        generate_final_report "$OUTPUT_DIR"
    fi

    log "Quality gate analysis completed with exit code: $exit_code"

    # Final exit decision
    if [ "$FAIL_ON_THRESHOLD" == "false" ]; then
        print_status "INFO" "Ignoring quality gate failure due to --no-fail flag"
        exit 0
    fi

    exit $exit_code
}

# Execute main function
main "$@"