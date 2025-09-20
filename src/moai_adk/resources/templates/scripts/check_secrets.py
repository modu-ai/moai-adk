#!/usr/bin/env python3
"""
MoAI-ADK Secrets Scanner
소스 코드에서 시크릿 정보 탐지 및 보안 검사
"""
import sys
import re
import argparse
from pathlib import Path
from typing import Dict, List, Tuple
import json


class SecretsScanner:
    def __init__(self):
        # 시크릿 패턴 정의
        self.secret_patterns = {
            'api_key': [
                r'api[_-]?key\s*[=:]\s*["\']([A-Za-z0-9_\-]{16,})["\']',
                r'apikey\s*[=:]\s*["\']([A-Za-z0-9_\-]{16,})["\']',
                r'API[_-]?KEY\s*[=:]\s*["\']([A-Za-z0-9_\-]{16,})["\']',
            ],
            'password': [
                r'password\s*[=:]\s*["\']([^"\']{6,})["\']',
                r'passwd\s*[=:]\s*["\']([^"\']{6,})["\']',
                r'pwd\s*[=:]\s*["\']([^"\']{6,})["\']',
            ],
            'token': [
                r'token\s*[=:]\s*["\']([A-Za-z0-9_\-\.]{20,})["\']',
                r'access[_-]?token\s*[=:]\s*["\']([A-Za-z0-9_\-\.]{20,})["\']',
                r'auth[_-]?token\s*[=:]\s*["\']([A-Za-z0-9_\-\.]{20,})["\']',
            ],
            'jwt': [
                r'jwt\s*[=:]\s*["\']([A-Za-z0-9_\-]{20,}\.[A-Za-z0-9_\-]{20,}\.[A-Za-z0-9_\-]{20,})["\']',
                r'Bearer\s+([A-Za-z0-9_\-]{20,}\.[A-Za-z0-9_\-]{20,}\.[A-Za-z0-9_\-]{20,})',
            ],
            'database_url': [
                r'DATABASE_URL\s*[=:]\s*["\']([^"\']*://[^"\']*)["\']',
                r'db[_-]?url\s*[=:]\s*["\']([^"\']*://[^"\']*)["\']',
                r'connection[_-]?string\s*[=:]\s*["\']([^"\']*://[^"\']*)["\']',
            ],
            'private_key': [
                r'-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----',
                r'private[_-]?key\s*[=:]\s*["\']([A-Za-z0-9+/=]{50,})["\']',
            ],
            'secret_key': [
                r'secret[_-]?key\s*[=:]\s*["\']([A-Za-z0-9_\-]{16,})["\']',
                r'SECRET[_-]?KEY\s*[=:]\s*["\']([A-Za-z0-9_\-]{16,})["\']',
            ],
            'aws_credentials': [
                r'AKIA[0-9A-Z]{16}',  # AWS Access Key ID
                r'aws[_-]?access[_-]?key\s*[=:]\s*["\']([A-Za-z0-9]{20})["\']',
                r'aws[_-]?secret[_-]?key\s*[=:]\s*["\']([A-Za-z0-9/+=]{40})["\']',
            ],
            'github_token': [
                r'gh[ps]_[A-Za-z0-9_]{36}',  # GitHub Personal Access Token
                r'github[_-]?token\s*[=:]\s*["\']([A-Za-z0-9_]{40})["\']',
            ],
            'email': [
                r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}',
            ]
        }
        
        # 허용된 테스트/예시 값들
        self.allowed_values = {
            'test@example.com',
            'user@example.com', 
            'admin@example.com',
            'test_password',
            'example_key',
            'your_api_key_here',
            'placeholder_token',
            'dummy_secret',
            'test_value'
        }
        
        # 스캔할 파일 확장자
        self.scan_extensions = {'.py', '.js', '.ts', '.jsx', '.tsx', '.json', '.yaml', '.yml', 
                               '.env', '.config', '.cfg', '.ini', '.md', '.txt'}
        
        # 제외할 디렉토리
        self.exclude_dirs = {'node_modules', '.git', '__pycache__', '.pytest_cache', 
                           'venv', 'env', '.venv', 'build', 'dist'}

    def scan_file(self, file_path: Path) -> List[Dict[str, any]]:
        """단일 파일에서 시크릿 스캔"""
        secrets = []
        
        try:
            content = file_path.read_text(encoding='utf-8', errors='ignore')
            lines = content.split('\n')
            
            for line_num, line in enumerate(lines, 1):
                for secret_type, patterns in self.secret_patterns.items():
                    for pattern in patterns:
                        matches = re.finditer(pattern, line, re.IGNORECASE)
                        for match in matches:
                            matched_value = match.group(1) if match.groups() else match.group(0)
                            
                            # 허용된 값들은 제외
                            if matched_value.lower() in {v.lower() for v in self.allowed_values}:
                                continue
                            
                            # 주석이나 문서인 경우 severity 낮춤
                            severity = self.determine_severity(line, secret_type, matched_value)
                            
                            secrets.append({
                                'type': secret_type,
                                'value': matched_value[:20] + '...' if len(matched_value) > 20 else matched_value,
                                'line': line_num,
                                'line_content': line.strip(),
                                'severity': severity,
                                'file': str(file_path)
                            })
        
        except (UnicodeDecodeError, PermissionError, OSError):
            pass
            
        return secrets

    def determine_severity(self, line: str, secret_type: str, value: str) -> str:
        """시크릿의 심각도 결정"""
        line_lower = line.lower().strip()
        
        # 주석이나 문서는 LOW
        if line_lower.startswith('#') or line_lower.startswith('//') or line_lower.startswith('*'):
            return 'LOW'
        
        # TODO, FIXME, 예시는 LOW
        if any(keyword in line_lower for keyword in ['todo', 'fixme', 'example', '예시', '샘플']):
            return 'LOW'
        
        # 테스트 파일은 MEDIUM
        if 'test' in str(line).lower():
            return 'MEDIUM'
        
        # 실제 시크릿 타입별 심각도
        high_risk_types = {'private_key', 'database_url', 'aws_credentials'}
        if secret_type in high_risk_types:
            return 'HIGH'
        
        # 값 길이 기반 심각도
        if len(value) > 50:
            return 'HIGH'
        elif len(value) > 20:
            return 'MEDIUM'
        else:
            return 'LOW'

    def scan_directory(self, directory: Path) -> Dict[str, List[Dict[str, any]]]:
        """디렉토리 전체 스캔"""
        all_secrets = {}
        
        for file_path in directory.rglob('*'):
            # 제외 디렉토리 스킵
            if any(exclude_dir in file_path.parts for exclude_dir in self.exclude_dirs):
                continue
                
            # 파일 확장자 확인
            if file_path.is_file() and file_path.suffix in self.scan_extensions:
                secrets = self.scan_file(file_path)
                if secrets:
                    all_secrets[str(file_path)] = secrets
        
        return all_secrets

    def generate_report(self, secrets: Dict[str, List[Dict[str, any]]]) -> Dict[str, any]:
        """보고서 생성"""
        total_secrets = sum(len(file_secrets) for file_secrets in secrets.values())
        
        severity_counts = {'HIGH': 0, 'MEDIUM': 0, 'LOW': 0}
        type_counts = {}
        
        for file_secrets in secrets.values():
            for secret in file_secrets:
                severity_counts[secret['severity']] += 1
                type_counts[secret['type']] = type_counts.get(secret['type'], 0) + 1
        
        report = {
            'summary': {
                'total_files_scanned': len(secrets),
                'total_secrets_found': total_secrets,
                'high_severity': severity_counts['HIGH'],
                'medium_severity': severity_counts['MEDIUM'],
                'low_severity': severity_counts['LOW']
            },
            'by_type': type_counts,
            'by_severity': severity_counts,
            'secrets': secrets
        }
        
        return report


