CREATE TABLE IF NOT EXISTS domain_trackings (
    id SERIAL PRIMARY KEY, 
    user_id UUID,
    domain_name TEXT NOT NULL,
    expires TIMESTAMP NOT NULL,
    issuer TEXT DEFAULT 'n/a',
    status TEXT NOT NULL,
    error TEXT,
    signature_algo TEXT,
    public_key_algo TEXT,
    dns_names TEXT,
    signature_fingerprint TEXT,
    FOREIGN KEY (user_id) REFERENCES auth.users (id)
);