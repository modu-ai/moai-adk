# IoT Resources & References - Enterprise Ecosystem

**Version**: 4.0.0 (2025-11-22)

---

## Protocols & Standards

### MQTT (Message Queuing Telemetry Transport)
- **Standard**: OASIS MQTT 3.1.1, MQTT 5.0 (2019)
- **Port**: 1883 (standard), 8883 (TLS)
- **Brokers**: Mosquitto, HiveMQ, Emqx, RabbitMQ
- **Best for**: Low-bandwidth IoT, battery-operated devices
- **Max message size**: Configurable (typically 256MB)
- **Key features**: Pub/Sub, QoS levels (0, 1, 2)

### CoAP (Constrained Application Protocol)
- **RFC**: 7252 (2014), RFC 8323 (8.03 over TCP)
- **Port**: 5683 (UDP), 5684 (DTLS)
- **Resources**: Single request-response per connection
- **Max payload**: 1024 bytes (limited by block transfers)
- **Best for**: Constrained devices, LTE-M, NB-IoT
- **Observe pattern**: Real-time notifications

### HTTP/HTTPS
- **Versions**: HTTP/1.1, HTTP/2, HTTP/3 (QUIC)
- **Typical overhead**: High (headers, connection management)
- **Best for**: Occasional updates, large payloads
- **Connection**: Stateless (no persistent connection)
- **Compression**: GZIP, Brotli supported

### LTE-M & NB-IoT
- **Coverage**: 20dB better than LTE
- **Bandwidth**: LTE-M 1Mbps, NB-IoT 250kbps
- **Latency**: LTE-M ~100ms, NB-IoT ~10s
- **Battery**: 6-10 years typical
- **Carriers**: Most major carriers worldwide

---

## Cloud IoT Platforms

### AWS IoT Core
- **Features**: Device registry, rules engine, fleet provisioning
- **Pricing**: Per message + GB data out
- **Limits**: 100,000 simultaneous connections/account
- **Protocols**: MQTT, HTTP, AMQP

### Azure IoT Hub
- **Features**: Device twin, device-to-cloud, cloud-to-device messaging
- **Pricing**: Per device per day
- **Limits**: 1,000,000 devices/hub
- **Protocols**: MQTT, AMQP, HTTP

### Google Cloud IoT
- **Features**: Cloud Pub/Sub integration, device management
- **Pricing**: Per message
- **Deprecated**: In favor of Google Cloud IoT Core alternatives

### ThingsBoard
- **Type**: Open-source IoT platform
- **Licensing**: Apache 2.0 (free), Enterprise ($$$)
- **Features**: Dashboard, rules engine, multi-tenancy
- **Deployment**: Cloud, on-prem, docker-compose

---

## Device Operating Systems

### MicroPython
- **Target**: Microcontrollers (ESP8266, ESP32, Pyboard)
- **Memory**: Typically 32KB-4MB RAM
- **Power**: Ultra-low, ideal for battery devices
- **Example**: Raspberry Pi Pico, WeMos D1 Mini

### CircuitPython
- **Target**: Adafruit boards, educational
- **Memory**: 32KB-2MB
- **Features**: USB boot protocol, easier debugging
- **Community**: Strong educational focus

### TinyOS
- **Language**: NesC
- **Target**: Wireless sensor networks
- **Memory**: 512B-64KB RAM
- **Use**: Research, WSN testbeds

### RIOT OS
- **Target**: Low-power IoT devices
- **Standards**: 6LoWPAN, CoAP, LwM2M support
- **License**: LGPL 2.1
- **Memory**: 4KB+ RAM

---

## Edge Computing Frameworks

### TensorFlow Lite
- **Model size**: 1-100MB (optimized)
- **Inference speed**: 10-1000ms (device dependent)
- **Supported platforms**: Android, iOS, Raspberry Pi, MCU
- **Quantization**: INT8 (4x smaller, slight accuracy loss)

### OpenVINO
- **Intel's framework for edge inference
- **Languages**: C++, Python
- **Platforms**: x86, ARM
- **Optimization**: Hardware-specific (GPU, VPU, FPGA)

