/**
 * @file Tests for help command implementation
 * @author MoAI Team
 * @tags @TEST:CLI-HELP-001 @SPEC:CLI-FOUNDATION-012
 */

import { beforeEach, describe, expect, it, vi } from 'vitest';
import { HelpCommand } from '../help';

describe('HelpCommand', () => {
  let helpCommand: HelpCommand;

  beforeEach(() => {
    helpCommand = new HelpCommand();
  });

  describe('Command Information Management', () => {
    it('should provide help for all CLI commands', () => {
      const availableCommands = helpCommand.getAvailableCommands();

      expect(availableCommands).toContain('init');
      expect(availableCommands).toContain('doctor');
      expect(availableCommands).toContain('restore');
      expect(availableCommands).toContain('status');
      expect(availableCommands).toContain('help');
    });

    it('should return detailed help for init command', () => {
      const initHelp = helpCommand.getCommandHelp('init');

      expect(initHelp).toBeDefined();
      expect(initHelp?.name).toBe('init');
      expect(initHelp?.description).toContain('Initialize');
      expect(initHelp?.options.length).toBeGreaterThan(0);
      expect(initHelp?.examples.length).toBeGreaterThan(0);
    });

    it('should return detailed help for status command', () => {
      const statusHelp = helpCommand.getCommandHelp('status');

      expect(statusHelp).toBeDefined();
      expect(statusHelp?.name).toBe('status');
      expect(statusHelp?.description).toContain('status');
      expect(statusHelp?.usage).toContain('moai status');
    });

    it('should return undefined for unknown command', () => {
      const unknownHelp = helpCommand.getCommandHelp('unknown');
      expect(unknownHelp).toBeUndefined();
    });
  });

  describe('Help Text Formatting', () => {
    it('should format general help text correctly', () => {
      const generalHelp = helpCommand.formatGeneralHelp();

      expect(generalHelp).toContain('MoAI-ADK');
      expect(generalHelp).toContain('Usage:');
      expect(generalHelp).toContain('Available commands:');
      expect(generalHelp).toContain('init');
      expect(generalHelp).toContain('status');
      expect(generalHelp).toContain('Examples:');
    });

    it('should format command-specific help text correctly', () => {
      const initHelp = helpCommand.getCommandHelp('init');
      expect(initHelp).toBeDefined();

      if (initHelp) {
        const formattedHelp = helpCommand.formatCommandHelp(initHelp);

        expect(formattedHelp).toContain('Command: init');
        expect(formattedHelp).toContain('Description:');
        expect(formattedHelp).toContain('Usage:');
        expect(formattedHelp).toContain('Options:');
        expect(formattedHelp).toContain('Examples:');
      }
    });
  });

  describe('Help Command Execution', () => {
    it('should run general help successfully', async () => {
      const result = await helpCommand.run({});

      expect(result.success).toBe(true);
      expect(result.helpText).toContain('MoAI-ADK');
      expect(result.command).toBeUndefined();
      expect(result.helpText).toContain('Usage: moai <command>');
    });

    it('should run command-specific help successfully', async () => {
      const result = await helpCommand.run({ command: 'init' });

      expect(result.success).toBe(true);
      expect(result.command).toBe('init');
      expect(result.helpText).toContain('Command: init');
      expect(result.helpText).toContain('Initialize a new MoAI-ADK project');
    });

    it('should handle unknown command gracefully', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      const result = await helpCommand.run({ command: 'unknown' });

      expect(result.success).toBe(false);
      expect(result.command).toBe('unknown');
      expect(result.helpText).toContain('Unknown command: unknown');

      consoleSpy.mockRestore();
    });
  });

  describe('Interface Validation', () => {
    it('should accept valid HelpOptions', () => {
      const validOptions = { command: 'init' };
      expect(validOptions).toEqual(
        expect.objectContaining({
          command: expect.any(String),
        })
      );
    });

    it('should define CommandHelp interface correctly', () => {
      const mockCommandHelp = {
        name: 'test',
        description: 'Test command',
        usage: 'moai test',
        options: [{ flag: '--test', description: 'Test option' }],
        examples: ['moai test'],
      };

      expect(mockCommandHelp).toEqual(
        expect.objectContaining({
          name: expect.any(String),
          description: expect.any(String),
          usage: expect.any(String),
          options: expect.any(Array),
          examples: expect.any(Array),
        })
      );
    });

    it('should define HelpResult interface correctly', () => {
      const mockResult = {
        success: true,
        command: 'init',
        helpText: 'Help text',
      };

      expect(mockResult).toEqual(
        expect.objectContaining({
          success: expect.any(Boolean),
          helpText: expect.any(String),
        })
      );
    });
  });
});
