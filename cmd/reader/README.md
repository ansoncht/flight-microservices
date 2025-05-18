# Flight Reader

This service is designed to periodically fetch flight information from an external API for various airports around the world. It supports fetching data for any specified airport.

## Features

- Fetches real-time flight data from an external API.
- Supports multiple airports by specifying the airport code.
- Periodically scheduled jobs to ensure up-to-date flight information.
- Provides a gRPC interface for further processing of flight data.

## Configuration

The service can be configured using environment variables to set parameters such as API URLs, authentication credentials, and server settings.

## Usage

To run the service, ensure that the necessary environment variables are set, and then execute the main application. The service will start fetching flight data based on the configured schedule.

## Endpoints

- **Fetch Flights**: Trigger a manual fetch of flights for a specified airport via the HTTP endpoint `/api/v1/fetch`.
