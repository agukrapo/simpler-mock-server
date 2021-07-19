# simpler-mock-server

A minimalistic mock server

### Usage

```
git clone https://github.com/agukrapo/simpler-mock-server.git
cd simpler-mock-server
make
```

Then add a file inside `responses` subdir according to the desired http method: DELETE, GET, PATCH, POST and PUT.

Be sure the file extension is inside `content-type-mapping.txt`.

### Default response status

Method | status
-------| ------
DELETE | 202
GET    | 200
PATCH  | 204
POST   | 201
PUT    | 204

Can be customized adding a prefix `{status}___` to the file:

	500___a3b69b44-d562-11eb-b8bc-0242ac130003.json

### Built-in endpoints

```
curl -X DELETE localhost:4321/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003
curl -X GET localhost:4321/health
curl -X PATCH localhost:4321/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003
curl -X POST localhost:4321/api/people/a3b69b44-d562-11eb-b8bc-0242ac130003
curl -X PUT localhost:4321/api/people/
```

### Run using Docker

```
docker build -t simpler-mock-server .
docker run -p 4321:4321 simpler-mock-server
```