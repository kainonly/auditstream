# microtools-gossh

Use gRPC to manage remote SSH clients

## Configuration

For configuration, please refer to `config/config.example.yml` and create `config/config.yml`

- **debug** `bool` Turn on debugging, that is `net/http/pprof`, and visit the address `http://localhost: 6060`
- **listen** `string` Microservice listening address

## Service

The service is based on gRPC to view `router/router.proto`

```
syntax = "proto3";

service Router {
    rpc Testing (TestingParameter) returns (Response) {
    }

    rpc Put (PutParameter) returns (Response) {
    }

    rpc Exec (ExecParameter) returns (ExecResponse) {
    }

    rpc Delete (DeleteParameter) returns (Response) {
    }

    rpc Get (GetParameter) returns (GetResponse) {
    }

    rpc All (NoParameter) returns (AllResponse) {
    }

    rpc Lists (ListsParameter) returns (ListsResponse) {
    }

    rpc Tunnels (TunnelsParameter) returns (Response) {
    }
}

message NoParameter {
}

message Response {
    uint32 error = 1;
    string msg = 2;
}

message TestingParameter {
    string host = 1;
    uint32 port = 2;
    string username = 3;
    string password = 4;
    string private_key = 5;
    string passphrase = 6;
}

message PutParameter {
    string identity = 1;
    string host = 2;
    uint32 port = 3;
    string username = 4;
    string password = 5;
    string private_key = 6;
    string passphrase = 7;
}

message ExecParameter {
    string identity = 1;
    string bash = 2;
}

message ExecResponse {
    uint32 error = 1;
    string msg = 2;
    string data = 3;
}

message DeleteParameter {
    string identity = 1;
}

message GetParameter {
    string identity = 1;
}

message GetResponse {
    uint32 error = 1;
    string msg = 2;
    Information data = 3;
}

message Information {
    string identity = 1;
    string host = 2;
    uint32 port = 3;
    string username = 4;
    string connected = 5;
    repeated TunnelOption tunnels = 6;
}

message TunnelOption {
    string src_ip = 1;
    uint32 src_port = 2;
    string dst_ip = 3;
    uint32 dst_port = 4;
}

message AllResponse {
    uint32 error = 1;
    string msg = 2;
    repeated string data = 3;
}

message ListsParameter {
    repeated string identity = 1;
}

message ListsResponse {
    uint32 error = 1;
    string msg = 2;
    repeated Information data = 3;
}

message TunnelsParameter {
    string identity = 1;
    repeated TunnelOption tunnels = 2;
}
```

#### rpc Testing (TestingParameter) returns (Response) {}

SSH client connection test

- TestingParameter
  - **host** `string`
  - **port** `uint32`
  - **username** `string`
  - **password** `string` SSH password, default empty
  - **private_key** `string` SSH private key (Base64)
  - **passphrase** `string` private key passphrase
- Response
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback

```golang
client := pb.NewRouterClient(conn)
response, err := client.Testing(
    context.Background(),
    &pb.TestingParameter{
        Host:       debug[0].Host,
        Port:       debug[0].Port,
        Username:   debug[0].Username,
        Password:   debug[0].Password,
        PrivateKey: debug[0].PrivateKey,
        Passphrase: debug[0].Passphrase,
    },
)
```

#### rpc Put (PutParameter) returns (Response) {}

New or updated SSH client connection

- PutParameter
  - **identity** `string` ssh identity code
  - **host** `string`
  - **port** `uint32`
  - **username** `string`
  - **password** `string` SSH password, default empty
  - **private_key** `string` SSH private key (Base64)
  - **passphrase** `string` private key passphrase
- Response
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback

```golang
client := pb.NewRouterClient(conn)
response, err := client.Put(
    context.Background(),
    &pb.PutParameter{
        Identity:   "test",
        Host:       debug[0].Host,
        Port:       debug[0].Port,
        Username:   debug[0].Username,
        Password:   debug[0].Password,
        PrivateKey: debug[0].PrivateKey,
        Passphrase: debug[0].Passphrase,
    },
)
```

