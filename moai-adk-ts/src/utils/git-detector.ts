// @CODE:INIT-004 | SPEC: SPEC-INIT-004.md | TEST: src/__tests__/utils/git-detector.test.ts
// Related: @SPEC:INIT-004

/**
 * @file Git detection and initialization utilities
 * @author MoAI Team
 * @tags @CODE:INIT-004:GIT-DETECTOR
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { type SimpleGit, simpleGit } from 'simple-git';

/**
 * Git remote information
 */
export interface GitRemote {
  name: string;
  url: string;
  type: 'fetch' | 'push';
}

/**
 * Git repository status
 */
export interface GitStatus {
  exists: boolean;
  commits: number;
  currentBranch: string;
  remotes: GitRemote[];
  githubUrl?: string;
}

/**
 * Detect Git repository status
 * @param cwd Current working directory
 * @returns Git status information
 */
export async function detectGitStatus(cwd: string): Promise<GitStatus> {
  // Check if .git directory exists
  const gitDir = path.join(cwd, '.git');
  if (!fs.existsSync(gitDir)) {
    return {
      exists: false,
      commits: 0,
      currentBranch: '',
      remotes: [],
    };
  }

  const git: SimpleGit = simpleGit(cwd);

  try {
    // Verify it's a valid Git repository
    const isRepo = await git.checkIsRepo();
    if (!isRepo) {
      return {
        exists: false,
        commits: 0,
        currentBranch: '',
        remotes: [],
      };
    }

    // Get commit count
    let commits = 0;
    try {
      const commitCount = await git.raw(['rev-list', '--count', 'HEAD']);
      commits = Number.parseInt(commitCount.trim(), 10);
    } catch {
      // No commits yet (new repository)
      commits = 0;
    }

    // Get current branch
    const branchInfo = await git.branch();
    const currentBranch = branchInfo.current;

    // Get remotes
    const remotesInfo = await git.getRemotes(true);
    const remotes: GitRemote[] = remotesInfo.flatMap(remote => {
      const result: GitRemote[] = [];
      if (remote.refs.fetch) {
        result.push({
          name: remote.name,
          url: remote.refs.fetch,
          type: 'fetch',
        });
      }
      if (remote.refs.push && remote.refs.push !== remote.refs.fetch) {
        result.push({
          name: remote.name,
          url: remote.refs.push,
          type: 'push',
        });
      }
      return result;
    });

    // Detect GitHub URL
    const githubUrl = detectGitHubRemote(remotes);

    const result: GitStatus = {
      exists: true,
      commits,
      currentBranch,
      remotes,
    };

    if (githubUrl) {
      result.githubUrl = githubUrl;
    }

    return result;
  } catch {
    // Return default status if any error occurs
    return {
      exists: false,
      commits: 0,
      currentBranch: '',
      remotes: [],
    };
  }
}

/**
 * Detect GitHub remote URL from remotes
 * @param remotes List of Git remotes
 * @returns GitHub URL if found, null otherwise
 */
export function detectGitHubRemote(remotes: GitRemote[]): string | null {
  // Prioritize 'origin' remote
  const originRemote = remotes.find(
    r => r.name === 'origin' && isGitHubUrl(r.url)
  );
  if (originRemote) {
    return originRemote.url;
  }

  // Find any GitHub remote
  const githubRemote = remotes.find(r => isGitHubUrl(r.url));
  if (githubRemote) {
    return githubRemote.url;
  }

  return null;
}

/**
 * Check if URL is a GitHub URL
 * @param url Git remote URL
 * @returns True if GitHub URL
 */
function isGitHubUrl(url: string): boolean {
  return (
    url.includes('github.com') &&
    (url.startsWith('https://github.com/') || url.startsWith('git@github.com:'))
  );
}

/**
 * Validate GitHub URL format
 * @param url GitHub URL to validate
 * @returns True if valid GitHub URL
 */
export function validateGitHubUrl(url: string): boolean {
  if (!url || url.trim() === '') {
    return false;
  }

  // HTTPS format: https://github.com/user/repo or https://github.com/user/repo.git
  const httpsPattern =
    /^https:\/\/github\.com\/[a-zA-Z0-9_-]+\/[a-zA-Z0-9_.-]+(?:\.git)?$/;
  if (httpsPattern.test(url)) {
    return true;
  }

  // SSH format: git@github.com:user/repo.git or git@github.com:user/repo
  const sshPattern =
    /^git@github\.com:[a-zA-Z0-9_-]+\/[a-zA-Z0-9_.-]+(?:\.git)?$/;
  if (sshPattern.test(url)) {
    return true;
  }

  return false;
}

/**
 * Initialize Git repository
 * @param cwd Current working directory
 * @throws Error if Git initialization fails
 */
export async function autoInitGit(cwd: string): Promise<void> {
  const git: SimpleGit = simpleGit(cwd);

  try {
    await git.init();
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    throw new Error(`Failed to initialize Git repository: ${message}`);
  }
}
