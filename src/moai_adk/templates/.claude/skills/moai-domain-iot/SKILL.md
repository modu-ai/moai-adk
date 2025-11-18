---
name: moai-domain-iot
version: 4.0.0
updated: 2025-11-19
status: stable
category: Domain
description: MQTT, edge computing, Industrial IoT, device management, sensor networks
author: Alfred SuperAgent
tags: [iot, mqtt, edge-computing, industrial-iot, device-management, sensor-networks]
allowed_tools: [WebSearch, WebFetch, Read, Bash]
---

# IoT Domain Mastery: From Sensors to Cloud Integration

## Overview

The Internet of Things has evolved from a futuristic concept to essential infrastructure powering smart homes, industrial facilities, agricultural systems, and healthcare devices. In 2025, the convergence of 5G networks, WiFi 6, and edge computing has fundamentally transformed IoT architecture, enabling real-time processing, ultra-low latency, and offline-capable systems.

This Skill provides production-ready patterns, architectural best practices, and implementation strategies for building scalable, secure IoT systems across diverse use cases. Whether you're deploying a single environmental sensor or orchestrating thousands of industrial machines, you'll learn the protocols, tools, and patterns that define modern IoT systems.

The shift from cloud-centric to edge-centric computing reflects a critical reality: processing data where it's generated (at the edge) reduces latency, bandwidth costs, and cloud dependency. Combined with robust MQTT communication, comprehensive device management, and industrial-grade security, this approach enables systems that operate reliably even under challenging network conditions.

---

## Quick Decision Tree

**Which IoT pattern do you need?**

```
Real-time sensor data collection?
  → IoT Architecture Fundamentals + MQTT Protocol

Edge processing & local decision-making?
  → Edge Computing + Cloud Integration

Industrial machine monitoring & predictive maintenance?
  → Industrial IoT (IIoT) + Device Management

Distributed sensor networks?
  → Sensor Networks + Best Practices

Large-scale device fleet management?
  → Device Management + Cloud Integration
```

---

## Section 1: IoT Architecture Fundamentals

### 1.1 System Components & Data Flow

Modern IoT systems consist of four integrated layers:

**Sensor/Actuator Layer** - Hardware endpoints that interact with physical world:
- **Sensors**: Continuously collect environmental data (temperature, pressure, light, motion, GPS)
- **Actuators**: Execute commands received from control systems (motors, relays, pumps)
- **Data rate**: Typically 10-100 measurements/second depending on sensor type
- **Power**: Battery-operated (months to years) or line-powered

**Network Layer** - Communication infrastructure connecting devices:
- **WiFi 6 (802.11ax)**: 1+ Gbps throughput, improved efficiency for IoT
- **Bluetooth Low Energy (BLE)**: Short-range (10-100m), extremely low power
- **LoRaWAN**: Long-range (2-15km), low bandwidth, ideal for remote sensors
- **NB-IoT & LTE-M**: Cellular IoT standards offering wide coverage
- **Wired alternatives**: Ethernet, industrial protocols (Modbus, Profibus)

**Edge/Gateway Layer** - Local processing and aggregation:
- Reduces cloud bandwidth by 40-60% through local filtering
- Enables offline operation when cloud connectivity fails
- Real-time decision-making with <50ms latency
- Devices: Raspberry Pi, NVIDIA Jetson, industrial controllers

**Cloud Layer** - Centralized analytics, storage, and integration:
- Time-series databases store metrics for historical analysis
- ML models identify patterns and predict failures
- Integration with enterprise systems (ERP, CRM)
- Geographic redundancy and disaster recovery

**Example 1: Sensor Data Collection Architecture**

```python
# Edge device collects sensor data, applies local processing
import asyncio
import json
from datetime import datetime
from dataclasses import dataclass

@dataclass
class SensorReading:
    timestamp: str
    sensor_id: str
    type: str  # temperature, humidity, pressure
    value: float
    unit: str

class EdgeDataCollector:
    def __init__(self, buffer_size: int = 100):
        self.buffer = []
        self.buffer_size = buffer_size
        self.local_stats = {}
    
    async def collect_sensor_data(self, sensor_id: str, value: float) -> None:
        """Collect and buffer sensor data with local processing"""
        reading = SensorReading(
            timestamp=datetime.utcnow().isoformat(),
            sensor_id=sensor_id,
            type="temperature",
            value=value,
            unit="celsius"
        )
        
        self.buffer.append(reading)
        
        # Local processing: Calculate moving average
        recent = [r.value for r in self.buffer[-10:]]
        self.local_stats[sensor_id] = {
            "current": value,
            "avg_10s": sum(recent) / len(recent),
            "min": min(recent),
            "max": max(recent)
        }
        
        # Send to cloud when buffer full (batching reduces overhead)
        if len(self.buffer) >= self.buffer_size:
            await self.send_to_cloud()
    
    async def send_to_cloud(self) -> None:
        """Batch send readings to cloud"""
        payload = {
            "batch_id": str(datetime.utcnow().timestamp()),
            "readings": [
                {
                    "timestamp": r.timestamp,
                    "sensor_id": r.sensor_id,
                    "value": r.value
                }
                for r in self.buffer
            ],
            "count": len(self.buffer)
        }
        
        # In production, use MQTT (see Section 2) or HTTP POST
        print(f"[CLOUD] Sending {payload['count']} readings")
        self.buffer.clear()

# Usage
async def main():
    collector = EdgeDataCollector()
    
    # Simulate sensor readings
    for i in range(150):
        await collector.collect_sensor_data(
            sensor_id="temp_01",
            value=20.5 + (i % 10) * 0.1
        )
        await asyncio.sleep(0.5)

asyncio.run(main())
```

### 1.2 Network Protocol Comparison

| Protocol | Range | Bandwidth | Power | Use Case |
|----------|-------|-----------|-------|----------|
| **WiFi 6** | 100-200m | 1+ Gbps | 2-5W | High-data IoT, video streaming |
| **Bluetooth LE** | 10-100m | 1-2 Mbps | 5-50mW | Wearables, personal devices |
| **LoRaWAN** | 2-15km | 50 Kbps | 50-100mW | Remote sensors, sparse networks |
| **NB-IoT** | 35km | 250 Kbps | 50-100mW | Cellular coverage, machine monitor |
| **Zigbee** | 10-100m | 250 Kbps | 20-50mW | Home automation, mesh networks |
| **Wired (Ethernet)** | 100m+ | 1+ Gbps | 1W | Industrial facilities, high reliability |

**Selection criteria**:
- **Distance requirements**: LoRaWAN for long-range, WiFi for short-range
- **Data volume**: High data → WiFi 6, Low data → LoRaWAN/NB-IoT
- **Power budget**: Battery-powered → Choose LE/LoRaWAN, Mains-powered → WiFi/Ethernet
- **Latency**: <100ms required → WiFi/Ethernet, >1s acceptable → LoRaWAN

### 1.3 Cloud Connectivity Patterns

**MQTT (Message Queue Telemetry Transport)**:
- Lightweight publish-subscribe model ideal for intermittent connectivity
- QoS levels ensure message delivery even with unreliable networks
- Sub-100ms latency for real-time applications

**AMQP (Advanced Message Queuing Protocol)**:
- Enterprise-grade with guaranteed delivery and persistence
- Higher overhead than MQTT (suitable for powerful gateways)

**HTTP/HTTPS REST APIs**:
- Synchronous request-response model
- Higher overhead per request
- Simpler to debug but less efficient for streaming data

