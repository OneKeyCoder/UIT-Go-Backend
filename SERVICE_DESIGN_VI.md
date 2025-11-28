## 5. THIẾT KẾ DỊCH VỤ CHI TIẾT

Tài liệu này mô tả chi tiết 5 dịch vụ trong hệ thống UIT-Go theo cách trực diện “API Endpoints”: mục đích, công nghệ, API HTTP, API gRPC, dữ liệu vào/ra, quy tắc nghiệp vụ, các chỉ số quan sát và lỗi thường gặp. Các API gRPC lấy từ thư mục `proto/` của repo.

Thuật ngữ nhanh (để khỏi nhầm)

-   API gRPC: Danh sách method được định nghĩa trong file `.proto` (vd: `AuthService.Authenticate`). Đây là “contract” giữa các service nội bộ.
-   API HTTP (Endpoints): Đường dẫn + method (vd: `POST /trip`) dành cho client/ops gọi qua HTTP.
-   Common middleware: Các thư viện nội bộ trong thư mục `common/` (Logger, Recovery, Prometheus metrics) dùng chung giữa các service. Đây không phải “công nghệ bên ngoài” mà là code chung của dự án.

Ghi chú chung

-   Xác thực: JWT (HS256). Endpoint được bảo vệ cần header: `Authorization: Bearer <access_token>`.
-   Lỗi HTTP: 4xx cho lỗi phía client, 5xx cho lỗi server. Body lỗi thống nhất: `{ "error": "...", "details": "..." }`.
-   Observability: mọi service có `/metrics`, `/health/live`, `/health/ready` để Prometheus/K8s/Compose kiểm tra.
-   Liên lạc giữa services: dùng gRPC qua mạng nội bộ.

---

## 5.1 API Gateway

Mục đích

-   Điểm vào duy nhất từ Internet; ủy quyền gọi tới microservices nội bộ qua gRPC.
-   Bảo vệ tuyến đầu: CORS, rate limit 100 req/phút/IP (burst 100), logging, metrics, health.

Công nghệ

-   Ngôn ngữ/Framework: Go + chi router v5
-   Thư viện: `didip/tollbooth` (rate limit), `prometheus/client_golang` (metrics)
-   Dùng chung: `common/middleware` (Logger, Recovery, Prometheus metrics)

API Endpoints (HTTP)

-   `GET /ping` → kiểm tra nhanh, 200 "pong".
-   `GET /health/live` | `GET /health/ready` → liveness/readiness.
-   `GET /metrics` → Prometheus metrics.
-   Legacy (tùy chọn): `POST /` → “Broker” endpoint kiểu cũ dùng cho demo trước đây (payload có field `action`). Không khuyến nghị dùng cho client mới; có thể loại bỏ khi không còn phụ thuộc.
-   `POST /grpc/auth` → Ủy quyền tới AuthService.Authenticate. Body:
    {
    "email": "user@example.com",
    "password": "secret"
    }
    Trả về: user + token pair (xem AuthService bên dưới).

-   Nhóm Location (yêu cầu JWT):

    -   `GET /location/me` → Lấy vị trí của chính user (gRPC GetLocation).
    -   `POST /location` → SetLocation.
    -   `GET /location/nearest?top_n=&radius=` → FindNearestUsers.
    -   `GET /location` → GetAllLocations (dùng cho admin/debug).

-   Nhóm Trip (yêu cầu JWT):
    -   `POST /trip` → CreateTrip.
    -   `PUT /trip/accept/{tripID}` | `PUT /trip/reject/{tripID}`.
    -   `GET /trip` → phân trang tất cả chuyến.
    -   `GET /trip/{tripID}` → chi tiết chuyến.
    -   `GET /trip/user` | `GET /trip/driver` → theo passenger/driver hiện tại.
    -   `GET /trip/suggested/{tripID}` → gợi ý tài xế.
    -   `PUT /trip/status/{tripID}` → UpdateTripStatus.
    -   `PUT /trip/cancel/{tripID}` → CancelTrip.
    -   `PUT /trip/review/{tripID}` | `GET /trip/review/{tripID}`.

Bảo mật & chính sách

-   Rate limit 100 req/phút/IP, cho phép burst 100.
-   JWT middleware ở các route `/location/*`, `/trip/*`.
-   CORS mở theo `http://*`, `https://*` (tinh chỉnh theo môi trường thực tế).

Observability

-   Metrics HTTP tự động từ middleware: `http_requests_total`, `http_request_duration_seconds{service="api-gateway"}`.

Lỗi thường gặp

-   401: token thiếu/hết hạn.
-   429: vượt rate limit.
-   502/504: backend gRPC không sẵn sàng hoặc timeout.

---

## 5.2 Authentication Service

Mục đích

-   Đăng ký, xác thực, phát hành/đổi mới JWT; validate token cho các dịch vụ khác. Quản lý hồ sơ cơ bản người dùng (giai đoạn 1).

