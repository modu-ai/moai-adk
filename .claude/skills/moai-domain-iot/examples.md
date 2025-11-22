# IoT Implementation Examples - 600+ Lines

**Version**: 4.0.0 (2025-11-22)

---

## 1. Basic MQTT Temperature Sensor

```python
import paho.mqtt.client as mqtt
import json
import time
from datetime import datetime

class TemperatureSensor:
    """Simple MQTT temperature sensor."""

    def __init__(self, broker='broker.hivemq.com', device_id='temp-sensor-01'):
        self.broker = broker
        self.device_id = device_id
        self.client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2)
        self.client.on_connect = self.on_connect
        self.client.on_message = self.on_message

    def on_connect(self, client, userdata, connect_flags, rc, properties):
        print(f"Connected with result code {rc}")
        client.subscribe(f"home/devices/{self.device_id}/command")

    def on_message(self, client, userdata, msg):
        """Handle incoming commands."""
        payload = json.loads(msg.payload.decode())
        if payload.get('action') == 'read':
            self.send_temperature()

    def send_temperature(self):
        """Read and publish temperature."""
        temp = 22.5  # Read from actual sensor
        payload = {
            'device_id': self.device_id,
            'temperature': temp,
            'timestamp': datetime.utcnow().isoformat(),
            'unit': 'celsius'
        }

        topic = f"home/devices/{self.device_id}/telemetry"
        self.client.publish(topic, json.dumps(payload), qos=1)

    def run(self):
        self.client.connect(self.broker, 1883, 60)
        self.client.loop_start()

        # Publish temperature every 5 minutes
        while True:
            self.send_temperature()
            time.sleep(300)
```

---

## 2. CoAP Server for Constrained Devices

```python
import asyncio
from aiocoap import resource, Context, Message, Code
import aiocoap

class SensorResource(resource.Resource):
    """CoAP resource for sensor data."""

    async def render_get(self, request):
        """Return sensor reading."""
        response_payload = b'{"sensor": "temperature", "value": 25.3}'
        return Message(code=Code.CONTENT, payload=response_payload)

    async def render_post(self, request):
        """Accept configuration."""
        command = request.payload.decode('utf-8')
        return Message(code=Code.CHANGED, payload=b'OK')

async def setup_coap_server():
    """Setup CoAP server."""
    root = resource.Site()
    root.add_child(('sensor',), SensorResource())

    context = await Context.create_server_context(root, bind=('::', 5683))
    print("CoAP server started on port 5683")

    await asyncio.Event().wait()

# Run: asyncio.run(setup_coap_server())
```

---

## 3. Device Shadow Management

```python
import json
from datetime import datetime
from typing import Dict, Any

class IoTDeviceShadow:
    """Manage device state synchronization."""

    def __init__(self, device_id: str):
        self.device_id = device_id
        self.shadow = {
            'state': {
                'desired': {'color': 'red', 'brightness': 100},
                'reported': {'color': 'green', 'brightness': 80}
            },
            'metadata': {'timestamp': datetime.utcnow().isoformat()},
            'version': 0
        }

    def update_desired(self, desired: Dict[str, Any]):
        """Cloud updates desired state."""
        self.shadow['state']['desired'].update(desired)
        self.shadow['version'] += 1
        self.calculate_delta()

    def update_reported(self, reported: Dict[str, Any]):
        """Device reports current state."""
        self.shadow['state']['reported'].update(reported)
        self.shadow['version'] += 1
        self.calculate_delta()

    def calculate_delta(self):
        """Calculate difference between desired and reported."""
        delta = {}
        desired = self.shadow['state']['desired']
        reported = self.shadow['state']['reported']

        for key, value in desired.items():
            if reported.get(key) != value:
                delta[key] = value

        self.shadow['state']['delta'] = delta
        return delta

    def get_delta_for_device(self) -> Dict:
        """Get only changes that need to be sent to device."""
        return self.shadow['state'].get('delta', {})

    def get_shadow(self) -> Dict:
        """Get complete shadow."""
        return self.shadow

# Usage example
shadow = IoTDeviceShadow('light-01')
shadow.update_desired({'brightness': 75})
delta = shadow.get_delta_for_device()
print(f"Changes to apply: {delta}")
```

---

## 4. Wireless Network Selection

