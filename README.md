# CEA Go AR Exception

A personal full-stack project that generates the **Charged Sales** AR exception report directly from Oracle. It does not authenticate users or store reports.

## What it does

- Runs one read-only, parameterized Oracle Charged Sales query.
- Filters by required date range and optional company ID and store code.
- Displays results in a React + TypeScript table.
- Downloads the current result as CSV.

## Requirements

- Go 1.27+
- Node.js 22+
- Oracle Database 10.2+ reachable from this computer
- An Oracle account with read-only access to the report tables

This project uses a pure-Go Oracle driver, so it does not require Oracle Instant Client or a C compiler.

## Configure Oracle once

Copy the example file, then open `.env` and fill in the same Oracle values used by a Laravel project:

```powershell
Copy-Item .env.example .env
notepad .env
```

The Go API loads `.env` automatically and safely builds the Oracle connection string. Use one value per line; do not add trailing backslashes. The password is URL-encoded by the driver, so you enter it normally in `.env`.

## Build and run

Double-click or run these files from the project root, in this order:

```powershell
.\build-frontend.bat
.\run-backend.bat
```

`build-frontend.bat` installs frontend packages if needed and creates `frontend/dist`. `run-backend.bat` starts the Go API and serves that built React app at `http://127.0.0.1:8012/`. There is no React development server in the normal workflow.

The health check remains available at `http://127.0.0.1:8012/health`.

## API request

`POST /api/reports/charged-sales`

```json
{
  "dateFrom": "2026-11-01",
  "dateTo": "2026-11-16",
  "companyID": 10,
  "storeCode": "001"
}
```

`companyID` and `storeCode` are optional. The Oracle query uses bind values, never SQL built from browser input.

## Deliberate limits

There is no authentication, saved-report history, separate application database, PDF/Excel export, or store lookup list. Keep the API on your own computer unless you add authentication before sharing it.
