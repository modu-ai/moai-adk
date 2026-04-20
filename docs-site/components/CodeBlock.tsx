"use client";

import { Check, Copy } from "lucide-react";
import { useState } from "react";

interface CodeBlockProps {
  children: React.ReactNode;
  className?: string;
  "data-language"?: string;
  "data-theme"?: string;
}

export function CodeBlock({
  children,
  className,
  "data-language": language,
}: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  // Extract code content from various structures
  const extractCodeContent = (node: React.ReactNode): string => {
    if (typeof node === "string") return node;
    if (Array.isArray(node)) return node.map(extractCodeContent).join("");
    if (typeof node === "object" && node !== null) {
      if ("props" in node) {
        const props = node.props as { children?: React.ReactNode };
        if (props.children) {
          return extractCodeContent(props.children);
        }
      }
    }
    return "";
  };

  const codeContent = extractCodeContent(children);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(codeContent);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative group my-6 rounded-lg overflow-hidden border border-zinc-700 bg-[#1a1a1a]">
      {/* macOS Window Header */}
      <div className="flex items-center justify-between px-4 py-2.5 bg-[#252526] border-b border-zinc-700">
        <div className="flex items-center gap-2">
          {/* macOS Window Controls */}
          <div className="flex gap-2">
            <div className="w-3 h-3 rounded-full bg-[#ff5f57] hover:bg-[#ff413f] transition-colors cursor-pointer" />
            <div className="w-3 h-3 rounded-full bg-[#febc2e] hover:bg-[#ffaa00] transition-colors cursor-pointer" />
            <div className="w-3 h-3 rounded-full bg-[#28c840] hover:bg-[#1faa2e] transition-colors cursor-pointer" />
          </div>
          {/* Language Badge */}
          {language && (
            <span className="ml-3 text-xs text-zinc-400 font-mono uppercase tracking-wide">
              {language}
            </span>
          )}
        </div>

        {/* Copy Button */}
        <button
          type="button"
          onClick={handleCopy}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium text-zinc-400 hover:text-zinc-100 hover:bg-zinc-700 transition-all opacity-0 group-hover:opacity-100"
          aria-label="Copy code"
        >
          {copied ? (
            <>
              <Check className="w-3.5 h-3.5" />
              Copied!
            </>
          ) : (
            <>
              <Copy className="w-3.5 h-3.5" />
              Copy
            </>
          )}
        </button>
      </div>

      {/* Code Content - VS Code Dark Style */}
      <div className="overflow-x-auto">
        <pre className={className}>
          <code>{children}</code>
        </pre>
      </div>
    </div>
  );
}