```python
import time

class WirelessNetworkSelector:
    """Intelligently select between WiFi and cellular."""

    def __init__(self):
        self.wifi_available = False
        self.cellular_available = False
        self.signal_strength = {'wifi': None, 'cellular': None}

    def scan_networks(self):
        """Scan available networks."""
        self.signal_strength['wifi'] = self.check_wifi_signal()
        self.signal_strength['cellular'] = self.check_cellular_signal()

    def check_wifi_signal(self) -> float:
        """Check WiFi signal strength (RSSI -30 to -90 dBm)."""
        return -65.0  # Stub

    def check_cellular_signal(self) -> float:
        """Check cellular signal strength (RSRP -90 to -140 dBm)."""
        return -110.0  # Stub

    def select_best_network(self) -> str:
        """Select best available network."""
        self.scan_networks()

        wifi_rssi = self.signal_strength['wifi'] or -100
        cellular_rsrp = self.signal_strength['cellular'] or -150

        # WiFi prefers lower power, cellular for reliability
        if wifi_rssi > -75:  # Good WiFi signal
            return 'wifi'
        elif cellular_rsrp > -120:  # Acceptable cellular
            return 'cellular'
        else:
            return 'none'  # No good connection

    def connect_to_network(self, network_type: str) -> bool:
        """Establish connection."""
        if network_type == 'wifi':
            return self.connect_wifi()
        elif network_type == 'cellular':
            return self.connect_cellular()
        return False

    def connect_wifi(self) -> bool:
        """Connect to WiFi."""
        print("Connecting to WiFi...")
        time.sleep(2)
        return True

    def connect_cellular(self) -> bool:
        """Connect to cellular."""
        print("Connecting to cellular...")
        time.sleep(3)
        return True
```

---

## 5. Battery-Optimized Sensor Node

```python
import machine
import time
from machine import Pin, ADC

class BatteryOptimizedSensor:
    """Minimal power consumption sensor node."""

    def __init__(self):
        self.sensor = ADC(Pin(26))  # ADC pin
        self.readings = []
        self.max_readings = 50

    def read_sensor(self) -> float:
        """Read analog sensor."""
        value = self.sensor.read_u16()
        return value / 65535.0 * 3.3  # Convert to voltage

    def collect_readings(self, num_samples: int = 10):
        """Collect multiple samples for averaging."""
        self.readings = []
        for _ in range(num_samples):
            self.readings.append(self.read_sensor())
            time.sleep(0.1)

    def get_average(self) -> float:
        """Calculate average of readings."""
        if not self.readings:
            return 0
        return sum(self.readings) / len(self.readings)

    def should_transmit(self, new_value: float, threshold: float = 0.05) -> bool:
        """Implement hysteresis - only transmit if significant change."""
        if not self.readings:
            return True

        old_average = self.get_average()
        change = abs(new_value - old_average) / old_average

        return change > threshold

    def enter_sleep_mode(self, duration_seconds: int):
        """Enter sleep to conserve battery."""
        # Disable all peripherals
        rtc = machine.RTC()
        rtc.irq(trigger=rtc.ALARM0, wake=machine.DEEPSLEEP)
        machine.deepsleep(duration_seconds * 1000)

    def power_consumption_profile(self) -> dict:
        """Estimate power usage."""
        return {
            'reading_30ms': 10,        # mA
            'transmission_500ms': 80,  # mA
            'sleep': 0.01,             # mA
            'daily_energy_mah': 200    # Total
        }
```

---

## 6. Real-time IoT Dashboard Backend

```python
from fastapi import FastAPI, WebSocket
from datetime import datetime
import json

app = FastAPI()

class IoTDashboard:
    """Real-time dashboard for IoT devices."""

    def __init__(self):
        self.devices = {}
        self.connections = set()

    def register_device(self, device_id: str, device_type: str):
        """Register device."""
        self.devices[device_id] = {
            'type': device_type,
            'last_update': datetime.utcnow(),
            'status': 'online',
            'data': {}
        }

    def update_device_data(self, device_id: str, data: dict):
        """Update device telemetry."""
        if device_id in self.devices:
            self.devices[device_id]['data'] = data
            self.devices[device_id]['last_update'] = datetime.utcnow()

    async def broadcast_update(self, device_id: str, data: dict):
        """Broadcast update to all connected dashboards."""
        message = {
            'device_id': device_id,
            'data': data,
            'timestamp': datetime.utcnow().isoformat()
        }

        # Send to all WebSocket connections
        for connection in self.connections:
            await connection.send_text(json.dumps(message))

    def get_device_status(self, device_id: str) -> dict:
        """Get device status."""
        if device_id not in self.devices:
            return None

        device = self.devices[device_id]
        return {
            'device_id': device_id,
            'type': device['type'],
            'status': device['status'],
            'last_update': device['last_update'].isoformat(),
            'data': device['data']
        }

dashboard = IoTDashboard()

@app.websocket("/ws/dashboard")
async def websocket_dashboard(websocket: WebSocket):
    await websocket.accept()
    dashboard.connections.add(websocket)

    try:
        while True:
            data = await websocket.receive_text()
            # Handle commands
    except Exception as e:
        dashboard.connections.remove(websocket)
```

---

## 7. NB-IoT Connection Manager

