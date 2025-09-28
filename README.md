# Go Fiber High-Throughput File Upload Backend

## Overview

This project is a backend application built with Go and the Fiber framework. It supports uploads of CSV/text files, processes them to add a boolean email flag column, and stores the processed files locally on disk. Redis is used for caching and queuing jobs to manage concurrency and scalability efficiently.

It exposes two main endpoints:

- **POST /api/upload** — Upload a CSV file. The server parses the file and adds a `true/false` flag column indicating if a row contains any valid email address. Returns a unique job ID.
- **GET /api/download/{id}** — Download the processed CSV by job ID. Returns 423 if processing is still underway, 400 if ID is invalid.

## Features

- Local file storage with clean separation of concerns (handlers, services, repositories)
- Redis for job queueing and state management
- Concurrent background worker processing CSV files
- Simple regex email validation
- Unit and integration test examples included

## Server:
- Use make commands to run tests & start server.