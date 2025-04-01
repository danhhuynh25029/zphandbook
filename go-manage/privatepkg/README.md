# Quản lý pkg với private repo github

* Set GOPRIVATE
  * Nếu muốn access tất cả module của tổ chức
```
    go env -w GOPRIVATE=github.com/{organization}
``` 
  * Sử dụng duy nhất một module
```
    go env -w GOPRIVATE=github.com/{organization}/{module_name}
``` 

* Set .gitconfig
```
    [url "https://{username}:{access_token}@github.com"]
           insteadOf = https://github.com
```