/**
 * Question Input Component
 *
 * Renders a single question input based on question type.
 * Handles text_input, select_single, and number_input types.
 */

import { Label } from "@/components/ui";
import { Input } from "@/components/ui";
import type { Question } from "@/types";

interface QuestionInputProps {
	question: Question;
	value?: string | boolean | number;
	onChange: (value: string | boolean | number) => void;
	error?: string;
	disabled?: boolean;
}

export function QuestionInput({
	question,
	value = "",
	onChange,
	error,
	disabled = false,
}: QuestionInputProps) {
	const fieldId = question.id;

	// Text input type
	if (question.type === "text_input") {
		return (
			<div className="space-y-2">
				<Label htmlFor={fieldId} className="text-sm font-medium">
					{question.question}
					{question.required && <span className="text-red-500 ml-1">*</span>}
				</Label>
				<Input
					id={fieldId}
					type="text"
					value={String(value)}
					onChange={(e) => onChange(e.target.value)}
					placeholder={
						question.smart_default !== undefined
							? String(question.smart_default)
							: undefined
					}
					disabled={disabled}
					className={error ? "border-red-500" : ""}
				/>
				{error && <p className="text-xs text-red-500">{error}</p>}
			</div>
		);
	}

	// Select input type
	if (question.type === "select_single") {
		return (
			<div className="space-y-2">
				<Label htmlFor={fieldId} className="text-sm font-medium">
					{question.question}
					{question.required && <span className="text-red-500 ml-1">*</span>}
				</Label>
				<select
					id={fieldId}
					value={String(value)}
					onChange={(e) => {
						const selectedOption = question.options.find(
							(opt) => String(opt.value) === e.target.value,
						);
						onChange(selectedOption ? selectedOption.value : e.target.value);
					}}
					disabled={disabled}
					className={`flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 ${error ? "border-red-500" : ""}`}
				>
					{!question.required && <option value="">Select an option</option>}
					{question.options.map((option) => (
						<option key={String(option.value)} value={String(option.value)}>
							{option.label}
						</option>
					))}
				</select>
				{error && <p className="text-xs text-red-500">{error}</p>}
			</div>
		);
	}

	// Number input type
	if (question.type === "number_input") {
		const numValue =
			typeof value === "number" ? value : Number.parseFloat(String(value)) || 0;
		return (
			<div className="space-y-2">
				<Label htmlFor={fieldId} className="text-sm font-medium">
					{question.question}
					{question.required && <span className="text-red-500 ml-1">*</span>}
				</Label>
				<Input
					id={fieldId}
					type="number"
					value={numValue}
					onChange={(e) => onChange(Number.parseFloat(e.target.value) || 0)}
					min={question.min}
					max={question.max}
					step={question.max && question.max > 1 ? "1" : "any"}
					disabled={disabled}
					className={error ? "border-red-500" : ""}
				/>
				{error && <p className="text-xs text-red-500">{error}</p>}
			</div>
		);
	}

	// Fallback for unknown types
	return (
		<div className="text-sm text-muted-foreground">
			Unknown question type: {question.type}
		</div>
	);
}
