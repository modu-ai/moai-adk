---
name: moai-domain-iot
description: IoT architecture, device management, and edge computing patterns for connected devices
allowed-tools: [Read, Bash, WebFetch]
---

# IoT Domain Expert

## Quick Reference (30 seconds)

IoT architecture encompasses device connectivity, data ingestion, edge computing, and cloud integration
using MQTT, CoAP, and LoRaWAN protocols. Design patterns enable billions of devices to communicate
securely, process data locally for latency-critical applications, and synchronize with cloud backends.

**Core Technologies** (November 2025):
- **Protocols**: MQTT (publish-subscribe), CoAP (constrained), LoRaWAN (long-range), NB-IoT (cellular)
- **Edge Computing**: AWS IoT Greengrass, Azure IoT Edge, Home Assistant, Node-RED
- **Cloud Platforms**: AWS IoT Core, Azure IoT Hub, Google IoT Core, Alibaba Cloud
- **Data Storage**: InfluxDB, TimescaleDB, Prometheus, Grafana

---

## Implementation Guide

### 1. Device Communication Protocols

**MQTT (Message Queuing Telemetry Transport)**:
```python
import paho.mqtt.client as mqtt
import ssl
import json

class IoTDevice:
    def __init__(self, client_id, broker_address, ca_cert, device_cert, device_key):
        self.client = mqtt.Client(client_id=client_id, protocol=mqtt.MQTTv311)

        # TLS/SSL Configuration
        self.client.tls_set(
            ca_certs=ca_cert,
            certfile=device_cert,
            keyfile=device_key,
            cert_reqs=ssl.CERT_REQUIRED,
            tls_version=ssl.PROTOCOL_TLSv1_2
        )

        self.client.on_connect = self._on_connect
        self.client.on_message = self._on_message
        self.broker = broker_address

    def _on_connect(self, client, userdata, flags, rc):
        if rc == 0:
            print("Connected to MQTT broker")
            client.subscribe("device/commands/#")
        else:
            print(f"Connection failed with code {rc}")

    def _on_message(self, client, userdata, msg):
        payload = json.loads(msg.payload.decode())
        self.handle_command(msg.topic, payload)

    def publish_sensor_data(self, sensor_id, temperature, humidity):
        payload = {
            "sensor_id": sensor_id,
            "temperature": temperature,
            "humidity": humidity,
            "timestamp": int(time.time())
        }
        self.client.publish(
            f"device/{sensor_id}/telemetry",
            json.dumps(payload),
            qos=1
        )

    def connect(self):
        self.client.connect(self.broker, 8883, keepalive=60)
        self.client.loop_start()

    def handle_command(self, topic, command):
        # Handle incoming commands from cloud
        pass
```

**CoAP (Constrained Application Protocol)**:
```python
from aiocoap import *
import asyncio

async def coap_client():
    """CoAP client for resource-constrained devices"""
    context = await Context.create_client_context()

    request = Message(code=GET, uri='coap://example.com/sensor/temperature')
    response = await context.request(request).response

    print(f"Temperature: {response.payload.decode('utf-8')}")
    await context.shutdown()

async def coap_server():
    """CoAP server running on edge device"""
    root = Resource()
    root.add_child(Site(), ['well-known'])

    site = Site(root)
    await asyncio.gather(
        site.render_to_transport(
            asyncio.DatagramEndpoint(site._proto_factory, ('0.0.0.0', 5683))
        )
    )
```

### 2. Edge Computing Architecture

**AWS IoT Greengrass Pattern**:
```python
# Lambda function running on edge device
import json
import greengrasssdk
import logging

logger = logging.getLogger()
logger.setLevel(logging.INFO)
client = greengrasssdk.client('iot-data')

def lambda_handler(event, context):
    """
    Process sensor data locally:
    - Filter outliers
    - Aggregate metrics
    - Cache during connectivity loss
    """

    # Local processing
    temperature = event.get('temperature')

    # Validate sensor reading
    if not (0 <= temperature <= 100):
        logger.warning(f"Invalid temperature: {temperature}")
        return {'status': 'invalid'}

    # Local threshold action
    if temperature > 80:
        trigger_cooling_system()

    # Send to cloud
    response = client.publish(
        topic='device/telemetry',
        qos=1,
        payload=json.dumps(event)
    )

    return {'status': 'processed'}

def trigger_cooling_system():
    """Local action without cloud dependency"""
    client.publish(
        topic='device/actions/cooling',
        qos=0,
        payload=json.dumps({'action': 'activate'})
    )
```

**Edge Data Pipeline**:
```
Sensor → Local Processing → Decision Logic → Cloud Sync
         (Filter, Aggregate)  (Threshold)    (Store, Analyze)

- Reduces latency: ms instead of seconds
- Saves bandwidth: 50-80% reduction
- Increases resilience: offline capability
```

### 3. Device Authentication & Security

