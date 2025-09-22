# Redis Package

Gói Redis cung cấp các kết nối và tiện ích để làm việc với Redis, hỗ trợ nhiều mô hình triển khai khác nhau.

## Tính năng

-   **Kết nối đơn node**: Kết nối đến một Redis server đơn lẻ
-   **Kết nối Cluster**: Hỗ trợ Redis Cluster với nhiều node
-   **Kết nối Sentinel**: Hỗ trợ Redis Sentinel cho high availability
-   **Mô hình Master-Slave**: Hỗ trợ phân tách đọc-ghi (read-write separation) cho cả Cluster và Sentinel
-   **Cấu hình linh hoạt**: Tùy chỉnh dễ dàng thông qua các cấu trúc Config

## Cài đặt

```bash
go get github.com/DuongSonn/go-libs/pkg/redis
```

## Sử dụng

### Kết nối đơn node

```go
import (
    "context"
    _redis "github.com/DuongSonn/go-libs/pkg/redis"
)

func main() {
    // Tạo cấu hình
    config := _redis.DefaultConfig()
    config.Host = "localhost"
    config.Port = 6379

    // Tạo kết nối
    conn := _redis.NewConnection(config)

    // Kết nối đến Redis
    if err := conn.Connect(context.Background()); err != nil {
        panic(err)
    }
    defer conn.Close()

    // Lấy client và sử dụng
    client := conn.GetClient()
    // Sử dụng client...
}
```

### Kết nối Cluster với Master-Slave

```go
import (
    "context"
    _redis "github.com/DuongSonn/go-libs/pkg/redis"
)

func main() {
    // Tạo cấu hình cho Redis Cluster
    config := _redis.DefaultClusterConfig()
    config.Addresses = []string{
        "localhost:7000",
        "localhost:7001",
        "localhost:7002",
    }
    config.UseSlaveConnection = true // Kích hoạt slave client
    config.SlaveReadOnly = true      // Đảm bảo slave chỉ dùng cho đọc

    // Tạo kết nối
    conn := _redis.NewClusterConnection(config)

    // Kết nối đến Redis Cluster
    if err := conn.Connect(context.Background()); err != nil {
        panic(err)
    }
    defer conn.Close()

    // Lấy master client cho thao tác ghi
    masterClient := conn.GetMasterClient()
    // Sử dụng masterClient để ghi dữ liệu...

    // Kiểm tra và sử dụng slave client cho thao tác đọc
    if conn.HasSlaveConnected() {
        slaveClient := conn.GetSlaveClient()
        // Sử dụng slaveClient để đọc dữ liệu...
    }
}
```

### Kết nối Sentinel với Master-Slave

```go
import (
    "context"
    _redis "github.com/DuongSonn/go-libs/pkg/redis"
)

func main() {
    // Tạo cấu hình cho Redis Sentinel
    config := _redis.DefaultSentinelConfig()
    config.MasterName = "mymaster"
    config.SentinelAddresses = []string{
        "localhost:26379",
        "localhost:26380",
        "localhost:26381",
    }
    config.UseSlaveConnection = true // Kích hoạt kết nối đến slave
    config.SlaveReadOnly = true      // Đảm bảo slave chỉ dùng cho đọc

    // Tạo kết nối
    conn := _redis.NewSentinelConnection(config)

    // Kết nối đến Redis sử dụng Sentinel
    if err := conn.Connect(context.Background()); err != nil {
        panic(err)
    }
    defer conn.Close()

    // Lấy master client cho thao tác ghi
    masterClient := conn.GetMasterClient()
    // Sử dụng masterClient để ghi dữ liệu...

    // Kiểm tra và sử dụng slave client cho thao tác đọc
    if conn.HasSlaveConnected() {
        slaveClient := conn.GetSlaveClient()
        // Sử dụng slaveClient để đọc dữ liệu...
    }
}
```

## Cấu trúc API

Package Redis được thiết kế với các interface thống nhất:

-   `RedisClient`: Interface cơ bản cho tất cả các loại kết nối Redis
-   `SingleNodeClient`: Interface cho kết nối đơn node
-   `ClusterClient`: Interface cho kết nối Cluster
-   `SentinelClient`: Interface cho kết nối Sentinel

Tất cả các loại kết nối đều tuân theo mẫu API thống nhất với các phương thức:

-   `Connect()`: Thiết lập kết nối
-   `Close()`: Đóng kết nối
-   `Ping()`: Kiểm tra kết nối
-   `IsHealthy()`: Kiểm tra tình trạng kết nối
-   `GetMasterClient()`: Lấy client cho thao tác ghi
-   `GetSlaveClient()`: Lấy client cho thao tác đọc (nếu có)
-   `HasSlaveConnected()`: Kiểm tra có kết nối slave không

## Xử lý lỗi và Failover

-   **Redis Cluster**: Tự động xử lý failover khi một master node bị down
-   **Redis Sentinel**: Tự động phát hiện khi master bị down và chuyển đổi sang master mới

## Yêu cầu

-   Go 1.24 hoặc cao hơn
-   github.com/redis/go-redis/v9