def print_report(report: Dict[str, any], verbose: bool = False):
    """보고서 출력"""
    summary = report['summary']
    
    print(f"\n🔍 MoAI-ADK 시크릿 스캔 보고서")
    print(f"{'='*50}")
    print(f"스캔된 파일: {summary['total_files_scanned']}개")
    print(f"발견된 시크릿: {summary['total_secrets_found']}개")
    print(f"  🔴 HIGH:   {summary['high_severity']}개")
    print(f"  🟡 MEDIUM: {summary['medium_severity']}개")
    print(f"  🟢 LOW:    {summary['low_severity']}개")
    
    # 위험도 평가
    risk_score = summary['high_severity'] * 3 + summary['medium_severity'] * 1
    if risk_score == 0:
        print("\n✅ 위험한 시크릿이 발견되지 않았습니다")
        risk_level = "SAFE"
    elif risk_score <= 2:
        print("\n⚠️ 낮은 위험도 - 확인 권장")
        risk_level = "LOW"
    elif risk_score <= 5:
        print("\n🚨 보통 위험도 - 즉시 검토 필요")
        risk_level = "MEDIUM"
    else:
        print("\n🔥 높은 위험도 - 긴급 조치 필요")
        risk_level = "HIGH"
    
    # 타입별 분포
    if report['by_type']:
        print(f"\n📊 시크릿 타입별 분포:")
        for secret_type, count in sorted(report['by_type'].items(), key=lambda x: x[1], reverse=True):
            print(f"  {secret_type}: {count}개")
    
    # 상세 내용 (HIGH와 MEDIUM만 또는 verbose 모드)
    if verbose or summary['high_severity'] > 0 or summary['medium_severity'] > 0:
        print(f"\n📋 발견된 시크릿:")
        
        for file_path, file_secrets in report['secrets'].items():
            # HIGH, MEDIUM만 표시 (verbose가 아닌 경우)
            filtered_secrets = file_secrets if verbose else [
                s for s in file_secrets if s['severity'] in ['HIGH', 'MEDIUM']
            ]
            
            if filtered_secrets:
                print(f"\n📄 {file_path}:")
                for secret in filtered_secrets:
                    severity_icon = {'HIGH': '🔴', 'MEDIUM': '🟡', 'LOW': '🟢'}[secret['severity']]
                    print(f"  {severity_icon} Line {secret['line']}: {secret['type']}")
                    print(f"     Value: {secret['value']}")
                    if verbose:
                        print(f"     Context: {secret['line_content']}")
    
    return risk_level


