# distributed

### 注册服务

=>registryservice

- 服务启动
- 服务卸载
- 服务发现
- 创建WEB Service
- 服务健康检查

```shell
go run cmd/registryservice/main.go
```

http://localhost:3000/services

### 日志服务

=>logservice

- 自定义logger
- 服务日志记录
- 客户端日志

```shell
go run cmd/logservice/main.go
```

http://localhost:4000

### 成绩服务

=>gradingservice

数据层

- MOCK-DATA
- 平均分计算
- 获取全部
- 根据ID获取

http://localhost:6000

### Portald服务

=>portal

业务层

- 模板
- 渲染

```shell
go run cmd/portal/main.go
```



业务Url：http://localhost:5000

```html
	/students
	/students/{:id}
	/students/{:id}/grades
```

![图片](https://github.com/x9nu/distributed/blob/main/docs/rendering.png)
