-- Enum
CREATE TYPE trip_status AS ENUM (
  'REQUESTED',
  'ACCEPTED',
  'STARTED',
  'COMPLETED',
  'CANCELLED'
);

-- Create table
CREATE TABLE trips (
  id SERIAL PRIMARY KEY,
  passenger_id INT NOT NULL,
  driver_id INT,
  origin_lat DOUBLE PRECISION NOT NULL,
  origin_lng DOUBLE PRECISION NOT NULL,
  dest_lat DOUBLE PRECISION NOT NULL,
  dest_lng DOUBLE PRECISION NOT NULL,
  status trip_status NOT NULL DEFAULT 'REQUESTED',
  distance DOUBLE PRECISION NOT NULL,
  fare DOUBLE PRECISION NOT NULL,
  payment_method VARCHAR(50) NOT NULL,
  rating INT,
  review TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  started_at TIMESTAMP NULL,
  completed_at TIMESTAMP NULL,
  cancelled_at TIMESTAMP NULL,
  cancel_by_user_id INT
);

-- Indexes 
CREATE INDEX idx_trips_passenger_id ON trips (passenger_id);
CREATE INDEX idx_trips_driver_id ON trips (driver_id);
CREATE INDEX idx_trips_status ON trips (status);