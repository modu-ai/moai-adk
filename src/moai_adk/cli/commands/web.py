"""Web CLI Command

CLI command for starting the MoAI Web Backend server.
Provides options for port, host, and browser opening.
"""

import webbrowser

import click


@click.command()
@click.option(
    "--port",
    "-p",
    type=int,
    default=9595,
    help="API server port number (default: 9595)",
    show_default=True,
)
@click.option(
    "--host",
    "-h",
    type=str,
    default="127.0.0.1",
    help="Server host address (default: 127.0.0.1)",
    show_default=True,
)
@click.option(
    "--open/--no-open",
    default=True,
    help="Open browser automatically (default: --open)",
)
def web(port: int, host: str, open: bool) -> None:
    """Start the MoAI Web Backend server

    Launches a FastAPI server providing REST API endpoints
    and WebSocket chat interface for MoAI-ADK.

    Examples:

        moai web                    # Start API on 9595, open browser to 9005

        moai web --port 8080        # Start API on port 8080

        moai web --no-open          # Start without opening browser
    """
    start_server(host=host, port=port, open_browser=open)


def start_server(
    host: str = "127.0.0.1",
    port: int = 9595,
    frontend_port: int = 9005,
    open_browser: bool = True,
    reload: bool = False,
) -> None:
    """Start the MoAI Web Backend server

    Args:
        host: Server host address
        port: API server port number
        frontend_port: Frontend dev server port number
        open_browser: Whether to open browser automatically
        reload: Enable auto-reload for development
    """
    try:
        import uvicorn
    except ImportError:
        click.echo(
            click.style(
                "Error: Web dependencies not installed. Run: pip install moai-adk[web]",
                fg="red",
            )
        )
        raise SystemExit(1)

    # Display startup message
    click.echo()
    click.echo(click.style("🗿 MoAI Web Dashboard", fg="cyan", bold=True))
    click.echo(click.style("=" * 44, fg="cyan"))
    click.echo(f"  API Server: {click.style(f'http://{host}:{port}', fg='green')}")
    click.echo(f"  Web UI:     {click.style(f'http://{host}:{frontend_port}', fg='yellow')}")
    click.echo(click.style("=" * 44, fg="cyan"))
    click.echo()
    click.echo(
        click.style("  Run frontend: ", fg="white") + click.style("cd src/moai_adk/web-ui && npm run dev", fg="cyan")
    )
    click.echo()
    click.echo("Press Ctrl+C to stop the server")
    click.echo()

    # Open browser if requested (opens Web UI frontend)
    if open_browser:
        import threading

        def _open_browser():
            import time

            time.sleep(1.5)  # Wait for server to start
            webbrowser.open(f"http://{host}:{frontend_port}")

        threading.Thread(target=_open_browser, daemon=True).start()

    # Start the server
    uvicorn.run(
        "moai_adk.web.server:create_app",
        host=host,
        port=port,
        reload=reload,
        factory=True,
        log_level="info",
    )
