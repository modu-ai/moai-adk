"use client";

import type { $NextraMetadata, Heading } from "nextra";
import { useMDXComponents } from "nextra-theme-docs";
import type { ReactNode } from "react";

interface MDXWrapperProps {
  toc: Heading[];
  metadata: $NextraMetadata;
  sourceCode: string;
  children: ReactNode;
}

export function MDXWrapper({
  toc,
  metadata,
  sourceCode,
  children,
}: MDXWrapperProps) {
  const { wrapper: Wrapper } = useMDXComponents();
  return (
    <Wrapper toc={toc} metadata={metadata} sourceCode={sourceCode}>
      {children}
    </Wrapper>
  );
}
