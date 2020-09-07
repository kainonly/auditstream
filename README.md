# elastic-transfer

Provide online and offline template data writing for elasticsearch

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/codexset/elastic-transfer?style=flat-square)](https://github.com/codexset/elastic-transfer)
[![Github Actions](https://img.shields.io/github/workflow/status/codexset/elastic-transfer/release?style=flat-square)](https://github.com/codexset/elastic-transfer/actions)
[![Image Size](https://img.shields.io/docker/image-size/kainonly/elastic-transfer?style=flat-square)](https://hub.docker.com/r/kainonly/elastic-transfer)
[![Docker Pulls](https://img.shields.io/docker/pulls/kainonly/elastic-transfer.svg?style=flat-square)](https://hub.docker.com/r/kainonly/elastic-transfer)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://raw.githubusercontent.com/codexset/elastic-transfer/master/LICENSE)

![guide](guide.svg)

## Setup

Example using docker compose

```yaml
version: "3.8"
services: 
  subscriber:
    image: kainonly/elastic-transfer
    restart: always
    volumes:
      - ./transfer/config:/app/config
    ports:
      - 6000:6000
```

## Configuration

For configuration, please refer to `config/config.example.yml`

- **debug** `bool` Start debugging, ie `net/http/pprof`, access address is`http://localhost:6060`
- **listen** `string` Microservice listening address
- **elastic** `object` Elasticsearch configuration
    - **addresses** `array` hosts
    - **username** `string`
    - **password** `string`
    - **cloud_id** `string` cloud id
    - **api_key** `string` api key
- **mq** `object`
    - **drive** `string` Contains: `amqp`
    - **url** `string` E.g `amqp://guest:guest@localhost:5672/`
    
## Service

The service is based on gRPC and you can view `router/router.proto`

```proto
syntax = "proto3";
package elastic.transfer;
service Router {
  rpc Get (GetParameter) returns (GetResponse) {
  }

  rpc Lists (ListsParameter) returns (ListsResponse) {
  }

  rpc All (NoParameter) returns (AllResponse) {
  }

  rpc Put (Information) returns (Response) {
  }

  rpc Delete (DeleteParameter) returns (Response) {
  }

  rpc Push (PushParameter) returns (Response) {
  }
}

message NoParameter {
}

message Response {
  uint32 error = 1;
  string msg = 2;
}

message Information {
  string identity = 1;
  string index = 2;
  string validate = 3;
  string topic = 4;
  string key = 5;
}

message GetParameter {
  string identity = 1;
}

message GetResponse {
  uint32 error = 1;
  string msg = 2;
  Information data = 3;
}

message ListsParameter {
  repeated string identity = 1;
}

message ListsResponse {
  uint32 error = 1;
  string msg = 2;
  repeated Information data = 3;
}

message AllResponse {
  uint32 error = 1;
  string msg = 2;
  repeated string data = 3;
}

message DeleteParameter {
  string identity = 1;
}

message PushParameter {
  string identity = 1;
  bytes data = 2;
}
```

#### rpc Get (GetParameter) returns (GetResponse) {}

Get transfer configuration

- GetParameter
  - **identity** `string` transfer id
- GetResponse
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback
  - **data** `Information` result
    - **identity** `string` transfer id
    - **index** `string` elasticsearch index
    - **validate** `string` JSON Schema
    - **topic** `string` Topic name of the message queue
    - **key** `string` The queue name of the message queue


```golang
client := pb.NewRouterClient(conn)
response, err := client.Get(context.Background(), &pb.GetParameter{
  Identity: "task",
})
```

#### rpc Lists (ListsParameter) returns (ListsResponse) {}

Lists transfer configuration

- ListsParameter
  - **identity** `string` transfer id
- ListsResponse
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback
  - **data** `[]Information` result
    - **identity** `string` transfer id
    - **index** `string` elasticsearch index
    - **validate** `string` JSON Schema
    - **topic** `string` Topic name of the message queue
    - **key** `string` The queue name of the message queue

```golang
client := pb.NewRouterClient(conn)
response, err := client.Lists(context.Background(), &pb.ListsParameter{
  Identity: []string{"task"},
})
```

#### rpc All (NoParameter) returns (AllResponse) {}

- NoParameter
- AllResponse
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback
  - **data** `[]string` transfer IDs

```golang
client := pb.NewRouterClient(conn)
response, err := client.All(context.Background(), &pb.NoParameter{})
```

#### rpc Put (Information) returns (Response) {}

- Information
  - **identity** `string` transfer id
  - **index** `string` elasticsearch index
  - **validate** `string` JSON Schema
  - **topic** `string` Topic name of the message queue
  - **key** `string` The queue name of the message queue
- Response
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback

```golang
client := pb.NewRouterClient(conn)
response, err := client.Put(context.Background(), &pb.Information{
  Identity: "task",
  Index:    "task-log",
  Validate: `{"type":"object","properties":{"name":{"type":"string"}}}`,
  Topic:    "sys.schedule",
  Key:      "",
})
```

#### rpc Delete (DeleteParameter) returns (Response) {}

- DeleteParameter
  - **identity** `string` transfer id
- Response
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback

```golang
client := pb.NewRouterClient(conn)
response, err := client.Delete(context.Background(), &pb.DeleteParameter{
  Identity: "task",
})
```

#### rpc Push (PushParameter) returns (Response) {}

- PushParameter
  - **identity** `string` transfer id
  - **data** `bytes` push data
- Response
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback

```golang
client := pb.NewRouterClient(conn)
response, err := client.Push(context.Background(), &pb.PushParameter{
  Identity: "task",
  Data:     []byte(`{"name":"kain"}`),
})
```