**WebSocket**:
- Bidirectional persistent connection
- Real-time command delivery
- Ideal for interactive devices (drones, robots)

---

## Section 2: MQTT Protocol - The IoT Communication Standard

### 2.1 MQTT Fundamentals

MQTT (Message Queue Telemetry Transport) is the dominant protocol for IoT communication, handling billions of messages daily across smart homes, industrial systems, and connected vehicles.

**Core concepts**:

- **Pub/Sub Model**: Publishers send messages to topics; subscribers receive from topics of interest. Decouples devices - sensors don't need to know which systems consume their data.

- **Broker**: Central message router (Mosquitto, Vernemq, HiveMQ) managing all communication. Single point of coordination.

- **Topic Hierarchy**: Use `/` separators like `building/floor1/room201/temperature`. Wildcards: `building/+/room201/+` matches any floor/sensor type.

- **QoS (Quality of Service)**:
  - **QoS 0**: At most once (fire and forget, lowest overhead)
  - **QoS 1**: At least once (message guaranteed, may duplicate)
  - **QoS 2**: Exactly once (guaranteed single delivery, highest overhead)

- **Retained Messages**: Broker keeps last message per topic, new subscribers receive latest state immediately.

- **Will Message**: Device specifies message published if it disconnects abruptly (e.g., "sensor_01/status: offline").

**Performance characteristics**:
- Typical payload: 1-10KB (supports up to 256MB)
- Message latency: 10-100ms (depends on network)
- Throughput: Single broker handles 100K-1M messages/second
- Connection overhead: ~2KB per device

**Example 2: Mosquitto MQTT Broker Setup**

```yaml
# /etc/mosquitto/mosquitto.conf
# Mosquitto 2.0+ configuration for production IoT deployment

# Network
listener 1883
protocol mqtt

# Secure MQTT (TLS/SSL)
listener 8883
protocol mqtt
cafile /etc/mosquitto/ca.crt
certfile /etc/mosquitto/broker.crt
keyfile /etc/mosquitto/broker.key
require_certificate false

# WebSocket support (for browser-based clients)
listener 9001
protocol websockets

# Authentication
allow_anonymous false
password_file /etc/mosquitto/passwd

# ACL - Access Control List
acl_file /etc/mosquitto/aclfile

# Message persistence
persistence true
persistence_location /var/lib/mosquitto/

# Maximum connections per client
max_connections -1

# Maximum queued messages per client
max_queued_messages 1000

# Message retention
retention_memory_limit 64000  # 64MB max retained messages

# Logging
log_dest file /var/log/mosquitto/mosquitto.log
log_dest syslog
log_type all

# Performance tuning
autosave_interval 1800  # Save persistence every 30 minutes
autosave_on_changes false
persistence_file mosquitto.db

# Thread pool for improved throughput
max_connections -1
```

### 2.2 Mosquitto Installation & Configuration

**Docker deployment (recommended for development)**:

```bash
# Docker Compose for complete MQTT stack
version: '3.8'

services:
  mosquitto:
    image: eclipse-mosquitto:2.0.18
    ports:
      - "1883:1883"    # MQTT
      - "8883:8883"    # MQTT with TLS
      - "9001:9001"    # WebSocket
    volumes:
      - ./mosquitto.conf:/mosquitto/config/mosquitto.conf
      - ./aclfile:/mosquitto/config/aclfile
      - ./passwd:/mosquitto/config/passwd
      - mosquitto-data:/mosquitto/data
      - mosquitto-logs:/mosquitto/log
    environment:
      - TZ=UTC
    healthcheck:
      test: ["CMD", "mosquitto_sub", "-h", "localhost", "-t", "#", "-W", "1"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  mosquitto-data:
  mosquitto-logs:
```

**Initialize user credentials**:

```bash
# Create password file (bcrypt hashed)
docker run --rm -it eclipse-mosquitto:2.0.18 \
  mosquitto_passwd -c /tmp/passwd sensor_user

# Copy to persistent volume
docker cp mosquitto:/tmp/passwd ./passwd
```

### 2.3 MQTT Client Implementation

**Example 3: Python MQTT Subscriber & Publisher**

```python
# pip install paho-mqtt==1.7.1
import asyncio
import json
import logging
from datetime import datetime
import paho.mqtt.client as mqtt

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class MQTTClient:
    def __init__(self, broker_host: str, broker_port: int = 1883, 
                 username: str = None, password: str = None):
        self.broker_host = broker_host
        self.broker_port = broker_port
        self.client = mqtt.Client(
            client_id=f"python-client-{datetime.now().timestamp()}",
            protocol=mqtt.MQTTv311,
            clean_session=True
        )
        
        # Set credentials
        if username and password:
            self.client.username_pw_set(username, password)
        
        # Register callbacks
        self.client.on_connect = self.on_connect
        self.client.on_message = self.on_message
        self.client.on_disconnect = self.on_disconnect
        self.client.on_publish = self.on_publish
        
        # Connection tracking
        self.is_connected = False
        self.subscriptions = {}
    
    def on_connect(self, client, userdata, flags, rc):
        """Connection callback"""
        if rc == 0:
            logger.info(f"Connected to {self.broker_host}:{self.broker_port}")
            self.is_connected = True
            
            # Resubscribe to topics on reconnection
            for topic, qos in self.subscriptions.items():
                client.subscribe(topic, qos)
                logger.info(f"Resubscribed to {topic}")
        else:
            logger.error(f"Connection failed with code {rc}")
    
    def on_message(self, client, userdata, msg):
        """Message received callback"""
        try:
            payload = json.loads(msg.payload.decode())
            logger.info(f"[{msg.topic}] QoS={msg.qos}: {payload}")
        except json.JSONDecodeError:
            logger.warning(f"Non-JSON message on {msg.topic}")
    
    def on_disconnect(self, client, userdata, rc):
        """Disconnection callback"""
        self.is_connected = False
        if rc != 0:
            logger.warning(f"Unexpected disconnection: code {rc}")
    
    def on_publish(self, client, userdata, mid):
        """Publish confirmation callback"""
        logger.debug(f"Message published (mid={mid})")
    
    def connect(self):
        """Establish connection to broker"""
        self.client.connect(self.broker_host, self.broker_port, keepalive=60)
        self.client.loop_start()
    
    def disconnect(self):
        """Gracefully disconnect from broker"""
        self.client.loop_stop()
        self.client.disconnect()
    
    def publish(self, topic: str, payload: dict, qos: int = 1, retain: bool = False):
        """Publish message to topic"""
        if not self.is_connected:
            logger.warning("Not connected, message queued")
        
        msg_info = self.client.publish(
            topic=topic,
            payload=json.dumps(payload),
            qos=qos,
            retain=retain
        )
        
        logger.info(f"Publishing to {topic} (QoS={qos}): {payload}")
        return msg_info
    
    def subscribe(self, topic: str, qos: int = 1, callback=None):
        """Subscribe to topic with optional callback"""
        self.subscriptions[topic] = qos
        
        if callback:
            # Wrap user callback
            original_on_message = self.client.on_message
            def wrapper(client, userdata, msg):
                if msg.topic.startswith(topic.replace('+', '').replace('#', '')):
                    callback(msg.topic, json.loads(msg.payload))
                original_on_message(client, userdata, msg)
            
            self.client.on_message = wrapper
        
        self.client.subscribe(topic, qos)
        logger.info(f"Subscribed to {topic} (QoS={qos})")

# Usage example
async def sensor_simulation():
    """Simulate temperature sensor publishing readings"""
    client = MQTTClient(
        broker_host="localhost",
        broker_port=1883,
        username="sensor_user",
        password="sensor_password"
    )
    
    client.connect()
    await asyncio.sleep(2)  # Wait for connection
    
    # Publish temperature readings
    for i in range(10):
        temperature = 20.0 + (i % 5) * 0.5
        payload = {
            "sensor_id": "temp_01",
            "temperature": temperature,
            "unit": "celsius",
            "timestamp": datetime.utcnow().isoformat()
        }
        
        client.publish(
            topic="building/floor1/room201/temperature",
            payload=payload,
            qos=1,
            retain=True
        )
        await asyncio.sleep(2)
    
    client.disconnect()

# async main for running
# asyncio.run(sensor_simulation())
```