#### rpc Exec (ExecParameter) returns (ExecResponse) {}

Execute remote shell command

- ExecParameter
  - **identity** `string` ssh identity code
  - **bash** `string` shell command
- ExecResponse
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback
  - **data** result

```golang
client := pb.NewRouterClient(conn)
response, err := client.Exec(
    context.Background(),
    &pb.ExecParameter{
        Identity: "test",
        Bash:     "uptime",
    },
)
```

#### rpc Delete (DeleteParameter) returns (Response) {}

Remove SSH client

- DeleteParameter
  - **identity** `string` ssh identity code
- Response
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback

```golang
client := pb.NewRouterClient(conn)
response, err := client.Delete(
    context.Background(),
    &pb.DeleteParameter{
        Identity: "test",
    },
)
```

#### rpc Get (GetParameter) returns (GetResponse) {}

Get the current information of the specified SSH

- GetParameter
  - **identity** `string` ssh identity code
- GetResponse
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback
  - **data** `Information` result
    - **identity** `string` ssh identity code
    - **host** `string`
    - **port** `uint32`
    - **username** `string`
    - **connected** ssh connected client version
    - **tunnels** `[]TunnelOption` ssh tunnels
      - **src_ip** `string` origin ip
      - **src_port** `uint32` origin port
      - **dst_ip** `string` target ip
      - **dst_port** `uint32` target port

```golang
client := pb.NewRouterClient(conn)
response, err := client.Get(
    context.Background(),
    &pb.GetParameter{
        Identity: "test",
    },
)
```

#### rpc All (NoParameter) returns (AllResponse) {}

Get all SSH client IDs

- NoParameter
- AllResponse
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback
  - **data** `[]string` SSH client IDs

```golang
client := pb.NewRouterClient(conn)
response, err := client.All(
    context.Background(),
    &pb.NoParameter{},
)
```

#### rpc Lists (ListsParameter) returns (ListsResponse) {}

Get current SSH information in batches

- ListsParameter
  - **identity** `[]string` ssh IDs code
- ListsResponse
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback
  - **data** `[]Information` result
    - **identity** `string` ssh identity code
    - **host** `string`
    - **port** `uint32`
    - **username** `string`
    - **connected** ssh connected client version
    - **tunnels** `[]TunnelOption` ssh tunnels
      - **src_ip** `string` origin ip
      - **src_port** `uint32` origin port
      - **dst_ip** `string` target ip
      - **dst_port** `uint32` target port

```golang
client := pb.NewRouterClient(conn)
response, err := client.Lists(
    context.Background(),
    &pb.ListsParameter{
        Identity: []string{"test", "other"},
    },
)
```

#### rpc Tunnels (TunnelsParameter) returns (Response) {}

Setting up an SSH tunnel

- TunnelsParameter
  - **identity** `string` ssh identity code
  - **tunnels** `[]TunnelOption` ssh tunnels
    - **src_ip** `string` origin ip
    - **src_port** `uint32` origin port
    - **dst_ip** `string` target ip
    - **dst_port** `uint32` target port
- Response
  - **error** `uint32` error code, `0` is normal
  - **msg** `string` error feedback

```golang
client := pb.NewRouterClient(conn)
response, err := client.Tunnels(
    context.Background(),
    &pb.TunnelsParameter{
        Identity: "test",
        Tunnels: []*pb.TunnelOption{
            &pb.TunnelOption{
                SrcIp:   "127.0.0.1",
                SrcPort: 3306,
                DstIp:   "127.0.0.1",
                DstPort: 3306,
            },
            &pb.TunnelOption{
                SrcIp:   "127.0.0.1",
                SrcPort: 9200,
                DstIp:   "127.0.0.1",
                DstPort: 9200,
            },
        },
    },
)
```