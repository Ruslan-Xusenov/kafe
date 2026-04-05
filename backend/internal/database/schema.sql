-- Create User Role Enum
CREATE TYPE user_role AS ENUM ('customer', 'cook', 'courier', 'admin');

-- Create Order Status Enum
CREATE TYPE order_status AS ENUM ('new', 'preparing', 'ready', 'on_way', 'delivered', 'cancelled');

-- Users Table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'customer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Categories Table
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    image_url TEXT,
    is_user_controlled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Products Table
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(12, 2) NOT NULL,
    image_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    unit VARCHAR(20) DEFAULT 'dona',
    min_quantity DECIMAL(12, 3) DEFAULT 1.0,
    quantity_step DECIMAL(12, 3) DEFAULT 1.0,
    has_mandatory_container BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Orders Table
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    total_price DECIMAL(12, 2) NOT NULL,
    status order_status DEFAULT 'new',
    address TEXT NOT NULL,
    phone VARCHAR(20) NOT NULL, -- Customer contact for this order
    lat DECIMAL(9, 6),
    lng DECIMAL(9, 6),
    courier_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    cook_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Order Items Table
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    quantity DECIMAL(12, 3) NOT NULL DEFAULT 1.000,
    price DECIMAL(12, 2) NOT NULL, -- Individual product price at times of purchase
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Settings Table
CREATE TABLE settings (
    key VARCHAR(50) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Initial Settings
INSERT INTO settings (key, value) VALUES ('container_price', '1000') ON CONFLICT (key) DO NOTHING;
INSERT INTO settings (key, value) VALUES ('container_product_id', '7') ON CONFLICT (key) DO NOTHING;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
