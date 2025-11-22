# IoT Optimization - Battery, Network & Edge Performance

**Version**: 4.0.0 (2025-11-22)
**Status**: Production Ready

---

## Battery Optimization Strategies

### Power Consumption Monitoring

```python
import time
from typing import Dict, List

class BatteryMonitor:
    """Monitor and optimize battery consumption on IoT devices."""

    def __init__(self, battery_capacity_mah: int):
        self.capacity = battery_capacity_mah
        self.consumption_log = []

    def log_operation(self, operation: str, duration_sec: float, current_ma: float):
        """Log power consumption for each operation."""
        energy_consumed = (current_ma * duration_sec) / 3600  # mAh
        self.consumption_log.append({
            'operation': operation,
            'duration_sec': duration_sec,
            'current_ma': current_ma,
            'energy_mah': energy_consumed
        })

    def estimate_battery_life(self) -> Dict:
        """Estimate battery life based on usage patterns."""
        daily_operations = self._get_daily_cycle()
        daily_consumption = sum(op['energy_mah'] for op in daily_operations)

        return {
            'daily_consumption_mah': daily_consumption,
            'estimated_days': self.capacity / daily_consumption,
            'top_consumers': self._get_top_consumers()
        }

    def _get_daily_cycle(self) -> List[Dict]:
        """Calculate average daily operation cycle."""
        # Example: sensor reading every 5 min = 288 times/day
        operations = [
            {'operation': 'sensor_read', 'frequency': 288, 'energy': 2},
            {'operation': 'mqtt_publish', 'frequency': 12, 'energy': 50},
            {'operation': 'sleep', 'frequency': 1, 'energy': 0.5}
        ]
        return operations

    def _get_top_consumers(self) -> List[tuple]:
        """Identify top power-consuming operations."""
        total_by_op = {}
        for log in self.consumption_log:
            op = log['operation']
            total_by_op[op] = total_by_op.get(op, 0) + log['energy_mah']

        return sorted(total_by_op.items(), key=lambda x: x[1], reverse=True)[:5]
```

### Sleep Mode Optimization

```python
import machine  # MicroPython

class SleepOptimizer:
    """Optimize device sleep cycles for minimal power drain."""

    def __init__(self, wake_interval_seconds: int = 300):  # 5 minutes
        self.wake_interval = wake_interval_seconds

    def configure_deep_sleep(self):
        """Configure ultra-low power sleep mode."""
        # Disable unnecessary components before sleep
        self.disable_wifi()
        self.disable_bluetooth()
        self.disable_peripherals()

        # Set wake timer
        machine.sleep(self.wake_interval * 1000)  # Convert to milliseconds

    def wake_and_transmit(self):
        """Wake, collect data, transmit, sleep."""
        # 1. Wake from sleep
        data = self.read_sensors()

        # 2. Connect to network (most power-consuming operation)
        self.connect_wifi()
        self.publish_mqtt(data)

        # 3. Disconnect before sleep
        self.disconnect_wifi()

        # 4. Return to sleep
        self.configure_deep_sleep()

    def calculate_duty_cycle(self) -> float:
        """Calculate device duty cycle (% active time)."""
        active_time_sec = 10  # Estimated active time
        cycle_time_sec = self.wake_interval + active_time_sec

        return (active_time_sec / cycle_time_sec) * 100

    # Stub methods
    def disable_wifi(self): pass
    def disable_bluetooth(self): pass
    def disable_peripherals(self): pass
    def read_sensors(self): return {}
    def connect_wifi(self): pass
    def publish_mqtt(self, data): pass
    def disconnect_wifi(self): pass
```

---

## Network Optimization

### Bandwidth Minimization Techniques

```python
import zlib
import struct

class BandwidthOptimizer:
    """Minimize bandwidth usage for battery and cost efficiency."""

    @staticmethod
    def compress_payload(data: bytes) -> bytes:
        """Compress data before transmission."""
        # Use zlib with maximum compression
        compressed = zlib.compress(data, level=9)
        compression_ratio = len(compressed) / len(data)

        if compression_ratio > 0.9:  # Not worth compressing
            return data

        return b'C' + compressed  # Prepend compression flag

    @staticmethod
    def batch_sensor_readings(readings: list, batch_size: int = 10) -> list:
        """Batch multiple sensor readings into single MQTT message."""
        batches = []
        for i in range(0, len(readings), batch_size):
            batch = readings[i:i+batch_size]
            batches.append({
                'timestamp': batch[0]['timestamp'],
                'count': len(batch),
                'readings': batch
            })
        return batches

    @staticmethod
    def delta_encoding(current_value: float, previous_value: float) -> bytes:
        """Use delta encoding to reduce data size (only send changes)."""
        delta = current_value - previous_value

        # Use variable-length encoding
        if -128 <= delta <= 127:
            return struct.pack('b', int(delta))  # 1 byte
        elif -32768 <= delta <= 32767:
            return struct.pack('>h', int(delta))  # 2 bytes
        else:
            return struct.pack('>f', delta)  # 4 bytes

    @staticmethod
    def adaptive_payload_size(network_quality: float, max_size: int = 256) -> int:
        """Adapt payload size based on network quality."""
        # If signal is poor, send smaller payloads
        if network_quality < 0.3:
            return max_size // 4
        elif network_quality < 0.6:
            return max_size // 2
        else:
            return max_size
```

