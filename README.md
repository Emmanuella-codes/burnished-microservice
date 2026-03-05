# Burnished Microservice

Small Go service for CV/resume processing. It accepts a PDF or DOCX, extracts text, and uses the DeepSeek API to:
- Optimize a resume for ATS as structured JSON (`format` mode)
- Provide a brutally honest critique (`roast` mode)
- Generate a cover letter from a CV + job description (`letter` mode)

## API
Base path: `/api/v1`

`POST /process` (auth required)
- Form-data fields:
  - `file` (PDF or DOCX)
  - `mode` (`format` | `roast` | `letter`)
  - `jobDescription` (required for `format` and `letter`)
- Response:
  - `format`: `formattedResume` JSON
  - `roast`: `feedback` string
  - `letter`: `coverLetter` string

`GET /health`
- Returns `{ "status": "ok", "time": "..." }`

## Configuration
Environment variables:
- `PORT` (default: `8080`)
- `DEEPSEEK_API_KEY` (required)
- `MAX_FILE_SIZE` (bytes, optional)
- `BURNISHED_WEB_API_KEY` (required for requests)
- `BURNISHED_WEB_WEBHOOK_URL` (optional, enables webhook responses)
- `WEBHOOK_SECRET` (optional, sent as Bearer token to webhook)

## Running
Local:
```sh
go run .
```

Docker:
```sh
docker build -t burnished-microservice .
docker run -p 8080:8080 -e DEEPSEEK_API_KEY=... -e BURNISHED_WEB_API_KEY=... burnished-microservice
```