### 2.4 MQTT Security - TLS/SSL & Authentication

**Example 4: MQTT Security Configuration**

```yaml
# MQTT Security Architecture
# Covers: Transport encryption, client authentication, message-level security

# 1. TLS/SSL Certificate Setup (self-signed for development)
openssl_config: |
  [ req ]
  distinguished_name = req_distinguished_name
  req_extensions = v3_req
  prompt = no

  [ req_distinguished_name ]
  C = US
  ST = California
  L = San Francisco
  O = IoT Platform
  CN = mqtt.example.com

  [ v3_req ]
  subjectAltName = @alt_names

  [ alt_names ]
  DNS.1 = mqtt.example.com
  DNS.2 = localhost
  IP.1 = 127.0.0.1
  IP.2 = 192.168.1.100

# 2. Mosquitto ACL Configuration
mosquitto_acl: |
  # Default deny all
  pattern write $SYS/#
  pattern read $SYS/broker/clients/#

  # Sensor devices - publish only
  user sensor_temp_01
  topic write building/floor1/room201/temperature
  topic write building/floor1/room201/status

  # Sensor devices - read control commands
  user sensor_temp_01
  topic read building/floor1/room201/cmd/+

  # Analytics service - subscribe all
  user analytics_service
  topic read building/#
  topic write analytics/results/#

  # Admin user - full access
  user admin
  topic readwrite #

# 3. Client Certificate Authentication
client_cert_auth: |
  # Enable client certificate verification
  require_certificate true
  use_identity_as_username true
  
  # Client certificate chain
  cafile /etc/mosquitto/certs/ca.crt
  certfile /etc/mosquitto/certs/client.crt
  keyfile /etc/mosquitto/certs/client.key
  
  # Verify client cert matches CN to username
  allow_anonymous false

# 4. Message-Level Security (application layer)
message_encryption: |
  # Encrypt sensitive payloads before publishing
  # Algorithm: AES-256-GCM
  
  import json
  from cryptography.hazmat.primitives.ciphers.aead import AESGCM
  import os
  
  class SecureMQTTPayload:
      def __init__(self, encryption_key: bytes):
          self.key = encryption_key
      
      def encrypt_payload(self, data: dict) -> dict:
          """Encrypt sensitive data before MQTT publish"""
          nonce = os.urandom(12)
          cipher = AESGCM(self.key)
          
          plaintext = json.dumps(data).encode()
          ciphertext = cipher.encrypt(nonce, plaintext, None)
          
          return {
              "nonce": nonce.hex(),
              "ciphertext": ciphertext.hex(),
              "algorithm": "AES-256-GCM"
          }
      
      def decrypt_payload(self, encrypted_data: dict) -> dict:
          """Decrypt received payload"""
          cipher = AESGCM(self.key)
          nonce = bytes.fromhex(encrypted_data["nonce"])
          ciphertext = bytes.fromhex(encrypted_data["ciphertext"])
          
          plaintext = cipher.decrypt(nonce, ciphertext, None)
          return json.loads(plaintext)
```

### 2.5 MQTT Performance & Optimization

**QoS Impact on Throughput**:

| QoS | Per-Message Overhead | Max Throughput* | Use Case |
|-----|---------------------|-----------------|----------|
| 0 | ~5 bytes | 50K msg/sec | Non-critical sensor data |
| 1 | ~20 bytes | 20K msg/sec | Standard IoT applications |
| 2 | ~50 bytes | 5K msg/sec | Financial transactions, critical alerts |

*Per-message throughput on typical home internet connection

**Optimization strategies**:

1. **Batch messages**: Send 10-100 readings in single message vs individual
2. **Topic filtering**: Use wildcards efficiently to reduce subscription overhead
3. **Message compression**: Gzip payloads >1KB
4. **Connection pooling**: Reuse MQTT connections for multiple logical clients

---

## Section 3: Edge Computing & Local Processing

### 3.1 Edge Computing Fundamentals

Edge computing processes data locally on devices near data collection points, reducing cloud dependency and enabling real-time decision-making. In 2025, edge has become essential due to:

- **Latency requirements**: <50ms response time impossible over internet
- **Bandwidth costs**: Processing locally saves 30-60% cloud bandwidth
- **Privacy**: Sensitive data never leaves facility
- **Resilience**: System functions even when cloud unavailable

**Edge device spectrum**:

- **Micro edge** (IoT device): Raspberry Pi Zero, Arduino
- **Lightweight edge**: Raspberry Pi 4, NVIDIA Jetson Nano
- **Industrial edge**: NVIDIA Jetson Xavier, x86 servers
- **Far edge**: Cloud region (50-100ms latency)

### 3.2 Edge Device Setup - Docker on Raspberry Pi

**Example 5: Docker Environment for Edge Computing**

```dockerfile
# Dockerfile.arm32v7
# Multi-stage build for Raspberry Pi (ARMv7)

FROM arm32v7/python:3.11-slim as builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    libssl-dev \
    libffi-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements and build wheels
COPY requirements.txt .
RUN pip install --user --no-cache-dir -r requirements.txt

# Final runtime stage (smaller image)
FROM arm32v7/python:3.11-slim

WORKDIR /app

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    mosquitto-clients \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy Python packages from builder
COPY --from=builder /root/.local /root/.local
ENV PATH=/root/.local/bin:$PATH

# Copy application code
COPY src/ ./src/
COPY config.yaml ./

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8080/health')" || exit 1

# Run application
CMD ["python", "-m", "uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8080"]
```

**Docker Compose for complete edge stack**:

```yaml
version: '3.8'

services:
  mqtt-broker:
    image: eclipse-mosquitto:2.0.18
    ports:
      - "1883:1883"
    volumes:
      - ./mosquitto.conf:/mosquitto/config/mosquitto.conf
      - mqtt-data:/mosquitto/data
    restart: unless-stopped
    
  edge-processor:
    build:
      context: .
      dockerfile: Dockerfile.arm32v7
    depends_on:
      - mqtt-broker
    environment:
      - MQTT_HOST=mqtt-broker
      - MQTT_PORT=1883
      - LOG_LEVEL=INFO
    volumes:
      - ./data:/app/data
      - ./models:/app/models
    ports:
      - "8080:8080"
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M

  time-series-db:
    image: arm32v7/influxdb:latest
    ports:
      - "8086:8086"
    environment:
      - INFLUXDB_DB=iot_metrics
      - INFLUXDB_ADMIN_USER=admin
      - INFLUXDB_ADMIN_PASSWORD=changeme
    volumes:
      - influxdb-data:/var/lib/influxdb
    restart: unless-stopped

volumes:
  mqtt-data:
  influxdb-data:
```

### 3.3 ML Inference on Edge - TensorFlow Lite

**Example 6: TensorFlow Lite Anomaly Detection on Edge**

