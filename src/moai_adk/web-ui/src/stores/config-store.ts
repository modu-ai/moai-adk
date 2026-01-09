/**
 * Config Store
 *
 * Zustand store for 0-Project configuration state management.
 * Handles loading, saving, and validating configuration.
 */

import { create } from "zustand";
import { persist } from "zustand/middleware";
import type {
	FullConfig,
	TabSchema,
	ConfigFormState,
	SaveConfigResponse,
	ValidateConfigResponse,
} from "@/types";
import {
	fetchSchema,
	fetchConfig,
	saveConfig as saveConfigApi,
	validateConfig as validateConfigApi,
} from "@/lib/config-api";

// ===============================
// Helper Functions
// ===============================

/**
 * Flatten nested config to flat form values
 */
function flattenConfig(
	config: FullConfig,
): Record<string, string | boolean | number> {
	const formValues: Record<string, string | boolean | number> = {};

	if (config.user?.name) formValues.user_name = config.user.name;
	if (config.language?.conversation_language) {
		formValues.conversation_language = config.language.conversation_language;
	}
	if (config.language?.agent_prompt_language) {
		formValues.agent_prompt_language = config.language.agent_prompt_language;
	}
	if (config.project?.name) formValues.project_name = config.project.name;
	if (config.github?.profile_name) {
		formValues.github_profile_name = config.github.profile_name;
	}
	if (config.project?.description) {
		formValues.project_description = config.project.description;
	}
	if (config.git_strategy?.mode) {
		formValues.git_strategy_mode = config.git_strategy.mode;
	}
	if (config.git_strategy?.workflow) {
		formValues.git_strategy_workflow = config.git_strategy.workflow;
	}
	if (config.git_strategy?.personal?.auto_checkpoint) {
		formValues.git_personal_auto_checkpoint =
			config.git_strategy.personal.auto_checkpoint;
	}
	if (config.git_strategy?.personal?.push_to_remote !== undefined) {
		formValues.git_personal_push_remote =
			config.git_strategy.personal.push_to_remote;
	}
	if (config.git_strategy?.team?.auto_pr !== undefined) {
		formValues.git_team_auto_pr = config.git_strategy.team.auto_pr;
	}
	if (config.git_strategy?.team?.draft_pr !== undefined) {
		formValues.git_team_draft_pr = config.git_strategy.team.draft_pr;
	}
	if (config.constitution?.test_coverage_target !== undefined) {
		formValues.test_coverage_target = config.constitution.test_coverage_target;
	}
	if (config.constitution?.enforce_tdd !== undefined) {
		formValues.enforce_tdd = config.constitution.enforce_tdd;
	}
	if (config.project?.documentation_mode) {
		formValues.documentation_mode = config.project.documentation_mode;
	}
	if (config.project?.documentation_depth) {
		formValues.documentation_depth = config.project.documentation_depth;
	}

	return formValues;
}

/**
 * Build nested config from flat form values
 */
function buildConfig(
	formValues: Record<string, string | boolean | number>,
): FullConfig {
	const configData: FullConfig = {};

	if (formValues.user_name) {
		configData.user = { name: String(formValues.user_name) };
	}
	if (formValues.conversation_language || formValues.agent_prompt_language) {
		configData.language = {
			conversation_language: String(formValues.conversation_language || "en"),
			agent_prompt_language: String(formValues.agent_prompt_language || "en"),
		};
	}
	if (formValues.project_name) {
		configData.project = {
			name: String(formValues.project_name),
			description: formValues.project_description
				? String(formValues.project_description)
				: "",
		};
	}
	if (formValues.github_profile_name) {
		configData.github = {
			profile_name: String(formValues.github_profile_name),
		};
	}
	if (formValues.documentation_mode) {
		if (!configData.project) configData.project = { name: "" };
		configData.project.documentation_mode = String(
			formValues.documentation_mode,
		);
		if (formValues.documentation_depth) {
			configData.project.documentation_depth = String(
				formValues.documentation_depth,
			);
		}
	}
	if (formValues.git_strategy_mode) {
		configData.git_strategy = {
			mode: String(formValues.git_strategy_mode),
		};
		if (formValues.git_strategy_workflow) {
			configData.git_strategy.workflow = String(
				formValues.git_strategy_workflow,
			);
		}
		if (
			formValues.git_strategy_mode === "personal" ||
			formValues.git_strategy_mode === "hybrid"
		) {
			configData.git_strategy.personal = {};
			if (formValues.git_personal_auto_checkpoint) {
				configData.git_strategy.personal.auto_checkpoint = String(
					formValues.git_personal_auto_checkpoint,
				);
			}
			if (formValues.git_personal_push_remote !== undefined) {
				configData.git_strategy.personal.push_to_remote = Boolean(
					formValues.git_personal_push_remote,
				);
			}
		}
		if (
			formValues.git_strategy_mode === "team" ||
			formValues.git_strategy_mode === "hybrid"
		) {
			configData.git_strategy.team = {};
			if (formValues.git_team_auto_pr !== undefined) {
				configData.git_strategy.team.auto_pr = Boolean(
					formValues.git_team_auto_pr,
				);
			}
			if (formValues.git_team_draft_pr !== undefined) {
				configData.git_strategy.team.draft_pr = Boolean(
					formValues.git_team_draft_pr,
				);
			}
		}
	}
	if (
		formValues.test_coverage_target !== undefined ||
		formValues.enforce_tdd !== undefined
	) {
		configData.constitution = {};
		if (formValues.test_coverage_target !== undefined) {
			configData.constitution.test_coverage_target = Number(
				formValues.test_coverage_target,
			);
		}
		if (formValues.enforce_tdd !== undefined) {
			configData.constitution.enforce_tdd = Boolean(formValues.enforce_tdd);
		}
	}

	return configData;
}

