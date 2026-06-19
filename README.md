# event-driven-backtesting-engine
ระบบ Event-Driven Backtesting สำหรับตลาดคริปโตที่ออกแบบให้ใกล้เคียงการเทรดจริง ลด Look-Ahead Bias และรองรับ Monte Carlo Simulation

## Tech Stack

- Go (Core Engine)
- Python (Data Pipeline)
- PostgreSQL 15
- TimescaleDB

## Database

Required:

- PostgreSQL 15
- TimescaleDB

Run:

1. CREATE EXTENSION IF NOT EXISTS timescaledb;
2. database/schema.sql
3. -- Convert candles table to hypertable
    SELECT create_hypertable(
        'candles',
        'timestamp',
        if_not_exists => TRUE
    );