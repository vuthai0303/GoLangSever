# AuthServer Guide

AuthServer là máy chủ chịu trách nhiệm định danh (Authentication) và cấp phát Access Token (JWT).

## Cấu trúc công nghệ
- **Framework HTTP:** `go-chi/chi/v5`
- **Database:** Sử dụng thư viện `database/sql` kết hợp `go-sqlite3` (Cơ sở dữ liệu lưu tại file `auth.db`)
- **API Documentation:** Tích hợp bộ mã nguồn mở `swaggo/swag` (Swagger UI Render trực tiếp ở phía Frontend HTML).

## Hướng dẫn Run Server
1. Mở công cụ Terminal (hoặc CMD/PowerShell) và di chuyển vào thư mục `AuthServer`
2. Tải toàn bộ thư viện cần thiết và Build mã:
   ```bash
   go mod tidy
   go build
   ```
3. Chạy Server:
   ```bash
   go run .
   ```
   > Server sẽ khởi động trên cổng nội bộ `8080`.

## Swagger UI Documentation
Để xem chi tiết API Documentation của server này:
1. Mở trình duyệt Web của bạn.
2. Truy cập đường dẫn: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
3. Mọi Endpoint của `AuthServer` đã được liệt kê chi tiết tại đây. Bạn có thể bấm vào từng nhánh API, ấn nút **"Try it out"** để điền thông số và test gọi điện lên Server mà không cần cài đặt thêm phần mềm như Postman hay cURL.
