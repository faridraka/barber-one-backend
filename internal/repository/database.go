package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ==========================================
// User Repository Methods
// ==========================================

func (r *Repository) CreateUser(ctx context.Context, name, phone string, email *string, passwordHash string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name, phone, email, password_hash, role)
		 VALUES ($1, $2, $3, $4, 'user')
		 RETURNING id`,
		name, phone, email, passwordHash,
	).Scan(&id)
	return id, err
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (map[string]any, error) {
	var userID int64
	var name, phone, role string
	var email *string
	err := r.db.QueryRow(ctx,
		`SELECT id, name, phone, email, role FROM users WHERE id = $1`,
		id,
	).Scan(&userID, &name, &phone, &email, &role)
	return map[string]any{
		"id":    userID,
		"name":  name,
		"phone": phone,
		"email": email,
		"role":  role,
	}, err
}

func (r *Repository) GetUserByIdentifier(ctx context.Context, identifier string) (map[string]any, error) {
	var userID int64
	var name, phone, role, passwordHash string
	var email *string
	err := r.db.QueryRow(ctx,
		`SELECT id, name, phone, email, role, password_hash
		 FROM users
		 WHERE phone = $1 OR lower(email) = lower($1)
		 LIMIT 1`,
		identifier,
	).Scan(&userID, &name, &phone, &email, &role, &passwordHash)
	return map[string]any{
		"id":             userID,
		"name":           name,
		"phone":          phone,
		"email":          email,
		"role":           role,
		"password_hash":  passwordHash,
	}, err
}

// ==========================================
// Service Repository Methods
// ==========================================

func (r *Repository) ListServices(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, price, duration_minutes
		 FROM services
		 ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []map[string]any
	for rows.Next() {
		var id int64
		var name string
		var description *string
		var price float64
		var duration int
		err := rows.Scan(&id, &name, &description, &price, &duration)
		if err != nil {
			return nil, err
		}
		services = append(services, map[string]any{
			"id":                id,
			"name":              name,
			"description":       description,
			"price":             price,
			"duration_minutes":  duration,
		})
	}
	return services, rows.Err()
}

func (r *Repository) GetServiceByID(ctx context.Context, id int64) (map[string]any, error) {
	var dbID int64
	var name string
	var description *string
	var price float64
	var duration int
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, price, duration_minutes FROM services WHERE id = $1`, id,
	).Scan(&dbID, &name, &description, &price, &duration)
	return map[string]any{
		"id":                dbID,
		"name":              name,
		"description":       description,
		"price":             price,
		"duration_minutes":  duration,
	}, err
}

// ==========================================
// Barber Repository Methods
// ==========================================

func (r *Repository) ListBarbers(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, rating, available FROM barber ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var barbers []map[string]any
	for rows.Next() {
		var id int64
		var name string
		var rating float64
		var available bool
		err := rows.Scan(&id, &name, &rating, &available)
		if err != nil {
			return nil, err
		}
		barbers = append(barbers, map[string]any{
			"id":        id,
			"name":      name,
			"rating":    rating,
			"available": available,
		})
	}
	return barbers, rows.Err()
}

func (r *Repository) GetBarberByID(ctx context.Context, id int64) (map[string]any, error) {
	var dbID int64
	var name string
	var rating float64
	var available bool
	err := r.db.QueryRow(ctx, `SELECT id, name, rating, available FROM barber WHERE id = $1`, id).
		Scan(&dbID, &name, &rating, &available)
	return map[string]any{
		"id":        dbID,
		"name":      name,
		"rating":    rating,
		"available": available,
	}, err
}

// ==========================================
// Booking Repository Methods
// ==========================================

func (r *Repository) CreateBooking(ctx context.Context, ticketCode string, userID *int64, customerName, customerPhone string, serviceID, barberID int64, bookingDate, bookingTime string, notes *string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO bookings
		 (ticket_code, user_id, customer_name, customer_phone, service_id, barber_id, booking_date, booking_time, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		ticketCode, userID, customerName, customerPhone,
		serviceID, barberID, bookingDate, bookingTime, notes,
	).Scan(&id)
	return id, err
}

