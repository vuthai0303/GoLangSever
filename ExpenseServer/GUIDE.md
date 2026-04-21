# ExpenseServer Guide

ExpenseServer là trung tâm lưu trữ toàn bộ các bút toán chi tiêu hàng ngày của người dùng thông qua giao diện truy xuất API GraphQL.

## Cấu trúc công nghệ
- **Framework HTTP:** `gin-gonic/gin`
- **GraphQL Engine:** `99designs/gqlgen` (Mô hình Schema-first nhằm giữ Code Base được đồng bộ, linh hoạt và tốn vi tính giới hạn nhất)
- **Database:** Sử dụng `database/sql` kết hợp `go-sqlite3` (Lưu lịch sử giao dịch tại file `expense.db` độc lập để tăng khả năng Scale-up)

## Hướng dẫn Run Server
1. Tiếp tục mở 1 cửa sổ phụ Terminal và di chuyển vào thư mục `ExpenseServer`
2. Compile Build và khởi động bằng lệnh:
   ```bash
   go mod tidy
   go build
   go run .
   ```
   > Server sẽ lắng nghe request HTTP tại cổng `8081`.

## Sử dụng GraphQL Playground (Môi trường Test)

Thay vì tích hợp Swagger cho GraphQL, Gqlgen có công cụ thay thế vượt trội hơn mang tên GraphQL Playground, cũng là giao diện Front-end, chiếm 0 điểm CPU Backend.
1. Sau khi dòng lệnh đã chạy được máy chủ, bạn mở [http://localhost:8081/](http://localhost:8081/) ở trình duyệt.
2. Nhằm chứng minh tính độc lập và bảo mật, giao diện sẽ chặn mọi request chưa được định danh. Việc của bạn là mở ứng dụng AuthServer (cổng 8080) thao tác API `/api/auth/signin` để nhận 1 **chuỗi JWT Token** cấp bởi AuthServer.
3. Ở dưới cùng tận cùng bên góc trái màn hình Playground có một nút **HTTP HEADERS**, bấm vào đó và chèn dòng Config JSON kèm Token của bạni:
   ```json
   {
     "Authorization": "Bearer <Dán_chuỗi_Token_của_bạn_vào_đây_sau_chữ_Bearer_với_dấu_cách>"
   }
   ```
4. Ở Layout phần bên trái, bạn có thể tự do sáng tạo các luồng thao tác. Nút **PLAY** ở giữa được dùng gọi lên Backend:
   ```graphql
   # Ví dụ 1: Xem tất cả tài khoản
   query GetAllAccounts {
     accounts { 
       id 
       name 
       type 
     }
   }
   
   # Ví dụ 2: Khởi tạo 1 giao dịch mới (Yêu cầu AccountID và CategoryID có thực)
   mutation CreateNewTx {
     createTransaction(input: {
       accountId: "1",
       categoryId: "2",
       amount: 50.0,
       notes: "Mua vé xem phim"
     })
   }
   ```

## 5. Tổng hợp tất cả các biến thể API có thể call với GraphQL hiện tại
Dưới đây là danh sách toàn bộ các Queries và Mutations dựa trên Schema mới nhất của hệ thống:

### Queries (Truy xuất dữ liệu)
Được sử dụng khi cần thiết lấy danh sách hoặc một mục từ DB.

```graphql
# 1. Lấy danh sách toàn bộ tài khoản (Accounts)
query GetAccounts {
  accounts {
    id
    userId
    name
    type
    defaultMoney
    amount
    iconId
    createdAt
    updatedAt
  }
}

# 2. Lấy danh sách các danh mục thu/chi (Categories)
query GetCategories {
  categories {
    id
    userId
    name
    categoriesType
    iconId
    createdAt
    updatedAt
  }
}

# 3. Lấy danh sách giao dịch (Transactions)
query GetTransactions {
  transactions {
    id
    userId
    accountName
    categoryName
    amount
    notes
    date
    createdAt
    updatedAt
  }
}

# 4. Lấy thống kê thu chi theo tháng (Monthly Expense Analysis)
# Có thể truyền parameter 'month' (ví dụ: "2026-04") hoặc bỏ trống để lấy hết.
query GetMonthlyAnalysis {
  monthlyExpenseAnalysis(month: "2026-04") {
    id
    userId
    spending
    income
    month
    createdAt
    updatedAt
  }
}
```

### Mutations (Thay đổi dữ liệu)
Được sử dụng cho các hành động thay đổi trạng thái (Thêm, Sửa, Xoá).

```graphql
# 1. Quản lý Tài Khoản (Accounts)
mutation {
  # Tạo tài khoản
  createAccount(input: {
    name: "Tiền mặt", 
    type: CASH, 
    defaultMoney: 1000.0,
    iconId: "wallet"
  }) { id name }
  
  # Cập nhật tài khoản
  updateAccount(id: "1", input: {
    name: "Tiền mặt (Mới)", 
    type: CASH, 
    defaultMoney: 2000.0
  }) { id name }
  
  # Xóa tài khoản
  deleteAccount(id: "1")
}

# 2. Quản lý Danh mục (Categories)
mutation {
  # Tạo danh mục
  createCategory(input: {
    name: "Ăn uống", 
    categoriesType: EXPENSE,
    iconId: "food"
  }) { id name }
  
  # Cập nhật danh mục
  updateCategory(id: "1", input: {
    name: "Đi chơi"
  }) { id name }
  
  # Xóa danh mục
  deleteCategory(id: "1")
}

# 3. Quản lý Giao dịch (Transactions)
mutation {
  # Tạo giao dịch (Sẽ tự động cập nhật amount của Account và Monthly Analysis)
  createTransaction(input: {
    accountId: "1", 
    categoryId: "1", 
    amount: 150.0, 
    notes: "Ăn trưa", 
    date: "2026-04-13" 
  }) { id accountName categoryName amount }
  
  # Cập nhật giao dịch
  updateTransaction(id: "1", input: {
    accountId: "1", 
    categoryId: "2", 
    amount: 200.0
  }) { id amount }
  
  # Xóa giao dịch
  deleteTransaction(id: "1")
}
```
