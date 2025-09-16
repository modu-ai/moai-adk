#!/usr/bin/env python3
"""
MoAI-ADK Performance Optimization Toolkit
Implements lazy loading, plugin architecture, and async operations

This script provides automated tools for implementing performance
optimizations identified in the comprehensive analysis.
"""

import asyncio
import importlib
import importlib.util
import time
import weakref
from pathlib import Path
from typing import Dict, List, Any, Optional, Callable, TypeVar, Generic
from dataclasses import dataclass, field
from collections import defaultdict
from functools import wraps, lru_cache
import logging
import threading
import psutil
from concurrent.futures import ThreadPoolExecutor
import json


T = TypeVar('T')


@dataclass
class LazyModule:
    """Represents a lazily loaded module"""
    name: str
    module_path: str
    load_time: Optional[float] = None
    memory_usage: Optional[float] = None
    _module: Optional[Any] = field(default=None, init=False)


@dataclass
class PluginInfo:
    """Plugin information and metadata"""
    name: str
    version: str
    description: str
    author: str
    hooks: List[str]
    dependencies: List[str] = field(default_factory=list)
    enabled: bool = True
    priority: int = 100
    _instance: Optional[Any] = field(default=None, init=False)


@dataclass
class PerformanceMetrics:
    """Performance monitoring metrics"""
    operation_name: str
    start_time: float
    end_time: Optional[float] = None
    memory_start: float = field(default_factory=lambda: psutil.Process().memory_info().rss)
    memory_end: Optional[float] = None
    cpu_percent: Optional[float] = None

    @property
    def duration_ms(self) -> Optional[float]:
        if self.end_time:
            return (self.end_time - self.start_time) * 1000
        return None

    @property
    def memory_delta_mb(self) -> Optional[float]:
        if self.memory_end:
            return (self.memory_end - self.memory_start) / 1024 / 1024
        return None


class LazyLoadManager:
    """Manages lazy loading of modules and components"""

    def __init__(self):
        self.modules: Dict[str, LazyModule] = {}
        self.load_stats = {}
        self._lock = threading.RLock()

    def register_module(self, name: str, module_path: str) -> LazyModule:
        """Register a module for lazy loading"""
        with self._lock:
            lazy_module = LazyModule(name=name, module_path=module_path)
            self.modules[name] = lazy_module
            return lazy_module

    def get_module(self, name: str) -> Any:
        """Get a module, loading it if necessary"""
        with self._lock:
            if name not in self.modules:
                raise ValueError(f"Module '{name}' not registered")

            lazy_module = self.modules[name]

            if lazy_module._module is None:
                start_time = time.perf_counter()
                start_memory = psutil.Process().memory_info().rss

                try:
                    if '.' in lazy_module.module_path:
                        # Import from package
                        lazy_module._module = importlib.import_module(lazy_module.module_path)
                    else:
                        # Import from file path
                        spec = importlib.util.spec_from_file_location(name, lazy_module.module_path)
                        module = importlib.util.module_from_spec(spec)
                        spec.loader.exec_module(module)
                        lazy_module._module = module

                    end_time = time.perf_counter()
                    end_memory = psutil.Process().memory_info().rss

                    lazy_module.load_time = (end_time - start_time) * 1000  # ms
                    lazy_module.memory_usage = (end_memory - start_memory) / 1024 / 1024  # MB

                    self.load_stats[name] = {
                        'load_time_ms': lazy_module.load_time,
                        'memory_usage_mb': lazy_module.memory_usage,
                        'loaded_at': time.time()
                    }

                    logging.debug(f"Lazy loaded {name}: {lazy_module.load_time:.2f}ms, {lazy_module.memory_usage:.2f}MB")

                except Exception as e:
                    logging.error(f"Failed to lazy load {name}: {e}")
                    raise

            return lazy_module._module

    def preload_modules(self, module_names: List[str]) -> Dict[str, bool]:
        """Preload specified modules in background"""
        results = {}

        def load_module(name: str) -> bool:
            try:
                self.get_module(name)
                return True
            except Exception as e:
                logging.error(f"Failed to preload {name}: {e}")
                return False

        with ThreadPoolExecutor(max_workers=4) as executor:
            future_to_name = {executor.submit(load_module, name): name for name in module_names}
            for future in future_to_name:
                name = future_to_name[future]
                results[name] = future.result()

        return results

    def get_load_statistics(self) -> Dict[str, Any]:
        """Get loading statistics for all modules"""
        return {
            'modules': dict(self.load_stats),
            'total_modules': len(self.modules),
            'loaded_modules': len([m for m in self.modules.values() if m._module is not None]),
            'average_load_time_ms': sum(stats['load_time_ms'] for stats in self.load_stats.values()) / max(len(self.load_stats), 1),
            'total_memory_usage_mb': sum(stats['memory_usage_mb'] for stats in self.load_stats.values())
        }