func (r *Repository) GetUserBookings(ctx context.Context, userID int64) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx, `
		SELECT b.id, b.ticket_code, b.status, b.customer_name, b.customer_phone,
		       b.service_id, s.name, s.price, s.duration_minutes,
		       b.barber_id, br.name, b.booking_date, b.booking_time, COALESCE(b.notes, ''),
		       COALESCE(t.total_amount, s.price), t.payment_method, t.payment_status, t.gateway_redirect_url, t.expires_at
		FROM bookings b
		JOIN services s ON s.id = b.service_id
		JOIN barber br ON br.id = b.barber_id
		LEFT JOIN LATERAL (
			SELECT total_amount, payment_method, payment_status, gateway_redirect_url, expires_at
			FROM transactions WHERE booking_id = b.id ORDER BY id DESC LIMIT 1
		) t ON TRUE
		WHERE b.user_id = $1
		ORDER BY b.booking_date DESC, b.booking_time DESC, b.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []map[string]any
	for rows.Next() {
		var id int64
		var ticketCode, status, customerName, customerPhone, notes string
		var serviceID int64
		var serviceName string
		var price float64
		var durationMinutes int
		var barberID int64
		var barberName string
		var bookingDate, bookingTime string
		var totalAmount float64
		var paymentMethod, paymentStatus, gatewayRedirectURL *string
		var expiresAt *time.Time

		err := rows.Scan(&id, &ticketCode, &status, &customerName, &customerPhone,
			&serviceID, &serviceName, &price, &durationMinutes,
			&barberID, &barberName, &bookingDate, &bookingTime, &notes,
			&totalAmount, &paymentMethod, &paymentStatus, &gatewayRedirectURL, &expiresAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, map[string]any{
			"id":                    id,
			"ticket_code":           ticketCode,
			"status":                status,
			"customer_name":         customerName,
			"customer_phone":        customerPhone,
			"service_id":            serviceID,
			"service_name":          serviceName,
			"price":                 price,
			"duration_minutes":      durationMinutes,
			"barber_id":             barberID,
			"barber_name":           barberName,
			"booking_date":          bookingDate,
			"booking_time":          bookingTime,
			"notes":                 notes,
			"total_amount":          totalAmount,
			"payment_method":        paymentMethod,
			"payment_status":        paymentStatus,
			"payment_redirect_url":  gatewayRedirectURL,
			"expires_at":            expiresAt,
		})
	}
	return bookings, rows.Err()
}

func (r *Repository) GetBookingByID(ctx context.Context, bookingID, userID int64) (map[string]any, error) {
	var id int64
	var ticketCode, status, customerName, customerPhone, notes string
	var serviceID int64
	var serviceName, barberName string
	var price float64
	var durationMinutes int
	var barberID int64
	var bookingDate, bookingTime string
	var chairNumber int
	var totalAmount float64
	var paymentMethod, paymentStatus *string
	var gatewayRedirectURL *string
	var expiresAt *time.Time

	err := r.db.QueryRow(ctx, `
		SELECT b.id, b.ticket_code, b.status, b.customer_name, b.customer_phone,
		       b.service_id, s.name, s.price, s.duration_minutes,
		       b.barber_id, br.name, b.booking_date, b.booking_time, b.chair_number, COALESCE(b.notes, ''),
		       COALESCE(t.total_amount, s.price), t.payment_method, t.payment_status, t.gateway_redirect_url, t.expires_at
		FROM bookings b
		JOIN services s ON s.id = b.service_id
		JOIN barber br ON br.id = b.barber_id
		LEFT JOIN LATERAL (
			SELECT total_amount, payment_method, payment_status, gateway_redirect_url, expires_at
			FROM transactions WHERE booking_id = b.id ORDER BY id DESC LIMIT 1
		) t ON TRUE
		WHERE b.id = $1 AND b.user_id = $2`, bookingID, userID).Scan(
		&id, &ticketCode, &status, &customerName, &customerPhone,
		&serviceID, &serviceName, &price, &durationMinutes,
		&barberID, &barberName, &bookingDate, &bookingTime, &chairNumber, &notes,
		&totalAmount, &paymentMethod, &paymentStatus, &gatewayRedirectURL, &expiresAt,
	)

	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":                    id,
		"ticket_code":           ticketCode,
		"status":                status,
		"customer_name":         customerName,
		"customer_phone":        customerPhone,
		"service_id":            serviceID,
		"service_name":          serviceName,
		"price":                 price,
		"duration_minutes":      durationMinutes,
		"barber_id":             barberID,
		"barber_name":           barberName,
		"booking_date":          bookingDate,
		"booking_time":          bookingTime,
		"chair_number":          chairNumber,
		"notes":                 notes,
		"total_amount":          totalAmount,
		"payment_method":        paymentMethod,
		"payment_status":        paymentStatus,
		"payment_redirect_url":  gatewayRedirectURL,
		"expires_at":            expiresAt,
	}, nil
}

func (r *Repository) UpdateBookingStatus(ctx context.Context, bookingID int64, status string) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE bookings SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		status, bookingID,
	)
	return tag.RowsAffected(), err
}

func (r *Repository) UpdateBooking(ctx context.Context, bookingID int64, bookingDate, bookingTime string) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE bookings SET booking_date = $1, booking_time = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`,
		bookingDate, bookingTime, bookingID,
	)
	return tag.RowsAffected(), err
}