### Connection Pooling for IoT

```python
import socket
from typing import Optional

class IoTConnectionPool:
    """Reuse connections to reduce connection overhead."""

    def __init__(self, broker_host: str, broker_port: int, pool_size: int = 5):
        self.host = broker_host
        self.port = broker_port
        self.pool = []
        self.pool_size = pool_size
        self.active_connections = 0

    def get_connection(self) -> socket.socket:
        """Get connection from pool or create new one."""
        if self.pool:
            return self.pool.pop()

        if self.active_connections < self.pool_size:
            conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            conn.connect((self.host, self.port))
            self.active_connections += 1
            return conn

        # Wait for available connection
        import time
        while not self.pool:
            time.sleep(0.1)

        return self.pool.pop()

    def return_connection(self, conn: socket.socket):
        """Return connection to pool for reuse."""
        self.pool.append(conn)

    def close_all(self):
        """Close all connections in pool."""
        for conn in self.pool:
            conn.close()
        self.active_connections = 0
        self.pool = []
```

---

## Edge Processing Optimization

### Lightweight Data Filtering at Edge

```python
class EdgeDataFilter:
    """Filter irrelevant data at edge to reduce cloud transmission."""

    def __init__(self, threshold_percent: int = 5):
        self.threshold = threshold_percent / 100
        self.last_value = None

    def should_transmit(self, value: float) -> bool:
        """Implement dead-band filtering (only send if change > threshold)."""
        if self.last_value is None:
            self.last_value = value
            return True

        percent_change = abs(value - self.last_value) / self.last_value
        if percent_change > self.threshold:
            self.last_value = value
            return True

        return False

    def filter_outliers(self, readings: list) -> list:
        """Remove outliers using IQR method."""
        import statistics

        if len(readings) < 4:
            return readings

        sorted_readings = sorted(readings)
        q1 = sorted_readings[len(readings)//4]
        q3 = sorted_readings[3*len(readings)//4]
        iqr = q3 - q1

        lower_bound = q1 - 1.5 * iqr
        upper_bound = q3 + 1.5 * iqr

        return [r for r in readings if lower_bound <= r <= upper_bound]
```

### Memory-Efficient Data Structures

```python
import array

class MemoryOptimizedSensor:
    """Use memory-efficient data structures on constrained devices."""

    def __init__(self, max_readings: int = 100):
        # Use array module instead of list (4x memory savings)
        self.readings = array.array('f')  # Float array
        self.max_readings = max_readings
        self.timestamps = array.array('I')  # Unsigned int for timestamps

    def add_reading(self, value: float, timestamp: int):
        """Add reading with automatic circular buffer behavior."""
        if len(self.readings) >= self.max_readings:
            # Remove oldest reading (circular buffer)
            self.readings.pop(0)
            self.timestamps.pop(0)

        self.readings.append(value)
        self.timestamps.append(timestamp)

    def get_statistics(self) -> dict:
        """Calculate statistics without creating new arrays."""
        if not self.readings:
            return {}

        n = len(self.readings)
        sum_val = sum(self.readings)
        mean = sum_val / n
        variance = sum((x - mean) ** 2 for x in self.readings) / n

        return {
            'count': n,
            'mean': mean,
            'min': min(self.readings),
            'max': max(self.readings),
            'std_dev': variance ** 0.5
        }

    def get_memory_usage_bytes(self) -> int:
        """Calculate memory usage (for monitoring)."""
        return len(self.readings) * 4 + len(self.timestamps) * 4  # Float=4, Int=4
```

---

## Network Resilience

### Retry Strategy with Exponential Backoff

```python
import time
import random

class ResilientNetworkClient:
    """Handle network failures gracefully with exponential backoff."""

    def __init__(self, max_retries: int = 5, base_delay: float = 1.0):
        self.max_retries = max_retries
        self.base_delay = base_delay

    def send_with_retry(self, data: bytes) -> bool:
        """Send data with automatic retry and backoff."""
        for attempt in range(self.max_retries):
            try:
                self._send_data(data)
                return True
            except (ConnectionError, TimeoutError) as e:
                if attempt == self.max_retries - 1:
                    print(f"Failed after {self.max_retries} attempts")
                    return False

                # Exponential backoff with jitter
                wait_time = self.base_delay * (2 ** attempt)
                jitter = random.uniform(0, wait_time * 0.1)
                total_wait = min(wait_time + jitter, 300)  # Max 5 minutes

                print(f"Attempt {attempt+1} failed, waiting {total_wait:.1f}s")
                time.sleep(total_wait)

        return False

    def _send_data(self, data: bytes):
        """Actual send operation (stub)."""
        pass
```

---

## Best Practices

**DO**:
- Monitor battery consumption regularly
- Use deep sleep between transmissions
- Batch data to reduce connection overhead
- Implement edge-side filtering
- Use compression for large payloads
- Monitor signal quality and adapt behavior
- Test on real hardware with real batteries

**DON'T**:
- Poll sensors too frequently
- Connect continuously (battery drain)
- Send uncompressed data
- Ignore network timeouts
- Skip outlier detection
- Assume stable network conditions
- Over-engineer for cloud processing

---

**Related Skills**: moai-domain-backend, moai-essentials-perf
