---
name: moai-domain-iot
description: IoT architecture, device management, and edge computing patterns for connected devices
allowed-tools: [Read, Bash, WebFetch]
---

# IoT Domain Expert

## Quick Reference

IoT architecture patterns for device connectivity, data ingestion, edge computing, and cloud integration with MQTT, CoAP, and LoRaWAN protocols.

**Core Technologies**:
- **Protocols**: MQTT, CoAP, LoRaWAN, NB-IoT
- **Edge**: AWS IoT Greengrass, Azure IoT Edge
- **Cloud**: AWS IoT Core, Azure IoT Hub, Google IoT Core
- **Data**: Time-series databases (InfluxDB, TimescaleDB)

---

## Implementation Guide

**MQTT Device Connection**:
```python
import paho.mqtt.client as mqtt

client = mqtt.Client()
client.connect("broker.example.com", 1883)
client.publish("sensor/temperature", "23.5")
```

**Edge Computing Pattern**:
```
Device → Edge Gateway (AWS Greengrass) → Cloud (AWS IoT Core)
├─ Local processing (reduce latency)
├─ Data filtering (reduce bandwidth)
└─ Offline operation (resilience)
```

---

## Best Practices

### ✅ DO
- Implement device authentication (X.509 certificates)
- Use TLS for all MQTT connections
- Design for intermittent connectivity
- Implement OTA firmware updates

### ❌ DON'T
- Use unencrypted protocols (security risk)
- Skip device provisioning workflow
- Ignore power consumption optimization

---

**Version**: 1.0.0 | **Last Updated**: 2025-11-21
