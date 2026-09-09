# Greenhouse IoT Backend API
Backend API for a greenhouse IoT system built with Go. This project provides REST APIs for device management and control, integrates with an MQTT broker for real-time device communication, and uses PostgreSQL for persistent data storage.

# The system is designed around a simple IoT communication flow:
```
Client / IoT Device
        |
        | HTTP Request
        v
+-------------------+
|   Go REST API     |
|      Echo         |
+---------+---------+
          |
          +----------------------+
          |                      |
          v                      v
   +-------------+        +-------------+
   | PostgreSQL  |        | MQTT Broker |
   +-------------+        +------+------+
                                 |
                                 | MQTT Publish
                                 v
                         Greenhouse Device
```

# Features
1. REST API built with Go and Echo
2. PostgreSQL database integration
3. MQTT broker integration using Eclipse Paho MQTT
4. Device control through MQTT
5. Device and device-report data models
6. Request validation
7. Centralized HTTP error handling
8. Request ID middleware
9. CORS support
10. Gzip response compression
11. Application logging
12. Panic recovery
13. Redis cache support
14. Configurable application host and port
15. MQTT connection status checking
16. Automatic MQTT reconnection

# Tech Stack
| Technology        | Purpose                          |
| ----------------- | -------------------------------- |
| Go                | Backend programming language     |
| Echo              | HTTP web framework               |
| PostgreSQL        | Persistent database              |
| MQTT              | Real-time communication protocol |
| Eclipse Paho MQTT | MQTT client library              |
| Mosquitto         | Local MQTT broker                |
| Redis             | Optional caching                 |
| Viper             | Configuration management         |
| Zerolog           | Structured logging               |

# Project Structure
```
green-house-api/
│
├── api/
│   ├── config/
│   │   └── redis/
│   │
│   ├── handler/
│   ├── middleware/
│   ├── mqtt/
│   ├── request/
│   ├── response/
│   ├── route/
│   └── usecase/
│
├── helper/
│   ├── cache/
│   ├── jwt/
│   ├── logger/
│   ├── postgre/
│   ├── response/
│   ├── validator/
│   └── viper/
│
├── model/
│   ├── device_identity.go
│   ├── device_report.go
│   ├── user.go
│   └── ...
│
├── repository/
│   ├── device_control_repo.go
│   ├── user_device_repo.go
│   └── ...
│
├── router/
│   └── router.go
│
├── main.go
├── go.mod
└── README.md
```

# The application separates responsibilities into several layers:
# Handler
Responsible for receiving HTTP requests, binding request payloads, validating input, and returning HTTP responses.

# Usecase
Contains application/business logic. The usecase coordinates operations between repositories and external services such as MQTT.

# Repository
Responsible for communication with the PostgreSQL database.

# MQTT
Contains the MQTT client implementation used by the application to connect to the MQTT broker and publish messages.

# Model
Contains the application's data structures and database-related models.

# Route
Registers and groups the application's HTTP endpoints.

# The application follows a layered architecture:
```
HTTP Request
     |
     v
  Router
     |
     v
  Handler
     |
     v
  Usecase
    / \
   /   \
  v     v
Repository   MQTT Client
   |             |
   v             v
PostgreSQL    MQTT Broker
```

This separation keeps the HTTP layer, business logic, database access, and MQTT communication independent from each other.

# MQTT Integration
MQTT is used for real-time communication between the backend and greenhouse devices.
The application uses Eclipse Paho MQTT as the MQTT client library.
The MQTT connection is initialized once when the application starts.

```
Application Startup
       |
       v
Read MQTT configuration
       |
       v
Connect to MQTT Broker
       |
       v
Create MQTT Client
       |
       v
Inject MQTT Client
       |
       v
Start HTTP Server
```

The MQTT client is then reused by subsequent requests.
This avoids creating a new MQTT connection for every API request.
When the application shuts down, the MQTT client is disconnected gracefully.
The MQTT client also enables automatic reconnection when the broker connection is lost.

# For a device control operation, the flow is:
```
Client
  |
  | POST device control request
  v
Handler
  |
  v
Usecase
  |
  | Build MQTT topic
  | Build JSON payload
  v
MQTT Client
  |
  | Publish
  v
MQTT Broker
  |
  v
Greenhouse Device
```

Example MQTT payload:

```
Topic : greenhouse/test
Url : /api/v1/report/device-control
Method : POST

{
  "device_id": 1,
  "switch_status": "OFF",
  "temperature": 28.9,
  "humidity": 77.1
}
```

The backend publishes the message to the broker and the greenhouse device can subscribe to the corresponding topic.

# MQTT Quality of Service
The current implementation uses MQTT QoS 0.
This means the message is delivered without requiring MQTT delivery acknowledgement.
This is suitable for the current assignment and simple real-time device control demonstration.
For production systems, the QoS level should be selected based on the required reliability and device-control semantics.

# The MQTT client is created during application startup:
```
mqtt, err := mqttClient.NewClient(brokerUrl, brokerClientID)
```

The client is then assigned to the application helper and passed through the router to the API layer.
```
main.go
   |
   v
MQTT Client
   |
   v
Router
   |
   v
Route
   |
   v
Usecase
```

The MQTT client is not recreated for every request.
This approach reduces connection overhead and allows the backend to maintain a persistent connection with the broker.

# Device Control
The device control functionality is responsible for sending commands to greenhouse devices through MQTT.
The request is handled by the following flow:
```
HTTP Request
     |
     v
Device Control Handler
     |
     v
Device Control Usecase
     |
     +----------------+
     |                |
     v                v
Repository        MQTT Client
     |                |
     v                v
PostgreSQL        MQTT Broker
```

The repository is responsible for device-related database operations, while the MQTT client handles communication with the MQTT broker.
This separation allows the business logic to coordinate both database operations and real-time device communication without coupling the repository directly to MQTT.

# Database
The application uses PostgreSQL as the primary persistent database.
The main device-related models include:

# Device Identity
Represents the registered greenhouse device.
Example attributes include:
```
id
code
name
batch
is_active
in_time
```

# Device Report
Stores information reported by a device.
Example attributes include:
```
id
device_id
switch_status
device_status
database_status
mqtt_status
temperature
humidity
```

The relationship can be represented as:
```
Device Identity
      |
      | 1
      |
      | N
      v
Device Reports
```

One device can have multiple reports over time.

# Configuration
The application reads its configuration through the project's configuration helper.
Important configuration groups include:
1. app
2. broker
3. database.postgre

# Example configuration values:
```
app.host
app.port
app.name
app.debug

broker.host
broker.port

database.postgre.db_master.*
database.postgre.db_main_master.*
database.postgre.db_report_master.*
```

Sensitive credentials should not be committed to the repository.

# Running the Application
Requirements

Make sure the following software is installed:
1. Go
2. PostgreSQL
3. Mosquitto MQTT Broker
4. Redis (optional, depending on cache configuration)

Steps :
1. Clone Repository
2. Install Go Dependencies
3. Configure PostgreSQL
Create the required PostgreSQL database and configure the database connection according to the application's configuration.
The application currently separates PostgreSQL connections into:
1. db_master
2. db_main_master
3. db_report_master

Here are the queries :
```
CREATE TABLE device_identities (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    batch VARCHAR(255) NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    in_time TIMESTAMP NOT null,
    created_by BIGINT NULL,
    created_at TIMESTAMP NULL,
    updated_by BIGINT NULL,
    updated_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,
    deleted_at TIMESTAMP NULL
);

CREATE TABLE device_reports (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL,
    switch_status BOOLEAN NOT NULL DEFAULT FALSE,
    device_status VARCHAR(255) NULL,
    database_status VARCHAR(255) NULL,
    mqtt_status VARCHAR(255) NULL,
    temperature DOUBLE PRECISION NULL,
    humidity DOUBLE PRECISION null,
    created_by BIGINT NULL,
    created_at TIMESTAMP NULL,
    updated_by BIGINT NULL,
    updated_at TIMESTAMP NULL,
    deleted_by BIGINT NULL,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_device_reports_deleted_at
    ON device_reports (deleted_at);

CREATE INDEX idx_device_reports_device_id
    ON device_reports (device_id);
```

Make sure the configured database credentials and hosts are available before starting the application.

# Running Mosquitto
Make sure the MQTT broker is running on the configured host and port.
For a local Mosquitto installation, the default MQTT port is 1883.

Verify that the broker is running.
You can also subscribe to MQTT messages using:
```
mosquitto_sub -h localhost -p 1883 -t "greenhouse/control/#" -v
```

