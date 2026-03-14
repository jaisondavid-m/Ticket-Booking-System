CREATE TABLE IF NOT EXISTS bookings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_bookings_user_id (user_id)
);

CREATE TABLE IF NOT EXISTS tickets (
    id BIGINT UNSIGNED NOT NULL,
    available INT NOT NULL,
    PRIMARY KEY (id)
);

INSERT INTO tickets (id, available)
VALUES (1, 100)
ON DUPLICATE KEY UPDATE available = VALUES(available);
