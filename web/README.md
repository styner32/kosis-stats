# KOSIS Stats Frontend

A TypeScript-based Single Page Application (SPA) for viewing financial reports from the KOSIS Stats backend.

## Features

- **Company Selection**: Browse and select companies from the database
- **Report Filtering**: Filter reports by company and year
- **Report Details**: View detailed report information including JSON data
- **TypeScript**: Fully typed with TypeScript for better development experience
- **Modern ES Modules**: Uses ES2020 modules for clean code organization

## Development

### Prerequisites

- Node.js 18+ and npm

### Setup

1. Install dependencies:
```bash
npm install
```

2. Build the TypeScript files:
```bash
npm run build
```

3. Serve the application:
```bash
npm run serve
```

Or for development with auto-rebuild:
```bash
npm run watch
# In another terminal:
npm run serve
```

Then open `http://localhost:3000` in your browser.

## Usage

### Option 1: Build and Serve

1. Build the TypeScript code:
```bash
npm run build
```

2. Serve with a local server:
```bash
npm run serve
```

### Option 2: Direct File Access (after build)

After building, you can open `index.html` directly in your browser. The frontend will attempt to connect to `http://localhost:8080`.

## Configuration

To change the backend API URL, edit `app.js`:

```javascript
const API_BASE_URL = 'http://localhost:8080';
```

## API Endpoints

The frontend expects the following API endpoints:

- `GET /api/v1/companies` - List all companies
- `GET /api/v1/reports?corp_code={code}` - List reports for a company
- `GET /api/v1/reports/:receipt_number` - Get report details
- `GET /health` - Health check

## Browser Support

Works in all modern browsers (Chrome, Firefox, Safari, Edge) that support:
- ES6+ JavaScript
- Fetch API
- CSS Grid and Flexbox