func (r *Repository) CheckSlotAvailability(ctx context.Context, barberID int64, bookingDate, bookingTime string, duration int) (bool, error) {
	var hasConflict bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM bookings b JOIN services s ON s.id = b.service_id
			WHERE b.barber_id = $1 AND b.booking_date = $2 AND b.status <> 'Batal'
			AND ($3::time < b.booking_time::time + s.duration_minutes * interval '1 minute'
			 AND b.booking_time::time < $3::time + $4 * interval '1 minute')
		)`, barberID, bookingDate, bookingTime, duration,
	).Scan(&hasConflict)
	return hasConflict, err
}

// ==========================================
// Transaction Repository Methods
// ==========================================

func (r *Repository) CreateTransaction(ctx context.Context, bookingID int64, amount float64, paymentMethod, gatewayOrderID, gatewayToken, gatewayRedirectURL string, expiresAt time.Time) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO transactions (booking_id, total_amount, payment_method, payment_status, gateway_order_id, gateway_token, gateway_redirect_url, expires_at, paid_at) 
		 VALUES ($1, $2, 'Midtrans', 'Menunggu', $3, $4, $5, $6, NULL)
		 RETURNING id`,
		bookingID, amount, gatewayOrderID, gatewayToken, gatewayRedirectURL, expiresAt,
	).Scan(&id)
	return id, err
}

