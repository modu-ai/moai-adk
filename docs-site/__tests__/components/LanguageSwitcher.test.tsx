// @TEST:NEXTRA-I18N-007
/**
 * LanguageSwitcher Component Tests
 *
 * Tests for language switcher UI component:
 * - Component rendering
 * - Locale selection
 * - Navigation behavior
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import LanguageSwitcher from '@/components/LanguageSwitcher';

// Mock functions
const mockPush = jest.fn();
const mockUseLocale = jest.fn();
const mockUsePathname = jest.fn();
const mockTranslation = jest.fn((key: string) => key);

// Mock next-intl hooks
jest.mock('next-intl', () => ({
  useLocale: () => mockUseLocale(),
  useTranslations: () => mockTranslation,
}));

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
  usePathname: () => mockUsePathname(),
}));

describe('LanguageSwitcher Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseLocale.mockReturnValue('ko');
    mockUsePathname.mockReturnValue('/ko/docs');
  });

  describe('component rendering', () => {
    it('should render without crashing', () => {
      const { container } = render(<LanguageSwitcher />);
      expect(container).toBeTruthy();
    });

    it('should render a select element', () => {
      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox');
      expect(select).toBeInTheDocument();
    });

    it('should have options for Korean and English', () => {
      render(<LanguageSwitcher />);
      const koOption = screen.getByRole('option', { name: '한국어' });
      const enOption = screen.getByRole('option', { name: 'English' });

      expect(koOption).toBeInTheDocument();
      expect(enOption).toBeInTheDocument();
      expect((koOption as HTMLOptionElement).value).toBe('ko');
      expect((enOption as HTMLOptionElement).value).toBe('en');
    });

    it('should display current locale as selected', () => {
      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox') as HTMLSelectElement;
      expect(select.value).toBe('ko');
    });

    it('should show English as selected when locale is en', () => {
      mockUseLocale.mockReturnValue('en');
      mockUsePathname.mockReturnValue('/en/docs');

      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox') as HTMLSelectElement;
      expect(select.value).toBe('en');
    });
  });

  describe('locale switching', () => {
    it('should call router.push when locale changes', () => {
      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox');

      fireEvent.change(select, { target: { value: 'en' } });

      expect(mockPush).toHaveBeenCalledTimes(1);
      expect(mockPush).toHaveBeenCalledWith('/en/docs');
    });

    it('should preserve current path when switching locale', () => {
      mockUsePathname.mockReturnValue('/ko/docs/guide');

      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox');

      fireEvent.change(select, { target: { value: 'en' } });

      expect(mockPush).toHaveBeenCalledWith('/en/docs/guide');
    });

    it('should handle root path correctly', () => {
      mockUsePathname.mockReturnValue('/ko');

      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox');

      fireEvent.change(select, { target: { value: 'en' } });

      expect(mockPush).toHaveBeenCalledWith('/en/');
    });

    it('should switch from English to Korean', () => {
      mockUseLocale.mockReturnValue('en');
      mockUsePathname.mockReturnValue('/en/docs');

      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox');

      fireEvent.change(select, { target: { value: 'ko' } });

      expect(mockPush).toHaveBeenCalledWith('/ko/docs');
    });
  });

  describe('accessibility', () => {
    it('should have aria-label for select element', () => {
      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox');

      expect(select).toHaveAttribute('aria-label', 'language');
    });

    it('should use translation for aria-label', () => {
      mockTranslation.mockReturnValue('Language Selection');

      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox');

      expect(select).toHaveAttribute('aria-label', 'Language Selection');
    });
  });

  describe('styling', () => {
    it('should have appropriate CSS classes', () => {
      render(<LanguageSwitcher />);
      const select = screen.getByRole('combobox');

      expect(select).toHaveClass('border', 'rounded', 'px-2', 'py-1');
    });
  });
});
