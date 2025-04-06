# Point-of-Sale (POS) Event Processing System

A scalable Point-of-Sale event processing system that handles various POS events through configurable plugins. The system uses Kafka for event streaming and PostgreSQL for data persistence.

## Features

- Event-driven architecture using Apache Kafka
- Pluggable architecture for event processing
- Real-time event generation and processing
- Configurable plugins with dynamic activation/deactivation
- Comprehensive test coverage
- Modern React TypeScript web interface for plugin management

## Plugins

1. **Employee Time Tracker**
   - Tracks employee login/logout events
   - Calculates time spent at terminals
   - Auto-logout on new terminal login

2. **Purchase Recommender**
   - Analyzes basket items
   - Provides purchase recommendations
   - Real-time recommendation generation

3. **Customer Lookup**
   - Processes customer identification events
   - Enriches events with customer data
   - Real-time customer data lookup

## Event Types

The system processes the following event types:

- `EMPLOYEE_LOGIN`: Employee login at a terminal
- `EMPLOYEE_LOGOUT`: Employee logout from a terminal
- `START_BASKET`: New transaction basket creation
- `CUSTOMER_IDENTIFY`: Customer identification
- `ADD_ITEM`: Item addition to basket
- `FINALIZE_SUBTOTAL`: Basket subtotal calculation
- `PAYMENT_COMPLETE`: Transaction completion

## Project Structure

```
/
├── cmd/
│   ├── producer/       # Event generator
│   └── server/        # Main server application
├── internal/
│   ├── models/        # Data models
│   └── plugins/       # Plugin implementations
├── pkg/
│   ├── kafka/         # Kafka utilities
│   └── database/      # Database utilities
├── web/              # React TypeScript frontend
│   ├── src/          # Source code
│   ├── public/       # Static assets
│   └── package.json  # Dependencies
└── docs/             # Documentation
```

## Prerequisites

- Go 1.24.1
- Apache Kafka
- PostgreSQL
- Node.js v20 LTS
- npm or yarn

## Getting Started

1. Clone the repository:

   ```bash
   git clone https://github.com/Piyushhbhutoria/tote-assignment.git
   cd tote-assignment
   ```

2. Install backend dependencies:

   ```bash
   go mod download
   ```

3. Run backend tests:

   ```bash
   go test -v ./...
   ```

4. Start the producer:

   ```bash
   go run cmd/producer/main.go
   ```

5. Start the server:

   ```bash
   go run cmd/server/main.go
   ```

6. Set up the web interface:

   ```bash
   cd web
   npm install    # or yarn install
   npm start     # or yarn start
   ```

## Web Interface

The web interface provides a user-friendly way to manage plugins and monitor the system. It's built with:

- React 18
- TypeScript
- Material-UI components
- Real-time updates

### Features

- View all available plugins
- Enable/disable plugins
- Configure plugin settings
- Monitor plugin statistics
- Real-time event monitoring

### Development

To start the web application in development mode:

```bash
cd web
npm install
npm start
```

The application will be available at `http://localhost:3000`.

To build for production:

```bash
npm run build
```

## Testing

The project includes comprehensive tests for all components. Tests are automatically run on:

- Push to master branch
- Pull requests to master branch

To run backend tests:

```bash
go test -v ./...
```
