-- =====================================================
-- Backtest Runs
-- =====================================================

CREATE TABLE backtest_runs (
    run_id BIGSERIAL PRIMARY KEY,

    strategy_name VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    timeframe VARCHAR(20) NOT NULL,

    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,

    total_return DOUBLE PRECISION,
    win_rate DOUBLE PRECISION,
    sharpe_ratio DOUBLE PRECISION,
    max_drawdown DOUBLE PRECISION
);

-- =====================================================
-- Portfolio
-- =====================================================

CREATE TABLE portfolio (
    portfolio_id BIGSERIAL PRIMARY KEY,

    run_id BIGINT NOT NULL,

    cash DOUBLE PRECISION NOT NULL,
    equity DOUBLE PRECISION NOT NULL,

    unrealized_pnl DOUBLE PRECISION,
    max_drawdown DOUBLE PRECISION,

    CONSTRAINT fk_portfolio_run
        FOREIGN KEY (run_id)
        REFERENCES backtest_runs(run_id)
);

-- =====================================================
-- Positions
-- =====================================================

CREATE TABLE positions (
    position_id BIGSERIAL PRIMARY KEY,

    portfolio_id BIGINT NOT NULL,

    symbol VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL,

    quantity DOUBLE PRECISION NOT NULL,

    entry_price DOUBLE PRECISION NOT NULL,
    current_price DOUBLE PRECISION NOT NULL,
    current_value DOUBLE PRECISION NOT NULL,

    unrealized_pnl DOUBLE PRECISION,
    realized_pnl DOUBLE PRECISION,

    CONSTRAINT fk_position_portfolio
        FOREIGN KEY (portfolio_id)
        REFERENCES portfolio(portfolio_id)
);

-- =====================================================
-- Orders
-- =====================================================

CREATE TABLE orders (
    order_id BIGSERIAL PRIMARY KEY,

    run_id BIGINT NOT NULL,

    symbol VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL,

    order_type VARCHAR(50) NOT NULL,

    quantity DOUBLE PRECISION NOT NULL,
    price DOUBLE PRECISION NOT NULL,

    status VARCHAR(50) NOT NULL,

    created_time TIMESTAMPTZ NOT NULL,
    filled_time TIMESTAMPTZ,

    CONSTRAINT fk_order_run
        FOREIGN KEY (run_id)
        REFERENCES backtest_runs(run_id)
);

-- =====================================================
-- Trades
-- =====================================================

CREATE TABLE trades (
    trade_id BIGSERIAL PRIMARY KEY,

    run_id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,

    symbol VARCHAR(50) NOT NULL,
    side VARCHAR(10) NOT NULL,

    executed_time TIMESTAMPTZ NOT NULL,

    executed_price DOUBLE PRECISION NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,

    CONSTRAINT fk_trade_run
        FOREIGN KEY (run_id)
        REFERENCES backtest_runs(run_id),

    CONSTRAINT fk_trade_order
        FOREIGN KEY (order_id)
        REFERENCES orders(order_id)
);

-- =====================================================
-- Candles
-- =====================================================

CREATE TABLE candles (
    timestamp TIMESTAMPTZ NOT NULL,

    symbol VARCHAR(50) NOT NULL,
    timeframe VARCHAR(20) NOT NULL,

    open DOUBLE PRECISION NOT NULL,
    high DOUBLE PRECISION NOT NULL,
    low DOUBLE PRECISION NOT NULL,
    close DOUBLE PRECISION NOT NULL,

    volume DOUBLE PRECISION NOT NULL,

    PRIMARY KEY (
        timestamp,
        symbol,
        timeframe
    )
);