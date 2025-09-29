/**
 * SPEC 메타데이터 검증 시스템
 * @file spec-validator.ts
 * @description SPEC 메타데이터의 유효성 검사, 의존성 검증, 상태 전환 규칙 적용
 *
 * @REQ:SPEC-VALIDATOR-001 SPEC 메타데이터 검증 요구사항
 * @DESIGN:VALIDATION-RULES-001 검증 규칙 설계
 * @TASK:METADATA-VALIDATION-001 메타데이터 검증 구현
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { execa } from 'execa';
import * as yaml from 'yaml';
import {
  type SpecDependencyNode,
  type SpecMetadata,
  type SpecMetadataValidation,
  SpecPriority,
  SpecStatus,
  STATUS_TRANSITIONS,
} from '../types/spec-metadata.js';

/**
 * SPEC 메타데이터 검증기 클래스
 */
export class SpecValidator {
  private specDirectory: string;
  private specCache: Map<string, SpecMetadata>;

  constructor(specDirectory: string = '.moai/specs') {
    this.specDirectory = specDirectory;
    this.specCache = new Map();
  }

  /**
   * SPEC 파일에서 메타데이터 추출
   * @param specFilePath SPEC 파일 경로
   * @returns 파싱된 메타데이터 또는 null
   */
  private extractMetadata(specFilePath: string): SpecMetadata | null {
    try {
      const content = fs.readFileSync(specFilePath, 'utf-8');
      const frontmatterMatch = content.match(/^---\n([\s\S]*?)\n---/);

      if (!frontmatterMatch) {
        return null;
      }

      const metadata = yaml.parse(frontmatterMatch[1]) as SpecMetadata;
      return metadata;
    } catch (error) {
      console.error(`Error parsing metadata from ${specFilePath}:`, error);
      return null;
    }
  }

  /**
   * 모든 SPEC 메타데이터 로드
   * @returns SPEC ID와 메타데이터 맵
   */
  private loadAllSpecMetadata(): Map<string, SpecMetadata> {
    const specs = new Map<string, SpecMetadata>();

    if (!fs.existsSync(this.specDirectory)) {
      return specs;
    }

    const specDirs = fs
      .readdirSync(this.specDirectory)
      .filter(dir => dir.startsWith('SPEC-'))
      .sort();

    for (const specDir of specDirs) {
      const specFile = path.join(this.specDirectory, specDir, 'spec.md');
      if (fs.existsSync(specFile)) {
        const metadata = this.extractMetadata(specFile);
        if (metadata) {
          specs.set(metadata.spec_id, metadata);
        }
      }
    }

    this.specCache = specs;
    return specs;
  }

  /**
   * SPEC ID 형식 검증
   * @param specId SPEC ID
   * @returns 유효성 여부
   */
  private validateSpecId(specId: string): boolean {
    const specIdPattern = /^SPEC-\d{3}$/;
    return specIdPattern.test(specId);
  }

  /**
   * 상태 값 검증
   * @param status 상태 값
   * @returns 유효성 여부
   */
  private validateStatus(status: string): boolean {
    return Object.values(SpecStatus).includes(status as SpecStatus);
  }

  /**
   * 우선순위 값 검증
   * @param priority 우선순위 값
   * @returns 유효성 여부
   */
  private validatePriority(priority: string): boolean {
    return Object.values(SpecPriority).includes(priority as SpecPriority);
  }

  /**
   * 상태 전환 유효성 검증
   * @param currentStatus 현재 상태
   * @param newStatus 새로운 상태
   * @returns 전환 가능 여부
   */
  validateStatusTransition(
    currentStatus: SpecStatus,
    newStatus: SpecStatus
  ): boolean {
    if (currentStatus === newStatus) {
      return true; // 같은 상태는 항상 허용
    }

    const allowedTransitions = STATUS_TRANSITIONS[currentStatus];
    return allowedTransitions.includes(newStatus);
  }

  /**
   * 의존성 검증
   * @param dependencies 의존성 목록
   * @param allSpecs 모든 SPEC 메타데이터
   * @returns 검증 결과
   */
  private validateDependencies(
    dependencies: string[] | undefined,
    allSpecs: Map<string, SpecMetadata>
  ): { valid: boolean; missing: string[] } {
    if (!dependencies || dependencies.length === 0) {
      return { valid: true, missing: [] };
    }

    const missing: string[] = [];
    for (const dep of dependencies) {
      if (!allSpecs.has(dep)) {
        missing.push(dep);
      }
    }

    return {
      valid: missing.length === 0,
      missing,
    };
  }