class PluginManager:
    """Extensible plugin system"""

    def __init__(self):
        self.plugins: Dict[str, PluginInfo] = {}
        self.hooks: Dict[str, List[PluginInfo]] = defaultdict(list)
        self._instances: Dict[str, Any] = {}
        self.lazy_loader = LazyLoadManager()

    def register_plugin(self, plugin_info: PluginInfo, module_path: str) -> bool:
        """Register a plugin"""
        try:
            # Validate plugin
            if plugin_info.name in self.plugins:
                logging.warning(f"Plugin {plugin_info.name} already registered")
                return False

            # Register for lazy loading
            self.lazy_loader.register_module(plugin_info.name, module_path)

            # Store plugin info
            self.plugins[plugin_info.name] = plugin_info

            # Register hooks
            for hook in plugin_info.hooks:
                self.hooks[hook].append(plugin_info)

            # Sort hooks by priority
            self.hooks[hook].sort(key=lambda p: p.priority)

            logging.info(f"Registered plugin: {plugin_info.name} v{plugin_info.version}")
            return True

        except Exception as e:
            logging.error(f"Failed to register plugin {plugin_info.name}: {e}")
            return False

    def get_plugin(self, name: str) -> Any:
        """Get a plugin instance"""
        if name not in self.plugins:
            raise ValueError(f"Plugin '{name}' not registered")

        plugin_info = self.plugins[name]

        if not plugin_info.enabled:
            raise ValueError(f"Plugin '{name}' is disabled")

        if name not in self._instances:
            # Load plugin module and create instance
            module = self.lazy_loader.get_module(name)

            # Look for plugin class
            plugin_class = getattr(module, 'Plugin', None) or getattr(module, f'{name.title()}Plugin', None)

            if plugin_class is None:
                raise ValueError(f"Plugin class not found in {name}")

            self._instances[name] = plugin_class()
            plugin_info._instance = self._instances[name]

        return self._instances[name]

    def call_hook(self, hook_name: str, *args, **kwargs) -> List[Any]:
        """Call all plugins registered for a hook"""
        results = []

        if hook_name not in self.hooks:
            return results

        for plugin_info in self.hooks[hook_name]:
            if not plugin_info.enabled:
                continue

            try:
                plugin = self.get_plugin(plugin_info.name)

                if hasattr(plugin, hook_name):
                    result = getattr(plugin, hook_name)(*args, **kwargs)
                    results.append(result)

            except Exception as e:
                logging.error(f"Plugin {plugin_info.name} failed in hook {hook_name}: {e}")

        return results

    def enable_plugin(self, name: str) -> bool:
        """Enable a plugin"""
        if name in self.plugins:
            self.plugins[name].enabled = True
            return True
        return False

    def disable_plugin(self, name: str) -> bool:
        """Disable a plugin"""
        if name in self.plugins:
            self.plugins[name].enabled = False
            # Remove instance to free memory
            if name in self._instances:
                del self._instances[name]
                self.plugins[name]._instance = None
            return True
        return False

    def list_plugins(self) -> Dict[str, Dict[str, Any]]:
        """List all registered plugins"""
        return {
            name: {
                'version': info.version,
                'description': info.description,
                'author': info.author,
                'hooks': info.hooks,
                'enabled': info.enabled,
                'loaded': name in self._instances
            }
            for name, info in self.plugins.items()
        }


