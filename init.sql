-- 1. Таблица users
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) CHECK (char_length(password_hash) > 0),
    auth_provider VARCHAR(20) CHECK (auth_provider IN ('local', 'google', 'vk')) DEFAULT 'local',
    external_provider_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_externalProviderId ON users(external_provider_id);

-- 2. Таблица places
CREATE TABLE places (
    place_id SERIAL PRIMARY KEY,
    place_name VARCHAR(255) NOT NULL,
    place_type VARCHAR(50) NOT NULL,
    place_description TEXT,
    place_location VARCHAR(255) NOT NULL,
    place_occupancy INT DEFAULT 0 CHECK (place_occupancy >= 0),
    place_capacity INT NOT NULL CHECK (place_capacity > 0),
    rating FLOAT DEFAULT 0.0 CHECK (rating BETWEEN 0 AND 5),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_places_type ON places(place_type);
CREATE INDEX idx_places_rating ON places(rating);
CREATE INDEX idx_places_location ON places(place_location);

-- 3. Таблица bookings
CREATE TABLE bookings (
    booking_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    place_id INT NOT NULL REFERENCES places(place_id) ON DELETE CASCADE,
    booking_status VARCHAR(50) CHECK (booking_status IN ('created','reserved','paid','cancelled','completed','no_show')) DEFAULT 'created',
    payment_state VARCHAR(50) CHECK (payment_state IN ('pending','paid','failed','refunded')) DEFAULT 'pending',
    booking_start_time TIMESTAMP NOT NULL,
    booking_end_time TIMESTAMP NOT NULL CHECK (booking_end_time > booking_start_time),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,
    check_in_time TIMESTAMP,
    check_out_time TIMESTAMP
);
CREATE INDEX idx_bookings_user ON bookings(user_id);
CREATE INDEX idx_bookings_place ON bookings(place_id);
CREATE INDEX idx_bookings_status ON bookings(booking_status);
CREATE INDEX idx_bookings_user_status ON bookings(user_id, booking_status);
CREATE INDEX idx_bookings_place_time ON bookings(place_id, booking_start_time, booking_end_time);

-- 4. Таблица payments
CREATE TABLE payments (
    payment_id SERIAL PRIMARY KEY,
    booking_id INT NOT NULL REFERENCES bookings(booking_id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    payment_status VARCHAR(50) CHECK (payment_status IN ('pending','completed','failed')) DEFAULT 'pending',
    payment_method VARCHAR(50) CHECK (payment_method IN ('card','sbp','on_site')),
    payment_provider VARCHAR(50) CHECK (payment_provider IN ('stripe','bank','sbp')),
    provider_transaction_id VARCHAR(255),
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) DEFAULT 'RUB',
    payment_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_payments_booking ON payments(booking_id);
CREATE INDEX idx_payments_user ON payments(user_id);
CREATE INDEX idx_payments_providerTransactionId ON payments(provider_transaction_id);

-- 5. Таблица notifications
CREATE TABLE notifications (
    notification_id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    booking_id INT REFERENCES bookings(booking_id) ON DELETE SET NULL,
    type VARCHAR(50) CHECK (type IN ('bookingConfirmed','reminder','paymentFailed')),
    channel VARCHAR(20) CHECK (channel IN ('push','email','sms')),
    status VARCHAR(20) CHECK (status IN ('pending','sent','failed')) DEFAULT 'pending',
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP
);
CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_booking ON notifications(booking_id);
CREATE INDEX idx_notifications_status ON notifications(status);

-- 6. Таблица reviews
CREATE TABLE reviews (
    review_id SERIAL PRIMARY KEY,
    place_id INT NOT NULL REFERENCES places(place_id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    rating INT CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    moderation_status VARCHAR(20) CHECK (moderation_status IN ('pending','approved','rejected')) DEFAULT 'pending'
);
CREATE INDEX idx_reviews_place ON reviews(place_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);

-- 7. Таблица favorites
CREATE TABLE favorites (
    user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    place_id INT REFERENCES places(place_id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, place_id)
);
CREATE INDEX idx_favorites_user ON favorites(user_id);
CREATE INDEX idx_favorites_place ON favorites(place_id);

-- 8. Таблица analytics_events
CREATE TABLE analytics_events (
    event_id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) CHECK (event_type IN ('search','bookingCreated','paymentCompleted')),
    user_id INT,
    place_id INT,
    booking_id INT,
    device_type VARCHAR(50),
    metadata JSONB,
    event_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_analytics_eventType ON analytics_events(event_type);
CREATE INDEX idx_analytics_user ON analytics_events(user_id);
CREATE INDEX idx_analytics_booking ON analytics_events(booking_id);
CREATE INDEX idx_analytics_time ON analytics_events(event_timestamp);

-- 9. Таблица availability_rules
CREATE TABLE availability_rules (
    rule_id SERIAL PRIMARY KEY,
    place_id INT NOT NULL REFERENCES places(place_id) ON DELETE CASCADE,
    day_of_week INT CHECK (day_of_week BETWEEN 1 AND 7),
    open_time TIME NOT NULL,
    close_time TIME NOT NULL CHECK (open_time < close_time),
    is_closed BOOLEAN DEFAULT FALSE
);
CREATE INDEX idx_availability_place ON availability_rules(place_id);

-- 10. Таблица refunds
CREATE TABLE refunds (
    refund_id SERIAL PRIMARY KEY,
    payment_id INT NOT NULL REFERENCES payments(payment_id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    canceled_at TIMESTAMP,
    status VARCHAR(50) CHECK (status IN ('pending','completed','failed')) DEFAULT 'pending',
    reason TEXT
);
CREATE INDEX idx_refunds_payment ON refunds(payment_id);
CREATE INDEX idx_refunds_status ON refunds(status);

-- 11. Таблица audit_logs
CREATE TABLE audit_logs (
    audit_id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) CHECK (event_type IN ('account_deleted','password_changed','email_changed','refund_requested','refund_completed','admin_login','failed_login')),
    actor_user_id INT REFERENCES users(user_id) ON DELETE SET NULL,
    target_user_id INT REFERENCES users(user_id) ON DELETE SET NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_audit_logs_eventType ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_actorUserId ON audit_logs(actor_user_id);
CREATE INDEX idx_audit_logs_targetUserId ON audit_logs(target_user_id);
CREATE INDEX idx_audit_logs_createdAt ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_actor_time ON audit_logs(actor_user_id, created_at);

-- НАПОЛНЕНИЕ БАЗОВЫМИ ДАННЫМИ (Seeding)
INSERT INTO users (first_name, last_name, email, password_hash) VALUES 
('Ivan', 'Ivanov', 'user@gmail.com', '$2y$10$randomhash123'),
('Alex', 'Petrov', 'alex@mail.com', '$2y$10$randomhash456');

INSERT INTO places (place_name, place_type, place_location, place_capacity, rating) VALUES 
('WorkPlace1', 'workTable', 'Russia, Moscow, Kiyevskaya St., 14', 1, 4.3),
('WorkPlace2', 'workTable', 'Russia, Moscow, Lenina St., 13', 1, 4.6);
