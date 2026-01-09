import { Menu, Moon, Sun, Settings } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui";
import { useUIStore, useProviderStore } from "@/stores";
import { cn } from "@/lib/utils";
import { ConfigDialog } from "@/components/config";

interface HeaderProps {
	isConnected: boolean;
}

export function Header({ isConnected }: HeaderProps) {
	const { toggleSidebar, theme, setTheme } = useUIStore();
	const { activeProvider, activeModel } = useProviderStore();
	const [configOpen, setConfigOpen] = useState(false);

	const toggleTheme = () => {
		if (theme === "dark") {
			setTheme("light");
		} else {
			setTheme("dark");
		}
	};

	return (
		<>
			<header className="flex-shrink-0 h-14 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
				<div className="flex h-full items-center px-4 gap-4">
					<Button
						variant="ghost"
						size="icon"
						onClick={toggleSidebar}
						aria-label="Toggle sidebar"
						className="shrink-0"
					>
						<Menu className="h-5 w-5" />
					</Button>

					<div className="flex items-center gap-2 shrink-0">
						<div className="flex items-center gap-1.5">
							<div className="w-7 h-7 rounded-lg bg-primary flex items-center justify-center">
								<span className="text-primary-foreground text-xs font-bold">
									M
								</span>
							</div>
							<span className="font-semibold">MoAI-ADK</span>
						</div>
					</div>

					<div className="flex-1 min-w-0" />

					<div className="flex items-center gap-2">
						{/* Connection Status */}
						<div
							className={cn(
								"flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-colors",
								isConnected
									? "bg-green-500/10 text-green-600 dark:text-green-400"
									: "bg-red-500/10 text-red-600 dark:text-red-400",
							)}
						>
							<span
								className={cn(
									"w-1.5 h-1.5 rounded-full animate-pulse",
									isConnected
										? "bg-green-600 dark:bg-green-400"
										: "bg-red-600 dark:bg-red-400",
								)}
							/>
							<span className="hidden sm:inline">
								{isConnected ? "Connected" : "Disconnected"}
							</span>
						</div>

						{/* Provider Info */}
						{activeProvider && activeModel && (
							<div className="hidden md:flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-muted/50 text-xs">
								<span className="capitalize font-medium">{activeProvider}</span>
								<span className="text-muted-foreground">/</span>
								<span className="text-muted-foreground truncate max-w-[100px]">
									{activeModel}
								</span>
							</div>
						)}

						{/* Theme Toggle */}
						<Button
							variant="ghost"
							size="icon"
							onClick={toggleTheme}
							aria-label="Toggle theme"
							className="shrink-0"
						>
							{theme === "dark" ? (
								<Sun className="h-4.5 w-4.5" />
							) : (
								<Moon className="h-4.5 w-4.5" />
							)}
						</Button>

						{/* Settings */}
						<Button
							variant="ghost"
							size="icon"
							onClick={() => setConfigOpen(true)}
							aria-label="Settings"
							className="shrink-0"
						>
							<Settings className="h-4.5 w-4.5" />
						</Button>
					</div>
				</div>
			</header>

			{/* Config Dialog */}
			<ConfigDialog open={configOpen} onOpenChange={setConfigOpen} />
		</>
	);
}
