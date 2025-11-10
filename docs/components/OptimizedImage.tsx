import React from 'react';
import Image from 'next/image';
import type { ImageProps } from 'next/image';

/**
 * Optimized Image Component for MoAI-ADK Documentation
 *
 * This component provides optimized image loading with proper fallbacks,
 * lazy loading, and Core Web Vitals optimizations.
 */

interface OptimizedImageProps extends Omit<ImageProps, 'blurDataURL'> {
  /**
   * Enable blur placeholder for better perceived performance
   */
  enableBlur?: boolean;

  /**
   * Priority loading for above-the-fold images
   */
  priority?: boolean;

  /**
   * Custom blur data URL (generated automatically if not provided)
   */
  blurDataURL?: string;

  /**
   * Critical image loading strategy
   */
  critical?: boolean;
}

const OptimizedImage: React.FC<OptimizedImageProps> = ({
  enableBlur = true,
  priority = false,
  blurDataURL,
  critical = false,
  alt,
  src,
  width,
  height,
  className = '',
  ...props
}) => {
  // Generate blur data URL for small images
  const generateBlurDataURL = (src: string): string => {
    // Simple gradient placeholder for demonstration
    // In production, you'd want to generate actual blurred versions
    return 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAiIGhlaWdodD0iMTAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CiAgPHJlY3Qgd2lkdGg9IjEwIiBoZWlnaHQ9IjEwIiBmaWxsPSIjZjBmMGYwIi8+CiAgPGNpcmNsZSBjeD0iNSIgY3k9IjUiIHI9IjIiIGZpbGw9IiNlNWU3ZWIiLz4KPC9zdmc+';
  };

  const imageProps: ImageProps = {
    src,
    width,
    height,
    alt: alt || '', // Always provide alt text for accessibility
    className: [
      'transition-opacity duration-300',
      critical ? 'opacity-100' : 'opacity-0 group-hover:opacity-100',
      className
    ].filter(Boolean).join(' '),
    priority: priority || critical,
    sizes: props.sizes || '(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw',
    quality: props.quality || 85,
    placeholder: enableBlur ? 'blur' : 'empty',
    blurDataURL: blurDataURL || (enableBlur ? generateBlurDataURL(src as string) : undefined),
    ...props,
  };

  // Add loading optimization
  if (!critical && !priority) {
    imageProps.loading = 'lazy';
  }

  return (
    <div className="relative overflow-hidden group">
      <Image {...imageProps} />
      {enableBlur && (
        <div className="absolute inset-0 bg-gradient-to-br from-gray-100 to-gray-200 opacity-0 group-hover:opacity-10 transition-opacity duration-300 pointer-events-none" />
      )}
    </div>
  );
};

export default OptimizedImage;