```python
# pip install tensorflow==2.16.0 numpy==1.24.3
import tensorflow as tf
import numpy as np
import asyncio
from collections import deque
from datetime import datetime

class EdgeMLModel:
    """Load and run TensorFlow Lite models on edge device"""
    
    def __init__(self, model_path: str, input_shape: tuple = (10, 5)):
        # Load TFLite model
        self.interpreter = tf.lite.Interpreter(model_path=model_path)
        self.interpreter.allocate_tensors()
        
        # Get input/output details
        self.input_details = self.interpreter.get_input_details()
        self.output_details = self.interpreter.get_output_details()
        
        self.input_shape = input_shape
        
        # Sliding window buffer for sequence input
        self.sensor_buffer = deque(maxlen=input_shape[0])
    
    def add_sensor_reading(self, reading: np.ndarray) -> None:
        """Add reading to buffer (reading shape: (5,) for 5 sensors)"""
        self.sensor_buffer.append(reading)
    
    def predict(self) -> dict:
        """Run inference on buffered readings"""
        if len(self.sensor_buffer) < self.input_shape[0]:
            return {"status": "buffering", "progress": len(self.sensor_buffer)}
        
        # Prepare input tensor
        input_data = np.array(list(self.sensor_buffer), dtype=np.float32)
        input_data = np.expand_dims(input_data, 0)  # Add batch dimension
        
        # Set input tensor
        self.interpreter.set_tensor(
            self.input_details[0]['index'],
            input_data
        )
        
        # Run inference (typically 10-50ms on Raspberry Pi)
        self.interpreter.invoke()
        
        # Get output
        output_data = self.interpreter.get_tensor(
            self.output_details[0]['index']
        )
        
        # Parse results
        anomaly_score = float(output_data[0][0])
        is_anomaly = anomaly_score > 0.7  # Threshold
        
        return {
            "status": "success",
            "anomaly_score": anomaly_score,
            "is_anomaly": is_anomaly,
            "timestamp": datetime.utcnow().isoformat(),
            "buffer_size": len(self.sensor_buffer)
        }

class EdgeProcessor:
    """Main edge processing pipeline"""
    
    def __init__(self, model_path: str):
        self.model = EdgeMLModel(model_path)
        self.readings_count = 0
    
    async def process_sensor_reading(self, sensor_id: str, 
                                     values: list) -> dict:
        """Process incoming sensor data"""
        self.readings_count += 1
        
        # Convert to numpy array
        reading = np.array(values, dtype=np.float32)
        
        # Add to model buffer
        self.model.add_sensor_reading(reading)
        
        # Run inference every N readings (batching)
        if self.readings_count % 5 == 0:
            result = self.model.predict()
            
            if result["status"] == "success" and result["is_anomaly"]:
                # Alert on anomaly detection
                return {
                    "alert": True,
                    "sensor_id": sensor_id,
                    "anomaly_score": result["anomaly_score"],
                    "action": "NOTIFY_OPERATOR"
                }
            
            return result
        
        return {"status": "buffering"}

# Usage
async def main():
    processor = EdgeProcessor("models/anomaly_detector.tflite")
    
    # Simulate 5 sensor readings
    for i in range(50):
        readings = [20.5 + i*0.01, 45.0 + i*0.02, 1013.25, 65.5, 0.8]
        result = await processor.process_sensor_reading("sensor_01", readings)
        
        if result.get("alert"):
            print(f"ANOMALY DETECTED: {result}")
        
        await asyncio.sleep(0.5)

# asyncio.run(main())
```

### 3.4 Offline Operation & Sync

Edge devices must function independently when cloud connectivity is unavailable. Implementation patterns:

**Offline data queue**:
- Buffer all readings locally in SQLite
- Automatically sync when connection restored
- Prevent duplicate entries (check timestamps)

**Conflict resolution**:
- Cloud has newer data → Use cloud version
- Edge has newer data → Upload edge version
- Both changed → Merge strategy (average, latest-wins, user-defined)

---

## Section 4: Industrial IoT (IIoT) - Enterprise-Grade Systems

### 4.1 IIoT Architecture & Machine Monitoring

Industrial IoT differs from consumer IoT through higher reliability requirements, specialized protocols, and integration with legacy manufacturing systems.

**Key characteristics**:
- **Availability**: 99.99% uptime (30 seconds/month downtime max)
- **Real-time**: <10ms latency for machine control
- **Safety-critical**: Failures can endanger personnel
- **Legacy integration**: Must interface with 20+ year old equipment
- **Audit trail**: Complete history of all commands/changes

**Machine monitoring use cases**:

1. **Predictive maintenance**: ML models detect bearing degradation from vibration
2. **Production optimization**: Real-time throughput monitoring
3. **Quality control**: In-line defect detection
4. **Energy management**: Peak usage identification and reduction

**Example 7: Industrial Vibration Analysis**

```python
# pip install numpy==1.24.3 scipy==1.11.0
import numpy as np
from scipy import signal, fft
from dataclasses import dataclass

@dataclass
class VibrationMetrics:
    rms: float  # Overall energy
    peak: float  # Maximum displacement
    crest_factor: float  # Peak / RMS (anomaly indicator)
    frequencies: dict  # Dominant frequencies
    bearing_health: str  # GOOD / CAUTION / CRITICAL

class BearingHealthMonitor:
    """Detect bearing degradation through vibration analysis"""
    
    def __init__(self, sampling_rate: int = 10000):
        self.sampling_rate = sampling_rate
        self.normal_crest_factor = 4.0  # Typical healthy bearing
        self.alarm_crest_factor = 8.0   # Degrading bearing
    
    def analyze_vibration(self, acceleration: np.ndarray) -> VibrationMetrics:
        """Analyze accelerometer data for bearing health"""
        
        # Time-domain metrics
        rms = np.sqrt(np.mean(acceleration ** 2))
        peak = np.max(np.abs(acceleration))
        crest_factor = peak / rms if rms > 0 else 0
        
        # Frequency-domain analysis (FFT)
        fft_values = np.abs(fft.fft(acceleration))
        frequencies = fft.fftfreq(len(acceleration), 1/self.sampling_rate)
        
        # Find dominant frequencies
        dominant_indices = np.argsort(fft_values)[-5:]
        dominant_freqs = {
            f"peak_{i}": {
                "frequency_hz": float(frequencies[idx]),
                "magnitude": float(fft_values[idx])
            }
            for i, idx in enumerate(dominant_indices)
        }
        
        # Bearing health diagnosis
        if crest_factor < self.normal_crest_factor * 1.2:
            health = "GOOD"
        elif crest_factor < self.alarm_crest_factor * 0.8:
            health = "CAUTION"
        else:
            health = "CRITICAL"
        
        return VibrationMetrics(
            rms=float(rms),
            peak=float(peak),
            crest_factor=float(crest_factor),
            frequencies=dominant_freqs,
            bearing_health=health
        )

# Usage
def main():
    monitor = BearingHealthMonitor()
    
    # Simulate healthy bearing
    t = np.linspace(0, 1, 10000)
    healthy = 0.5 * np.sin(2*np.pi*50*t) + 0.1*np.random.randn(10000)
    
    metrics = monitor.analyze_vibration(healthy)
    print(f"Healthy bearing: CF={metrics.crest_factor:.2f}, Status={metrics.bearing_health}")
    
    # Simulate degraded bearing (high frequency components)
    degraded = 0.5 * np.sin(2*np.pi*50*t) + 2.0*np.sin(2*np.pi*500*t) + 0.2*np.random.randn(10000)
    
    metrics = monitor.analyze_vibration(degraded)
    print(f"Degraded bearing: CF={metrics.crest_factor:.2f}, Status={metrics.bearing_health}")

# main()
```

