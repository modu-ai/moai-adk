/**
 * Batch Questions Component
 *
 * Renders a batch of questions with header and conditional visibility.
 */

import type { Batch } from "@/types";
import { QuestionInput } from "./question-input";

interface BatchQuestionsProps {
	batch: Batch;
	formValues: Record<string, string | boolean | number>;
	errors: Record<string, string>;
	onChange: (questionId: string, value: string | boolean | number) => void;
	isVisible?: boolean;
}

export function BatchQuestions({
	batch,
	formValues,
	errors,
	onChange,
	isVisible = true,
}: BatchQuestionsProps) {
	if (!isVisible) {
		return null;
	}

	return (
		<div className="space-y-4">
			{batch.header && (
				<h3 className="text-lg font-semibold text-foreground">
					{batch.header}
					<span className="ml-2 text-sm font-normal text-muted-foreground">
						({batch.batch_number} / {batch.total_batches})
					</span>
				</h3>
			)}

			<div className="space-y-4">
				{batch.questions.map((question) => (
					<QuestionInput
						key={question.id}
						question={question}
						value={formValues[question.id]}
						onChange={(value) => onChange(question.id, value)}
						error={errors[question.id]}
					/>
				))}
			</div>
		</div>
	);
}
