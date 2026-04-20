"use client";

import { Github, Star } from "lucide-react";
import * as React from "react";

import { cn } from "@/lib/utils";

// Re-export the existing LanguageSelector component
export { default as LanguageSelector } from "./LanguageSelector";

// GitHub repository configuration
const GITHUB_REPO = "modu-ai/moai-adk";
const GITHUB_URL = `https://github.com/${GITHUB_REPO}`;

/**
 * GitHubStarBadge component for displaying repository star count
 *
 * Fetches star count from GitHub API and displays it with the GitHub icon.
 * Falls back to static display if API fetch fails.
 *
 * @example
 * ```tsx
 * import { GitHubStarBadge } from '@/components/navbar'
 *
 * <GitHubStarBadge />
 * ```
 */
export function GitHubStarBadge({ className }: { className?: string }) {
  const [starCount, setStarCount] = React.useState<number | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState(false);

  React.useEffect(() => {
    // Fetch star count from GitHub API
    const fetchStarCount = async () => {
      try {
        const response = await fetch(
          `https://api.github.com/repos/${GITHUB_REPO}`,
          {
            headers: {
              Accept: "application/vnd.github.v3+json",
            },
            // Cache for 5 minutes
            next: { revalidate: 300 },
          },
        );

        if (!response.ok) {
          throw new Error("Failed to fetch star count");
        }

        const data = (await response.json()) as { stargazers_count: number };
        setStarCount(data.stargazers_count);
        setError(false);
      } catch {
        setError(true);
      } finally {
        setIsLoading(false);
      }
    };

    fetchStarCount();
  }, []);

  // Format star count (e.g., 1234 -> 1.2k)
  const formatStarCount = (count: number): string => {
    if (count >= 1000) {
      return `${(count / 1000).toFixed(1)}k`;
    }
    return count.toString();
  };

  return (
    <a
      href={GITHUB_URL}
      target="_blank"
      rel="noopener noreferrer"
      aria-label="Star us on GitHub"
      className={cn(
        "x:inline-flex x:items-center x:gap-1.5 x:shrink-0 x:whitespace-nowrap",
        "x:rounded-md x:p-2 x:text-sm",
        "x:transition-colors",
        "x:text-gray-600 x:hover:text-black x:dark:text-gray-400 x:dark:hover:text-gray-200",
        className,
      )}
    >
      <Github className="x:size-5" aria-hidden="true" />
      <Star className="x:size-3.5 x:fill-current" aria-hidden="true" />
      {isLoading ? (
        <span className="x:text-xs">...</span>
      ) : error ? null : (
        <span className="x:text-xs x:font-medium">
          {formatStarCount(starCount ?? 0)}
        </span>
      )}
    </a>
  );
}

/**
 * GitHubStarButton component for a compact "Star" button
 *
 * Smaller variant focused on the star action without count display.
 * Useful for tight spaces or mobile views.
 */
export function GitHubStarButton({ className }: { className?: string }) {
  return (
    <a
      href={`${GITHUB_URL}`}
      target="_blank"
      rel="noopener noreferrer"
      aria-label="Star us on GitHub"
      className={cn(
        "inline-flex items-center justify-center",
        "rounded-md p-2",
        "transition-colors",
        "hover:bg-accent hover:text-accent-foreground",
        "focus-visible:outline-none focus-visible:ring-2",
        "focus-visible:ring-ring focus-visible:ring-offset-2",
        className,
      )}
    >
      <Star className="h-4 w-4" aria-hidden="true" />
      <span className="sr-only">Star us on GitHub</span>
    </a>
  );
}
