# Web GUI Feature Demo

## Overview
This document demonstrates how the web GUI feature works for the TODO Tracker CLI.

## Starting the Web Server
To start the web GUI, users run:

```bash
todo serve
```

By default, this starts a web server on http://localhost:8080

Options:
- `-p, --port` - Specify port (default: 8080)
- `-H, --host` - Specify host (default: localhost)

Example with custom port:
```bash
todo serve -p 3000
```

## Web Interface Features

### 1. Dashboard View (/)
- Shows statistics overview
- Displays recent TODOs
- Navigation to other sections

### 2. TODO List (/todos)
- Complete list of all TODOs
- Filtering by status, type, priority
- Search functionality
- Links to individual TODO details

### 3. TODO Detail (/todo/{id})
- View detailed TODO information
- Edit TODO properties (status, priority, assignee, content)
- Save changes through form submission

### 4. API Endpoints
- `/api/todos` - Get list of TODOs (JSON)
- `/api/todo/{id}` - Update specific TODO

## Technical Implementation

The web server is built using Go's standard `net/http` package with:
- Template-based HTML rendering
- Form handling for TODO updates
- Static file serving for CSS
- REST-like API endpoints

## Benefits
- Browser-based interface for easier TODO management
- Visual representation of statistics
- Simplified editing workflow
- Accessible from any device on the network