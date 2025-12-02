-- Set search path to public (admin tenant)
SET search_path TO public;

-- Create Dummy Customer
INSERT INTO customers (id, first_name, last_name, email, phone, driver_license_id)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'John',
    'Doe',
    'john.doe@example.com',
    '+15550102030',
    'DL12345678'
) ON CONFLICT (email) DO NOTHING;

-- Create Backdated Booking (Pending)
INSERT INTO bookings (
    id,
    car_id,
    customer_id,
    start_time,
    end_time,
    status,
    total_amount_cents,
    deposit_amount_cents
) VALUES (
    'b1ffcd88-8d1c-5ff9-cc7e-7cc0ce491b22',
    '5e5b8875-50c5-4a48-a178-f057f7f1d155', -- Car ID from previous step
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- Customer ID
    NOW() - INTERVAL '3 days',
    NOW(),
    'pending',
    15000, -- $150.00
    50000  -- $500.00 Deposit
);