**X.509 Certificate-Based Auth**:
```bash
#!/bin/bash
# Generate device certificate and key

# Create private key
openssl genrsa -out device.key 2048

# Create certificate signing request
openssl req -new -key device.key -out device.csr \
  -subj "/C=US/ST=CA/L=SF/O=MyCompany/CN=device-001"

# Sign with CA (AWS IoT creates this)
openssl x509 -req -in device.csr \
  -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out device.crt \
  -days 365 -sha256

# Verify certificate
openssl x509 -in device.crt -text -noout
```

**Device Provisioning Workflow**:
```python
class DeviceProvisioner:
    def __init__(self, api_endpoint):
        self.endpoint = api_endpoint

    def register_device(self, device_id, certificate_pem):
        """Register device with cloud backend"""
        response = requests.post(
            f"{self.endpoint}/devices",
            json={
                "device_id": device_id,
                "certificate": certificate_pem,
                "status": "active"
            },
            headers={"Authorization": "Bearer " + self.get_token()}
        )
        return response.json()

    def rotate_certificate(self, device_id, new_cert_pem):
        """Rotate device certificate (security best practice)"""
        response = requests.patch(
            f"{self.endpoint}/devices/{device_id}/certificate",
            json={"certificate": new_cert_pem}
        )
        return response.status_code == 200
```

### 4. Time-Series Data Storage

**InfluxDB Pattern**:
```python
from influxdb_client import InfluxDBClient
from influxdb_client.client.write_api import SYNCHRONOUS
from datetime import datetime

client = InfluxDBClient(url="http://localhost:8086", token="your-token", org="myorg")
write_api = client.write_api(write_options=SYNCHRONOUS)

def store_sensor_metrics(device_id, temperature, humidity):
    """Store IoT metrics in time-series database"""
    from influxdb_client.client.write.point import Point

    point = (
        Point("sensor_reading")
        .tag("device_id", device_id)
        .tag("location", "warehouse-1")
        .field("temperature", temperature)
        .field("humidity", humidity)
        .time(datetime.utcnow())
    )

    write_api.write(bucket="iot_data", record=point)

def query_temperature_trend(device_id, hours=24):
    """Query historical data"""
    query_api = client.query_api()

    query = f'''
    from(bucket:"iot_data")
      |> range(start: -{hours}h)
      |> filter(fn: (r) => r.device_id == "{device_id}")
      |> filter(fn: (r) => r._field == "temperature")
      |> aggregateWindow(every: 1h, fn: mean)
    '''

    result = query_api.query(query)
    return result
```

### 5. Fleet Management & OTA Updates

**Over-The-Air (OTA) Update Pattern**:
```python
import hashlib
import requests

class OTAManager:
    def __init__(self, s3_bucket, device_registry):
        self.bucket = s3_bucket
        self.registry = device_registry

    def publish_firmware_update(self, firmware_version, binary_url):
        """Publish new firmware to device fleet"""
        firmware_hash = self.calculate_hash(binary_url)

        # Store metadata
        self.registry.store_firmware_metadata({
            "version": firmware_version,
            "url": binary_url,
            "hash": firmware_hash,
            "release_date": datetime.utcnow().isoformat()
        })

        # Notify devices via MQTT
        self.notify_devices(firmware_version, binary_url)

    def calculate_hash(self, file_path):
        sha256 = hashlib.sha256()
        with open(file_path, 'rb') as f:
            for chunk in iter(lambda: f.read(4096), b''):
                sha256.update(chunk)
        return sha256.hexdigest()

    def verify_update_on_device(self, device_id, expected_hash):
        """Verify firmware integrity on device"""
        # Device sends its firmware hash
        device_hash = self.get_device_firmware_hash(device_id)
        return device_hash == expected_hash
```

---

## Best Practices

### ✅ DO
- **Authenticate all devices**: Use X.509 certificates (not API keys)
- **Encrypt communications**: TLS 1.2+ for all MQTT/CoAP connections
- **Design for offline**: Queue data locally during connectivity loss
- **Implement OTA updates**: Secure firmware update mechanism
- **Monitor device health**: Connection status, battery, error rates
- **Partition by time**: Use time-series databases for sensor data
- **Set QoS appropriately**: QoS 1 for important telemetry, QoS 0 for frequent metrics

### ❌ DON'T
- Transmit credentials in plaintext
- Ignore device provisioning workflow
- Neglect power consumption optimization
- Store unbounded sensor data in relational databases
- Update firmware without verification
- Use default credentials or passwords
- Deploy without monitoring and alerting

---

## Works Well With

- `moai-domain-devops` (Fleet management, CI/CD)
- `moai-domain-database` (Time-series data modeling)
- `moai-domain-security` (Device authentication, encryption)
- `moai-cloud-aws-advanced` (AWS IoT Core integration)

---

**Version**: 2.0.0 | **Last Updated**: 2025-11-21 | **Lines**: 245