Công nghệ & dữ liệu

-   Ngôn ngữ/Framework: Go + chi router
-   gRPC server
-   PostgreSQL: bảng `users`, (tùy chọn) lưu refresh tokens
-   Bảo mật: bcrypt hash password

API gRPC (service methods, `proto/auth/auth.proto`)

-   `Authenticate(AuthRequest) → AuthResponse`
-   `ValidateToken(ValidateTokenRequest) → ValidateTokenResponse`
-   `RefreshToken(RefreshTokenRequest) → AuthResponse`

API HTTP (endpoints nội bộ để debug/automation)

-   `POST /register` → tạo user mới. Body tối thiểu: email, password, first_name, last_name.
-   `POST /authenticate` → như gRPC Authenticate.
-   `POST /refresh` → như gRPC RefreshToken.
-   `POST /validate` → như gRPC ValidateToken.
-   `PATCH /change-password` → yêu cầu JWT; trường: `old_password`, `new_password`.

Quy tắc nghiệp vụ chính

-   Access token 24h; Refresh token 7 ngày.
-   Bắt buộc email unique; mật khẩu được băm bằng bcrypt (cost cấu hình).
-   Đăng nhập thành công phát sự kiện log qua RabbitMQ (routing `log.INFO`).

Observability

-   `/metrics` Prometheus; counter đăng nhập thành công/thất bại (qua middleware + log).

Lỗi & mã lỗi

-   400: input không hợp lệ.
-   401: sai thông tin hoặc token không hợp lệ.
-   409: email đã tồn tại (khi register).

---

## 5.3 Trip Service

Mục đích

-   Quản lý vòng đời chuyến đi: tạo yêu cầu, tài xế nhận/chối, cập nhật trạng thái, hủy chuyến, đánh giá. Tính cước dựa theo quãng đường và thời gian ước lượng từ HERE API.

Công nghệ & dữ liệu

-   Ngôn ngữ/Framework: Go + chi router
-   gRPC server
-   PostgreSQL: bảng `trips`
-   HERE API: ước tính distance/duration
-   Pricing: 5 USD/km (demo)

Trạng thái (enum `TripStatus`)

-   REQUESTED → ACCEPTED → STARTED → COMPLETED; có thể CANCELLED ở bất kỳ lúc nào bởi passenger/driver/system.

API gRPC (`proto/trip/trip.proto` – rút gọn)

-   `CreateTrip(CreateTripRequest) → CreateTripResponse` // tạo yêu cầu
-   `AcceptTrip(AcceptTripRequest) → MessageResponse`
-   `RejectTrip(RejectTripRequest) → MessageResponse`
-   `GetSuggestedDriver(TripIDRequest) → GetSuggestedDriverResponse`
-   `GetTripDetail(TripIDRequest) → GetTripDetailResponse`
-   `GetTripsByPassenger(GetTripsByUserIDRequest) → TripsResponse`
-   `GetTripsByDriver(GetTripsByUserIDRequest) → TripsResponse`
-   `GetAllTrips(GetAllTripsRequest) → PageResponse`
-   `UpdateTripStatus(UpdateTripStatusRequest) → MessageResponse`
-   `CancelTrip(CancelTripRequest) → MessageResponse`
-   `SubmitReview(SubmitReviewRequest) → MessageResponse`
-   `GetTripReview(TripIDRequest) → GetTripReviewResponse`

API HTTP (endpoints nội bộ phục vụ debug)

-   `POST /trip/create`, `PUT /trip/accept`, `PUT /trip/reject`, `GET /trip/suggested/{trip_id}`, `GET /trip/{trip_id}/{user_id}`, `GET /trip/passenger/{id}`, `GET /trip/driver/{id}`, `GET /trips/{page}/{limit}`, `PUT /trip/update`, `PUT /trip/cancel`, `PUT /trip/review`, `GET /trip/review/{trip_id}`.

Quy tắc nghiệp vụ chính

-   Tính cước: `fare = distance_km * 5` (đơn vị USD – có thể cấu hình theo vùng).
-   Driver không được Accept khi trip đã ACCEPTED/STARTED/COMPLETED/CANCELLED.
-   Passenger chỉ được hủy khi chưa COMPLETED; lưu `cancel_by_user_id`.
-   Review chỉ tạo sau COMPLETED; rating 1–5.

Observability

-   Metrics HTTP/gRPC; log thay đổi trạng thái; có thể emit event sang Logger.

Lỗi & cạnh biên

-   404: Trip không tồn tại/không thuộc quyền xem.
-   409: thay đổi trạng thái không hợp lệ (đi ngược state machine).
-   424: lỗi khi gọi HERE API để ước lượng (tùy chọn dùng mã này hoặc 502).

---

## 5.4 Location Service