### 4.2 OPC-UA - Industrial Standard Protocol

OPC-UA (OLE for Process Control - Unified Architecture) is the standard for machine-to-machine communication in manufacturing.

**Key features**:
- **Information model**: Describes equipment, data types, and relationships
- **Real-time data**: 100+ messages/second
- **Security**: Certificates, encryption, role-based access
- **Legacy support**: Bridges old industrial equipment

**Example 8: OPC-UA Client Implementation**

```python
# pip install opcua==0.98.13
from opcua import Client
from opcua.common import callbacks
import asyncio
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class IndustrialOPCUAClient:
    """Connect to OPC-UA servers in manufacturing equipment"""
    
    def __init__(self, server_url: str = "opc.tcp://localhost:4840"):
        self.server_url = server_url
        self.client = None
        self.subscriptions = {}
    
    async def connect(self, username: str = None, password: str = None):
        """Establish secure connection to OPC-UA server"""
        self.client = Client(self.server_url)
        
        if username and password:
            self.client.set_user(username)
            self.client.set_password(password)
        
        try:
            self.client.connect()
            logger.info(f"Connected to {self.server_url}")
        except Exception as e:
            logger.error(f"Connection failed: {e}")
            raise
    
    async def read_variable(self, node_id: str):
        """Read single variable from OPC-UA server"""
        try:
            node = self.client.get_node(node_id)
            value = node.get_value()
            logger.info(f"Read {node_id}: {value}")
            return value
        except Exception as e:
            logger.error(f"Read failed for {node_id}: {e}")
            return None
    
    async def write_variable(self, node_id: str, value, data_type=None):
        """Write to OPC-UA variable"""
        try:
            node = self.client.get_node(node_id)
            if data_type:
                node.set_value(value, data_type)
            else:
                node.set_value(value)
            logger.info(f"Wrote {node_id}: {value}")
        except Exception as e:
            logger.error(f"Write failed for {node_id}: {e}")
    
    async def subscribe_variable(self, node_id: str, callback=None):
        """Subscribe to variable changes with callback"""
        sub = self.client.create_subscription(100, None)
        
        def handler(node, value, data):
            logger.info(f"Value changed: {node_id} = {value}")
            if callback:
                asyncio.run(callback(node_id, value))
        
        try:
            node = self.client.get_node(node_id)
            handle = sub.subscribe_data_change(node, handler)
            self.subscriptions[node_id] = (sub, handle)
            logger.info(f"Subscribed to {node_id}")
        except Exception as e:
            logger.error(f"Subscription failed for {node_id}: {e}")
    
    async def disconnect(self):
        """Gracefully disconnect"""
        for node_id, (sub, handle) in self.subscriptions.items():
            sub.unsubscribe(handle)
        
        if self.client:
            self.client.disconnect()
            logger.info("Disconnected from OPC-UA server")

# Usage
async def main():
    opcua_client = IndustrialOPCUAClient("opc.tcp://plc.example.com:4840")
    
    try:
        await opcua_client.connect(username="operator", password="secure_pass")
        
        # Read machine temperature
        temp = await opcua_client.read_variable("ns=2;s=Machine.Temperature")
        print(f"Current temperature: {temp}°C")
        
        # Subscribe to pressure changes
        async def on_pressure_change(node_id, value):
            print(f"Pressure changed to {value} bar")
        
        await opcua_client.subscribe_variable(
            "ns=2;s=Machine.Pressure",
            callback=on_pressure_change
        )
        
        await asyncio.sleep(60)
        
    finally:
        await opcua_client.disconnect()

# asyncio.run(main())
```

### 4.3 IIoT Security Architecture

**Example 9: Industrial IoT Security Framework**

```yaml
# IIoT Security Architecture (Comprehensive)
# Covers: Network segmentation, access control, audit logging, threat detection

# 1. Network Architecture
network_design: |
  # Segmented network with firewalls and VLANs
  
  [Internet] → [DMZ] → [Corporate LAN] → [MES VLAN] → [PLC VLAN] → [Devices]
                 ↓          (80)              (40)          (20)
             Firewall #1  Firewall #2
  
  VLAN Routing:
    MES VLAN (192.168.40.0/24): Manufacturing Execution System
    PLC VLAN (192.168.20.0/24): Industrial controllers
    Corporate (192.168.80.0/24): Business systems
  
  Access Rules:
    - MES → PLC: Limited (write only critical commands)
    - PLC → Cloud: Outbound only (no inbound)
    - Devices ↔ Each other: None (star topology)

# 2. Device Authentication
device_auth: |
  # Certificate-based mutual authentication
  
  Device Provisioning:
    1. Factory: Install unique certificate + private key
    2. Registration: Device sends CSR to IoT gateway
    3. Approval: Admin reviews + signs certificate
    4. Deployment: Device uses cert for all communication
  
  Certificate Properties:
    - Validity: 1-5 years (depends on device lifecycle)
    - Key size: 2048-bit RSA minimum (4096-bit recommended)
    - Revocation: CRL checked quarterly, OCSP for critical devices
    - Hardware storage: TPM/secure enclave when available

# 3. Message Integrity & Confidentiality
message_security: |
  # Protect data in transit and at rest
  
  Transport Layer:
    Protocol: TLS 1.3 (mandatory)
    Ciphers: TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305
    Certificates: X.509 v3 (mutual authentication)
  
  Application Layer:
    Sensitive fields: JSON Web Encryption (JWE)
    Algorithm: ECDH-ES+AES-256-KW + A256GCM
    Payload example:
      {
        "device_id": "PLC-001",
        "command": "set_pressure",
        "value_encrypted": "<JWE token>"
      }
  
  At Rest (Edge Database):
    Database: SQLite with SQLCipher
    Key derivation: PBKDF2 (64K iterations)
    Encryption: ChaCha20-Poly1305

# 4. Access Control & Authorization
access_control: |
  # Role-based access control (RBAC)
  
  Roles:
    - Operator: Start/stop machines, monitor
    - Technician: Modify settings, diagnose
    - Engineer: Write firmware, change configuration
    - Administrator: Full system access
  
  Permissions Matrix:
    Resource: temperature_sensor_01
    Operator: READ
    Technician: READ, WRITE(limits)
    Engineer: READ, WRITE, DELETE
    Administrator: *
  
  Enforcement:
    - Check role at gateway before forwarding to device
    - Device maintains local policy (defense in depth)
    - Audit log all decisions

# 5. Threat Detection & Intrusion Prevention
threat_detection: |
  # Detect suspicious behavior
  
  Anomalies Detected:
    - Unexpected command sequence (write permission change)
    - Out-of-range values (sensor reading 999°C)
    - Protocol violations (malformed MQTT)
    - Geographic anomaly (login from new country)
    - Frequency anomaly (10x normal message rate)
  
  Response Actions:
    Severity 1 (Critical):
      - Isolate device immediately
      - Alert security team
      - Capture full traffic for forensics
    
    Severity 2 (High):
      - Rate limit connection
      - Require additional authentication
      - Log all future activity
    
    Severity 3 (Medium):
      - Increase monitoring frequency
      - Add to watchlist (1 week)

# 6. Audit Logging & Compliance
audit_logging: |
  # Complete audit trail for regulatory compliance
  
  Logged Events:
    - Device authentication (success/failure)
    - Configuration changes + author
    - Command execution + parameters
    - Access denials + reason
    - Firmware updates + hash
    - Certificate expirations
  
  Retention:
    - Online: 1 year
    - Archive: 7 years (for compliance)
  
  Immutable storage:
    - Write-once filesystem (WORM)
    - Cryptographic signatures
    - Time-series database with tamper detection
```

