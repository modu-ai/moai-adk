/**
 * Config Validation Schemas
 *
 * Zod validation schemas for 0-Project configuration.
 * Provides runtime type validation for config API responses and form data.
 */

import { z } from "zod";
import type {
	TabSchema as TabSchemaType,
	FullConfig as FullConfigType,
	SaveConfigResponse as SaveConfigResponseType,
} from "@/types";

// ===============================
// Question Schemas
// ===============================

export const QuestionOptionSchema = z.object({
	label: z.string(),
	value: z.union([z.string(), z.boolean(), z.number()]),
	description: z.string().optional(),
});

export const QuestionSchema = z.object({
	id: z.string(),
	question: z.string(),
	type: z.enum(["text_input", "select_single", "number_input"]),
	required: z.boolean(),
	options: z.array(QuestionOptionSchema).default([]),
	smart_default: z.union([z.string(), z.boolean(), z.number()]).optional(),
	min: z.number().optional(),
	max: z.number().optional(),
	show_if: z.string().optional(),
	conditional_mapping: z.record(z.string(), z.array(z.string())).optional(),
	smart_default_mapping: z
		.record(z.string(), z.union([z.string(), z.boolean()]))
		.optional(),
});

export const BatchSchema = z.object({
	id: z.string(),
	header: z.string(),
	batch_number: z.number(),
	total_batches: z.number(),
	questions: z.array(QuestionSchema).default([]),
	show_if: z.string().optional(),
});

export const TabSchema = z.object({
	id: z.string(),
	label: z.string(),
	description: z.string(),
	batches: z.array(BatchSchema).default([]),
});

export const TabSchemaResponseSchema = z.object({
	version: z.literal("3.0.0"),
	tabs: z.array(TabSchema),
});

// ===============================
// Config Schemas
// ===============================

export const UserConfigSchema = z.object({
	name: z.string().min(1, "User name is required"),
});

export const LanguageConfigSchema = z.object({
	conversation_language: z.enum(["ko", "en", "ja", "zh"]),
	agent_prompt_language: z.enum(["ko", "en"]).default("en"),
	conversation_language_name: z.string().optional(),
});

export const ProjectConfigSchema = z.object({
	name: z.string().min(1, "Project name is required"),
	description: z.string().default(""),
	language: z.string().optional(),
	locale: z.string().optional(),
	template_version: z.string().optional(),
	documentation_mode: z.enum(["skip", "full_now", "minimal"]).default("skip"),
	documentation_depth: z.enum(["quick", "standard", "deep"]).optional(),
});

export const GitHubConfigSchema = z.object({
	profile_name: z.string().optional(),
});

export const GitStrategyPersonalConfigSchema = z.object({
	workflow: z.enum(["github-flow", "git-flow", "trunk-based"]).optional(),
	auto_checkpoint: z.enum(["disabled", "event-driven", "manual"]).optional(),
	push_to_remote: z.boolean().default(false),
});

export const GitStrategyTeamConfigSchema = z.object({
	workflow: z.enum(["github-flow", "git-flow"]).optional(),
	auto_pr: z.boolean().default(false),
	draft_pr: z.boolean().default(false),
});

export const GitStrategyConfigSchema = z.object({
	mode: z.enum(["manual", "personal", "team", "hybrid"]),
	workflow: z.string().optional(),
	personal: GitStrategyPersonalConfigSchema.optional(),
	team: GitStrategyTeamConfigSchema.optional(),
});

export const ConstitutionConfigSchema = z.object({
	test_coverage_target: z.number().min(0).max(100).default(85),
	enforce_tdd: z.boolean().default(true),
});

export const MoAIConfigSchema = z.object({
	version: z.string().optional(),
});

export const FullConfigSchema = z.object({
	version: z.string().optional(),
	user: UserConfigSchema.optional(),
	language: LanguageConfigSchema.optional(),
	project: ProjectConfigSchema.optional(),
	github: GitHubConfigSchema.optional(),
	git_strategy: GitStrategyConfigSchema.optional(),
	constitution: ConstitutionConfigSchema.optional(),
	moai: MoAIConfigSchema.optional(),
});

// ===============================
// API Response Schemas
// ===============================

export const SaveConfigResponseSchema = z.object({
	success: z.boolean(),
	message: z.string().optional(),
	backup_path: z.string().optional(),
});

export const ValidateConfigResponseSchema = z.object({
	valid: z.boolean(),
	missing_fields: z.array(z.string()).default([]),
	errors: z.array(z.string()).default([]),
});

export const ApiErrorSchema = z.object({
	code: z.string(),
	message: z.string(),
	detail: z
		.object({
			message: z.string().optional(),
			missing_fields: z.array(z.string()).optional(),
		})
		.optional(),
});

export const ApiResponseSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
	z.object({
		success: z.boolean(),
		data: dataSchema.optional(),
		error: ApiErrorSchema.optional(),
	});

// ===============================
// Form Input Schemas
// ===============================

export const ConfigFormDataSchema = z.object({
	// Tab 1: Quick Start
	user_name: z.string().min(1),
	conversation_language: z.enum(["ko", "en", "ja", "zh"]),
	agent_prompt_language: z.enum(["ko", "en"]).default("en"),
	project_name: z.string().min(1),
	github_profile_name: z.string().optional(),
	project_description: z.string().optional(),
	git_strategy_mode: z.enum(["manual", "personal", "team", "hybrid"]),
	git_strategy_workflow: z.string().optional(),
	git_personal_auto_checkpoint: z.string().optional(),
	git_personal_push_remote: z.boolean().default(false),
	git_team_auto_pr: z.boolean().default(false),
	git_team_draft_pr: z.boolean().default(false),
	test_coverage_target: z.number().min(0).max(100),
	enforce_tdd: z.boolean().default(true),

	// Tab 2: Documentation
	documentation_mode: z.enum(["skip", "full_now", "minimal"]),
	documentation_depth: z.enum(["quick", "standard", "deep"]).optional(),
});

export type ConfigFormData = z.infer<typeof ConfigFormDataSchema>;

// ===============================
// Validation Helper Functions
// ===============================

/**
 * Validate tab schema response from API
 */
export function validateTabSchema(data: unknown): TabSchemaType {
	const result = TabSchemaResponseSchema.safeParse(data);
	if (!result.success) {
		console.error("Invalid tab schema:", result.error);
		throw new Error("Invalid tab schema format");
	}
	return result.data;
}

/**
 * Validate full config response from API
 */
export function validateFullConfig(data: unknown): FullConfigType {
	const result = FullConfigSchema.safeParse(data);
	if (!result.success) {
		console.error("Invalid config:", result.error);
		throw new Error("Invalid config format");
	}
	return result.data;
}

/**
 * Validate save config response from API
 */
export function validateSaveConfigResponse(
	data: unknown,
): SaveConfigResponseType {
	const result = SaveConfigResponseSchema.safeParse(data);
	if (!result.success) {
		console.error("Invalid save response:", result.error);
		throw new Error("Invalid save response format");
	}
	return result.data;
}

/**
 * Validate config form data
 */
export function validateConfigFormData(data: unknown): ConfigFormData {
	return ConfigFormDataSchema.parse(data);
}
