CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('CLIENT', 'COURIER', 'ADMIN')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE clients (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    address VARCHAR(255)
);

CREATE TABLE couriers (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('AVAILABLE', 'BUSY', 'OFFLINE')),
    location GEOMETRY(Point, 4326),
    rating DECIMAL(3, 2) DEFAULT 0.0
);

CREATE INDEX couriers_location_idx ON couriers USING GIST (location);

CREATE TABLE admins (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    permissions JSONB
);

CREATE TABLE orders (
    id UUID PRIMARY KEY,
    client_id UUID NOT NULL REFERENCES clients(user_id) ON DELETE CASCADE,
    courier_id UUID REFERENCES couriers(user_id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('CREATED', 'ASSIGNED', 'IN_TRANSIT', 'DELIVERED', 'CANCELED')),
    delivery_address VARCHAR(255) NOT NULL,
    delivery_coords GEOMETRY(Point, 4326) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_status_logs (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX order_status_logs_order_id_idx ON order_status_logs (order_id);

CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX notifications_user_id_idx ON notifications (user_id);

CREATE TABLE ratings (
    id UUID PRIMARY KEY,
    courier_id UUID NOT NULL REFERENCES couriers(user_id) ON DELETE CASCADE,
    order_id UUID NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ratings_courier_id_idx ON ratings (courier_id);