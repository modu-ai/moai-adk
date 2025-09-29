import { defineConfig } from 'tsup';
import { existsSync } from 'fs';

// hooks 빌드는 templates/.claude/hooks/moai/ 에 존재하는 test_hook.ts 파일 사용
// 실제 TypeScript 소스 파일이 있는 템플릿 경로 확인
const hasHookSources = existsSync('./templates/.claude/hooks/moai/test_hook.ts');

export default defineConfig({
  entry: hasHookSources ? {
    'hooks/moai/test_hook': 'templates/.claude/hooks/moai/test_hook.ts',
  } : {},
  format: ['cjs'],
  target: 'node18',
  outDir: 'dist',
  clean: false, // main build already cleans
  sourcemap: false, // hooks don't need sourcemaps
  minify: false, // keep readable for debugging
  splitting: false, // each hook should be standalone
  bundle: true, // bundle dependencies for standalone execution
  external: [], // don't externalize anything for hooks
  platform: 'node',
  shims: true, // add import.meta.url and __dirname shims
});