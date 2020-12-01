# Elastic Transfer

Provide online and offline template data writing for elasticsearch

[![Github Actions](https://img.shields.io/github/workflow/status/kain-lab/elastic-transfer/release?style=flat-square)](https://github.com/kain-lab/elastic-transfer/actions)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/kain-lab/elastic-transfer?style=flat-square)](https://github.com/kain-lab/elastic-transfer)
[![Image Size](https://img.shields.io/docker/image-size/kainonly/elastic-transfer?style=flat-square)](https://hub.docker.com/r/kainonly/elastic-transfer)
[![Docker Pulls](https://img.shields.io/docker/pulls/kainonly/elastic-transfer.svg?style=flat-square)](https://hub.docker.com/r/kainonly/elastic-transfer)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://raw.githubusercontent.com/kain-lab/elastic-transfer/master/LICENSE)

## Setup

Example using docker compose

```yaml
version: "3.8"
services: 
  transfer:
    image: kainonly/elastic-transfer
    restart: always
    volumes:
      - ./transfer/config:/app/config
    ports:
      - 6000:6000
      - 8080:8080
```

## Configuration

For configuration, please refer to `config/config.example.yml`

- **debug** `string` Start debugging, ie `net/http/pprof`, access address is`http://localhost:6060`
- **listen** `string` grpc server listening address
- **gateway** `string` API gateway server listening address
- **elastic** `object` Elasticsearch configuration
    - **addresses** `array` hosts
    - **username** `string`
    - **password** `string`
    - **cloud_id** `string` cloud id
    - **api_key** `string` api key
- **queue** `object`
    - **drive** `string` Contains: `amqp`
    - **option** `object` (amqp) 
        - **url** `string` E.g `amqp://guest:guest@localhost:5672/`
    
## Service

The service is based on gRPC to view `api/api.proto`

```proto
syntax = "proto3";
package elastic.transfer;
option go_package = "elastic-transfer/gen/go/elastic/transfer";
import "google/protobuf/empty.proto";
import "google/api/annotations.proto";

service API {
  rpc Get (ID) returns (Data) {
    option (google.api.http) = {
      get: "/transfer",
    };
  }
  rpc Lists (IDs) returns (DataLists) {
    option (google.api.http) = {
      post: "/transfers",
      body: "*"
    };
  }
  rpc All (google.protobuf.Empty) returns (IDs) {
    option (google.api.http) = {
      get: "/transfers",
    };
  }
  rpc Put (Data) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      put: "/transfer",
      body: "*",
    };
  }
  rpc Delete (ID) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/transfer",
    };
  }
  rpc Push (Body) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      post: "/push",
      body: "*"
    };
  }
}

message Data {
  string id = 1;
  string index = 2;
  string validate = 3;
  string topic = 4;
  string key = 5;
}

message ID {
  string id = 1;
}

message IDs {
  repeated string ids = 1;
}

message DataLists {
  repeated Data data = 1;
}

message DeleteParameter {
  string identity = 1;
}

message Body {
  string id = 1;
  bytes content = 2;
}
```

## Get (ID) returns (Data)

Get transfer configuration

### RPC

- **ID**
  - **id** `string` transfer id
- **Data**
  - **id** `string` transfer id
  - **index** `string` elasticsearch index
  - **validate** `string` JSON Schema
  - **topic** `string` Topic name of the message queue
  - **key** `string` The queue name of the message queue


```golang
client := pb.NewAPIClient(conn)
response, err := client.Get(context.Background(), &pb.ID{
		Id: "debug",
})
```

### API Gateway

- **PUT** `/client`

```http
GET /transfer?id=debug HTTP/1.1
Host: localhost:8080
```

## Lists (IDs) returns (DataLists)

Lists transfer configuration

### RPC

- **IDs**
  - **ids** `[]string` transfer id
- **DataLists**
  - **data** `[]Data`
    - **id** `string` transfer id
    - **index** `string` elasticsearch index
    - **validate** `string` JSON Schema
    - **topic** `string` Topic name of the message queue
    - **key** `string` The queue name of the message queue

```golang
client := pb.NewAPIClient(conn)
response, err := client.Lists(context.Background(), &pb.IDs{
  Ids: []string{"debug"},
})
```

### API Gateway

- **POST** `/transfers`

```http
POST /transfers HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
    "ids":["debug"]
}
```

## All (google.protobuf.Empty) returns (IDs)

Get all transfer configuration identifiers

### RPC

- **IDs**
  - **ids** `[]string` transfer id

```golang
client := pb.NewAPIClient(conn)
response, err := client.All(context.Background(), &empty.Empty{})
```

### API Gateway

- **GET** `/transfers`

```http
GET /transfers HTTP/1.1
Host: localhost:8080
```

## Put (Data) returns (google.protobuf.Empty)

Put transfer configuration

### RPC

- **Data**
  - **id** `string` transfer id
  - **index** `string` elasticsearch index
  - **validate** `string` JSON Schema
  - **topic** `string` Topic name of the message queue
  - **key** `string` The queue name of the message queue

```golang
client := pb.NewAPIClient(conn)
response, err := client.Put(context.Background(), &pb.Data{
  Id:       "debug",
  Index:    "debug-logs-alpha",
  Validate: `{"type":"object"}`,
  Topic:    "logs.debug",
  Key:      "",
})
```

### API Gateway

- **PUT** `/transfer`

```http
PUT /transfer HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
    "id": "debug",
    "index": "debug-logs-alpha",
    "validate": "{\"type\":\"object\"}",
    "topic": "logs.debug",
    "key": ""
}
```

## Delete (ID) returns (google.protobuf.Empty)

Remove transfer configuration

### RPC

- **ID**
  - **id** `string` transfer id

```golang
client := pb.NewAPIClient(conn)
response, err := client.Delete(context.Background(), &pb.ID{
  Id: "debug",
})
```

### API Gateway

- **DELETE** `/transfer`

```http
DELETE /transfer?id=debug HTTP/1.1
Host: localhost:8080
```

## Push (Body) returns (google.protobuf.Empty)

Push content to transfer

### RPC

- **Body**
  - **id** `string` transfer id
  - **content** `bytes` push content

```golang
client := pb.NewAPIClient(conn)
response, err := client.Push(context.Background(), &pb.Body{
  Id:      "debug",
  Content: []byte(`{"name":"kain"}`),
})
```

### API Gateway

- **POST** `/push`

```http
POST /push HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
    "id": "debug",
    "content": "eyJuYW1lIjoiYXBpIn0="
}
```