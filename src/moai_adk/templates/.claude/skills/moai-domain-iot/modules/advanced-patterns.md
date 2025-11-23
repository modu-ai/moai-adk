# Advanced IoT Patterns - Enterprise Edge Computing & Device Management

**Version**: 4.0.0 (2025-11-22)
**Status**: Production Ready

---

## Advanced Device Management

### MQTT Broker with Hierarchical Topic Organization

```python
import paho.mqtt.client as mqtt
import json
from datetime import datetime

class EnterpriseIoTBroker:
    """Enterprise-grade MQTT broker with hierarchical topics and security."""

    def __init__(self, broker_host='localhost', broker_port=8883, ca_certs='ca.crt'):
        self.client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, client_id='iot_hub')
        self.client.tls_set(ca_certs=ca_certs)
        self.client.username_pw_set('admin', 'password')

        self.client.on_connect = self.on_connect
        self.client.on_message = self.on_message

        self.client.connect(broker_host, broker_port, keepalive=60)

    def on_connect(self, client, userdata, connect_flags, rc, properties):
        """Subscribe to all device topics on connection."""
        print(f"Connected with result code {rc}")
        client.subscribe("devices/+/+/telemetry", qos=1)  # device/type/id/telemetry
        client.subscribe("devices/+/+/status", qos=1)

    def on_message(self, client, userdata, msg):
        """Process incoming messages with validation."""
        try:
            topic_parts = msg.topic.split('/')
            if len(topic_parts) == 4:
                device_type, device_id, message_type = topic_parts[1:4]
                payload = json.loads(msg.payload.decode())

                self.process_device_message(
                    device_type, device_id, message_type, payload
                )
        except json.JSONDecodeError:
            print(f"Invalid JSON from {msg.topic}")

    def publish_command(self, device_type: str, device_id: str, command: dict):
        """Publish command to device with QoS=1 (at least once)."""
        topic = f"devices/{device_type}/{device_id}/commands"
        payload = json.dumps({
            'timestamp': datetime.utcnow().isoformat(),
            'command': command
        })
        self.client.publish(topic, payload, qos=1, retain=False)

    def process_device_message(self, device_type, device_id, message_type, payload):
        """Process different message types from devices."""
        if message_type == 'telemetry':
            self.store_telemetry(device_type, device_id, payload)
        elif message_type == 'status':
            self.update_device_status(device_type, device_id, payload)

    def run(self):
        """Start MQTT loop."""
        self.client.loop_forever()
```

### Device Shadow Pattern for Offline Capability

```python
import json
from datetime import datetime
from typing import Dict, Any

class DeviceShadow:
    """Device shadow for resilient offline operation."""

    def __init__(self, device_id: str, backend_storage):
        self.device_id = device_id
        self.storage = backend_storage
        self.shadow = {
            'state': {
                'desired': {},
                'reported': {},
                'delta': {}
            },
            'metadata': {
                'desiredTimestamp': None,
                'reportedTimestamp': None
            },
            'version': 0
        }

    def update_desired_state(self, desired_state: Dict[str, Any]):
        """Update desired state (cloud to device)."""
        self.shadow['state']['desired'].update(desired_state)
        self.shadow['metadata']['desiredTimestamp'] = datetime.utcnow().isoformat()
        self.shadow['version'] += 1
        self.compute_delta()
        self.storage.save_shadow(self.device_id, self.shadow)

    def update_reported_state(self, reported_state: Dict[str, Any]):
        """Update reported state (device to cloud)."""
        self.shadow['state']['reported'].update(reported_state)
        self.shadow['metadata']['reportedTimestamp'] = datetime.utcnow().isoformat()
        self.shadow['version'] += 1
        self.compute_delta()
        self.storage.save_shadow(self.device_id, self.shadow)

    def compute_delta(self):
        """Compute delta (desired - reported)."""
        desired = self.shadow['state']['desired']
        reported = self.shadow['state']['reported']

        self.shadow['state']['delta'] = {
            k: v for k, v in desired.items()
            if k not in reported or reported[k] != v
        }

    def get_delta_for_device(self):
        """Get only delta to send to device (bandwidth efficient)."""
        return self.shadow['state']['delta']
```

### Over-The-Air (OTA) Update Management

```python
import hashlib
from typing import BinaryIO

class OTAUpdateManager:
    """Manage firmware updates across fleet of devices."""

    def __init__(self, storage_backend):
        self.storage = storage_backend

    def prepare_firmware_package(self, firmware_path: str, version: str):
        """Prepare and sign firmware for distribution."""
        with open(firmware_path, 'rb') as f:
            firmware_content = f.read()

        # Calculate checksum for integrity verification
        checksum = hashlib.sha256(firmware_content).hexdigest()

        firmware_metadata = {
            'version': version,
            'size': len(firmware_content),
            'checksum': checksum,
            'timestamp': datetime.utcnow().isoformat(),
            'rollout_percentage': 10  # Gradual rollout
        }

        self.storage.store_firmware(version, firmware_content)
        self.storage.store_metadata(version, firmware_metadata)

        return firmware_metadata

    def rollout_firmware(self, device_filter: Dict, version: str):
        """Initiate gradual firmware rollout to devices."""
        devices = self.get_matching_devices(device_filter)
        firmware_meta = self.storage.get_metadata(version)

        update_commands = []
        for device in devices:
            update_cmd = {
                'action': 'update',
                'firmware_version': version,
                'checksum': firmware_meta['checksum'],
                'download_url': f'https://ota.example.com/fw/{version}',
                'max_retry': 5,
                'timeout_seconds': 300
            }
            update_commands.append((device['id'], update_cmd))

        return update_commands

    def monitor_rollout(self, version: str):
        """Monitor OTA rollout progress and handle failures."""
        results = self.storage.get_update_results(version)

        return {
            'version': version,
            'total_devices': len(results),
            'successful': sum(1 for r in results if r['status'] == 'success'),
            'failed': sum(1 for r in results if r['status'] == 'failed'),
            'pending': sum(1 for r in results if r['status'] == 'pending'),
            'success_rate': sum(1 for r in results if r['status'] == 'success') / len(results) if results else 0
        }
```

