# Objective

The goal is to develop a Point-of-Sale (POS) subsystem to enhance the overall buying experience. Your subsystem will process events generated by a test rig you create. These events will subsequently be consumed and acted upon by specialized programs called "plugins" (details to follow). Actions taken by these specialized programs (plugins) may include publishing more events, writing to a database, making API calls, or a combination of these.

# Technical Specifications

- Event Bus: Utilize Apache Kafka as the event bus.

- Event Generation: Create a program to generate and publish bulk events to Apache Kafka. Events will be of various types (more details to follow).

- Event Consumption: Write consumers to process these events using at least two different programming languages from the following options: GoLang, Python, Node.js, and Java.

- Database Choices: For database requirements, you can choose between MySQL and PostgreSQL for relational databases. MongoDB and Redis are also options where applicable.

- Plugins: Create plugins that consume events. Each plugin can either be active or inactive. When inactive, it should not consume events.

- Plugin Configuration: Each plugin will contain settings to determine the suitability of an event for further processing (details to follow).

- Actions: If an event qualifies based on the plugin's settings and the plugin is active, actions may include:

- Making an HTTP API call

- Writing to a database

- Generating and publishing one or more new events

- Writing to a debug log

# Frontend

Develop a straightforward frontend to display the list of plugins. This UI should allow the toggling of individual plugins and modification of event eligibility settings. Changes should take immediate effect.

# Development Practices

You are encouraged to demonstrate development best practices such as TDD, CI/CD, Technical Documentation, Git Hygiene etc. The more expertise you showcase, the higher your working session will be rated.

# Deployment

You can deploy the system on your local computer or an external free service.

# Reporting

Provide a daily written update in addition to our daily check-in session. This will help us evaluate your written communication skills.

# Github

We will provide a private Github repo where you will be expected to commit your progress at regular intervals.

# Desired Plugins

- Employee Time Tracker: Logs and calculates total time spent by employees at a POS terminal. Auto-logout should occur if an employee logs in at a new terminal without explicitly logging out first.

- Purchase Recommender: Detects item addition events and makes hard-coded or dynamic recommendations for additional items.

- Customer Lookup: Identifies customer identifier events, fetches customer data from a remote system, and publishes it back on the bus.

- Fraud Detection: Maintains a state of multiple events and publishes an alert based on configured rules.

- Age Verification: Determines if age verification is required when certain items are purchased.

# Test Events

Your test rig should generate the following dummy events in a logical sequence, additional events can be added:

Employee login

Start basket

Customer identifier

Add item to basket

Finalize subtotals

Payment completion

Employee logout

For scheme/fields for these fields, assume and add them.

# Point-of-Sale (POS) Subsystem

# Setup Instructions

## Prerequisites

- Docker and Docker Compose
- Go 1.21 or later
- PostgreSQL client (optional, for direct DB access)

## Local Development Setup

1. Clone the repository:

```bash
git clone https://github.com/Piyushhbhutoria/tote-assignment.git
cd tote-assignment
```

2. Start the infrastructure services:

```bash
docker-compose up -d
```

3. Install Go dependencies:

```bash
go mod tidy
```

4. Run the producer (in one terminal):

```bash
go run cmd/producer/main.go
```

5. Run the consumer (in another terminal):

```bash
go run cmd/consumer/main.go
```

## Project Structure

```
/
├── cmd/                    # Application entry points
│   ├── producer/          # Test rig (event generator)
│   ├── consumer/          # Main consumer
│   └── consumer2/         # Secondary consumer
├── internal/              # Private application code
│   ├── models/           # Shared data models
│   ├── plugins/          # Plugin implementations
│   └── config/           # Configuration management
├── pkg/                  # Public libraries
│   ├── kafka/           # Kafka utilities
│   └── database/        # Database utilities
├── web/                 # React frontend
├── docker/              # Docker configurations
└── docs/                # Documentation
```

## Available Plugins

1. Employee Time Tracker
   - Tracks employee login/logout events
   - Calculates time spent at terminals
   - Handles auto-logout scenarios

2. Purchase Recommender
   - Analyzes basket items
   - Generates purchase recommendations
   - Real-time suggestion updates

3. Customer Lookup
   - Processes customer identification
   - Fetches customer data
   - Enriches events with customer information

4. Fraud Detection
   - Monitors transaction patterns
   - Applies configurable rules
   - Generates fraud alerts

5. Age Verification
   - Checks age-restricted items
   - Enforces age verification
   - Manages compliance requirements

## Monitoring

The system can be monitored through Prometheus at <http://localhost:9090>

## Database Access

PostgreSQL connection details:

- Host: localhost
- Port: 5432
- Database: pos_system
- Username: pos
- Password: pos123

## Contributing

1. Create a new branch for your feature
2. Make your changes
3. Write/update tests
4. Create a pull request

## License

[MIT License](LICENSE)
