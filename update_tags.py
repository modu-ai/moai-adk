#!/usr/bin/env python3
"""
@TASK:TAG-SYNC-001 프로젝트 전체 @TAG 스캔 및 SQLite DB 갱신
"""

import os
import re
import sqlite3
import sys
from pathlib import Path
from datetime import datetime


def initialize_db(db_path):
    """SQLite 태그 데이터베이스 초기화"""
    conn = sqlite3.connect(db_path)

    # 테이블 생성
    conn.executescript("""
        CREATE TABLE IF NOT EXISTS tags (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            category TEXT NOT NULL,
            identifier TEXT NOT NULL,
            description TEXT,
            file_path TEXT NOT NULL,
            line_number INTEGER,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(category, identifier)
        );

        CREATE TABLE IF NOT EXISTS tag_references (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            source_tag TEXT NOT NULL,
            target_tag TEXT NOT NULL,
            reference_type TEXT DEFAULT 'chain',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(source_tag, target_tag)
        );

        CREATE TABLE IF NOT EXISTS statistics (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            stat_name TEXT UNIQUE NOT NULL,
            stat_value TEXT NOT NULL,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    """)

    return conn


def scan_tags(project_root):
    """프로젝트에서 @TAG 패턴 스캔"""
    tag_pattern = re.compile(r'@([A-Z]+):([A-Z-]+[0-9]*)', re.MULTILINE)
    tags = []

    # 스캔할 파일 확장자
    extensions = {'.py', '.md', '.toml', '.json', '.txt', '.yml', '.yaml', '.sh'}

    # 제외할 디렉토리
    exclude_dirs = {
        '.git', '__pycache__', '.pytest_cache', 'node_modules',
        '.env', 'venv', 'test-*-env', 'build', 'dist', '.ruff_cache'
    }

    for root, dirs, files in os.walk(project_root):
        # 제외 디렉토리 스킵
        dirs[:] = [d for d in dirs if not any(exclude in d for exclude in exclude_dirs)]

        for file in files:
            if any(file.endswith(ext) for ext in extensions):
                file_path = os.path.join(root, file)
                rel_path = os.path.relpath(file_path, project_root)

                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        content = f.read()

                    for line_num, line in enumerate(content.split('\n'), 1):
                        matches = tag_pattern.findall(line)
                        for category, identifier in matches:
                            tags.append({
                                'category': category,
                                'identifier': identifier,
                                'full_tag': f'@{category}:{identifier}',
                                'file_path': rel_path,
                                'line_number': line_num,
                                'line_content': line.strip()
                            })
                except Exception as e:
                    print(f"Warning: Could not read {rel_path}: {e}")

    return tags


def detect_tag_chains(tags):
    """태그 체인 관계 감지"""
    chains = []

    # Primary chain 패턴: REQ → DESIGN → TASK → TEST
    primary_chain = ['REQ', 'DESIGN', 'TASK', 'TEST']

    # 태그를 식별자별로 그룹화
    tag_groups = {}
    for tag in tags:
        base_id = re.sub(r'-\d+$', '', tag['identifier'])  # 숫자 접미사 제거
        if base_id not in tag_groups:
            tag_groups[base_id] = []
        tag_groups[base_id].append(tag)

    # 체인 관계 생성
    for base_id, group_tags in tag_groups.items():
        categories_present = {tag['category'] for tag in group_tags}

        # Primary chain 검사
        for i in range(len(primary_chain) - 1):
            source_cat = primary_chain[i]
            target_cat = primary_chain[i + 1]

            if source_cat in categories_present and target_cat in categories_present:
                source_tag = f"@{source_cat}:{base_id}"
                target_tag = f"@{target_cat}:{base_id}"
                chains.append({
                    'source_tag': source_tag,
                    'target_tag': target_tag,
                    'reference_type': 'primary_chain'
                })

    return chains


def update_database(db_path, tags, chains):
    """데이터베이스 갱신"""
    conn = initialize_db(db_path)

    try:
        # 기존 데이터 삭제
        conn.execute("DELETE FROM tags")
        conn.execute("DELETE FROM tag_references")

        # 태그 삽입
        unique_tags = {}
        for tag in tags:
            key = (tag['category'], tag['identifier'])
            if key not in unique_tags:
                unique_tags[key] = tag

        for tag in unique_tags.values():
            conn.execute("""
                INSERT OR REPLACE INTO tags (category, identifier, description, file_path, line_number)
                VALUES (?, ?, ?, ?, ?)
            """, (
                tag['category'],
                tag['identifier'],
                tag['line_content'][:200],  # 설명으로 라인 내용 사용 (200자 제한)
                tag['file_path'],
                tag['line_number']
            ))

        # 체인 관계 삽입
        for chain in chains:
            conn.execute("""
                INSERT OR REPLACE INTO tag_references (source_tag, target_tag, reference_type)
                VALUES (?, ?, ?)
            """, (chain['source_tag'], chain['target_tag'], chain['reference_type']))

        # 통계 업데이트
        stats = [
            ('total_tags', str(len(unique_tags))),
            ('total_chains', str(len(chains))),
            ('categories', str(len(set(tag['category'] for tag in unique_tags.values())))),
            ('last_updated', datetime.now().isoformat())
        ]

        for stat_name, stat_value in stats:
            conn.execute("""
                INSERT OR REPLACE INTO statistics (stat_name, stat_value, updated_at)
                VALUES (?, ?, CURRENT_TIMESTAMP)
            """, (stat_name, stat_value))

        conn.commit()

        # 결과 출력
        print(f"✅ SQLite TAG 데이터베이스 갱신 완료!")
        print(f"   - 총 태그: {len(unique_tags):,}개")
        print(f"   - 태그 체인: {len(chains):,}개")
        print(f"   - 카테고리: {len(set(tag['category'] for tag in unique_tags.values()))}개")
        print(f"   - 데이터베이스: {db_path}")

        return True

    except Exception as e:
        conn.rollback()
        print(f"❌ 데이터베이스 갱신 실패: {e}")
        return False
    finally:
        conn.close()


def main():
    project_root = Path(__file__).parent
    db_path = project_root / '.moai' / 'indexes' / 'tags.db'

    print("🔍 프로젝트 @TAG 스캔 시작...")
    tags = scan_tags(str(project_root))

    print("🔗 TAG 체인 관계 분석 중...")
    chains = detect_tag_chains(tags)

    print("💾 SQLite 데이터베이스 갱신 중...")
    success = update_database(str(db_path), tags, chains)

    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())