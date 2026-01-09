/**
 * Config Types
 *
 * Type definitions for 0-Project configuration dialog.
 * Matches the Tab Schema v3.0.0 backend structure.
 */

// ===============================
// Question Types
// ===============================

export type QuestionType = "text_input" | "select_single" | "number_input";

export interface QuestionOption {
	label: string;
	value: string | boolean | number;
	description?: string;
}

export interface Question {
	id: string;
	question: string;
	type: QuestionType;
	required: boolean;
	options: QuestionOption[];
	smart_default?: string | boolean | number;
	min?: number;
	max?: number;
	show_if?: string;
	conditional_mapping?: Record<string, string[]>;
	smart_default_mapping?: Record<string, string | boolean>;
}

export interface Batch {
	id: string;
	header: string;
	batch_number: number;
	total_batches: number;
	questions: Question[];
	show_if?: string;
}

export interface Tab {
	id: string;
	label: string;
	description: string;
	batches: Batch[];
}

export interface TabSchema {
	version: string;
	tabs: Tab[];
}

// ===============================
// Config Types
// ===============================

export interface UserConfig {
	name: string;
}

export interface LanguageConfig {
	conversation_language: string;
	agent_prompt_language?: string;
	conversation_language_name?: string;
}

export interface ProjectConfig {
	name: string;
	description?: string;
	language?: string;
	locale?: string;
	template_version?: string;
	documentation_mode?: string;
	documentation_depth?: string;
}

export interface GitHubConfig {
	profile_name?: string;
}

export interface GitStrategyPersonalConfig {
	workflow?: string;
	auto_checkpoint?: string;
	push_to_remote?: boolean;
}

export interface GitStrategyTeamConfig {
	workflow?: string;
	auto_pr?: boolean;
	draft_pr?: boolean;
}

export interface GitStrategyConfig {
	mode: string;
	workflow?: string;
	personal?: GitStrategyPersonalConfig;
	team?: GitStrategyTeamConfig;
}

export interface ConstitutionConfig {
	test_coverage_target?: number;
	enforce_tdd?: boolean;
}

export interface MoAIConfig {
	version?: string;
}

export interface FullConfig {
	version?: string;
	user?: UserConfig;
	language?: LanguageConfig;
	project?: ProjectConfig;
	github?: GitHubConfig;
	git_strategy?: GitStrategyConfig;
	constitution?: ConstitutionConfig;
	moai?: MoAIConfig;
}

// ===============================
// API Types
// ===============================

export interface SaveConfigResponse {
	success: boolean;
	message?: string;
	backup_path?: string;
}

export interface ValidateConfigResponse {
	valid: boolean;
	missing_fields?: string[];
	errors?: string[];
}

export interface ApiError {
	code: string;
	message: string;
	detail?: {
		message?: string;
		missing_fields?: string[];
	};
}

export interface ApiResponse<T> {
	success: boolean;
	data?: T;
	error?: ApiError;
}

// ===============================
// UI State Types
// ===============================

export interface ConfigFormState {
	currentTab: string;
	isDirty: boolean;
	isSaving: boolean;
	isValid: boolean;
	errors: Record<string, string>;
}

export type TabId =
	| "tab_1_quick_start"
	| "tab_2_documentation"
	| "tab_3_git_automation";
