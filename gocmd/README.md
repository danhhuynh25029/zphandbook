# Golang Command Line

## Quản lý biến môi trường Golang
* Xem environment của Go
```
    go env
```
* Sửa environment của Golang

```
    go env -w {env}={value}
```

## Quản lý module trong Go
* Khởi tạo go.mod
```
    go mod init {module_name}
```
* Cài đặt dependency
```
    go get {module_name}
```
 
=> Trong lệnh go get còn các argument 

| flag | Mô tả                                         |
|------|-----------------------------------------------|
| -t   | Sử dụng khi cần download và test module       |
| -x   | Hiển thị lệnh được thưc thi. Sử dụng để debug |    
| -u   | Được sử dụng khi cần update module            |

* Cài đặt và update module trong file go.mod

Có thể sử dụng một trong 2 lệnh 

```
    go mod download
```
=> Cài đặt các module chưa được download với version tương ứng trong file go.mod 

hoặc
```
    go mod tidy
```
=> Update các version mới cho module trong file go.mod và loại bỏ các module không sử dụng

# Build Binary Go

```
    go build -o {binary_name} .
```
