# GoLang Microservices: Auth & Expense Server

Tài liệu này mô tả chi tiết về kiến trúc, công cụ được dùng và luồng chạy (workflow) của 2 servers trong dự án. Mục tiêu hệ thống là chạy mượt mà trên server cấu hình tối thiểu (1 CPU, 1GB RAM) của Oracle Cloud.

## 1. Kiến trúc Tổng Quan (Architecture)

Hệ thống được chia làm 2 microservices hoàn toàn biệt lập để dễ bảo trì và phân tải:
- **AuthServer**: Cung cấp các Endpoint RESTful chuẩn để định danh người dùng.
- **ExpenseServer**: Cung cấp GraphQL query/mutation phục vụ chức năng chính là quản lý sổ thu/chi.

Hai server này **không giao tiếp trực tiếp qua HTTP** đối với mỗi request (như vậy sẽ làm nghẽn và tốn tài nguyên mạng/CPU). Thay vào đó, chúng chia sẻ chung công thức nhận diện JWT (thông qua cùng một Secret Key).
- **Cơ sở dữ liệu (Database)**: Cả hai sử dụng hệ quản trị CSDL nhúng `SQLite`, tạo thành 2 file vật lý riêng biệt là `auth.db` và `expense.db`. SQLite không tốn RAM chạy nền như MySQL/PostgreSQL nên cực kì phù hợp cho máy ảo Cấu hình thấp. 

## 2. Công Nghệ Sử Dụng (Tech Stack)

### AuthServer
1. **Ngôn ngữ**: Golang
2. **Web Framework**: Thư viện `go-chi/chi/v5` nhẹ nhàng, dễ config Router kết hợp module HTTP thuần `net/http`.
3. **Database Driver**: `database/sql` + `github.com/mattn/go-sqlite3` tối đa hóa hiệu năng truy xuất bằng lệnh Raw SQL.
4. **Security & Auth**: 
   - Mã hóa mật khẩu: Thư viện `golang.org/x/crypto/bcrypt`.
   - Cấp phát Token: Thư viện `github.com/golang-jwt/jwt/v5`.

### ExpenseServer
1. **Ngôn ngữ**: Golang
2. **Web Framework**: Thư viện `github.com/gin-gonic/gin` hiệu suất cao, cung cấp Middleware tiện lợi kết nối Client và System.
3. **GraphQL Framework**: Thư viện `github.com/99designs/gqlgen` (Cách tiếp cận tự generate code GraphQL ưu việt `Schema-First`).
4. **Database Driver**: Tương tự như hệ thống kia, tiếp cận Raw SQL qua module `database/sql`.

## 3. Workflow & Data Flow

1. **Bước 1: Đăng Ký / Đăng Nhập (Logic tại AuthServer)**
   - Người dùng mới gửi yêu cầu REST API vào Endpoint `POST /api/auth/signup`.
   - Server băm mật khẩu ra hash và lưu vào bảng `users`.
   - Người dùng gọi API `POST /api/auth/signin` để lấy `Access Token`. Bảng Payload của JWT sẽ được chèn thêm trường `user_id`.

2. **Bước 2: Nạp Token vào GraphQL (Logic tại ExpenseServer)**
   - Web / App Client nhận thẻ Token và gắn vào HTTP Header `Authorization: Bearer <TOKEN>` trong mọi lệnh GraphQL POST.
   - Khi request đi qua lớp mạng của `Gin`, nó sẽ vấp phải `AuthMiddleware`. Gin kiểm tra tính hợp lệ của token và Decode nó trong phần nghìn giây. 
   - Nếu token lỗi/hết hạn, API Gin từ chối và trả về 401 Unauthorized nhanh chóng mà không cần tốn tài nguyên kích hoạt logic đồ sộ chằng chịt của GraphQL phía sau. 
   - Nếu token hợp lệ, Gin lôi `user_id` ra và nạp (Inject) vào Context của GraphQL HTTP Request chặn cuối cùng.

3. **Bước 3: Xử lý Logic Database (Logic bên trong ExpenseServer Resolver)**
   - Gqlgen đón lấy Context. Các hàm Resolver sử dụng `user_id` ẩn này để query SQL trích xuất chỉ những Account, Expense, Category dành riêng cho tài khoản User tương ứng từ SQLite lên.

## 4. Hướng dẫn chạy và sử dụng dự án (Run Guide)

Do hai server hoàn toàn tách biệt, bạn cần mở **2 Tab Terminal** để chạy song song. Ở máy ảo Linux (chạy nền) bạn có thể dùng công cụ Systemd hoặc lệnh `nohup`.

**Tab 1 (Cửa sổ khởi chạy AuthServer)**
```bash
cd AuthServer
go mod tidy
go run .
# Nếu thành công Terminal sẽ báo: AuthServer running on :8080
```

**Tab 2 (Cửa sổ khởi chạy ExpenseServer)**
```bash
cd ExpenseServer
go mod tidy
go run .
# Nếu thành công Terminal sẽ báo: ExpenseServer running on :8081
```

### 4.2. Test API với AuthServer (REST)
Tại Tab Terminal số 3 hoặc qua công cụ **Postman / Insomnia**. Call tới cổng 8080.
**1. Tạo tài khoản mới**
```bash
curl -X POST http://localhost:8080/api/auth/signup \
-H "Content-Type: application/json" \
-d '{"username": "admin123", "password": "password123"}'
```

**2. Đăng nhập để lấy Token phân quyền**
```bash
curl -X POST http://localhost:8080/api/auth/signin \
-H "Content-Type: application/json" \
-d '{"username": "admin123", "password": "password123"}'
```
> Kết quả là máy chủ cấp cho bạn một chuỗi Token. Bạn hãy Copy đoạn mã đó lại.

### 4.3. Test GraphQL với ExpenseServer
ExpenseServer đã được code sẵn **GraphQL Playground** hiển thị ngay trên Web.

1. Mở cửa sổ ẩn danh trình duyệt và truy cập ở link: http://localhost:8081/
2. Giao diện đồ họa để viết Query hiện ra. Bạn chưa có quyền nên gửi query sẽ báo lỗi! Hãy ấn mục `HTTP HEADERS` ở sát góc dưới cùng bên trái.
3. Cài đặt Headers với chuỗi token bạn lấy được ở trên:
```json
{
  "Authorization": "Bearer <Dán_Token_Dài_Của_Bạn_Vào_Đây>"
}
```
4. Bắt đầu viết lệnh Query ở khung bên trái và bấm nút **PLAY**. 
Ví dụ bạn nhập lệnh tạo một Ví để cất giữ nguồn tiền (Account):
```graphql
mutation CreateMyAccount {
  createAccount(input: {
    name: "Tài khoản Sinh Lời Vietcombank",
    type: "bank"
  }) {
    id
    name
    userId
  }
}
```
5. Đọc danh sách Ví tiền mà bạn vừa tạo:
```graphql
query GetAllAccounts {
  accounts {
    id
    name
  }
}
```