class AsyncOperationManager:
    """Manages async operations for I/O intensive tasks"""

    def __init__(self, max_workers: int = 4):
        self.max_workers = max_workers
        self.executor = ThreadPoolExecutor(max_workers=max_workers)

    async def parallel_file_operations(self, operations: List[Callable]) -> List[Any]:
        """Execute file operations in parallel"""
        loop = asyncio.get_event_loop()

        tasks = []
        for operation in operations:
            task = loop.run_in_executor(self.executor, operation)
            tasks.append(task)

        return await asyncio.gather(*tasks, return_exceptions=True)

    async def batch_process_files(self, file_paths: List[Path], processor: Callable[[Path], Any]) -> Dict[str, Any]:
        """Process multiple files in parallel"""
        def process_file(path: Path) -> tuple[str, Any]:
            try:
                result = processor(path)
                return str(path), result
            except Exception as e:
                return str(path), {'error': str(e)}

        operations = [lambda p=path: process_file(p) for path in file_paths]
        results = await self.parallel_file_operations(operations)

        return {path: result for path, result in results if not isinstance(result, Exception)}

    async def async_import_analysis(self, file_paths: List[Path]) -> Dict[str, Dict]:
        """Analyze imports in multiple files asynchronously"""
        from migration_toolkit import ImportAnalyzer

        analyzer = ImportAnalyzer()

        def analyze_file(path: Path) -> tuple[str, Dict]:
            result = analyzer.analyze_file_imports(path)
            return str(path), result

        operations = [lambda p=path: analyze_file(p) for path in file_paths]
        results = await self.parallel_file_operations(operations)

        return {path: result for path, result in results if not isinstance(result, Exception)}

    def shutdown(self):
        """Shutdown the executor"""
        self.executor.shutdown(wait=True)


class ResourceManager:
    """Manages memory usage and resource cleanup"""

    def __init__(self, memory_threshold_mb: int = 100):
        self.memory_threshold = memory_threshold_mb * 1024 * 1024  # Convert to bytes
        self.resource_registry: Dict[str, weakref.ref] = {}
        self.cleanup_callbacks: Dict[str, Callable] = {}
        self._monitoring = False

    def register_resource(self, name: str, resource: Any, cleanup_callback: Optional[Callable] = None):
        """Register a resource for monitoring"""
        self.resource_registry[name] = weakref.ref(resource)
        if cleanup_callback:
            self.cleanup_callbacks[name] = cleanup_callback

    def start_monitoring(self):
        """Start memory monitoring"""
        self._monitoring = True
        threading.Thread(target=self._monitor_memory, daemon=True).start()

    def stop_monitoring(self):
        """Stop memory monitoring"""
        self._monitoring = False

    def _monitor_memory(self):
        """Monitor memory usage and trigger cleanup if needed"""
        while self._monitoring:
            try:
                current_memory = psutil.Process().memory_info().rss

                if current_memory > self.memory_threshold:
                    logging.warning(f"Memory usage high: {current_memory / 1024 / 1024:.2f}MB")
                    self.cleanup_resources()

                time.sleep(5)  # Check every 5 seconds

            except Exception as e:
                logging.error(f"Memory monitoring error: {e}")

    def cleanup_resources(self):
        """Clean up unused resources"""
        cleaned = 0

        for name, resource_ref in list(self.resource_registry.items()):
            if resource_ref() is None:  # Resource was garbage collected
                # Run cleanup callback if available
                if name in self.cleanup_callbacks:
                    try:
                        self.cleanup_callbacks[name]()
                        cleaned += 1
                    except Exception as e:
                        logging.error(f"Cleanup callback failed for {name}: {e}")

                # Remove from registry
                del self.resource_registry[name]
                if name in self.cleanup_callbacks:
                    del self.cleanup_callbacks[name]

        if cleaned > 0:
            logging.info(f"Cleaned up {cleaned} resources")

    def get_memory_usage(self) -> Dict[str, float]:
        """Get current memory usage statistics"""
        process = psutil.Process()
        memory_info = process.memory_info()

        return {
            'rss_mb': memory_info.rss / 1024 / 1024,
            'vms_mb': memory_info.vms / 1024 / 1024,
            'percent': process.memory_percent(),
            'available_mb': psutil.virtual_memory().available / 1024 / 1024
        }


