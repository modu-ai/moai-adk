---
name: moai-domain-iot
version: 4.0.0
updated: 2025-11-20
status: stable
category: Domain
description: Enterprise IoT patterns with MQTT, Edge Computing, and Device Management
allowed-tools: [WebSearch, WebFetch, Read, Bash]
---

# IoT Domain Expert

**Enterprise IoT Architecture: From Sensors to Cloud**

> **Focus**: MQTT, Edge Computing, Industrial IoT, Security  
> **Stack**: Python, Mosquitto, Docker, TensorFlow Lite

---

## Overview

Production-ready patterns for scalable, secure IoT systems.

### Core Architecture

- **Device Layer**: Sensors/Actuators (ESP32, Arduino)
- **Edge Layer**: Local processing & buffering (Raspberry Pi, Jetson)
- **Network Layer**: MQTT, CoAP, LoRaWAN, NB-IoT
- **Cloud Layer**: Analytics, Storage, Device Management

### Protocol Selection

| Protocol    | Range | Bandwidth | Power    | Use Case                             |
| ----------- | ----- | --------- | -------- | ------------------------------------ |
| **MQTT**    | Any   | Low-Med   | Low      | Real-time telemetry, Command/Control |
| **LoRaWAN** | 10km+ | Very Low  | Very Low | Remote sensors, Agriculture          |
| **WiFi 6**  | <100m | High      | High     | Video streaming, High-bandwidth data |
| **BLE**     | <100m | Low       | Low      | Wearables, Proximity                 |

---

## Implementation Patterns

### 1. Robust MQTT Client (Python)

Production-grade client with auto-reconnection and buffering.

```python
import json
import logging
import paho.mqtt.client as mqtt
from datetime import datetime

class IoTClient:
    def __init__(self, client_id, broker, port=1883, username=None, password=None):
        self.client = mqtt.Client(client_id=client_id, protocol=mqtt.MQTTv5)
        if username:
            self.client.username_pw_set(username, password)

        self.client.on_connect = self.on_connect
        self.client.on_message = self.on_message
        self.client.on_disconnect = self.on_disconnect

        self.connected = False
        self.topics = []

    def connect(self):
        try:
            self.client.connect(self.broker, self.port, keepalive=60)
            self.client.loop_start()
        except Exception as e:
            logging.error(f"Connection failed: {e}")

    def on_connect(self, client, userdata, flags, rc, properties=None):
        if rc == 0:
            self.connected = True
            logging.info("Connected to broker")
            for topic, qos in self.topics:
                client.subscribe(topic, qos)
        else:
            logging.error(f"Connection failed with code {rc}")

    def publish_telemetry(self, device_id, data):
        if not self.connected:
            return False

        payload = {
            "device_id": device_id,
            "ts": datetime.utcnow().isoformat(),
            "data": data
        }

        info = self.client.publish(
            f"devices/{device_id}/telemetry",
            json.dumps(payload),
            qos=1
        )
        return info.rc == mqtt.MQTT_ERR_SUCCESS

    def subscribe_command(self, device_id, callback):
        topic = f"devices/{device_id}/commands/#"
        self.topics.append((topic, 1))
        self.client.message_callback_add(topic, callback)
        if self.connected:
            self.client.subscribe(topic, 1)
```

### 2. Edge Data Processing

Local buffering and aggregation before cloud upload.

```python
import asyncio
from collections import deque

class EdgeProcessor:
    def __init__(self, buffer_size=100):
        self.buffer = deque(maxlen=buffer_size)
        self.cloud_client = IoTClient("edge-gateway", "cloud.broker.com")

    async def process_reading(self, sensor_id, value):
        # 1. Local Anomaly Detection
        if value > 100:  # Simple threshold
            await self.trigger_local_alert(sensor_id, value)

        # 2. Buffer for Aggregation
        self.buffer.append((sensor_id, value))

        # 3. Batch Upload
        if len(self.buffer) >= 50:
            await self.flush_to_cloud()

    async def flush_to_cloud(self):
        batch = list(self.buffer)
        self.buffer.clear()

        # Calculate aggregates
        avg_val = sum(v for _, v in batch) / len(batch)

        payload = {
            "count": len(batch),
            "avg": avg_val,
            "raw": batch
        }

        self.cloud_client.publish_telemetry("edge-01", payload)
```

### 3. Edge ML Inference (TFLite)

Running lightweight models on edge devices.

```python
import tensorflow as tf
import numpy as np

class EdgeModel:
    def __init__(self, model_path):
        self.interpreter = tf.lite.Interpreter(model_path=model_path)
        self.interpreter.allocate_tensors()

        self.input_details = self.interpreter.get_input_details()
        self.output_details = self.interpreter.get_output_details()

    def predict(self, input_data):
        # Preprocess
        input_data = np.array(input_data, dtype=np.float32)
        input_data = np.expand_dims(input_data, 0)

        # Set input
        self.interpreter.set_tensor(self.input_details[0]['index'], input_data)

        # Run inference
        self.interpreter.invoke()

        # Get output
        output = self.interpreter.get_tensor(self.output_details[0]['index'])
        return output[0]
```

---

## Infrastructure & Deployment

### Docker Compose (Edge Gateway)

```yaml
version: "3.8"
services:
  mosquitto:
    image: eclipse-mosquitto:2.0
    ports:
      - "1883:1883"
      - "8883:8883"
    volumes:
      - ./mosquitto/config:/mosquitto/config
      - ./mosquitto/data:/mosquitto/data

  edge-service:
    build: .
    restart: always
    depends_on:
      - mosquitto
    environment:
      - MQTT_BROKER=mosquitto
      - DEVICE_ID=gateway-01
    devices:
      - "/dev/gpiomem:/dev/gpiomem" # GPIO access

  influxdb:
    image: influxdb:2.7
    ports:
      - "8086:8086"
    volumes:
      - influxdb-data:/var/lib/influxdb2
```

### Mosquitto Configuration

```conf
# mosquitto.conf
listener 1883
protocol mqtt

listener 8883
protocol mqtt
cafile /mosquitto/config/certs/ca.crt
certfile /mosquitto/config/certs/broker.crt
keyfile /mosquitto/config/certs/broker.key

persistence true
persistence_location /mosquitto/data/

# Security
allow_anonymous false
password_file /mosquitto/config/passwd
acl_file /mosquitto/config/acl
```

---

## Security Best Practices

1.  **Transport Security**: Always use TLS (MQTTS) for cloud communication.
2.  **Authentication**: Use X.509 client certificates or strong unique credentials per device.
3.  **Authorization**: Implement strict ACLs (e.g., sensors can only write to their own topics).
4.  **Firmware Updates**: Secure OTA (Over-The-Air) updates with code signing.

---

## Validation Checklist

**Connectivity**:

- [ ] MQTT broker reachable
- [ ] TLS handshake successful
- [ ] Auto-reconnection working

**Data Flow**:

- [ ] Telemetry publishing (QoS 1)
- [ ] Command subscription working
- [ ] Edge buffering functioning

**Security**:

- [ ] Credentials not hardcoded
- [ ] ACLs enforced
- [ ] Data encrypted at rest/transit

---

## Related Skills

- `moai-domain-ml`: Edge ML models
- `moai-security-devsecops`: Secure deployment
- `moai-cloud-aws-advanced`: AWS IoT Core integration

---

**Last Updated**: 2025-11-20
