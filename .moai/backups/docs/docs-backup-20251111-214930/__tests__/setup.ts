/**
 * Jest 환경 설정 파일
 */

// 모의 설정
global.console = {
  ...console,
  log: jest.fn(),
  error: jest.fn(),
  warn: jest.fn(),
  info: jest.fn(),
  debug: jest.fn(),
};

// 모의 DOM 환경
Object.defineProperty(global, 'window', {
  value: {
    location: {
      href: 'http://localhost:3000',
    },
  },
  writable: true,
});

Object.defineProperty(global, 'document', {
  value: {
    createElement: jest.fn(),
    getElementById: jest.fn(),
    querySelector: jest.fn(),
  },
  writable: true,
});

// 경고 메시지 숨기기
jest.spyOn(console, 'warn').mockImplementation(() => {});