```python
class NBIoTDevice:
    """Manage NB-IoT cellular connection."""

    def __init__(self, apn='nbiot.vodafone.com'):
        self.apn = apn
        self.connected = False
        self.signal_quality = 0

    def initialize_module(self):
        """Initialize cellular module with AT commands."""
        at_commands = [
            ('AT', 'OK'),                              # Test
            ('AT+CPIN?', '+CPIN: READY'),              # Check SIM
            (f'AT+CGDCONT=1,"IP","{self.apn}"', 'OK'),  # Set APN
            ('AT+CGACT=1,1', 'OK'),                    # Activate context
            ('AT+COPS=0,0', 'OK'),                     # Auto operator
        ]

        for cmd, expected in at_commands:
            response = self.send_at_command(cmd)
            if expected not in response:
                print(f"Failed: {cmd}")
                return False

        self.connected = True
        return True

    def send_at_command(self, command: str, timeout: int = 5) -> str:
        """Send AT command to modem."""
        # Simulated response
        return "OK"

    def get_signal_quality(self) -> dict:
        """Get NB-IoT signal metrics."""
        # AT+NUESTATS returns various parameters
        response = self.send_at_command('AT+NUESTATS')

        return {
            'rsrp': -100,      # Reference Signal Received Power
            'rsrq': -10,       # Reference Signal Received Quality
            'snr': 5,          # Signal-to-Noise Ratio
            'bars': 3          # 1-4 bar representation
        }

    def estimate_data_cost(self, bytes_sent: int) -> float:
        """Estimate NB-IoT data cost."""
        # NB-IoT typically 5-50 bytes per transmission
        # Cost structure varies by operator
        cost_per_mb = 0.05  # Example: $0.05 per MB
        mb = bytes_sent / (1024 * 1024)
        return mb * cost_per_mb

    def close_connection(self):
        """Close NB-IoT connection."""
        self.send_at_command('AT+CGACT=0,1')  # Deactivate
        self.connected = False
```

---

## 8. Multi-Protocol Gateway

```python
import json
from typing import Dict, Any

class IoTGateway:
    """Support multiple IoT protocols (MQTT, CoAP, HTTP)."""

    def __init__(self):
        self.devices = {}
        self.protocol_handlers = {
            'mqtt': self.handle_mqtt,
            'coap': self.handle_coap,
            'http': self.handle_http
        }

    def register_device(self, device_id: str, protocol: str):
        """Register device with specific protocol."""
        self.devices[device_id] = {'protocol': protocol, 'last_seen': None}

    def process_message(self, device_id: str, payload: bytes) -> bool:
        """Process incoming message from device."""
        if device_id not in self.devices:
            return False

        protocol = self.devices[device_id]['protocol']
        handler = self.protocol_handlers.get(protocol)

        if handler:
            return handler(device_id, payload)

        return False

    def handle_mqtt(self, device_id: str, payload: bytes) -> bool:
        """Handle MQTT message."""
        data = json.loads(payload.decode())
        print(f"MQTT from {device_id}: {data}")
        return True

    def handle_coap(self, device_id: str, payload: bytes) -> bool:
        """Handle CoAP message."""
        print(f"CoAP from {device_id}")
        return True

    def handle_http(self, device_id: str, payload: bytes) -> bool:
        """Handle HTTP message."""
        print(f"HTTP from {device_id}")
        return True

    def send_command(self, device_id: str, command: Dict[str, Any]) -> bool:
        """Send command to device."""
        if device_id not in self.devices:
            return False

        protocol = self.devices[device_id]['protocol']
        # Route command based on protocol
        print(f"Sending {command} to {device_id} via {protocol}")
        return True
```

---

## 9. Time-Series Data Storage for IoT

```python
from collections import defaultdict
from datetime import datetime, timedelta

class IoTTimeSeriesStore:
    """Simple in-memory time series storage."""

    def __init__(self, retention_hours: int = 24):
        self.data = defaultdict(list)
        self.retention_hours = retention_hours

    def add_measurement(self, device_id: str, value: float, timestamp=None):
        """Add measurement to time series."""
        if timestamp is None:
            timestamp = datetime.utcnow()

        self.data[device_id].append({
            'timestamp': timestamp,
            'value': value
        })

        # Clean old data
        self.cleanup_old_data(device_id)

    def cleanup_old_data(self, device_id: str):
        """Remove measurements older than retention period."""
        cutoff = datetime.utcnow() - timedelta(hours=self.retention_hours)
        self.data[device_id] = [
            m for m in self.data[device_id]
            if m['timestamp'] > cutoff
        ]

    def get_metrics(self, device_id: str) -> dict:
        """Calculate statistics for device data."""
        if device_id not in self.data or not self.data[device_id]:
            return {}

        values = [m['value'] for m in self.data[device_id]]

        return {
            'count': len(values),
            'min': min(values),
            'max': max(values),
            'avg': sum(values) / len(values),
            'last_value': values[-1],
            'last_timestamp': self.data[device_id][-1]['timestamp'].isoformat()
        }

    def query_range(self, device_id: str, start: datetime, end: datetime) -> list:
        """Query measurements in time range."""
        if device_id not in self.data:
            return []

        return [
            m for m in self.data[device_id]
            if start <= m['timestamp'] <= end
        ]
```

---

**Related Skills**: moai-domain-backend, moai-domain-database, moai-essentials-perf