def main():
    parser = argparse.ArgumentParser(description='MoAI-ADK Secrets Scanner')
    parser.add_argument('--directory', '-d', type=Path, default=Path.cwd(),
                       help='스캔할 디렉토리 (기본값: 현재 디렉토리)')
    parser.add_argument('--verbose', '-v', action='store_true',
                       help='상세한 출력 (LOW 레벨 시크릿도 표시)')
    parser.add_argument('--output', '-o', type=Path,
                       help='JSON 보고서 출력 파일')
    parser.add_argument('--fail-on', choices=['HIGH', 'MEDIUM', 'LOW'], default='HIGH',
                       help='실패 처리할 최소 심각도 레벨')
    
    args = parser.parse_args()
    
    print(f"🔍 시크릿 스캔 시작: {args.directory}")
    
    scanner = SecretsScanner()
    secrets = scanner.scan_directory(args.directory)
    report = scanner.generate_report(secrets)
    
    risk_level = print_report(report, args.verbose)
    
    # JSON 보고서 저장
    if args.output:
        args.output.write_text(json.dumps(report, indent=2, ensure_ascii=False))
        print(f"\n📄 보고서가 {args.output}에 저장되었습니다")
    
    # Exit code 결정
    fail_levels = {'HIGH': ['HIGH'], 'MEDIUM': ['HIGH', 'MEDIUM'], 'LOW': ['HIGH', 'MEDIUM', 'LOW']}
    should_fail = risk_level in fail_levels[args.fail_on]
    
    if should_fail:
        print(f"\n❌ {args.fail_on} 레벨 이상의 시크릿이 발견되어 실패 처리합니다")
        sys.exit(1)
    else:
        print(f"\n✅ {args.fail_on} 레벨 이상의 위험한 시크릿이 없습니다")
        sys.exit(0)


if __name__ == '__main__':
    main()