Mục đích

-   Theo dõi vị trí thời gian thực và tìm người gần nhất. Dữ liệu lưu bộ nhớ (Valkey/Redis) với TTL để tự làm mới, tối ưu tốc độ truy vấn geospatial.

Công nghệ & dữ liệu

-   Ngôn ngữ/Framework: Go + Gin
-   gRPC server
-   Valkey/Redis 7.2. Cấu trúc:
    -   GEOSET `geo:users` (chỉ mục vị trí)
    -   STRING `{user_id}` → JSON Location (kèm speed/heading/timestamp), TTL mặc định 3600 giây.

API gRPC (`proto/location/location.proto`)

-   `SetLocation(SetLocationRequest) → SetLocationResponse`
-   `GetLocation(GetLocationRequest) → GetLocationResponse`
-   `FindNearestUsers(FindNearestUsersRequest) → FindNearestUsersResponse`
-   `GetAllLocations(GetAllLocationsRequest) → GetAllLocationsResponse`

API HTTP

-   `POST /location` → SetCurrentLocation
-   `GET /location?user_id=` → GetCurrentLocation
-   `GET /location/nearest?user_id=&top_n=&radius=`
-   `GET /location/all`

Quy tắc nghiệp vụ chính

-   Chỉ chấp nhận `latitude ∈ [-90, 90]`, `longitude ∈ [-180, 180]`.
-   Khi SetLocation: cập nhật GEOSET + JSON, gia hạn TTL.
-   FindNearest mặc định `top_n=10`, `radius=10km`; loại bỏ chính user khỏi kết quả.

Observability

-   Middleware tự đếm `location_service_http_requests_total` và `location_service_http_request_duration_seconds`.

Lỗi & cạnh biên

-   404: không có vị trí (hết TTL hoặc chưa từng set).
-   400: tham số radius/top_n không hợp lệ.

---

## 5.5 Logger Service

Mục đích

-   Ghi log tập trung vào MongoDB, hỗ trợ ingest trực tiếp qua gRPC/HTTP hoặc gián tiếp qua RabbitMQ (event-driven). Dùng cho audit, debug và phân tích.

Công nghệ & dữ liệu

-   Ngôn ngữ/Framework: Go + chi router
-   gRPC server
-   MongoDB (collection `logs`)
-   RabbitMQ consumer: exchange `logs_topic`, routing `log.INFO`

API gRPC (`proto/logger/logger.proto`)

-   `WriteLog(LogRequest{name, data}) → LogResponse{success, message}`
-   `GetLogs(GetLogsRequest{limit, level?}) → GetLogsResponse{logs[]}`

API HTTP

-   `POST /log` → ghi một log entry (dùng trong demo hoặc khi không tiện gRPC).

Luồng xử lý

-   Trực tiếp: service → gRPC `WriteLog` → Mongo.
-   Bất đồng bộ: service publish RabbitMQ → Logger consumer → Mongo (không làm chậm request chính).

Observability

-   `/metrics`, health endpoints; có thể tạo index Mongo trên `created_at` để query nhanh.

Lỗi & cạnh biên

-   503: Mongo down; Logger tạm thời queue nội bộ/thông báo lỗi.

---

## Phụ lục: Payload mẫu

Authentication

-   Request: { "email": "a@b.c", "password": "xxx" }
-   Response: {
    "success": true,
    "user": { "id": 1, "email": "a@b.c", "first_name": "A", "last_name": "B", "role": "passenger" },
    "tokens": { "access_token": "...", "refresh_token": "...", "access_expires_at": 1730000000, "refresh_expires_at": 1736000000 }
    }

Location

-   SetLocation request: { "user_id": 10, "role": "driver", "latitude": 10.78, "longitude": 106.65, "speed": 12.3, "heading": "NE", "timestamp": "2025-11-02T10:20:30Z" }

Trip

-   CreateTrip request: { "passenger_id": 10, "origin_lat": 10.78, "origin_lng": 106.65, "dest_lat": 10.76, "dest_lng": 106.70, "payment_method": "cash" }
-   UpdateStatus request: { "trip_id": 123, "driver_id": 45, "status": "STARTED" }

---

## Hướng dẫn kiểm thử nhanh

-   Unit test: handler/service cho mỗi service (happy path + 1–2 edge cases như token sai, tham số thiếu, DB lỗi).
-   Health: `curl` hoặc probe `/health/ready` trước khi gửi lưu lượng thật.
-   Rate limit: test burst 100 requests → không trả 429 trong burst, sau đó refill 100/min.
-   Observability: kiểm tra Prometheus thấy series tăng, Grafana dashboard hiển thị latency p95.

---

Kết thúc mục 5. Nội dung bám sát hiện trạng mã nguồn (branch Modernize) và có thể mở rộng khi thêm User Service độc lập trong giai đoạn tiếp theo.
