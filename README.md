# Weplanx Transfer

日志传输器，配合日志采集器将相应的消息管道数据传输至支援的日志系统中。

> 项目继承于 [elastic-transfer](https://github.com/weplanx/log-transfer/tree/elastic-transfer) 延续开发
> 新版本将以 `v*.*.*` 形式发布

## 部署服务

日志传输器采用更广泛 gRPC 进行服务通讯，通过 NATS JetStream 处理消息流，除此之外还需要 MongoDB 作为配置存储介质

> 需要注意的是 NATS 与 MongoDB 仅支持集群方式（原因是 NATS JetStream 至少需要3节点的最小集群，而为了配置的一致性 MongoDB 至少采用副本集方式）

镜像源主要有：

- ghcr.io/weplanx/transfer:latest
- ccr.ccs.tencentyun.com/weplanx/transfer:latest（国内）

案例将使用 Kubernetes 部署编排，复制部署内容（需要根据情况做修改）：

1. 设置配置

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: transfer.cfg
data:
  config.yml: |
    address: ":6000"
    tls: <TLS配置，非必须>
      cert:
      key:
    namespace: <命名空间>
    database:
      uri: mongodb://<username>:<password>@<host>:<port>/<database>?authSource=<authSource>
      name: <数据库名>
      collection: <默认集合>
    nats:
      hosts: [ ]
      nkey:
```

2. 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: transfer
  name: transfer-deploy
spec:
  replicas: 2
  selector:
    matchLabels:
      app: transfer
  template:
    metadata:
      labels:
        app: transfer
    spec:
      containers:
        - image: ccr.ccs.tencentyun.com/weplanx/transfer:latest
          imagePullPolicy: Always
          name: transfer
          ports:
            - containerPort: 6000
          volumeMounts:
            - name: config
              mountPath: "/app/config"
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: transfer.cfg
            items:
              - key: "config.yml"
                path: "config.yml"
```

3. 设置入口，服务网关推荐采用 traefik 做更多处理

```yaml
apiVersion: v1
kind: Service
metadata:
  name: transfer-svc
spec:
  ports:
    - port: 6000
      protocol: TCP
  selector:
    app: transfer
```

## 滚动更新

复制模板内容，并需要自行定制触发条件，原理是每次patch将模板中 `${tag}` 替换为版本执行

```yml
spec:
  template:
    spec:
      containers:
        - image: ccr.ccs.tencentyun.com/weplanx/transfer:${tag}
          name: api
```

例如：在 Github Actions
中 `patch deployment transfer-deploy --patch "$(sed "s/\${tag}/${{steps.meta.outputs.version}}/" < ./config/patch.yml)"`，国内可使用**Coding持续部署**或**云效流水线**等。

## License

[BSD-3-Clause License](https://github.com/weplanx/transfer/blob/main/LICENSE)