### Apache MXNet
- **Lightweight ML framework
- **Edge optimization**: Model compression, pruning
- **Language bindings**: Python, R, C++

---

## Local Development & Testing

### MQTT Test Tools
```bash
# Command-line MQTT client
mosquitto_pub -h localhost -t "test/topic" -m "hello"
mosquitto_sub -h localhost -t "test/topic"

# Docker MQTT broker
docker run -p 1883:1883 eclipse-mosquitto:latest
```

### CoAP Clients
- **libcoap**: C library with CLI
- **coap-cli**: Node.js command-line tool
- **Postman**: Has CoAP support (v9.0+)

### Firmware Flashing Tools
- **esptool.py**: For ESP8266/ESP32
- **pyocd**: ARM Cortex-M devices
- **avrdude**: Atmel microcontrollers

---

## Performance Benchmarks

### Data Transmission Sizes (typical)
- MQTT temperature reading: 50-80 bytes
- CoAP sensor update: 30-50 bytes (binary)
- JSON REST API: 200-500 bytes
- Compression savings: 40-70% with gzip/zlib

### Battery Life Estimates
- Sleep only (no processing): 10 years
- 1 reading/day, WiFi TX: 1-2 years
- 1 reading/hour, LTE-M: 3-5 years
- 1 reading/min, Bluetooth: 3-6 months

### Network Latency (typical)
- Local WiFi: 10-50ms
- Cloud MQTT: 50-200ms
- LTE-M: 100-500ms
- NB-IoT: 1-10 seconds

---

## Community Resources

### GitHub Projects
- **MQTT.js**: JavaScript MQTT client
- **paho-mqtt**: Python reference implementation
- **CoAP-18**: Python CoAP library
- **Home Assistant**: Open-source home automation (IoT integration)

### Tutorials & Guides
- **HiveMQ**: MQTT essentials (mqtt-essentials.com)
- **IOT.Eclipse**: CoAP protocol introduction
- **TensorFlow Lite Micro**: Getting started with edge AI
- **Arduino Project Hub**: 1000+ IoT projects

### Organizations & Standards
- **OpenConnectivity Foundation (OCF)**: IoT interoperability
- **IETF**: MQTT, CoAP, 6LoWPAN standards
- **OASIS**: MQTT standards body
- **IEEE 802.15.4**: Wireless sensor network standard

---

## Context7 Integration

### Related Libraries & Tools
- [Mosquitto](/eclipse-mosquitto/mosquitto): MQTT broker and client library
- [MQTT.js](/mqttjs/MQTT.js): JavaScript MQTT client
- [paho-mqtt](/eclipse/paho.mqtt.python): Python MQTT client
- [TensorFlow Lite](/tensorflow/tflite): ML on edge devices
- [CoAP](/obgm/libcoap): C library for CoAP protocol

### Official Documentation
- [MQTT Specification](https://mqtt.org/mqtt-specification)
- [CoAP RFC 7252](https://datatracker.ietf.org/doc/html/rfc7252)
- [TensorFlow Lite Documentation](https://www.tensorflow.org/lite/guide)
- [Arduino IoT Guide](https://docs.arduino.cc/cloud/)

### Version-Specific Guides
**Latest stable versions** (2025-11-22):
- MQTT: 5.0 (OASIS standard)
- CoAP: RFC 8323 (CoAP over TCP/TLS)
- ESP8266: 3.1 SDK, MicroPython 1.21+
- ESP32: IDF 5.1+, MicroPython 1.21+

---

## Recommended Learning Path

1. **Start**: Basic MQTT with mosquitto_pub/sub
2. **Learn**: Protocol comparison (MQTT vs CoAP vs HTTP)
3. **Build**: Simple sensor with Arduino + MQTT
4. **Scale**: Multi-device MQTT with QoS handling
5. **Optimize**: Battery optimization strategies
6. **Advance**: Device shadows, OTA updates
7. **Edge**: TensorFlow Lite inference on MCU
8. **Enterprise**: Cloud platform integration

---

**Related Skills**: moai-domain-backend, moai-domain-database, moai-essentials-perf, moai-domain-devops