  /**
   * 순환 의존성 검증
   * @param specId 검증할 SPEC ID
   * @param allSpecs 모든 SPEC 메타데이터
   * @returns 순환 의존성 존재 여부 (true = 순환 없음)
   */
  private validateNoCyclicDependencies(
    specId: string,
    allSpecs: Map<string, SpecMetadata>
  ): boolean {
    const visited = new Set<string>();
    const recursionStack = new Set<string>();

    const hasCycle = (currentId: string): boolean => {
      if (recursionStack.has(currentId)) {
        return true; // 순환 의존성 발견
      }
      if (visited.has(currentId)) {
        return false; // 이미 방문한 노드
      }

      visited.add(currentId);
      recursionStack.add(currentId);

      const spec = allSpecs.get(currentId);
      if (spec?.dependencies) {
        for (const dep of spec.dependencies) {
          if (hasCycle(dep)) {
            return true;
          }
        }
      }

      recursionStack.delete(currentId);
      return false;
    };

    return !hasCycle(specId);
  }

  /**
   * 단일 SPEC 메타데이터 검증
   * @param metadata 검증할 메타데이터
   * @param allSpecs 모든 SPEC 메타데이터 (의존성 검증용)
   * @returns 검증 결과
   */
  validateMetadata(
    metadata: SpecMetadata,
    allSpecs?: Map<string, SpecMetadata>
  ): SpecMetadataValidation {
    const errors: string[] = [];
    const warnings: string[] = [];

    // 기본 필드 검증
    if (!this.validateSpecId(metadata.spec_id)) {
      errors.push(
        `Invalid spec_id format: ${metadata.spec_id}. Expected format: SPEC-XXX`
      );
    }

    if (!this.validateStatus(metadata.status)) {
      errors.push(
        `Invalid status: ${metadata.status}. Must be one of: ${Object.values(SpecStatus).join(', ')}`
      );
    }

    if (!this.validatePriority(metadata.priority)) {
      errors.push(
        `Invalid priority: ${metadata.priority}. Must be one of: ${Object.values(SpecPriority).join(', ')}`
      );
    }

    // 의존성 검증 (allSpecs가 제공된 경우)
    if (allSpecs) {
      const depValidation = this.validateDependencies(
        metadata.dependencies,
        allSpecs
      );
      if (!depValidation.valid) {
        errors.push(
          `Missing dependencies: ${depValidation.missing.join(', ')}`
        );
      }

      // 순환 의존성 검증
      if (!this.validateNoCyclicDependencies(metadata.spec_id, allSpecs)) {
        errors.push(`Circular dependency detected for ${metadata.spec_id}`);
      }
    }

    // 경고 검사
    if (!metadata.tags || metadata.tags.length === 0) {
      warnings.push(
        'No tags specified. Consider adding tags for better organization.'
      );
    }

    if (
      metadata.status === SpecStatus.ACTIVE &&
      (!metadata.dependencies || metadata.dependencies.length === 0)
    ) {
      warnings.push(
        'Active SPEC without dependencies. Verify if this is intentional.'
      );
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings,
    };
  }

  /**
   * 모든 SPEC 메타데이터 검증
   * @returns 검증 결과 맵
   */
  validateAllSpecs(): Map<string, SpecMetadataValidation> {
    const allSpecs = this.loadAllSpecMetadata();
    const results = new Map<string, SpecMetadataValidation>();

    for (const [specId, metadata] of allSpecs) {
      const validation = this.validateMetadata(metadata, allSpecs);
      results.set(specId, validation);
    }

    return results;
  }

  /**
   * 의존성 그래프 생성
   * @returns 의존성 그래프 노드 맵
   */
  buildDependencyGraph(): Map<string, SpecDependencyNode> {
    const allSpecs = this.loadAllSpecMetadata();
    const graph = new Map<string, SpecDependencyNode>();

    // 모든 SPEC에 대해 노드 초기화
    for (const [specId, metadata] of allSpecs) {
      graph.set(specId, {
        spec_id: specId,
        metadata,
        dependencies: [],
        dependents: [],
      });
    }

    // 의존성 관계 구축
    for (const [specId, metadata] of allSpecs) {
      const node = graph.get(specId)!;

      if (metadata.dependencies) {
        for (const depId of metadata.dependencies) {
          const depSpec = allSpecs.get(depId);
          if (depSpec) {
            node.dependencies.push({
              spec_id: depId,
              status: depSpec.status,
              satisfied: depSpec.status === SpecStatus.COMPLETED,
            });

            // 역방향 관계도 설정
            const depNode = graph.get(depId);
            if (depNode) {
              depNode.dependents.push(specId);
            }
          }
        }
      }
    }

    return graph;
  }

  /**
   * 구현 가능한 SPEC 목록 반환
   * @returns 의존성이 모두 만족된 SPEC 목록
   */
  getReadyToImplementSpecs(): SpecMetadata[] {
    const graph = this.buildDependencyGraph();
    const readySpecs: SpecMetadata[] = [];

    for (const [_specId, node] of graph) {
      if (node.metadata.status === SpecStatus.DRAFT) {
        const allDependenciesSatisfied = node.dependencies.every(
          dep => dep.satisfied
        );
        if (allDependenciesSatisfied) {
          readySpecs.push(node.metadata);
        }
      }
    }

    // 우선순위별 정렬
    readySpecs.sort((a, b) => {
      const priorityOrder = { high: 3, medium: 2, low: 1 };
      return priorityOrder[b.priority] - priorityOrder[a.priority];
    });

    return readySpecs;
  }