---

## Section 5: Device Management at Scale

### 5.1 Device Provisioning & Configuration

Managing thousands of devices requires automated provisioning with zero manual intervention.

**Provisioning workflow**:

1. **Factory provisioning**: Device receives unique certificate
2. **Registration**: Device contacts IoT platform with CSR
3. **Approval**: Admin or automated system reviews
4. **Configuration**: Device downloads settings (WiFi, MQTT, policies)
5. **Deployment**: Device ready for field use

**Example 10: Azure IoT Hub Device Provisioning Service (DPS)**

```python
# pip install azure-iot-device==2.14.0
import asyncio
from azure.iot.device import IoTHubDeviceClient, X509
from azure.iot.device.provisioning.model import RegistrationResult
from azure.iot.device.provisioning.aio import ProvisioningDeviceClient

class IoTDeviceProvisioning:
    """Provision IoT devices to cloud platform"""
    
    def __init__(self, dps_endpoint: str, dps_id_scope: str,
                 cert_path: str, key_path: str):
        self.dps_endpoint = dps_endpoint
        self.dps_id_scope = dps_id_scope
        self.cert_path = cert_path
        self.key_path = key_path
    
    async def provision_device(self, device_id: str) -> dict:
        """Register device with DPS, receive IoT Hub connection string"""
        
        # Create X.509 credentials
        x509 = X509(
            cert_file=self.cert_path,
            key_file=self.key_path
        )
        
        # Create provisioning client
        provisioning_client = ProvisioningDeviceClient.create_from_x509_certificate(
            provisioning_host=self.dps_endpoint,
            registration_id=device_id,
            id_scope=self.dps_id_scope,
            x509=x509
        )
        
        try:
            # Register device
            registration_result = await provisioning_client.register()
            
            if registration_result.registration_state.assigned_hub:
                # Provisioning successful
                hub_host = registration_result.registration_state.assigned_hub
                device_id = registration_result.registration_state.device_id
                
                return {
                    "status": "provisioned",
                    "hub_host": hub_host,
                    "device_id": device_id,
                    "registered_at": registration_result.registration_state.created_date_time
                }
            else:
                return {
                    "status": "pending",
                    "message": registration_result.registration_state.status
                }
        
        except Exception as e:
            return {
                "status": "failed",
                "error": str(e)
            }
    
    async def connect_to_hub(self, hub_host: str, device_id: str) -> IoTHubDeviceClient:
        """Connect provisioned device to IoT Hub"""
        
        # Create device client
        device_client = IoTHubDeviceClient.create_from_x509_certificate(
            hostname=hub_host,
            device_id=device_id,
            x509=X509(
                cert_file=self.cert_path,
                key_file=self.key_path
            )
        )
        
        try:
            await device_client.connect()
            return device_client
        except Exception as e:
            print(f"Connection failed: {e}")
            return None

# Usage
async def main():
    provisioning = IoTDeviceProvisioning(
        dps_endpoint="global.azure-devices-provisioning.net",
        dps_id_scope="0ne00XXXXXXX",
        cert_path="/path/to/device-cert.pem",
        key_path="/path/to/device-key.pem"
    )
    
    # Provision device
    result = await provisioning.provision_device("device-001")
    print(f"Provisioning result: {result}")
    
    if result["status"] == "provisioned":
        # Connect to assigned hub
        client = await provisioning.connect_to_hub(
            result["hub_host"],
            result["device_id"]
        )
        
        if client:
            # Ready to send/receive messages
            print("Device connected and ready")
            await asyncio.sleep(10)
            await client.disconnect()
```

### 5.2 Firmware Updates (OTA)

Over-the-Air updates enable fixing bugs and deploying features without physical access.

**OTA workflow**:
1. Admin uploads firmware + hash to server
2. Device checks for updates periodically
3. Downloads firmware in chunks (network resilience)
4. Verifies cryptographic signature
5. Writes to boot partition
6. Reboots into new firmware
7. Rolls back if boot fails

### 5.3 Device Health Monitoring

**Metrics to track**:
- **Connectivity**: Online/offline status, connection duration
- **Resource usage**: CPU, memory, disk, battery level
- **Communication**: Message count, latency, error rate
- **Application**: Business metrics (devices operational)

---

## Section 6: Sensor Networks & Data Collection

### 6.1 Multi-Sensor Data Fusion

Real-world systems integrate data from multiple sensor types to form a comprehensive picture.

**Sensor types in IoT**:

| Sensor | Use Case | Range | Accuracy | Cost |
|--------|----------|-------|----------|------|
| **Temperature (DHT22)** | Environmental, HVAC | -40 to 80°C | ±0.5°C | $5 |
| **Humidity** | Climate control | 0-100% | ±3% | $5 |
| **Pressure (BMP280)** | Altitude, weather | 300-1100 hPa | ±1 hPa | $3 |
| **GPS** | Location tracking | Global | ±5m | $15 |
| **Accelerometer (IMU)** | Vibration, motion | ±16g | ±0.01g | $8 |
| **Light (LUX)** | Ambient light | 1-100K lux | ±10% | $2 |
| **Soil moisture** | Agriculture | 0-100% | ±3% | $8 |
| **Gas sensor (MQ-135)** | Air quality | 10-1000 ppm | ±5% | $10 |

**Example 11: Multi-Sensor Fusion Algorithm**

```python
# pip install numpy==1.24.3
import numpy as np
from dataclasses import dataclass
from datetime import datetime

@dataclass
class SensorReading:
    timestamp: str
    sensor_id: str
    value: float
    unit: str
    confidence: float = 1.0  # 0-1

class SensorFusion:
    """Combine multiple sensor inputs for robust estimates"""
    
    def __init__(self, fusion_method: str = "weighted_average"):
        self.readings = []
        self.fusion_method = fusion_method
    
    def add_reading(self, sensor_id: str, value: float,
                   unit: str, confidence: float = 1.0):
        """Add sensor reading with confidence weight"""
        reading = SensorReading(
            timestamp=datetime.utcnow().isoformat(),
            sensor_id=sensor_id,
            value=value,
            unit=unit,
            confidence=confidence
        )
        self.readings.append(reading)
    
    def fuse(self) -> dict:
        """Fuse multiple readings into single estimate"""
        
        if not self.readings:
            return {"error": "No readings"}
        
        values = [r.value for r in self.readings]
        confidences = [r.confidence for r in self.readings]
        
        if self.fusion_method == "weighted_average":
            # Higher confidence readings weighted more
            total_confidence = sum(confidences)
            fused_value = sum(v * c for v, c in zip(values, confidences)) / total_confidence
            
            # Fused confidence = average confidence
            fused_confidence = np.mean(confidences)
        
        elif self.fusion_method == "median":
            # Robust to outliers
            fused_value = np.median(values)
            fused_confidence = np.mean(confidences)
        
        else:  # simple_average
            fused_value = np.mean(values)
            fused_confidence = np.mean(confidences)
        
        return {
            "fused_value": float(fused_value),
            "fused_confidence": float(fused_confidence),
            "reading_count": len(self.readings),
            "unit": self.readings[0].unit,
            "timestamp": datetime.utcnow().isoformat(),
            "readings_used": [r.sensor_id for r in self.readings]
        }

# Usage
def main():
    fusion = SensorFusion(fusion_method="weighted_average")
    
    # Add readings from multiple temperature sensors
    fusion.add_reading("temp_01", 20.5, "celsius", confidence=0.95)  # New sensor
    fusion.add_reading("temp_02", 20.3, "celsius", confidence=0.85)  # Older sensor
    fusion.add_reading("temp_03", 21.2, "celsius", confidence=0.6)   # Unreliable
    
    result = fusion.fuse()
    print(f"Fused temperature: {result['fused_value']:.2f}°C (confidence: {result['fused_confidence']:.2f})")
```

