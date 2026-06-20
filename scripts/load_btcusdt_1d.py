import argparse
import json
import sys
import time
from datetime import datetime, timedelta, timezone
from decimal import Decimal, InvalidOperation
from pathlib import Path
from urllib.parse import urlencode
from urllib.request import urlopen

try:
    import psycopg
except ImportError:
    psycopg = None


SYMBOL = "BTCUSDT"
INTERVAL = "1d"
TIMEFRAME = "1d"
BINANCE_KLINES_URL = "https://api.binance.com/api/v3/klines"
START_DATE = datetime(2017, 8, 17, tzinfo=timezone.utc)
RAW_JSON_PATH = Path("data/raw/binance/BTCUSDT/1d/klines.json")
KLINE_FIELD_COUNT = 12
REQUEST_LIMIT = 1000


def parse_args():
    parser = argparse.ArgumentParser(
        description="Fetch Binance BTCUSDT 1d candles and import them into TimescaleDB."
    )
    parser.add_argument("--db-host", default="localhost")
    parser.add_argument("--db-port", default=5432, type=int)
    parser.add_argument("--db-name", required=True)
    parser.add_argument("--db-user", required=True)
    parser.add_argument("--db-password", required=True)
    return parser.parse_args()


def to_milliseconds(value):
    return int(value.timestamp() * 1000)


def from_milliseconds(value):
    return datetime.fromtimestamp(value / 1000, tz=timezone.utc)


def last_closed_daily_open(now=None):
    now = now or datetime.now(timezone.utc)
    today_open = datetime(now.year, now.month, now.day, tzinfo=timezone.utc)
    return today_open - timedelta(days=1)


def fetch_klines():
    rows = []
    start_ms = to_milliseconds(START_DATE)
    end_ms = to_milliseconds(last_closed_daily_open())

    while start_ms <= end_ms:
        params = {
            "symbol": SYMBOL,
            "interval": INTERVAL,
            "startTime": start_ms,
            "endTime": end_ms,
            "limit": REQUEST_LIMIT,
        }
        url = f"{BINANCE_KLINES_URL}?{urlencode(params)}"

        with urlopen(url, timeout=30) as response:
            batch = json.loads(response.read().decode("utf-8"))

        if not batch:
            break

        rows.extend(batch)
        next_open_time = int(batch[-1][0]) + 1

        if next_open_time <= start_ms:
            raise ValueError("Binance pagination did not advance")

        start_ms = next_open_time
        time.sleep(0.2)

    return rows


def save_raw_json(rows):
    RAW_JSON_PATH.parent.mkdir(parents=True, exist_ok=True)
    with RAW_JSON_PATH.open("w", encoding="utf-8") as file:
        json.dump(rows, file, indent=2)


def parse_decimal(value, field_name):
    try:
        return Decimal(value)
    except (InvalidOperation, TypeError):
        raise ValueError(f"invalid decimal value for {field_name}: {value!r}") from None


def normalize_and_validate(rows):
    if not rows:
        raise ValueError("no candles fetched")

    normalized = []
    previous_timestamp = None
    seen_timestamps = set()

    for index, row in enumerate(rows):
        if not isinstance(row, list) or len(row) != KLINE_FIELD_COUNT:
            raise ValueError(f"row {index} does not have {KLINE_FIELD_COUNT} kline fields")

        open_time = int(row[0])
        timestamp = from_milliseconds(open_time)

        if timestamp.hour != 0 or timestamp.minute != 0 or timestamp.second != 0:
            raise ValueError(f"row {index} timestamp is not a UTC daily open: {timestamp}")

        if previous_timestamp is not None and timestamp <= previous_timestamp:
            raise ValueError(f"row {index} timestamp is not strictly increasing: {timestamp}")

        if timestamp in seen_timestamps:
            raise ValueError(f"duplicate timestamp found: {timestamp}")

        open_price = parse_decimal(row[1], "open")
        high_price = parse_decimal(row[2], "high")
        low_price = parse_decimal(row[3], "low")
        close_price = parse_decimal(row[4], "close")
        volume = parse_decimal(row[5], "volume")

        if open_price <= 0 or high_price <= 0 or low_price <= 0 or close_price <= 0:
            raise ValueError(f"row {index} contains non-positive OHLC values")

        if volume < 0:
            raise ValueError(f"row {index} contains negative volume")

        if high_price < max(open_price, close_price, low_price):
            raise ValueError(f"row {index} has invalid high value")

        if low_price > min(open_price, close_price, high_price):
            raise ValueError(f"row {index} has invalid low value")

        normalized.append(
            (
                timestamp,
                SYMBOL,
                TIMEFRAME,
                open_price,
                high_price,
                low_price,
                close_price,
                volume,
            )
        )

        seen_timestamps.add(timestamp)
        previous_timestamp = timestamp

    return normalized


def import_candles(args, candles):
    if psycopg is None:
        raise RuntimeError(
            "missing dependency: install psycopg before importing into TimescaleDB "
            "(for example: pip install psycopg[binary])"
        )

    insert_sql = """
        INSERT INTO candles (
            timestamp, symbol, timeframe, open, high, low, close, volume
        )
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (timestamp, symbol, timeframe)
        DO UPDATE SET
            open = EXCLUDED.open,
            high = EXCLUDED.high,
            low = EXCLUDED.low,
            close = EXCLUDED.close,
            volume = EXCLUDED.volume;
    """

    with psycopg.connect(
        host=args.db_host,
        port=args.db_port,
        dbname=args.db_name,
        user=args.db_user,
        password=args.db_password,
    ) as connection:
        with connection.cursor() as cursor:
            cursor.executemany(insert_sql, candles)
        connection.commit()


def main():
    args = parse_args()

    rows = fetch_klines()
    save_raw_json(rows)
    candles = normalize_and_validate(rows)
    import_candles(args, candles)

    print("BTCUSDT 1d load complete")
    print(f"raw rows fetched: {len(rows)}")
    print(f"first timestamp: {candles[0][0].isoformat()}")
    print(f"last timestamp: {candles[-1][0].isoformat()}")
    print(f"rows imported: {len(candles)}")
    print(f"raw json path: {RAW_JSON_PATH}")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"error: {error}", file=sys.stderr)
        sys.exit(1)
