// @CODE:TYPE-SAFETY-001 |
// Related: @CODE:TYPE-001:INFRA, @DOC:TYPE-001

/**
 * Custom Error classes with type-safe properties
 * Resolves TS2353 errors by properly extending Error
 */

export class ValidationError extends Error {
  public readonly pattern?: string | undefined;
  public readonly vulnerabilities?: string[] | undefined;
  public readonly context?: Record<string, unknown> | undefined;

  constructor(
    message: string,
    options?: {
      pattern?: string | undefined;
      vulnerabilities?: string[] | undefined;
      context?: Record<string, unknown> | undefined;
    }
  ) {
    super(message);
    this.name = 'ValidationError';
    this.pattern = options?.pattern;
    this.vulnerabilities = options?.vulnerabilities;
    this.context = options?.context;
    Object.setPrototypeOf(this, ValidationError.prototype);
  }
}

export class InstallationError extends Error {
  public readonly error?: Error | undefined;
  public readonly errorMessage?: string | undefined;
  public readonly projectPath?: string | undefined;
  public readonly phase?: string | undefined;
  public readonly context?: Record<string, unknown> | undefined;

  constructor(
    message: string,
    options?: {
      error?: Error | undefined;
      errorMessage?: string | undefined;
      projectPath?: string | undefined;
      phase?: string | undefined;
      context?: Record<string, unknown> | undefined;
    }
  ) {
    super(message);
    this.name = 'InstallationError';
    this.error = options?.error;
    this.errorMessage = options?.errorMessage;
    this.projectPath = options?.projectPath;
    this.phase = options?.phase;
    this.context = options?.context;
    Object.setPrototypeOf(this, InstallationError.prototype);
  }
}

export class TemplateError extends Error {
  public readonly error?: Error | undefined;
  public readonly templatePath?: string | undefined;
  public readonly context?: Record<string, unknown> | undefined;

  constructor(
    message: string,
    options?: {
      error?: Error | undefined;
      templatePath?: string | undefined;
      context?: Record<string, unknown> | undefined;
    }
  ) {
    super(message);
    this.name = 'TemplateError';
    this.error = options?.error;
    this.templatePath = options?.templatePath;
    this.context = options?.context;
    Object.setPrototypeOf(this, TemplateError.prototype);
  }
}

export class ResourceError extends Error {
  public readonly error?: Error | undefined;
  public readonly resourcePath?: string | undefined;
  public readonly context?: Record<string, unknown> | undefined;

  constructor(
    message: string,
    options?: {
      error?: Error | undefined;
      resourcePath?: string | undefined;
      context?: Record<string, unknown> | undefined;
    }
  ) {
    super(message);
    this.name = 'ResourceError';
    this.error = options?.error;
    this.resourcePath = options?.resourcePath;
    this.context = options?.context;
    Object.setPrototypeOf(this, ResourceError.prototype);
  }
}

export class PhaseError extends Error {
  public readonly error?: Error | undefined;
  public readonly phase?: string | undefined;
  public readonly context?: Record<string, unknown> | undefined;

  constructor(
    message: string,
    options?: {
      error?: Error | undefined;
      phase?: string | undefined;
      context?: Record<string, unknown> | undefined;
    }
  ) {
    super(message);
    this.name = 'PhaseError';
    this.error = options?.error;
    this.phase = options?.phase;
    this.context = options?.context;
    Object.setPrototypeOf(this, PhaseError.prototype);
  }
}

/**
 * Type guard to check if an error is of a specific custom type
 */
export function isValidationError(error: unknown): error is ValidationError {
  return error instanceof ValidationError;
}

export function isInstallationError(
  error: unknown
): error is InstallationError {
  return error instanceof InstallationError;
}

export function isTemplateError(error: unknown): error is TemplateError {
  return error instanceof TemplateError;
}

export function isResourceError(error: unknown): error is ResourceError {
  return error instanceof ResourceError;
}

export function isPhaseError(error: unknown): error is PhaseError {
  return error instanceof PhaseError;
}

/**
 * Convert unknown error to standard Error
 */
export function toError(error: unknown): Error {
  if (error instanceof Error) {
    return error;
  }
  return new Error(String(error));
}

/**
 * Get error message safely
 */
export function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}