---

## Advanced Edge Computing

### Edge Computing with TensorFlow Lite

```python
import tensorflow as tf
import numpy as np

class EdgeMLModel:
    """Run ML models on edge devices with TensorFlow Lite."""

    def __init__(self, model_path: str):
        self.interpreter = tf.lite.Interpreter(model_path=model_path)
        self.interpreter.allocate_tensors()

        self.input_details = self.interpreter.get_input_details()
        self.output_details = self.interpreter.get_output_details()

    def predict(self, input_data: np.ndarray) -> np.ndarray:
        """Run inference on edge device."""
        # Prepare input (convert to right dtype and shape)
        input_data = input_data.astype(self.input_details[0]['dtype'])
        self.interpreter.set_tensor(self.input_details[0]['index'], input_data)

        # Run inference
        self.interpreter.invoke()

        # Get output
        output = self.interpreter.get_tensor(self.output_details[0]['index'])
        return output

    def get_model_info(self):
        """Get model size and resource requirements."""
        return {
            'input_shape': self.input_details[0]['shape'],
            'output_shape': self.output_details[0]['shape'],
            'input_dtype': str(self.input_details[0]['dtype']),
            'output_dtype': str(self.output_details[0]['dtype'])
        }
```

### Stream Processing with Apache Kafka

```python
from kafka import KafkaConsumer, KafkaProducer
import json

class EdgeStreamProcessor:
    """Process IoT streams in real-time on edge."""

    def __init__(self, kafka_servers: list):
        self.consumer = KafkaConsumer(
            'iot-telemetry',
            bootstrap_servers=kafka_servers,
            group_id='edge-processor',
            value_deserializer=lambda m: json.loads(m.decode('utf-8')),
            auto_offset_reset='latest'
        )

        self.producer = KafkaProducer(
            bootstrap_servers=kafka_servers,
            value_serializer=lambda v: json.dumps(v).encode('utf-8')
        )

    def process_stream(self):
        """Process incoming IoT data with simple rules engine."""
        for message in self.consumer:
            data = message.value

            # Apply edge processing rules
            processed = self.apply_rules(data)

            # Send processed data upstream
            self.producer.send('processed-telemetry', value=processed)

    def apply_rules(self, data):
        """Apply simple processing rules (filtering, aggregation)."""
        if data['temperature'] > 50:  # Alert on high temperature
            return {
                'original': data,
                'alert': 'HIGH_TEMPERATURE',
                'processed_time': datetime.utcnow().isoformat()
            }
        return {'original': data, 'processed_time': datetime.utcnow().isoformat()}
```

---

## Advanced Network Protocols

### CoAP (Constrained Application Protocol) Server

```python
from aiocoap import resource, Context
import asyncio

class IotResourceHandler(resource.Resource):
    """Handle CoAP requests for constrained devices."""

    async def render_get(self, request):
        """Handle GET requests (sensor readings)."""
        # Return sensor data in CBOR format (smaller than JSON)
        return resource.Response(content=b'{"temp": 25.5, "humidity": 60}')

    async def render_post(self, request):
        """Handle POST requests (device commands)."""
        payload = request.payload
        # Process command and return acknowledgment
        return resource.Response(code=aiocoap.CHANGED, payload=b'OK')

async def start_coap_server():
    """Start CoAP server for IoT devices."""
    root = resource.Site()
    root.add_child(('sensor',), IotResourceHandler())

    context = await Context.create_server_context(root)
    await asyncio.Event().wait()  # Run forever
```

### NB-IoT/LTE-M Support for Cellular IoT

```python
class CellularIoTDevice:
    """Support for cellular IoT protocols (NB-IoT, LTE-M)."""

    def __init__(self, apn='nbiot.vodafone.com'):
        self.apn = apn
        self.signal_quality = None

    def initialize_cellular(self):
        """Initialize cellular connection."""
        # AT commands for cellular module
        at_commands = [
            'AT+CPIN?',  # Check SIM
            f'AT+CGDCONT=1,"IP","{self.apn}"',  # Set APN
            'AT+CGACT=1,1',  # Activate context
            'AT+COPS=0,0'  # Select operator
        ]
        return at_commands

    def get_signal_quality(self):
        """Monitor signal quality (important for reliability)."""
        # RSRP range: -140 to -44 dBm
        # RSRQ range: -20 to -3 dB
        return {
            'rsrp': -100,  # Reference Signal Received Power
            'rsrq': -10,   # Reference Signal Received Quality
            'signal_bars': 3
        }
```

---

## Best Practices

**DO**:
- Use device shadows for offline resilience
- Implement gradual OTA rollouts (don't rush)
- Monitor signal quality and battery health
- Use TLS 1.3 for all connections
- Implement device certificate rotation
- Use batch operations to reduce network calls
- Compress data before transmission

**DON'T**:
- Deploy firmware updates to all devices at once
- Store sensitive data on devices
- Ignore security in favor of speed
- Skip device health monitoring
- Use plaintext MQTT (always use TLS)
- Implement custom cryptography
- Deploy without testing on real hardware

---

**Related Skills**: moai-domain-backend, moai-domain-devops, moai-essentials-perf