class PerformanceMonitor:
    """Comprehensive performance monitoring"""

    def __init__(self):
        self.metrics: List[PerformanceMetrics] = []
        self.benchmarks: Dict[str, List[float]] = defaultdict(list)

    def start_operation(self, name: str) -> PerformanceMetrics:
        """Start monitoring an operation"""
        metric = PerformanceMetrics(
            operation_name=name,
            start_time=time.perf_counter()
        )
        self.metrics.append(metric)
        return metric

    def end_operation(self, metric: PerformanceMetrics) -> PerformanceMetrics:
        """End monitoring an operation"""
        metric.end_time = time.perf_counter()
        metric.memory_end = psutil.Process().memory_info().rss
        metric.cpu_percent = psutil.Process().cpu_percent()

        # Add to benchmarks
        if metric.duration_ms:
            self.benchmarks[metric.operation_name].append(metric.duration_ms)

        return metric

    def benchmark_operation(self, name: str, operation: Callable, iterations: int = 10) -> Dict[str, Any]:
        """Benchmark an operation multiple times"""
        durations = []
        memory_deltas = []

        for _ in range(iterations):
            metric = self.start_operation(f"{name}_benchmark")

            try:
                operation()
            except Exception as e:
                logging.error(f"Benchmark operation {name} failed: {e}")
                continue

            self.end_operation(metric)

            if metric.duration_ms:
                durations.append(metric.duration_ms)
            if metric.memory_delta_mb:
                memory_deltas.append(metric.memory_delta_mb)

        return {
            'operation': name,
            'iterations': len(durations),
            'avg_duration_ms': sum(durations) / len(durations) if durations else 0,
            'min_duration_ms': min(durations) if durations else 0,
            'max_duration_ms': max(durations) if durations else 0,
            'avg_memory_delta_mb': sum(memory_deltas) / len(memory_deltas) if memory_deltas else 0,
            'total_memory_delta_mb': sum(memory_deltas) if memory_deltas else 0
        }

    def get_performance_report(self) -> Dict[str, Any]:
        """Generate comprehensive performance report"""
        completed_metrics = [m for m in self.metrics if m.end_time is not None]

        if not completed_metrics:
            return {'error': 'No completed operations to report'}

        operation_stats = defaultdict(list)
        for metric in completed_metrics:
            operation_stats[metric.operation_name].append(metric)

        report = {
            'summary': {
                'total_operations': len(completed_metrics),
                'unique_operations': len(operation_stats),
                'avg_duration_ms': sum(m.duration_ms for m in completed_metrics if m.duration_ms) / len(completed_metrics),
                'total_memory_delta_mb': sum(m.memory_delta_mb for m in completed_metrics if m.memory_delta_mb)
            },
            'operations': {}
        }

        for op_name, metrics in operation_stats.items():
            durations = [m.duration_ms for m in metrics if m.duration_ms]
            memory_deltas = [m.memory_delta_mb for m in metrics if m.memory_delta_mb]

            report['operations'][op_name] = {
                'count': len(metrics),
                'avg_duration_ms': sum(durations) / len(durations) if durations else 0,
                'min_duration_ms': min(durations) if durations else 0,
                'max_duration_ms': max(durations) if durations else 0,
                'avg_memory_delta_mb': sum(memory_deltas) / len(memory_deltas) if memory_deltas else 0
            }

        return report


def lazy_property(func):
    """Decorator for lazy properties"""
    attr_name = '_' + func.__name__

    @property
    @wraps(func)
    def wrapper(self):
        if not hasattr(self, attr_name):
            setattr(self, attr_name, func(self))
        return getattr(self, attr_name)

    return wrapper


def performance_monitor(name: str = None):
    """Decorator for performance monitoring"""
    def decorator(func):
        operation_name = name or func.__name__

        @wraps(func)
        def wrapper(*args, **kwargs):
            monitor = PerformanceMonitor()
            metric = monitor.start_operation(operation_name)

            try:
                result = func(*args, **kwargs)
                return result
            finally:
                monitor.end_operation(metric)
                logging.debug(f"Operation {operation_name}: {metric.duration_ms:.2f}ms")

        return wrapper
    return decorator