func (r *Repository) UpdateTransactionPaymentStatus(ctx context.Context, gatewayOrderID, paymentStatus string, paidAt time.Time) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE transactions SET payment_method='Midtrans', payment_status=$1, paid_at=CASE WHEN $2 THEN CURRENT_TIMESTAMP ELSE paid_at END WHERE gateway_order_id=$3`,
		paymentStatus, paymentStatus == "Lunas", gatewayOrderID,
	)
	return tag.RowsAffected(), err
}

func (r *Repository) GetTransactionByGatewayOrderID(ctx context.Context, gatewayOrderID string) (map[string]any, error) {
	var id int64
	var paymentStatus string
	var redirectURL *string
	var expiresAt *time.Time
	err := r.db.QueryRow(ctx,
		`SELECT id, payment_status, COALESCE(gateway_redirect_url, ''), expires_at
		 FROM transactions WHERE booking_id = (SELECT booking_id FROM transactions WHERE gateway_order_id=$1) ORDER BY id DESC LIMIT 1`,
		gatewayOrderID,
	).Scan(&id, &paymentStatus, &redirectURL, &expiresAt)
	return map[string]any{
		"id":             id,
		"payment_status": paymentStatus,
		"redirect_url":   redirectURL,
		"expires_at":     expiresAt,
	}, err
}

func (r *Repository) GetTransactionForUpdate(ctx context.Context, bookingID int64) (map[string]any, error) {
	var id int64
	var paymentStatus string
	var redirectURL *string
	var expiresAt *time.Time
	err := r.db.QueryRow(ctx,
		`SELECT id, payment_status, COALESCE(gateway_redirect_url, ''), expires_at
		 FROM transactions WHERE booking_id = $1 ORDER BY id DESC LIMIT 1 FOR UPDATE`,
		bookingID,
	).Scan(&id, &paymentStatus, &redirectURL, &expiresAt)
	return map[string]any{
		"id":             id,
		"payment_status": paymentStatus,
		"redirect_url":   redirectURL,
		"expires_at":     expiresAt,
	}, err
}

// ==========================================
// Review Repository Methods
// ==========================================

func (r *Repository) CreateReview(ctx context.Context, bookingID int64, rating int, comment *string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO reviews (booking_id, rating, comment)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		bookingID, rating, comment,
	).Scan(&id)
	return id, err
}

func (r *Repository) ListReviews(ctx context.Context, barberID *int64) ([]map[string]any, error) {
	query := `
		SELECT rv.id, rv.booking_id, rv.rating, COALESCE(rv.comment, ''), rv.created_at,
		       b.customer_name, s.name, br.name
		FROM reviews rv
		JOIN bookings b ON b.id = rv.booking_id
		JOIN services s ON s.id = b.service_id
		JOIN barber br ON br.id = b.barber_id`
	args := []any{}
	if barberID != nil {
		query += ` WHERE br.id = $1`
		args = append(args, *barberID)
	}
	query += ` ORDER BY rv.created_at DESC, rv.id DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []map[string]any
	for rows.Next() {
		var id, bookingID int64
		var rating int
		var comment string
		var createdAt time.Time
		var customerName, serviceName, barberName string
		err := rows.Scan(&id, &bookingID, &rating, &comment, &createdAt,
			&customerName, &serviceName, &barberName)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, map[string]any{
			"id":           id,
			"booking_id":   bookingID,
			"rating":       rating,
			"comment":      comment,
			"created_at":   createdAt,
			"customer_name": customerName,
			"service_name": serviceName,
			"barber_name":  barberName,
		})
	}
	return reviews, rows.Err()
}

// ==========================================
// Availability Repository Methods
// ==========================================

func (r *Repository) GetBookedSlots(ctx context.Context, barberID int64, bookingDate string) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx,
		`SELECT b.booking_time, s.duration_minutes
		 FROM bookings b JOIN services s ON s.id = b.service_id
		 WHERE b.barber_id = $1 AND b.booking_date = $2 AND b.status <> 'Batal'`,
		barberID, bookingDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []map[string]any
	for rows.Next() {
		var bookingTime string
		var duration int
		err := rows.Scan(&bookingTime, &duration)
		if err != nil {
			return nil, err
		}
		slots = append(slots, map[string]any{
			"booking_time":    bookingTime,
			"duration_minutes": duration,
		})
	}
	return slots, rows.Err()
}

func (r *Repository) CheckServiceDuration(ctx context.Context, serviceID int64) (int, error) {
	var duration int
	err := r.db.QueryRow(ctx,
		`SELECT duration_minutes FROM services WHERE id = $1`,
		serviceID,
	).Scan(&duration)
	return duration, err
}

func (r *Repository) GetBarberDetails(ctx context.Context, barberID int64) (bool, error) {
	var available bool
	err := r.db.QueryRow(ctx,
		`SELECT available FROM barber WHERE id = $1`,
		barberID,
	).Scan(&available)
	return available, err
}

// ==========================================
// Cashier Repository Methods
// ==========================================