### 6.2 Anomaly Detection in Sensor Data

**Example 12: Statistical Anomaly Detection**

```python
# pip install numpy==1.24.3 scipy==1.11.0
import numpy as np
from collections import deque
from scipy import stats

class AnomalyDetector:
    """Detect unusual sensor readings using statistical methods"""
    
    def __init__(self, window_size: int = 100,
                 z_score_threshold: float = 3.0):
        self.window = deque(maxlen=window_size)
        self.z_score_threshold = z_score_threshold
    
    def add_reading(self, value: float) -> dict:
        """Process incoming reading and detect anomalies"""
        
        self.window.append(value)
        
        if len(self.window) < 10:
            return {"status": "insufficient_data"}
        
        # Calculate statistics
        mean = np.mean(self.window)
        std = np.std(self.window)
        
        # Z-score: How many standard deviations from mean
        z_score = (value - mean) / std if std > 0 else 0
        
        # Determine if anomaly
        is_anomaly = abs(z_score) > self.z_score_threshold
        
        return {
            "value": value,
            "mean": float(mean),
            "std": float(std),
            "z_score": float(z_score),
            "is_anomaly": is_anomaly,
            "severity": "HIGH" if abs(z_score) > 5 else ("MEDIUM" if is_anomaly else "NORMAL")
        }

# Usage
detector = AnomalyDetector()

# Normal readings
for i in range(100):
    result = detector.add_reading(20.0 + np.random.normal(0, 0.5))

# Inject anomaly
result = detector.add_reading(35.0)  # Sudden jump
print(f"Anomaly detection: {result}")
```

### 6.3 Data Quality & Calibration

**Sensor calibration best practices**:

1. **Factory calibration**: Manufacturer-provided baseline
2. **Field calibration**: Periodic comparison with reference
3. **Cross-calibration**: Compare multiple sensors in same environment
4. **Drift compensation**: Adjust for gradual drift over time

**Maintenance schedule**:
- **Monthly**: Visual inspection, connectivity check
- **Quarterly**: Accuracy verification against reference
- **Annually**: Replace filters, recalibrate if drift detected

---

## Section 7: Cloud Integration & Scalability

### 7.1 Cloud Platform Selection

**Comparison of major IoT cloud platforms** (as of 2025):

| Platform | Strengths | Best For |
|----------|-----------|----------|
| **AWS IoT Core** | Massive scale, ML integration, wide ecosystem | Enterprise, high-volume |
| **Azure IoT Hub** | Enterprise integration, strong auth, hybrid | Corporate systems, local+cloud |
| **Google Cloud IoT** | Data analytics, BigQuery integration | Analytics-heavy use cases |
| **Mosquitto** | Open-source, lightweight, self-hosted | Full control, privacy-focused |
| **InfluxDB Cloud** | Time-series optimized, fast queries | High-frequency metrics |

### 7.2 Time-Series Database Selection

**Example 13: InfluxDB for IoT Metrics**

```python
# pip install influxdb-client==1.18.0
from influxdb_client import InfluxDBClient, Point
from influxdb_client.client.write_api import SYNCHRONOUS
from datetime import datetime
import asyncio

class TimeSeriesDatabase:
    """Store and query IoT metrics in InfluxDB"""
    
    def __init__(self, url: str, token: str, org: str, bucket: str):
        self.client = InfluxDBClient(
            url=url,
            token=token,
            org=org
        )
        self.write_api = self.client.write_api(write_type=SYNCHRONOUS)
        self.query_api = self.client.query_api()
        self.bucket = bucket
        self.org = org
    
    async def write_metric(self, measurement: str, tags: dict,
                          fields: dict, timestamp: str = None):
        """Write metric to InfluxDB"""
        
        point = Point(measurement)
        
        # Add tags (indexed, efficient filtering)
        for key, value in tags.items():
            point.tag(key, value)
        
        # Add fields (the actual data values)
        for key, value in fields.items():
            if isinstance(value, (int, float)):
                point.field(key, value)
        
        # Set timestamp
        if timestamp:
            point.time(timestamp)
        
        try:
            self.write_api.write(bucket=self.bucket, org=self.org, record=point)
        except Exception as e:
            print(f"Write failed: {e}")
    
    async def query_metrics(self, measurement: str, sensor_id: str,
                           time_range: str = "-1h") -> list:
        """Query historical metrics"""
        
        query = f"""
        from(bucket:"{self.bucket}")
          |> range(start: {time_range})
          |> filter(fn: (r) => r._measurement == "{measurement}")
          |> filter(fn: (r) => r.sensor_id == "{sensor_id}")
          |> sort(columns: ["_time"], desc: false)
        """
        
        try:
            result = self.query_api.query(query)
            
            data = []
            for table in result:
                for record in table.records:
                    data.append({
                        "timestamp": record.values.get("_time"),
                        "value": record.values.get("_value"),
                        "field": record.values.get("_field")
                    })
            
            return data
        except Exception as e:
            print(f"Query failed: {e}")
            return []
    
    async def get_latest_values(self, measurement: str) -> dict:
        """Get latest value per tag combination"""
        
        query = f"""
        from(bucket:"{self.bucket}")
          |> range(start: -1h)
          |> filter(fn: (r) => r._measurement == "{measurement}")
          |> last()
        """
        
        try:
            result = self.query_api.query(query)
            
            latest = {}
            for table in result:
                for record in table.records:
                    sensor_id = record.values.get("sensor_id", "unknown")
                    latest[sensor_id] = record.values.get("_value")
            
            return latest
        except Exception as e:
            print(f"Query failed: {e}")
            return {}

# Usage
async def main():
    tsdb = TimeSeriesDatabase(
        url="http://localhost:8086",
        token="your-token-here",
        org="your-org",
        bucket="iot_metrics"
    )
    
    # Write sensor readings
    for i in range(100):
        await tsdb.write_metric(
            measurement="temperature",
            tags={"sensor_id": "temp_01", "location": "building_A"},
            fields={"value": 20.0 + i*0.1, "humidity": 65.0},
            timestamp=datetime.utcnow().isoformat()
        )
    
    # Query historical data
    data = await tsdb.query_metrics("temperature", "temp_01", "-24h")
    print(f"Retrieved {len(data)} readings from last 24h")
    
    # Get latest values
    latest = await tsdb.get_latest_values("temperature")
    print(f"Latest temperatures: {latest}")
```

---

## Section 8: Real-World Applications & Use Cases

### 8.1 Smart Home Automation

**Example 14: Home Automation System Architecture**

