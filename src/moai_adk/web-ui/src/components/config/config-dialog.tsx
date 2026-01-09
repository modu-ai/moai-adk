/**
 * Config Dialog Component
 *
 * Main 0-Project configuration dialog with tabs.
 * Integrates schema loading, form rendering, and save functionality.
 */

import { useEffect, useState } from "react";
import { useConfigStore } from "@/stores";
import { ConfigForm } from "./config-form";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui";
import { Badge } from "@/components/ui";

interface ConfigDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export function ConfigDialog({ open, onOpenChange }: ConfigDialogProps) {
	const {
		schema,
		currentTab,
		isSaving,
		isDirty,
		errors,
		loadSchema,
		loadConfig,
		saveConfig,
		setFieldValue,
		setCurrentTab,
		clearErrors,
		resetForm,
	} = useConfigStore();

	const [isInitialized, setIsInitialized] = useState(false);

	// Load schema and config on mount
	useEffect(() => {
		if (open && !isInitialized) {
			loadSchema();
			loadConfig();
			setIsInitialized(true);
		}
	}, [open, isInitialized, loadSchema, loadConfig]);

	// Clear errors when dialog opens
	useEffect(() => {
		if (open) {
			clearErrors();
		}
	}, [open, clearErrors]);

	const handleSave = async () => {
		const result = await saveConfig();
		if (result?.success) {
			onOpenChange(false);
		}
	};

	const handleCancel = () => {
		resetForm();
		onOpenChange(false);
	};

	const handleTabChange = (tabId: string) => {
		setCurrentTab(tabId);
	};

	const handleFieldValueChange = (
		questionId: string,
		value: string | boolean | number,
	) => {
		setFieldValue(questionId, value);
	};

	if (!schema || schema.tabs.length === 0) {
		return (
			<Dialog open={open} onOpenChange={onOpenChange}>
				<DialogContent className="sm:max-w-[600px]">
					<DialogHeader>
						<DialogTitle>0-Project Configuration</DialogTitle>
					</DialogHeader>
					<div className="py-6 text-center text-muted-foreground">
						Loading configuration schema...
					</div>
				</DialogContent>
			</Dialog>
		);
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
				<DialogHeader className="space-y-2">
					<div className="flex items-center justify-between">
						<DialogTitle>0-Project Configuration</DialogTitle>
						{isDirty && <Badge variant="outline">Unsaved changes</Badge>}
					</div>
					{schema.version && (
						<p className="text-sm text-muted-foreground">
							Schema v{schema.version}
						</p>
					)}
				</DialogHeader>

				<Tabs value={currentTab} onValueChange={handleTabChange}>
					<TabsList className="w-full justify-start">
						{schema.tabs.map((tab) => (
							<TabsTrigger key={tab.id} value={tab.id}>
								{tab.label}
							</TabsTrigger>
						))}
					</TabsList>

					{schema.tabs.map((tab) => (
						<TabsContent key={tab.id} value={tab.id} className="mt-4">
							<ConfigForm
								tab={tab}
								formValues={useConfigStore.getState().formValues}
								errors={errors}
								isSaving={isSaving}
								onChange={handleFieldValueChange}
								onSave={handleSave}
								onCancel={handleCancel}
							/>
						</TabsContent>
					))}
				</Tabs>
			</DialogContent>
		</Dialog>
	);
}