func (r *Repository) GetCashierQueue(ctx context.Context, date string) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx, `
		SELECT b.id, b.ticket_code, b.status, b.user_id IS NULL, b.customer_name, b.customer_phone,
		       s.name, s.duration_minutes, b.barber_id, br.name, b.chair_number, b.booking_date, b.booking_time,
		       COALESCE(t.total_amount, s.price), t.payment_method, t.payment_status, b.started_at
		FROM bookings b
		JOIN services s ON s.id = b.service_id
		JOIN barber br ON br.id = b.barber_id
		LEFT JOIN LATERAL (
			SELECT total_amount, payment_method, payment_status
			FROM transactions WHERE booking_id = b.id ORDER BY id DESC LIMIT 1
		) t ON TRUE
		WHERE b.booking_date = $1 AND b.status <> 'Batal'
		  AND (b.status <> 'Menunggu' OR t.payment_status = 'Lunas')
		ORDER BY b.booking_time`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queue []map[string]any
	for rows.Next() {
		var id int64
		var ticketCode, status string
		var isWalkIn bool
		var customerName, customerPhone, serviceName string
		var durationMinutes int
		var barberID int64
		var barberName string
		var chairNumber int
		var bookingDate, bookingTime string
		var totalAmount float64
		var paymentMethod, paymentStatus *string
		var startedAt *time.Time

		err := rows.Scan(&id, &ticketCode, &status, &isWalkIn, &customerName, &customerPhone,
			&serviceName, &durationMinutes, &barberID, &barberName, &chairNumber, &bookingDate, &bookingTime,
			&totalAmount, &paymentMethod, &paymentStatus, &startedAt)
		if err != nil {
			return nil, err
		}
		queue = append(queue, map[string]any{
			"id":               id,
			"ticket_code":      ticketCode,
			"status":           status,
			"is_walk_in":       isWalkIn,
			"customer_name":    customerName,
			"customer_phone":   customerPhone,
			"service_name":     serviceName,
			"duration_minutes": durationMinutes,
			"barber_id":        barberID,
			"barber_name":      barberName,
			"chair_number":     chairNumber,
			"booking_date":     bookingDate,
			"booking_time":     bookingTime,
			"total_amount":     totalAmount,
			"payment_method":   paymentMethod,
			"payment_status":   paymentStatus,
			"started_at":       startedAt,
		})
	}
	return queue, rows.Err()
}

func (r *Repository) GetCashierSummary(ctx context.Context, date string) (map[string]any, error) {
	var waiting, inService, done int
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE b.status = 'Dikonfirmasi' OR (b.status = 'Menunggu' AND t.payment_status = 'Lunas')) AS waiting,
			COUNT(*) FILTER (WHERE b.status = 'Sedang Dilayani') AS in_service,
			COUNT(*) FILTER (WHERE b.status = 'Selesai') AS done
		FROM bookings b
		LEFT JOIN LATERAL (
			SELECT payment_status FROM transactions WHERE booking_id = b.id ORDER BY id DESC LIMIT 1
		) t ON TRUE
		WHERE b.booking_date = $1`, date,
	).Scan(&waiting, &inService, &done)
	return map[string]any{
		"waiting":   waiting,
		"in_service": inService,
		"done":      done,
	}, err
}

func (r *Repository) GetRevenueSummary(ctx context.Context, date string) (map[string]any, error) {
	var revenue float64
	var cashCount, nonCashCount int
	err := r.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(t.total_amount), 0) AS revenue,
			COUNT(*) FILTER (WHERE t.payment_method = 'Tunai') AS cash_count,
			COUNT(*) FILTER (WHERE t.payment_method <> 'Tunai') AS non_cash_count
		FROM transactions t JOIN bookings b ON b.id = t.booking_id
		WHERE b.booking_date = $1 AND t.payment_status = 'Lunas'`, date,
	).Scan(&revenue, &cashCount, &nonCashCount)
	return map[string]any{
		"revenue":      revenue,
		"cash_count":   cashCount,
		"non_cash_count": nonCashCount,
	}, err
}

// ==========================================
// Worker Repository Methods
// ==========================================