  /**
   * ripgrep을 사용한 SPEC 검색
   * @param pattern 검색 패턴
   * @param options ripgrep 옵션
   * @returns 검색 결과
   */
  async searchSpecs(
    pattern: string,
    options: string[] = []
  ): Promise<string[]> {
    try {
      const { stdout } = await execa('rg', [
        pattern,
        this.specDirectory,
        '--type',
        'md',
        ...options,
      ]);
      return stdout.split('\n').filter(line => line.trim());
    } catch (_error) {
      // ripgrep이 설치되지 않은 경우 fallback
      console.warn('ripgrep not found, falling back to filesystem search');
      return [];
    }
  }

  /**
   * 특정 상태의 SPEC 목록 조회 (ripgrep 기반)
   * @param status 검색할 상태
   * @returns SPEC ID 목록
   */
  async getSpecsByStatus(status: SpecStatus): Promise<string[]> {
    try {
      const { stdout } = await execa('rg', [
        `status: ${status}`,
        this.specDirectory,
        '--type',
        'md',
        '-l', // 파일명만 출력
      ]);

      const files = stdout.split('\n').filter(line => line.trim());
      const specIds: string[] = [];

      for (const file of files) {
        const { stdout: specIdOutput } = await execa('rg', [
          'spec_id: (SPEC-\\d{3})',
          file,
          '--only-matching',
          '--replace',
          '$1',
        ]);
        if (specIdOutput.trim()) {
          specIds.push(specIdOutput.trim());
        }
      }

      return specIds;
    } catch (_error) {
      console.warn('ripgrep search failed, falling back to filesystem scan');
      return this.getSpecsByStatusFallback(status);
    }
  }

  /**
   * ripgrep 실패 시 fallback 메서드
   * @param status 검색할 상태
   * @returns SPEC ID 목록
   */
  private getSpecsByStatusFallback(status: SpecStatus): string[] {
    const allSpecs = this.loadAllSpecMetadata();
    return Array.from(allSpecs.values())
      .filter(spec => spec.status === status)
      .map(spec => spec.spec_id);
  }

  /**
   * 태그별 SPEC 검색 (ripgrep 기반)
   * @param tag 검색할 태그
   * @returns SPEC ID 목록
   */
  async getSpecsByTag(tag: string): Promise<string[]> {
    try {
      const { stdout } = await execa('rg', [
        `^\\s*-\\s+${tag}`,
        this.specDirectory,
        '--type',
        'md',
        '-l',
      ]);

      const files = stdout.split('\n').filter(line => line.trim());
      const specIds: string[] = [];

      for (const file of files) {
        const { stdout: specIdOutput } = await execa('rg', [
          'spec_id: (SPEC-\\d{3})',
          file,
          '--only-matching',
          '--replace',
          '$1',
        ]);
        if (specIdOutput.trim()) {
          specIds.push(specIdOutput.trim());
        }
      }

      return specIds;
    } catch (_error) {
      console.warn('ripgrep tag search failed');
      return [];
    }
  }

  /**
   * SPEC 상태 업데이트
   * @param specId SPEC ID
   * @param newStatus 새로운 상태
   * @returns 업데이트 성공 여부
   */
  updateSpecStatus(specId: string, newStatus: SpecStatus): boolean {
    const specFile = path.join(this.specDirectory, specId, 'spec.md');

    if (!fs.existsSync(specFile)) {
      console.error(`SPEC file not found: ${specFile}`);
      return false;
    }

    try {
      const content = fs.readFileSync(specFile, 'utf-8');
      const frontmatterMatch = content.match(/^(---\n)([\s\S]*?)(\n---)/);

      if (!frontmatterMatch) {
        console.error(`No frontmatter found in ${specFile}`);
        return false;
      }

      const metadata = yaml.parse(frontmatterMatch[2]) as SpecMetadata;

      // 상태 전환 검증
      if (!this.validateStatusTransition(metadata.status, newStatus)) {
        console.error(
          `Invalid status transition from ${metadata.status} to ${newStatus}`
        );
        return false;
      }

      // 메타데이터 업데이트
      metadata.status = newStatus;
      metadata.updated = new Date().toISOString().split('T')[0];

      // 파일 업데이트
      const newFrontmatter = yaml.stringify(metadata);
      const newContent = content.replace(
        frontmatterMatch[0],
        `---\n${newFrontmatter}---`
      );

      fs.writeFileSync(specFile, newContent, 'utf-8');

      // 캐시 업데이트
      this.specCache.set(specId, metadata);

      return true;
    } catch (error) {
      console.error(`Error updating SPEC status:`, error);
      return false;
    }
  }
}
