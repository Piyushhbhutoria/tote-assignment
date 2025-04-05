-- Plugins table to store plugin configurations
CREATE TABLE IF NOT EXISTS plugins (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT false,
    config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Employees table for tracking employee sessions
CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    employee_id VARCHAR(100) NOT NULL UNIQUE,
    current_terminal_id VARCHAR(100),
    last_login TIMESTAMP WITH TIME ZONE,
    last_logout TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Employee sessions for time tracking
CREATE TABLE IF NOT EXISTS employee_sessions (
    id SERIAL PRIMARY KEY,
    employee_id VARCHAR(100) NOT NULL,
    terminal_id VARCHAR(100) NOT NULL,
    login_time TIMESTAMP WITH TIME ZONE NOT NULL,
    logout_time TIMESTAMP WITH TIME ZONE,
    duration_minutes INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Customers table
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    customer_id VARCHAR(100) NOT NULL UNIQUE,
    data JSONB,
    last_seen TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Items table for product information
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    item_id VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    requires_age_verification BOOLEAN DEFAULT false,
    minimum_age INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Baskets table for tracking shopping sessions
CREATE TABLE IF NOT EXISTS baskets (
    id SERIAL PRIMARY KEY,
    basket_id VARCHAR(100) NOT NULL UNIQUE,
    terminal_id VARCHAR(100) NOT NULL,
    employee_id VARCHAR(100) NOT NULL,
    customer_id VARCHAR(100),
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Basket items for tracking items in baskets
CREATE TABLE IF NOT EXISTS basket_items (
    id SERIAL PRIMARY KEY,
    basket_id VARCHAR(100) NOT NULL,
    item_id VARCHAR(100) NOT NULL,
    quantity INTEGER NOT NULL,
    price_at_time DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Fraud alerts for tracking suspicious activities
CREATE TABLE IF NOT EXISTS fraud_alerts (
    id SERIAL PRIMARY KEY,
    basket_id VARCHAR(100) NOT NULL,
    alert_type VARCHAR(100) NOT NULL,
    description TEXT,
    severity VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Item recommendations for the purchase recommender
CREATE TABLE IF NOT EXISTS item_recommendations (
    id SERIAL PRIMARY KEY,
    source_item_id VARCHAR(100) NOT NULL,
    recommended_item_id VARCHAR(100) NOT NULL,
    confidence_score DECIMAL(5,4) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_item_id, recommended_item_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_employee_sessions_employee_id ON employee_sessions(employee_id);
CREATE INDEX IF NOT EXISTS idx_basket_items_basket_id ON basket_items(basket_id);
CREATE INDEX IF NOT EXISTS idx_fraud_alerts_basket_id ON fraud_alerts(basket_id);
CREATE INDEX IF NOT EXISTS idx_item_recommendations_source_item ON item_recommendations(source_item_id); 