class OptimizedMoAIADK:
    """Example of optimized MoAI-ADK with performance enhancements"""

    def __init__(self):
        self.lazy_loader = LazyLoadManager()
        self.plugin_manager = PluginManager()
        self.async_manager = AsyncOperationManager()
        self.resource_manager = ResourceManager()
        self.performance_monitor = PerformanceMonitor()

        # Register core modules for lazy loading
        self._setup_lazy_modules()

        # Start resource monitoring
        self.resource_manager.start_monitoring()

    def _setup_lazy_modules(self):
        """Setup lazy loading for core modules"""
        modules = {
            'config': 'moai_adk.config.settings',
            'logger': 'moai_adk.utils.logger',
            'installer': 'moai_adk.install.installer',
            'validator': 'moai_adk.core.validation.validator',
            'security': 'moai_adk.core.security.validator'
        }

        for name, path in modules.items():
            self.lazy_loader.register_module(name, path)

    @lazy_property
    def config(self):
        """Lazy loaded configuration"""
        return self.lazy_loader.get_module('config')

    @lazy_property
    def logger(self):
        """Lazy loaded logger"""
        return self.lazy_loader.get_module('logger')

    @performance_monitor('initialization')
    def initialize(self, config_path: Optional[str] = None):
        """Initialize MoAI-ADK with performance monitoring"""
        # Preload critical modules in background
        critical_modules = ['config', 'logger', 'security']
        self.lazy_loader.preload_modules(critical_modules)

        # Call plugin hooks
        self.plugin_manager.call_hook('on_initialize', config_path=config_path)

    async def install_async(self, project_path: Path, options: Dict[str, Any]) -> Dict[str, Any]:
        """Async installation with parallel operations"""
        installer = self.lazy_loader.get_module('installer')

        # Break down installation into parallel tasks
        tasks = [
            self._validate_environment(project_path),
            self._prepare_resources(options),
            self._check_dependencies()
        ]

        results = await asyncio.gather(*tasks, return_exceptions=True)

        # Execute main installation
        result = installer.install(project_path, options)

        # Call plugin hooks
        self.plugin_manager.call_hook('on_install_complete', result=result)

        return result

    async def _validate_environment(self, project_path: Path) -> Dict[str, bool]:
        """Validate environment asynchronously"""
        validator = self.lazy_loader.get_module('validator')
        # Simulate async validation
        await asyncio.sleep(0.1)
        return validator.validate_environment(project_path)

    async def _prepare_resources(self, options: Dict[str, Any]) -> bool:
        """Prepare resources asynchronously"""
        # Simulate async resource preparation
        await asyncio.sleep(0.1)
        return True

    async def _check_dependencies(self) -> Dict[str, str]:
        """Check dependencies asynchronously"""
        # Simulate async dependency check
        await asyncio.sleep(0.1)
        return {'status': 'ok'}

    def shutdown(self):
        """Clean shutdown of all components"""
        self.resource_manager.stop_monitoring()
        self.async_manager.shutdown()
        self.resource_manager.cleanup_resources()


def main():
    """Demonstration of optimization features"""
    logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

    # Create optimized MoAI-ADK instance
    moai = OptimizedMoAIADK()

    # Demonstrate lazy loading
    print("=== Lazy Loading Demo ===")
    start_time = time.perf_counter()
    config = moai.config
    load_time = (time.perf_counter() - start_time) * 1000
    print(f"Config loaded in {load_time:.2f}ms")

    # Show loading statistics
    stats = moai.lazy_loader.get_load_statistics()
    print(f"Lazy loading stats: {json.dumps(stats, indent=2)}")

    # Demonstrate plugin system
    print("\n=== Plugin System Demo ===")
    plugin_info = PluginInfo(
        name="sample_plugin",
        version="1.0.0",
        description="Sample plugin for demo",
        author="MoAI Team",
        hooks=["on_initialize", "on_install_complete"]
    )
    # Note: Would need actual plugin file for full demo
    print(f"Plugin registry: {moai.plugin_manager.list_plugins()}")

    # Demonstrate performance monitoring
    print("\n=== Performance Monitoring Demo ===")
    @performance_monitor('demo_operation')
    def demo_operation():
        time.sleep(0.1)  # Simulate work
        return "completed"

    result = demo_operation()
    print(f"Demo operation result: {result}")

    # Show performance report
    report = moai.performance_monitor.get_performance_report()
    print(f"Performance report: {json.dumps(report, indent=2)}")

    # Memory usage report
    memory_usage = moai.resource_manager.get_memory_usage()
    print(f"Memory usage: {json.dumps(memory_usage, indent=2)}")

    # Clean shutdown
    moai.shutdown()
    print("\n✅ Optimization demo completed successfully!")


if __name__ == "__main__":
    main()