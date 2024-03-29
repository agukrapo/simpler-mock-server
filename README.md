# simpler-mock-server

A minimalistic mock http server

## Usage

```
git clone https://github.com/agukrapo/simpler-mock-server.git
cd simpler-mock-server
make
```

Then add a file inside `responses` subdir according to the desired http method: DELETE, GET, PATCH, POST and PUT.

Be sure the file extension is inside `content-type-mapping.txt`.

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

## Example endpoints

Executing `make run` in the project directory will make these endpoints available

```
curl -X DELETE localhost:4321/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003
curl -X GET localhost:4321/health
curl -X PATCH localhost:4321/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003
curl -X POST localhost:4321/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003
curl -X PUT localhost:4321/api/people/
```

## Environment Variables

- `PORT` (default: `4321`)
- `ADDRESS` (default: `:$PORT`)
- `LOG_LEVEL` (default: `debug`)
- `RESPONSES_DIR` (default: `./.sms_responses`)
- `EXTENSION_CONTENT_TYPE_MAP` (default: `txt:text/plain,json:application/json,yaml:text/yaml,xml:application/xml,html:text/html,csv:text/csv`)
- `METHOD_STATUS_MAP` (default: `DELETE:202,GET:200,PATCH:204,POST:201,PUT:204`)


# TODO
- Better README
- CLI help
