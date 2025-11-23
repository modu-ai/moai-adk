      'return',
      'dest',
      'destination',
    ];

    for (const param of suspiciousParams) {
      if (url.searchParams.has(param)) {
        const paramValue = url.searchParams.get(param);
        if (paramValue && this.isSuspiciousURL(paramValue)) {
          return this.createInvalidResult(
            'SUSPICIOUS_PARAMETER',
            `Suspicious parameter ${param} detected`
          );
        }
      }
    }

    // Check URL length (very long URLs can be suspicious)
    if (url.toString().length > 2048) {
      return this.createInvalidResult(
        'URL_TOO_LONG',
        'URL exceeds maximum allowed length'
      );
    }

    return { isValid: true };
  }

  private isSuspiciousURL(url: string): boolean {
    const suspiciousPatterns = [
      /localhost/i,
      /127\.0\.0\.1/,
      /0x[0-9a-f]+/i,
      /internal/i,
      /private/i,
      /metadata/i,
      /169\.254\./, // Link-local
      /file:\/\//,
    ];

    return suspiciousPatterns.some(pattern => pattern.test(url));
  }

  async makeSecureRequest(url: string, options: RequestOptions = {}): Promise<SecureRequestResult> {
    // Validate URL first
    const validationResult = await this.validateURL(url);
    if (!validationResult.isValid) {
      throw new SSRFError('URL_VALIDATION_FAILED', validationResult.reason || 'Invalid URL');
    }

    // Check request cache
    const cacheKey = this.generateCacheKey(url, options);
    const cached = this.requestCache.get(cacheKey);
    if (cached && !this.isCacheExpired(cached)) {
      return cached.result;
    }

    try {
      // Make secure request
      const startTime = Date.now();
      const response = await this.makeRequest(url, {
        timeout: this.config.timeoutMs,
        maxRedirects: this.config.maxRedirects,
        ...options,
      });
      
      const responseTime = Date.now() - startTime;

      // Validate response
      const responseValidation = await this.validateResponse(response);
      if (!responseValidation.isValid) {
        throw new SSRFError('RESPONSE_VALIDATION_FAILED', responseValidation.reason || 'Invalid response');
      }

      const result: SecureRequestResult = {
        success: true,
        data: response.data,
        status: response.status,
        headers: response.headers,
        responseTime,
        url: validationResult.normalizedURL,
      };

      // Cache result
      this.requestCache.set(cacheKey, {
        result,
        timestamp: Date.now(),
      });

      // Log request
      this.logger.log('SECURE_REQUEST_SUCCESS', {
        url: validationResult.normalizedURL,
        responseTime,
        status: response.status,
        riskScore: validationResult.riskScore,
      });

      return result;
    } catch (error) {
      // Log failed request
      this.logger.log('SECURE_REQUEST_FAILED', {
        url: url,
        error: error.message,
        type: error.name,
      });

      throw error;
    }
  }

  private async makeRequest(url: string, options: RequestOptions): Promise<RequestResponse> {
    const axios = require('axios');
    
    const response = await axios({
      url,
      method: options.method || 'GET',
      timeout: options.timeout,
      maxRedirects: options.maxRedirects,
      headers: {
        'User-Agent': 'SecureRequest/1.0',
        ...options.headers,
      },
      responseType: 'arraybuffer', // Handle binary responses
      maxContentLength: this.config.maxResponseSize,
    });

    return {
      data: response.data,
      status: response.status,
      headers: response.headers,
    };
  }

  private async validateResponse(response: RequestResponse): Promise<ValidationResult> {
    // Check response size
    if (Buffer.byteLength(response.data) > this.config.maxResponseSize) {
      return {
        isValid: false,
        reason: `Response size exceeds maximum allowed size`,
      };
    }

    // Check content type (optional based on requirements)
    const contentType = response.headers['content-type'];
    if (contentType && this.isSuspiciousContentType(contentType)) {
      return {
        isValid: false,
        reason: `Suspicious content type: ${contentType}`,
      };
    }

    return { isValid: true };
  }

  private isSuspiciousContentType(contentType: string): boolean {
    const suspiciousTypes = [
      'application/octet-stream',
      'application/x-executable',
      'application/x-msdownload',
      'application/x-msdos-program',
    ];

    return suspiciousTypes.some(type => contentType.includes(type));
  }

  private calculateRiskScore(url: URL): number {
    let score = 0;

    // URL length factor
    if (url.toString().length > 1000) score += 10;
    if (url.toString().length > 2000) score += 20;

    // Protocol factor
    if (url.protocol !== 'https:') score += 5;

    // Domain complexity factor
    if (url.hostname.split('.').length > 4) score += 5;

    // Query parameter factor
    if (url.searchParams.size > 10) score += 5;
    if (url.searchParams.size > 20) score += 15;

    return Math.min(score, 100); // Cap at 100
  }

  private createInvalidResult(reason: string, details?: string): URLValidationResult {
    return {
      isValid: false,
      reason,
      details,
      riskScore: 100, // High risk for invalid URLs
    };
  }
}

// Error classes
export class SSRFError extends Error {
  constructor(
    public code: string,
    message: string,
    public details?: any
  ) {
    super(message);
    this.name = 'SSRFError';
  }
}

// Types
interface URLValidationResult {
  isValid: boolean;
  normalizedURL?: string;
  reason?: string;
  details?: string;
  riskScore?: number;
}

interface SecureRequestResult {
  success: boolean;
  data: Buffer;
  status: number;
  headers: Record<string, string>;
  responseTime: number;
  url: string;
}

interface RequestResponse {
  data: Buffer;
  status: number;
  headers: Record<string, string>;
}

interface RequestOptions {
  method?: string;
  timeout?: number;
  maxRedirects?: number;
  headers?: Record<string, string>;
}

interface ValidationResult {
  isValid: boolean;
  reason?: string;
}

interface RequestCache {
  result: SecureRequestResult;
  timestamp: number;
}
```




## Reference & Resources

See [reference.md](reference.md) for detailed API reference and official documentation.