The subscriber will listen for all greenhouse device control messages.

# Running the Backend
Start the application using:
```
go run main.go
```

The application will:
1. Load application configuration
2. Initialize the Echo HTTP server
3. Configure middleware
4. Connect to the MQTT broker
5. Initialize PostgreSQL connections
6. Initialize optional Redis cache
7. Register API routes
8. Start the HTTP server

# Testing MQTT Device Control
Start an MQTT subscriber:
```
mosquitto_sub -h localhost -p 1883 -t "greenhouse/control/#" -v
```

Then send a device-control request to the backend.
The backend will:
```
HTTP Request
     |
     v
Validate Request
     |
     v
Device Control Usecase
     |
     v
Create MQTT Topic
     |
     v
Publish JSON Payload
     |
     v
Mosquitto Broker
     |
     v
MQTT Subscriber / Device
```

# Error Handling
The application provides centralized HTTP error handling.
Errors are returned in a consistent JSON structure:

```
{
  "code": 400,
  "data": "record not found",
  "field": null,
  "message": "record not found",
  "status": "error"
}
```

The application also uses recovery middleware to prevent an unexpected panic inside an HTTP handler from terminating the entire service.
When a panic occurs, the application records the error and returns an HTTP 500 response.

# Logging
The application provides application logging and panic logging.
Logs are written to the application's logs directory.
The application also includes request-related information when running in debug mode, including:
1. request_id
2. user_agent
3. remote
4. method
5. path
6. query
7. status
8. latency

# Middleware
The HTTP server uses several Echo middlewares, including:
1. CORS
2. Request ID
3. Gzip
4. Body limit
5. Recovery
6. Custom recovery
7. Request logging in debug mode

The default request body limit is configured to 200 MB.

# Design Decisions
1. Layered Architecture
The application separates HTTP handling, business logic, database access, and MQTT communication.
This makes the code easier to maintain and allows individual components to be tested independently.

```
Handler
   |
   v
Usecase
   |
   +------> Repository ------> PostgreSQL
   |
   +------> MQTT Client -----> MQTT Broker
```

2. Persistent MQTT Connection
The MQTT connection is established once during application startup rather than connecting and disconnecting for every request.
This reduces connection overhead and is more suitable for a backend that may handle frequent device-control requests.

3. MQTT as a Communication Layer
The backend does not communicate directly with a physical device.
Instead:

```
Backend
   |
   v
MQTT Broker
   |
   v
Device
```

This allows devices to communicate asynchronously and makes it possible to add additional devices without tightly coupling them to the backend.

4. Repository Separation
Database operations are kept inside repository components.
This prevents SQL/database implementation details from leaking into the HTTP handler layer.

5. Usecase as Application Coordinator
The usecase layer is responsible for coordinating business operations.
For example, device control may involve:

```
Validate Device
      |
      v
Process Command
      |
      v
Publish MQTT Message
      |
      v
Return Result
```

This keeps business rules outside the HTTP handler.

# Real-Time Communication
MQTT provides the real-time communication mechanism between the backend and greenhouse devices.
Unlike a traditional request/response-only architecture:

Client -> Backend -> Database

the system can also communicate asynchronously:

Backend -> MQTT Broker -> Device

This architecture is useful for IoT systems because devices can subscribe to topics and receive commands without the backend needing to maintain a direct connection to each physical device.

# Future Improvements
The current implementation can be extended with:
1. MQTT topic subscription for incoming sensor data
2. Sensor data ingestion through MQTT
3. Device online/offline status tracking
4. MQTT Last Will and Testament (LWT)
5. Authentication and authorization
6. Database migrations
7. Automated unit and integration tests
8. Docker Compose for PostgreSQL, Mosquitto, Redis, and the API
9. API documentation using OpenAPI/Swagger
10. Graceful HTTP server shutdown
11. Health check for PostgreSQL and MQTT connection
12. Message retry or QoS strategy depending on device requirements
13. Structured observability using metrics and tracing

# Project Status
This project was developed as a backend engineering assignment for an IoT greenhouse system.
The implementation focuses on:
1. Clean backend structure
2. REST API design
3. PostgreSQL integration
4. MQTT integration
5. Device control
6. Real-time communication concepts
7. Error handling
8. Logging
9. Maintainable code organization
