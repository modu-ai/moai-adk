/**
 * Config Form Component
 *
 * Renders a complete configuration form with batches and conditional logic.
 * Handles smart defaults and conditional batch visibility.
 */

import { useEffect, useState } from "react";
import type { Tab } from "@/types";
import { BatchQuestions } from "./batch-questions";
import { Button } from "@/components/ui";

interface ConfigFormProps {
	tab: Tab;
	formValues: Record<string, string | boolean | number>;
	errors: Record<string, string>;
	isSaving?: boolean;
	onChange: (questionId: string, value: string | boolean | number) => void;
	onSave?: () => void;
	onCancel?: () => void;
}

/**
 * Check if a batch should be visible based on show_if condition
 */
function isBatchVisible(
	batch: { show_if?: string },
	formValues: Record<string, string | boolean | number>,
): boolean {
	if (!batch.show_if) {
		return true;
	}

	// Parse simple conditions like "git_strategy_mode==personal"
	const [fieldId, expectedValue] = batch.show_if.split("==");
	if (!fieldId || !expectedValue) {
		return true;
	}

	const currentValue = formValues[fieldId];
	return String(currentValue) === expectedValue.trim();
}

/**
 * Apply smart defaults to form values based on schema
 */
function applySmartDefaults(
	questions: unknown[],
	formValues: Record<string, string | boolean | number>,
): Record<string, string | boolean | number> {
	const result = { ...formValues };

	for (const question of questions as {
		id: string;
		smart_default?: string | boolean | number;
	}[]) {
		if (
			question.smart_default !== undefined &&
			result[question.id] === undefined
		) {
			result[question.id] = question.smart_default;
		}
	}

	return result;
}

export function ConfigForm({
	tab,
	formValues,
	errors,
	isSaving = false,
	onChange,
	onSave,
	onCancel,
}: ConfigFormProps) {
	const [localValues, setLocalValues] = useState(formValues);

	// Apply smart defaults when tab loads
	useEffect(() => {
		let values = { ...formValues };
		for (const batch of tab.batches) {
			values = applySmartDefaults(batch.questions, values);
		}
		setLocalValues(values);
	}, [tab.id]);

	// Update local values when formValues change
	useEffect(() => {
		setLocalValues(formValues);
	}, [formValues]);

	const handleChange = (
		questionId: string,
		value: string | boolean | number,
	) => {
		const newValues = { ...localValues, [questionId]: value };
		setLocalValues(newValues);
		onChange(questionId, value);
	};

	const hasRequiredFields = tab.batches.some((batch) =>
		batch.questions.some((q) => q.required && !localValues[q.id]),
	);

	return (
		<div className="space-y-6">
			{tab.description && (
				<div className="text-sm text-muted-foreground">{tab.description}</div>
			)}

			{tab.batches.map((batch) => (
				<BatchQuestions
					key={batch.id}
					batch={batch}
					formValues={localValues}
					errors={errors}
					onChange={handleChange}
					isVisible={isBatchVisible(batch, localValues)}
				/>
			))}

			{(onSave || onCancel) && (
				<div className="flex gap-3 pt-4 border-t">
					{onCancel && (
						<Button
							type="button"
							variant="outline"
							onClick={onCancel}
							disabled={isSaving}
						>
							Cancel
						</Button>
					)}
					{onSave && (
						<Button
							type="button"
							onClick={onSave}
							disabled={isSaving || hasRequiredFields}
						>
							{isSaving ? "Saving..." : "Save Configuration"}
						</Button>
					)}
				</div>
			)}
		</div>
	);
}
