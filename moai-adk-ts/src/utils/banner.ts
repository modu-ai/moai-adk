// @CODE:UTIL-001 |
// Related: @CODE:BANNER-001:UI

/**
 * @file MoAI-ADK Banner display utility
 * @author MoAI Team
 */

import chalk from 'chalk';
import { getCurrentVersion } from './version';
import { logger } from './/winston-logger.js';

/**
 * Check if terminal supports color output
 * @returns True if colors are supported
 * @tags UTIL:COLOR-SUPPORT-001
 */
function supportsColor(): boolean {
  return (
    process.stdout.isTTY &&
    process.env.TERM !== 'dumb' &&
    process.env.NO_COLOR === undefined
  );
}

/**
 * Apply zinc monochrome color to ASCII art line
 * @param line - Line to colorize
 * @returns Colorized line
 * @tags UTIL:COLORIZE-001
 */
function applyZincColor(line: string): string {
  if (!supportsColor() || !line.trim()) {
    return line;
  }

  // Zinc-400 monochrome color - #a1a1aa (RGB: 161, 161, 170)
  return chalk.rgb(161, 161, 170)(line);
}

/**
 * Get MoAI logo ASCII art
 * @returns Array of logo lines
 * @tags UTIL:LOGO-ASCII-001
 */
function getMoaiLogo(): string[] {
  return [
    '███╗   ███╗          █████╗ ██╗      █████╗ ██████╗ ██╗  ██╗',
    '████╗ ████║ ██████╗ ██╔══██╗██║     ██╔══██╗██╔══██╗██║ ██╔╝',
    '██╔████╔██║██╔═══██ ███████║██║ ███ ███████║██║  ██║█████╔╝ ',
    '██║╚██╔╝██║██║   ██║██╔══██║██║ ╚══ ██╔══██║██║  ██║██╔═██╗ ',
    '██║ ╚═╝ ██║╚██████╔╝██║  ██║██║     ██║  ██║██████╔╝██║  ██╗',
    '╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝',
  ];
}

/**
 * Create the complete MoAI banner
 * @param options Banner options
 * @returns Complete banner string
 * @tags @CODE:CREATE-BANNER-001:API
 */
export function createBanner(
  options: { version?: string; showUsage?: boolean } = {}
): string {
  const { version = getCurrentVersion(), showUsage = false } = options;

  const moaiLines = getMoaiLogo();
  const bannerLines: string[] = [];

  // Empty line at top
  bannerLines.push('');

  // MoAI-ADK logo with zinc monochrome color
  for (const line of moaiLines) {
    const coloredLine = applyZincColor(line);
    bannerLines.push(coloredLine);
  }

  // Bottom border matching logo width (60 chars for MoAI-ADK)
  const border = supportsColor()
    ? applyZincColor('═'.repeat(60))
    : '═'.repeat(60);
  bannerLines.push(border);
  bannerLines.push('');

  // Description with version in one line
  const description = `🗿 MoAI-ADK: Modu-AI's Agentic Development kit (v${version}) 🚀`;

  bannerLines.push(supportsColor() ? applyZincColor(description) : description);
  bannerLines.push('');

  // Usage info if requested
  if (showUsage) {
    bannerLines.push('');
    bannerLines.push('Usage: moai [options] [command]');
    bannerLines.push('');
    bannerLines.push('Commands:');
    bannerLines.push(
      '  doctor [options]                 Run system diagnostics'
    );
    bannerLines.push(
      '  init [options] [project]         Initialize a new MoAI-ADK project'
    );
    bannerLines.push(
      '  restore [options] <backup-path>  Restore MoAI-ADK from a backup directory'
    );
    bannerLines.push(
      '  status [options]                 Show MoAI-ADK project status'
    );
    bannerLines.push(
      '  help [command]                   Show help for MoAI-ADK commands'
    );
  }

  // Footer
  bannerLines.push('');
  const footer = 'copyleft 2024, Modu-AI / 모두의AI (https://mo.ai.kr)';
  bannerLines.push(supportsColor() ? chalk.gray(footer) : footer);
  bannerLines.push('');

  return bannerLines.join('\n');
}

/**
 * Print the MoAI-ADK banner to stdout
 * @param options Banner options
 * @tags @CODE:PRINT-BANNER-001:API
 */
export function printBanner(
  options: { version?: string; showUsage?: boolean } = {}
): void {
  logger.info(createBanner(options));
}

/**
 * Create a simple header for commands
 * @param title Command title
 * @param subtitle Optional subtitle
 * @returns Formatted header
 * @tags @CODE:CREATE-HEADER-001:API
 */
export function createHeader(title: string, subtitle?: string): string {
  const lines: string[] = [];

  const formattedTitle = supportsColor()
    ? chalk.blue.bold(`🚀 ${title}`)
    : `🚀 ${title}`;

  lines.push(formattedTitle);

  if (subtitle) {
    const formattedSubtitle = supportsColor() ? chalk.blue(subtitle) : subtitle;
    lines.push(formattedSubtitle);
  }

  lines.push('');

  return lines.join('\n');
}