```python
# Smart home coordinator - integrates lighting, HVAC, security
import asyncio
from datetime import datetime
from typing import Dict, List

class SmartHomeController:
    """Orchestrate smart home devices"""
    
    def __init__(self, mqtt_host: str):
        self.mqtt_host = mqtt_host
        self.devices = {}
        self.automation_rules = []
        self.occupancy_state = {}
    
    async def register_device(self, device_id: str, device_type: str,
                             location: str, capabilities: List[str]):
        """Register smart home device"""
        self.devices[device_id] = {
            "type": device_type,
            "location": location,
            "capabilities": capabilities,
            "state": "disconnected"
        }
    
    async def create_automation_rule(self, rule_name: str,
                                    condition: dict,
                                    action: dict):
        """Create automation rule (e.g., "If sunset, turn on lights")"""
        rule = {
            "name": rule_name,
            "condition": condition,  # {"type": "time", "value": "sunset"}
            "action": action,        # {"device": "light_01", "command": "on"}
            "enabled": True
        }
        self.automation_rules.append(rule)
    
    async def on_occupancy_change(self, location: str, occupied: bool):
        """Handle room occupancy detection"""
        self.occupancy_state[location] = {
            "occupied": occupied,
            "timestamp": datetime.utcnow().isoformat()
        }
        
        if occupied:
            # Turn on lights
            await self.send_command("light_01", "on")
            # Set comfortable temperature
            await self.send_command("thermostat_01", {"setpoint": 22.0})
        else:
            # Turn off lights after 5 minutes
            await asyncio.sleep(300)
            await self.send_command("light_01", "off")
    
    async def send_command(self, device_id: str, command: str | dict):
        """Send command to device via MQTT"""
        topic = f"home/devices/{device_id}/command"
        payload = {"command": command, "timestamp": datetime.utcnow().isoformat()}
        print(f"[{topic}] {payload}")

# Usage
async def main():
    home = SmartHomeController("mqtt.home.local")
    
    # Register devices
    await home.register_device("light_01", "light", "living_room", ["on/off", "brightness"])
    await home.register_device("thermostat_01", "thermostat", "living_room", ["temp_setpoint"])
    
    # Create automation: "Goodnight" scene
    await home.create_automation_rule(
        rule_name="goodnight",
        condition={"type": "button", "device": "goodnight_button"},
        action={"command": "set_scene", "scene": "sleep"}
    )
```

### 8.2 Smart City Infrastructure

**Smart city metrics**:
- **Traffic**: Vehicle count, flow rate, congestion index
- **Air quality**: PM2.5, NO2, O3 levels (WHO standards)
- **Parking**: Availability, utilization rates
- **Utilities**: Water consumption, power demand, waste levels

### 8.3 Precision Agriculture

**Example 15: Precision Agriculture Data Pipeline**

```python
# Precision agriculture - optimize crop yield through sensor data
import asyncio
from datetime import datetime, timedelta

class PrecisionAgricultureSystem:
    """Monitor and optimize agricultural operations"""
    
    def __init__(self):
        self.field_zones = {}
        self.irrigation_schedule = {}
        self.ml_model = None  # Load pre-trained yield prediction model
    
    async def monitor_field_zone(self, zone_id: str,
                                 soil_moisture: float,
                                 soil_temp: float,
                                 rainfall: float) -> dict:
        """Analyze field conditions and make irrigation decisions"""
        
        # Optimal ranges
        optimal_moisture = (40, 60)  # 40-60% saturation
        
        # Decision logic
        if soil_moisture < optimal_moisture[0]:
            # Soil dry - irrigation needed
            irrigation_hours = (optimal_moisture[1] - soil_moisture) * 0.2
            action = "IRRIGATE"
        elif soil_moisture > optimal_moisture[1]:
            # Soil saturated - stop irrigation
            action = "STOP_IRRIGATION"
            irrigation_hours = 0
        else:
            action = "MONITOR"
            irrigation_hours = 0
        
        # Predict crop yield impact
        estimated_impact = 1.0
        if soil_moisture < optimal_moisture[0]:
            estimated_impact = 0.7  # 30% yield loss if underwatered
        
        return {
            "zone_id": zone_id,
            "soil_moisture": soil_moisture,
            "soil_temperature": soil_temp,
            "rainfall": rainfall,
            "recommendation": action,
            "irrigation_hours": irrigation_hours,
            "estimated_yield_impact": estimated_impact,
            "timestamp": datetime.utcnow().isoformat()
        }

# Usage
async def main():
    ag = PrecisionAgricultureSystem()
    
    # Simulate field monitoring
    result = await ag.monitor_field_zone(
        zone_id="field_a_1",
        soil_moisture=35.0,  # Below optimal
        soil_temp=18.5,
        rainfall=0.0
    )
    print(f"Irrigation decision: {result}")
```

---

## Section 9: Best Practices & Production Patterns

### 9.1 Scalability Principles

**Design for scale from day one**:

1. **Stateless edge devices**: Device retains no session state
2. **Load balancing**: Multiple brokers behind load balancer
3. **Horizontal scaling**: Add devices without changing architecture
4. **Rate limiting**: Protect backend from device storms
5. **Graceful degradation**: System works even if some components offline

### 9.2 Security-First Approach

**Security checklist**:
- All communication encrypted (TLS 1.3)
- Certificate-based authentication (no passwords)
- Role-based access control enforced
- Audit logging of all sensitive operations
- Regular security patches (monthly minimum)
- Penetration testing (quarterly)

### 9.3 Power Efficiency

**Battery-powered device optimization**:

| Technique | Power Savings |
|-----------|--------------|
| **Reduced sampling rate** | 40-60% |
| **Local processing** | 20-30% (avoid cloud transmissions) |
| **Sleep modes** | 70-90% (when inactive) |
| **Efficient protocols** (BLE > WiFi) | 50-80% |
| **Data compression** | 10-30% |

### 9.4 Reliability & Fault Tolerance

**High-availability patterns**:

1. **Message queuing**: Local buffer if cloud unavailable
2. **Circuit breaker**: Stop retry attempts after N failures
3. **Exponential backoff**: 1s → 2s → 4s → 8s delay between retries
4. **Health monitoring**: Detect and alert on failures
5. **Auto-recovery**: Automatic restart of failed services

---

## TRUST 5 Compliance

| Principle | Implementation |
|-----------|---|
| **Test-First** | Unit tests for sensor drivers, MQTT handlers, edge algorithms |
| **Readable** | Clear sensor schemas, documented protocols, architectural diagrams |
| **Unified** | Standard MQTT format, common device interfaces |
| **Secured** | TLS, certificates, ACL, encrypted storage |
| **Trackable** | Device provenance, firmware versions, command audit log |

---

## Quick Reference Table

| Pattern | When to Use | Example |
|---------|------------|---------|
| **MQTT QoS 0** | Non-critical sensor data | Temperature readings |
| **MQTT QoS 1** | Standard IoT | Device status, metrics |
| **MQTT QoS 2** | Critical commands | Emergency shutdown |
| **Edge ML** | <50ms latency required | Real-time anomaly detection |
| **Time-series DB** | High-frequency metrics | Sensor data collection |
| **Cloud ML** | Complex patterns | Predictive maintenance |

---

## Related Skills

- `moai-lang-python` - Python 3.13 implementation patterns
- `moai-domain-backend` - API design for IoT services
- `moai-domain-security` - OWASP Top 10 for IoT
- `moai-essentials-perf` - Performance optimization techniques

## Resources

- **MQTT Specification**: https://mqtt.org/mqtt-specification
- **Mosquitto Documentation**: https://mosquitto.org/documentation/
- **OPC-UA Standard**: https://opcfoundation.org/
- **IEC 62443**: Industrial Cybersecurity Standard
- **NIST IoT Framework**: https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-184.pdf

---

**Version**: 4.0.0  
**Last Updated**: 2025-11-19  
**Status**: Stable - Production Ready  
**Author**: Alfred SuperAgent  
**Language**: English (Technical Infrastructure)
