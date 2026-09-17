# Symbol Web

Symbol Web is a Go web app that turns user text into ASCII art using selectable banner fonts. It includes a small AI layer for smart suggestions and banner recommendations, while still working in a deterministic mock mode when no LLM backend is configured.

## Features

- Text-to-ASCII conversion with multiple banner styles
- Banner recommendation based on text input
- AI-powered suggestion and variation APIs
- Simple HTML frontend with JavaScript-based interactions
- Mock mode for local development without API keys

## Project structure

- [cmd/server/main.go](cmd/server/main.go) – server entry point
- [internal/ai](internal/ai) – AI client, mock logic, prompt parsing, recommendation logic
- [internal/ascii](internal/ascii) – ASCII renderer and font loader
- [internal/handlers](internal/handlers) – HTTP handlers for pages and JSON APIs
- [static/css/style.css](static/css/style.css) – frontend styling
- [static/js](static/js) – browser-side suggestion/recommendation/variation logic
- [templates](templates) – HTML templates
- [shadow.txt](shadow.txt), [standard.txt](standard.txt), [thinkertoy.txt](thinkertoy.txt) – ASCII banner font files

## Requirements

- Go 1.22 or newer
- A browser for the app UI
- Optional: an OpenAI-compatible or Ollama-compatible LLM backend

## Run locally

From the project root:

```bash
cd /Users/aigerim/symbol-web
go run ./cmd/server
```

Then open:

```text
http://localhost:8080
```

Important: run from the project root, not from inside the server folder. The app resolves paths relative to the root when it starts.

## Environment variables

The app reads the following variables from the environment:

```bash
LLM_BASE_URL
LLM_MODEL
LLM_API_KEY
```

- If `LLM_BASE_URL` is empty, the app runs in mock mode.
- If set, the app will call the configured OpenAI-compatible backend.

Example:

```bash
export LLM_BASE_URL=http://localhost:11434
export LLM_MODEL=llama3.1
export LLM_API_KEY=
```

## Endpoints

### Web UI

- `GET /` – main generator page
- `POST /symbol-art` – generate ASCII art from text and selected banner

### AI API

- `POST /api/suggest/` – returns suggestion strings
- `POST /api/recommend-banner` – returns banner recommendation metadata
- `POST /api/variations` – returns creative variations for the text

## Development notes

- The HTML templates live in [templates](templates)
- Static assets live in [static](static)
- The server uses `template.ParseGlob` with paths resolved from the project root
- The app is intentionally tolerant of missing LLM config, so it can run in a fully local mock mode

## License

This project is intended for local development and experimentation.
