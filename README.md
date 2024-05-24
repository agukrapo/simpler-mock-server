# simpler-mock-server

SMS is a minimalistic mock http server that uses a filesystem as backend.

## Usage

```
go install github.com/agukrapo/simpler-mock-server/cmd/sms@latest
mkdir .sms_responses
sms
```

## Default response status

| Method | status |
|--------|--------|
| DELETE | 202    |
| GET    | 200    |
| PATCH  | 204    |
| POST   | 201    |
| PUT    | 204    |

Can be customized adding a prefix `{status}___` to the file:

	500___a3b69b44-d562-11eb-b8bc-0242ac130003.json

## Environment Variables

- `PORT` (default: `4321`)
- `ADDRESS` (default: `:$PORT`)
- `LOG_LEVEL` (default: `debug`)
- `RESPONSES_DIR` - Directory where the response files are located (default: `./.sms_responses`)
- `EXTENSION_MIME_TYPE_MAP` - File extension to http request Accept MIME type, e.g. `txt:text/plain`
- `METHOD_STATUS_MAP` - Request http method to response http status (default: `DELETE:202,GET:200,PATCH:204,POST:201,PUT:204`)


# TODO
- Better README
