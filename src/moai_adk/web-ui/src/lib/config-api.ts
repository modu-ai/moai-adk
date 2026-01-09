/**
 * Config API Client
 *
 * Dedicated API client functions for 0-Project configuration endpoints.
 * Provides typed HTTP requests to backend config API.
 */

import type {
	TabSchema,
	FullConfig,
	SaveConfigResponse,
	ValidateConfigResponse,
	ApiResponse,
} from "@/types";

const API_BASE = "/api";

// ===============================
// API Client Functions
// ===============================

/**
 * GET /api/schema
 * Fetch tab schema v3.0.0 for configuration dialog
 */
export async function fetchSchema(): Promise<ApiResponse<TabSchema>> {
	try {
		const response = await fetch(`${API_BASE}/schema`);

		if (!response.ok) {
			const error = await response.json().catch(() => ({
				code: "UNKNOWN_ERROR",
				message: response.statusText,
			}));
			return { success: false, error };
		}

		const data = await response.json();
		return { success: true, data };
	} catch (error) {
		return {
			success: false,
			error: {
				code: "NETWORK_ERROR",
				message: error instanceof Error ? error.message : "Network error",
			},
		};
	}
}

/**
 * GET /api/config
 * Fetch current configuration from backend
 */
export async function fetchConfig(): Promise<ApiResponse<FullConfig>> {
	try {
		const response = await fetch(`${API_BASE}/config`, {
			headers: {
				"Content-Type": "application/json",
			},
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({
				code: "UNKNOWN_ERROR",
				message: response.statusText,
			}));
			return { success: false, error };
		}

		const data = await response.json();
		return { success: true, data };
	} catch (error) {
		return {
			success: false,
			error: {
				code: "NETWORK_ERROR",
				message: error instanceof Error ? error.message : "Network error",
			},
		};
	}
}

/**
 * POST /api/config
 * Save configuration to backend with backup creation
 */
export async function saveConfig(
	configData: FullConfig,
): Promise<ApiResponse<SaveConfigResponse>> {
	try {
		const response = await fetch(`${API_BASE}/config`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(configData),
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({
				code: "UNKNOWN_ERROR",
				message: response.statusText,
			}));
			return { success: false, error };
		}

		const data = await response.json();
		return { success: true, data };
	} catch (error) {
		return {
			success: false,
			error: {
				code: "NETWORK_ERROR",
				message: error instanceof Error ? error.message : "Network error",
			},
		};
	}
}

/**
 * POST /api/validate
 * Validate configuration without saving
 */
export async function validateConfig(
	configData: FullConfig,
): Promise<ApiResponse<ValidateConfigResponse>> {
	try {
		const response = await fetch(`${API_BASE}/validate`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(configData),
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({
				code: "UNKNOWN_ERROR",
				message: response.statusText,
			}));
			return { success: false, error };
		}

		const data = await response.json();
		return { success: true, data };
	} catch (error) {
		return {
			success: false,
			error: {
				code: "NETWORK_ERROR",
				message: error instanceof Error ? error.message : "Network error",
			},
		};
	}
}

// ===============================
// Helper Functions
// ===============================

/**
 * Check if API response is successful
 */
export function isApiSuccess<T>(
	response: ApiResponse<T>,
): response is ApiResponse<T> & { success: true; data: T } {
	return response.success === true && response.data !== undefined;
}

/**
 * Extract error message from API response
 */
export function getApiErrorMessage(response: ApiResponse<never>): string {
	if (!response.success && response.error) {
		return response.error.message;
	}
	return "Unknown error occurred";
}
