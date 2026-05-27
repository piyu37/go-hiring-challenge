ALTER TABLE products ALTER COLUMN category_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_code ON products(code);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
