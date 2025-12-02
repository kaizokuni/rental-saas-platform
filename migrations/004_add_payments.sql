CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL, -- Foreign key to bookings table (when created)
    stripe_intent_id TEXT NOT NULL,
    amount_cents INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending_auth', 'authorized', 'captured', 'voided', 'failed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
