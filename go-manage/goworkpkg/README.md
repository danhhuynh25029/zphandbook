# GOWORK

Quản lý nhiều module trong một repo duy nhất

* Khơi tạo 2 project với gocmd
```
    mkdir errorlog && cd errorlog && go mod init errorlog 
    mkdir view && cd view && go mod init view
```
* Khởi tạo file go.work
```
    go work init errorlog view
```


* Thêm module vào file go.work
```
    go work use ./{module_name}
```

