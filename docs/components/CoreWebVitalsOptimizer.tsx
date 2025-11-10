'use client';

import React, { useEffect } from 'react';
import { onCLS, onFID, onFCP, onLCP, onTTFB } from 'web-vitals';

/**
 * Core Web Vitals Optimizer Component
 *
 * This component monitors and reports Core Web Vitals metrics
 * to help identify and fix performance issues in production.
 */

interface CoreWebVitalsMetrics {
  cls: number;
  fid: number;
  fcp: number;
  lcp: number;
  ttfb: number;
}

interface WebVitalsOptimizerProps {
  /**
   * Callback function to receive metrics
   */
  onMetrics?: (metrics: CoreWebVitalsMetrics) => void;

  /**
   * Enable debug logging in development
   */
  debug?: boolean;

  /**
   * Sample rate for data collection (0-1)
   */
  sampleRate?: number;

  /**
   * Custom endpoint for sending metrics
   */
  endpoint?: string;
}

const CoreWebVitalsOptimizer: React.FC<WebVitalsOptimizerProps> = ({
  onMetrics,
  debug = process.env.NODE_ENV === 'development',
  sampleRate = 1,
  endpoint
}) => {
  useEffect(() => {
    // Only collect metrics if we should sample this session
    if (Math.random() > sampleRate) {
      return;
    }

    const metrics: Partial<CoreWebVitalsMetrics> = {};

    const sendToAnalytics = (metric: any) => {
      // Store metric
      switch (metric.name) {
        case 'CLS':
          metrics.cls = metric.value;
          break;
        case 'FID':
          metrics.fid = metric.value;
          break;
        case 'FCP':
          metrics.fcp = metric.value;
          break;
        case 'LCP':
          metrics.lcp = metric.value;
          break;
        case 'TTFB':
          metrics.ttfb = metric.value;
          break;
      }

      if (debug) {
        console.log(`📊 [Web Vitals] ${metric.name}:`, metric.value, metric);
      }

      // Check if we have all metrics
      if (Object.keys(metrics).length === 5) {
        const completeMetrics = metrics as CoreWebVitalsMetrics;

        // Call custom callback if provided
        if (onMetrics) {
          onMetrics(completeMetrics);
        }

        // Send to analytics endpoint if provided
        if (endpoint && typeof fetch !== 'undefined') {
          fetch(endpoint, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              ...completeMetrics,
              url: window.location.href,
              userAgent: navigator.userAgent,
              timestamp: new Date().toISOString(),
            }),
          }).catch(error => {
            if (debug) {
              console.warn('Failed to send web vitals metrics:', error);
            }
          });
        }

        // Analyze and report performance issues
        analyzePerformance(completeMetrics, debug);
      }
    };

    // Measure all Core Web Vitals
    try {
      onCLS(sendToAnalytics);
      onFID(sendToAnalytics);
      onFCP(sendToAnalytics);
      onLCP(sendToAnalytics);
      onTTFB(sendToAnalytics);
    } catch (error) {
      if (debug) {
        console.error('Error measuring web vitals:', error);
      }
    }
  }, [onMetrics, debug, sampleRate, endpoint]);

  return null; // This component doesn't render anything
};

/**
 * Analyze performance metrics and provide recommendations
 */
function analyzePerformance(metrics: CoreWebVitalsMetrics, debug: boolean = false) {
  const issues: Array<{
    metric: string;
    threshold: number;
    actual: number;
    severity: 'good' | 'needs-improvement' | 'poor';
    recommendation: string;
  }> = [];

  // LCP (Largest Contentful Paint) - Target: <2.5s
  const lcpSeverity = metrics.lcp <= 2500 ? 'good' : metrics.lcp <= 4000 ? 'needs-improvement' : 'poor';
  if (lcpSeverity !== 'good') {
    issues.push({
      metric: 'LCP',
      threshold: 2500,
      actual: metrics.lcp,
      severity: lcpSeverity,
      recommendation: 'Optimize images, reduce server response time, eliminate render-blocking resources'
    });
  }

  // FID (First Input Delay) - Target: <100ms
  const fidSeverity = metrics.fid <= 100 ? 'good' : metrics.fid <= 300 ? 'needs-improvement' : 'poor';
  if (fidSeverity !== 'good') {
    issues.push({
      metric: 'FID',
      threshold: 100,
      actual: metrics.fid,
      severity: fidSeverity,
      recommendation: 'Reduce JavaScript execution time, break up long tasks, use web workers'
    });
  }

  // CLS (Cumulative Layout Shift) - Target: <0.1
  const clsSeverity = metrics.cls <= 0.1 ? 'good' : metrics.cls <= 0.25 ? 'needs-improvement' : 'poor';
  if (clsSeverity !== 'good') {
    issues.push({
      metric: 'CLS',
      threshold: 0.1,
      actual: metrics.cls,
      severity: clsSeverity,
      recommendation: 'Include size attributes on images and videos, avoid inserting content above existing content'
    });
  }

  // FCP (First Contentful Paint) - Target: <1.8s
  const fcpSeverity = metrics.fcp <= 1800 ? 'good' : metrics.fcp <= 3000 ? 'needs-improvement' : 'poor';
  if (fcpSeverity !== 'good') {
    issues.push({
      metric: 'FCP',
      threshold: 1800,
      actual: metrics.fcp,
      severity: fcpSeverity,
      recommendation: 'Reduce server response time, eliminate render-blocking resources, optimize CSS'
    });
  }

  // TTFB (Time to First Byte) - Target: <800ms
  const ttfbSeverity = metrics.ttfb <= 800 ? 'good' : metrics.ttfb <= 1800 ? 'needs-improvement' : 'poor';
  if (ttfbSeverity !== 'good') {
    issues.push({
      metric: 'TTFB',
      threshold: 800,
      actual: metrics.ttfb,
      severity: ttfbSeverity,
      recommendation: 'Improve server response time, use CDN, optimize backend performance'
    });
  }

  // Report issues in development
  if (debug && issues.length > 0) {
    console.group('🚨 Core Web Vitals Issues Detected');
    issues.forEach(issue => {
      const emoji = issue.severity === 'poor' ? '🔴' : '🟡';
      console.log(`${emoji} ${issue.metric}: ${issue.actual.toFixed(2)} (target: ${issue.threshold})`);
      console.log(`💡 ${issue.recommendation}`);
    });
    console.groupEnd();
  }

  // Store in localStorage for debugging
  if (typeof window !== 'undefined' && debug) {
    const report = {
      timestamp: new Date().toISOString(),
      metrics,
      issues,
      url: window.location.href
    };

    const reports = JSON.parse(localStorage.getItem('web-vitals-reports') || '[]');
    reports.push(report);

    // Keep only last 10 reports
    if (reports.length > 10) {
      reports.splice(0, reports.length - 10);
    }

    localStorage.setItem('web-vitals-reports', JSON.stringify(reports));
  }

  return issues;
}

/**
 * Utility to get performance reports from localStorage
 */
export function getPerformanceReports() {
  if (typeof window === 'undefined') return [];

  try {
    return JSON.parse(localStorage.getItem('web-vitals-reports') || '[]');
  } catch {
    return [];
  }
}

/**
 * Utility to clear performance reports
 */
export function clearPerformanceReports() {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('web-vitals-reports');
  }
}

export default CoreWebVitalsOptimizer;