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
    default=8080,
    help="Server port number (default: 8080)",
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

        moai web                    # Start on default port 8080

        moai web --port 3000        # Start on port 3000

        moai web --no-open          # Start without opening browser
    """
    start_server(host=host, port=port, open_browser=open)


def start_server(
    host: str = "127.0.0.1",
    port: int = 8080,
    open_browser: bool = True,
    reload: bool = False,
) -> None:
    """Start the MoAI Web Backend server

    Args:
        host: Server host address
        port: Server port number
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

    # Import here to avoid loading FastAPI if web deps not installed
    from moai_adk.web.config import WebConfig

    # Create configuration
    config = WebConfig(host=host, port=port)

    # Display startup message
    click.echo()
    click.echo(click.style("MoAI Web Backend", fg="cyan", bold=True))
    click.echo(click.style("=" * 40, fg="cyan"))
    click.echo(f"  Host: {click.style(host, fg='green')}")
    click.echo(f"  Port: {click.style(str(port), fg='green')}")
    click.echo(f"  URL:  {click.style(f'http://{host}:{port}', fg='yellow')}")
    click.echo(click.style("=" * 40, fg="cyan"))
    click.echo()
    click.echo("Press Ctrl+C to stop the server")
    click.echo()

    # Open browser if requested
    if open_browser:
        import threading

        def _open_browser():
            import time

            time.sleep(1.5)  # Wait for server to start
            webbrowser.open(f"http://{host}:{port}/docs")

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