// ===============================
// Config Store
// ===============================

interface ConfigStore extends ConfigFormState {
	// Data
	schema: TabSchema | null;
	config: FullConfig | null;
	formValues: Record<string, string | boolean | number>;

	// Actions
	loadSchema: () => Promise<void>;
	loadConfig: () => Promise<void>;
	saveConfig: () => Promise<SaveConfigResponse | null>;
	validateConfig: () => Promise<ValidateConfigResponse | null>;
	setFieldValue: (fieldId: string, value: string | boolean | number) => void;
	setFormValues: (values: Record<string, string | boolean | number>) => void;
	resetForm: () => void;
	setCurrentTab: (tabId: string) => void;
	clearErrors: () => void;
}

export const useConfigStore = create<ConfigStore>()(
	persist(
		(set, get) => ({
			// Initial state
			currentTab: "tab_1_quick_start",
			isDirty: false,
			isSaving: false,
			isValid: false,
			errors: {},
			schema: null,
			config: null,
			formValues: {},

			// Load tab schema from API
			loadSchema: async () => {
				const response = await fetchSchema();

				if (response.success && response.data) {
					set({ schema: response.data });
				} else {
					set((state) => ({
						errors: {
							...state.errors,
							schema: response.error?.message || "Failed to load schema",
						},
					}));
				}
			},

			// Load current config from API
			loadConfig: async () => {
				const response = await fetchConfig();

				if (response.success && response.data) {
					const config = response.data;
					const formValues = flattenConfig(config);
					set({ config, formValues, isDirty: false });
				}
			},

			// Save config to API
			saveConfig: async () => {
				const { formValues } = get();

				set({ isSaving: true, errors: {} });

				const configData = buildConfig(formValues);
				const response = await saveConfigApi(configData);

				set({ isSaving: false });

				if (response.success && response.data) {
					set({ isDirty: false, config: configData });
					return response.data;
				}
				set((state) => ({
					errors: {
						...state.errors,
						save: response.error?.message || "Failed to save config",
					},
				}));
				return null;
			},

			// Validate config without saving
			validateConfig: async () => {
				const { formValues } = get();
				const configData = buildConfig(formValues);
				const response = await validateConfigApi(configData);

				if (response.success && response.data) {
					set({ isValid: response.data.valid });
					return response.data;
				}
				return null;
			},

			// Set a single form field value
			setFieldValue: (fieldId: string, value: string | boolean | number) => {
				set((state) => ({
					formValues: { ...state.formValues, [fieldId]: value },
					isDirty: true,
					isValid: false, // Re-validate on next change
				}));
			},

			// Set multiple form values at once
			setFormValues: (values: Record<string, string | boolean | number>) => {
				set((state) => ({
					formValues: { ...state.formValues, ...values },
					isDirty: true,
					isValid: false,
				}));
			},

			// Reset form to initial state
			resetForm: () => {
				const { config } = get();
				const formValues = config ? flattenConfig(config) : {};
				set({ formValues, isDirty: false, errors: {} });
			},

			// Set current tab
			setCurrentTab: (tabId: string) => {
				set({ currentTab: tabId });
			},

			// Clear all errors
			clearErrors: () => {
				set({ errors: {} });
			},
		}),
		{
			name: "moai-config-storage",
			partialize: (state) => ({
				formValues: state.formValues,
				currentTab: state.currentTab,
			}),
		},
	),
);
