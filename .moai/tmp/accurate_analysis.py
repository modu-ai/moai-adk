#!/usr/bin/env python3
"""더 정확한 명령형/선언형 분석"""
import re
from pathlib import Path

def analyze_imperative_vs_declarative(content: str) -> tuple[int, int, float]:
    """명령형 vs 선언형 문장 개수 및 비율"""
    
    # 명령형 패턴 (동사로 시작하는 명령문)
    imperative_patterns = [
        r'^\s*-\s+(Follow|Execute|Perform|Run|Do|Take|Complete|Check|Verify|Ensure|Call|Invoke|Pass|Create|Update|Delete|Add|Remove|Load|Save|Print|Display|Show|Ask|Present|Collect|Merge|Archive|Terminate|Delegate|Orchestrate|Coordinate)',
        r'^\s*\*\*(Step|Phase|Action|Next|Then|First|Second|Third):\*\*',
        r'^\s*(IF|WHEN|WHILE|FOR EACH|UNLESS)',
        r'^\*\*.*\*\*:\s*[A-Z]',  # Bold labels followed by instructions
    ]
    
    # 선언형 패턴
    declarative_patterns = [
        r'^\s*-\s+(You are|Your role|Your responsibility|You should|You must|You will|You need|The agent)',
        r'^\s*You (are|should|must|will|need to)',
        r'^\s*The .* (is|should|must|will) responsible for',
    ]
    
    imperative_count = 0
    declarative_count = 0
    
    for line in content.split('\n'):
        line = line.strip()
        if not line or line.startswith('#') or line.startswith('```'):
            continue
            
        for pattern in imperative_patterns:
            if re.match(pattern, line, re.IGNORECASE):
                imperative_count += 1
                break
        
        for pattern in declarative_patterns:
            if re.match(pattern, line, re.IGNORECASE):
                declarative_count += 1
                break
    
    total = imperative_count + declarative_count
    if total == 0:
        return 0, 0, 0.5
    
    ratio = imperative_count / total
    return imperative_count, declarative_count, ratio

def main():
    base_dir = Path('/Users/goos/MoAI/MoAI-ADK/.claude')
    
    agent_files = sorted((base_dir / 'agents/alfred').glob('*.md'))
    command_files = sorted((base_dir / 'commands/alfred').glob('*.md'))
    
    print("=" * 100)
    print("📊 정확한 명령형/선언형 분석")
    print("=" * 100)
    print()
    
    needs_refactoring = []
    already_good = []
    
    for file_path in agent_files + command_files:
        content = file_path.read_text(encoding='utf-8')
        
        # YAML frontmatter 제외
        if content.startswith('---'):
            yaml_end = content.find('---', 3)
            if yaml_end != -1:
                content = content[yaml_end+3:]
        
        imp_count, decl_count, ratio = analyze_imperative_vs_declarative(content)
        
        file_info = {
            'name': file_path.name,
            'imperative': imp_count,
            'declarative': decl_count,
            'ratio': ratio
        }
        
        # 선언형 문장이 5개 이상이면 리팩토링 필요
        if decl_count >= 5 or ratio < 0.6:
            needs_refactoring.append(file_info)
        else:
            already_good.append(file_info)
    
    if needs_refactoring:
        print("🔴 리팩토링 필요 (선언형 문장 5개 이상 또는 명령형 비율 < 60%)")
        print("=" * 100)
        for info in sorted(needs_refactoring, key=lambda x: -x['declarative']):
            print(f"  {info['name']:40} | 명령형: {info['imperative']:3} | 선언형: {info['declarative']:3} | 비율: {info['ratio']:5.1%}")
        print()
    
    if already_good:
        print("✅ 이미 양호 (선언형 < 5개 AND 명령형 비율 >= 60%)")
        print("=" * 100)
        for info in sorted(already_good, key=lambda x: -x['ratio']):
            print(f"  {info['name']:40} | 명령형: {info['imperative']:3} | 선언형: {info['declarative']:3} | 비율: {info['ratio']:5.1%}")
        print()
    
    print("=" * 100)
    print("📈 요약")
    print("=" * 100)
    print(f"  🔴 리팩토링 필요: {len(needs_refactoring)}개")
    print(f"  ✅ 이미 양호: {len(already_good)}개")
    print(f"  📊 전체: {len(needs_refactoring) + len(already_good)}개")
    print()

if __name__ == '__main__':
    main()