func (r *Repository) GetDueBookings(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx, `
		SELECT b.id, b.barber_id
		FROM bookings b
		LEFT JOIN LATERAL (
			SELECT payment_status FROM transactions WHERE booking_id = b.id ORDER BY id DESC LIMIT 1
		) t ON TRUE
		WHERE b.booking_date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')::date
		  AND b.booking_date + b.booking_time::time <= CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta'
		  AND (b.status = 'Dikonfirmasi' OR b.status = 'Menunggu')
		  AND t.payment_status = 'Lunas'
		  AND NOT EXISTS (
			SELECT 1 FROM bookings busy WHERE busy.barber_id = b.barber_id AND busy.status = 'Sedang Dilayani'
		  )
		ORDER BY b.booking_time, b.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var due []map[string]any
	for rows.Next() {
		var id, barberID int64
		err := rows.Scan(&id, &barberID)
		if err != nil {
			return nil, err
		}
		due = append(due, map[string]any{
			"id":       id,
			"barber_id": barberID,
		})
	}
	return due, rows.Err()
}

func (r *Repository) GetUsedChairs(ctx context.Context) ([]int, error) {
	rows, err := r.db.Query(ctx, `SELECT chair_number FROM bookings WHERE status = 'Sedang Dilayani'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chairs []int
	for rows.Next() {
		var chair int
		if err := rows.Scan(&chair); err != nil {
			return nil, err
		}
		chairs = append(chairs, chair)
	}
	return chairs, rows.Err()
}

func (r *Repository) UpdateBookingChair(ctx context.Context, bookingID, chairID int64) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE bookings SET status = 'Sedang Dilayani', chair_number = $1, started_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND status IN ('Menunggu', 'Dikonfirmasi')`,
		chairID, bookingID,
	)
	return tag.RowsAffected(), err
}

func (r *Repository) ExpirePayments(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		WITH expired AS (
			SELECT id
			FROM transactions
			WHERE payment_status = 'Menunggu'
			  AND expires_at IS NOT NULL
			  AND expires_at <= CURRENT_TIMESTAMP - INTERVAL '1 minute'
			ORDER BY expires_at
			FOR UPDATE SKIP LOCKED
			LIMIT 100
		), failed AS (
			UPDATE transactions t
			SET payment_status = 'Gagal'
			FROM expired e
			WHERE t.id = e.id AND t.payment_status = 'Menunggu'
			RETURNING t.booking_id
		)
		UPDATE bookings b
		SET status = 'Batal', updated_at = CURRENT_TIMESTAMP
		FROM failed f
		WHERE b.id = f.booking_id AND b.status = 'Menunggu'
	`)
	return tag.RowsAffected(), err
}

func (r *Repository) GetBookingForUpdate(ctx context.Context, bookingID, userID int64) (map[string]any, error) {
	var id int64
	var ticketCode, status string
	var serviceID int64
	var durationMinutes int
	var barberID int64
	var bookingDate, bookingTime string
	var createdAt time.Time
	err := r.db.QueryRow(ctx,
		`SELECT b.id, b.ticket_code, b.status, b.service_id, s.duration_minutes,
		       b.barber_id, b.booking_date, b.booking_time, b.created_at
		 FROM bookings b JOIN services s ON s.id = b.service_id
		 WHERE b.id = $1 AND b.user_id = $2 FOR UPDATE OF b`, bookingID, userID,
	).Scan(&id, &ticketCode, &status, &serviceID, &durationMinutes,
		&barberID, &bookingDate, &bookingTime, &createdAt)
	return map[string]any{
		"id":             id,
		"ticket_code":    ticketCode,
		"status":         status,
		"service_id":     serviceID,
		"duration_minutes": durationMinutes,
		"barber_id":      barberID,
		"booking_date":   bookingDate,
		"booking_time":   bookingTime,
		"created_at":     createdAt,
	}, err
}

func (r *Repository) CanChangeBooking(ctx context.Context, bookingID, userID int64) (map[string]any, error) {
	var status string
	var createdAt time.Time
	err := r.db.QueryRow(ctx,
		`SELECT b.status, b.created_at FROM bookings b
		 WHERE b.id = $1 AND b.user_id = $2`,
		bookingID, userID,
	).Scan(&status, &createdAt)
	return map[string]any{
		"status":     status,
		"created_at": createdAt,
	}, err
}

