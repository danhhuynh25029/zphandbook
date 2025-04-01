# buff-sample

Bước 1 : 
* Cài đặt buf-cli từ document
  * https://buf.build/docs/cli/installation
* Cài đặt : protoc-gen-go
  * https://grpc.io/docs/languages/go/quickstart

Bước 2:
* Chạy lệnh
```
  buf generate
```
Bước 3:
* Cài đặt kafka bằng docker

```
  docker compose up -d
```

Bước 4:
* Chạy producer:
```
  go run  service/producer/producer.go
```
* Chạy consumer:
```
  go run service/consumer/consumer.go
```