func (r *Repository) CreateWalkInBooking(ctx context.Context, ticketCode, customerName, customerPhone string, serviceID, barberID int64, bookingDate, bookingTime string, cashierID int64, paymentMethod string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO bookings (ticket_code, customer_name, customer_phone, service_id, barber_id, booking_date, booking_time, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'Dikonfirmasi')
		 RETURNING id`,
		ticketCode, customerName, customerPhone, serviceID, barberID, bookingDate, bookingTime,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	// Insert transaction for walk-in (already paid)
	_, err = r.db.Exec(ctx,
		`INSERT INTO transactions (booking_id, cashier_id, total_amount, payment_method, payment_status, paid_at)
		 VALUES ($1, $2, $3, $4, 'Lunas', CURRENT_TIMESTAMP)`,
		id, cashierID, serviceID, paymentMethod,
	)
	return id, err
}

func (r *Repository) GetServicePrice(ctx context.Context, serviceID int64) (float64, error) {
	var price float64
	err := r.db.QueryRow(ctx,
		`SELECT price FROM services WHERE id = $1`,
		serviceID,
	).Scan(&price)
	return price, err
}

// ==========================================
// Ticket Repository Methods
// ==========================================

func (r *Repository) GetMyTickets(ctx context.Context, userID int64) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx,
		`SELECT b.id, b.ticket_code, b.status, b.booking_date, b.booking_time, b.chair_number,
		        COALESCE(b.notes, ''), s.name, s.description, s.duration_minutes, s.price,
		        COALESCE(t.total_amount, s.price), br.name, br.rating, t.payment_method, t.payment_status,
		        t.gateway_redirect_url, t.expires_at
		 FROM bookings b
		 JOIN services s ON s.id = b.service_id
		 JOIN barber br ON br.id = b.barber_id
		 LEFT JOIN LATERAL (
		     SELECT total_amount, payment_method, payment_status, gateway_redirect_url, expires_at
		     FROM transactions WHERE booking_id = b.id ORDER BY id DESC LIMIT 1
		 ) t ON TRUE
		 WHERE b.user_id = $1
		 ORDER BY b.booking_date ASC, b.booking_time ASC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []map[string]any
	for rows.Next() {
		var id int64
		var ticketCode, status string
		var bookingDate, bookingTime string
		var chairNumber int
		var notes, serviceName string
		var description *string
		var durationMinutes int
		var price float64
		var totalAmount float64
		var barberName string
		var barberRating float64
		var paymentMethod, paymentStatus, gatewayRedirectURL *string
		var expiresAt *time.Time

		err := rows.Scan(&id, &ticketCode, &status, &bookingDate, &bookingTime, &chairNumber,
			&notes, &serviceName, &description, &durationMinutes, &price,
			&totalAmount, &barberName, &barberRating, &paymentMethod, &paymentStatus,
			&gatewayRedirectURL, &expiresAt)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, map[string]any{
			"id":                    id,
			"ticket_code":           ticketCode,
			"status":                status,
			"booking_date":          bookingDate,
			"booking_time":          bookingTime,
			"chair_number":          chairNumber,
			"notes":                 notes,
			"service_name":          serviceName,
			"service_description":   description,
			"duration_minutes":      durationMinutes,
			"price":                 price,
			"total_amount":          totalAmount,
			"barber_name":           barberName,
			"barber_rating":         barberRating,
			"payment_method":        paymentMethod,
			"payment_status":        paymentStatus,
			"payment_redirect_url":  gatewayRedirectURL,
			"expires_at":            expiresAt,
		})
	}
	return tickets, rows.Err()
}

func (r *Repository) GetUserProfile(ctx context.Context, userID int64) (map[string]any, error) {
	var id int64
	var name, phone, role string
	var email *string
	err := r.db.QueryRow(ctx,
		`SELECT id, name, phone, email, role FROM users WHERE id = $1`, userID,
	).Scan(&id, &name, &phone, &email, &role)
	return map[string]any{
		"id":    id,
		"name":  name,
		"phone": phone,
		"email": email,
		"role":  role,
	}, err
}
