#!/usr/bin/env python3
"""
UIT-Go Backend Load Test - Location Service
Tests concurrent location updates with authentication token
"""

import asyncio
import aiohttp
import time
import statistics
from typing import List, Dict, Any
import argparse
import random

# Test configuration
BASE_URL = "http://localhost:8080"
LOCATION_ENDPOINT = f"{BASE_URL}/location/"

# Authentication token for user ID 3 (jane.smith@example.com, role: driver)
AUTH_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJlbWFpbCI6ImphbmUuc21pdGhAZXhhbXBsZS5jb20iLCJyb2xlIjoiZHJpdmVyIiwiZXhwIjoxNzY1MjkxMzcwLCJuYmYiOjE3NjUyMDQ5NzAsImlhdCI6MTc2NTIwNDk3MH0.52X1eNmQEHugm3KouBZW6VWORLZXBV4YM2v_0XCVDtg"

# Saigon area coordinates for realistic testing
SAIGON_CENTER_LAT = 10.7769
SAIGON_CENTER_LON = 106.7009
COORDINATE_VARIANCE = 0.05  # ~5km radius


class LoadTestResult:
    def __init__(self):
        self.successful_requests = 0
        self.failed_requests = 0
        self.response_times: List[float] = []
        self.status_codes: Dict[int, int] = {}
        self.errors: Dict[str, int] = {}
        self.start_time = 0
        self.end_time = 0

    def add_success(self, response_time: float, status_code: int):
        self.successful_requests += 1
        self.response_times.append(response_time)
        self.status_codes[status_code] = self.status_codes.get(status_code, 0) + 1

    def add_failure(self, error_type: str, response_time: float = 0, status_code: int = 0):
        self.failed_requests += 1
        if response_time > 0:
            self.response_times.append(response_time)
        if status_code > 0:
            self.status_codes[status_code] = self.status_codes.get(status_code, 0) + 1
        self.errors[error_type] = self.errors.get(error_type, 0) + 1

    def calculate_stats(self) -> Dict[str, Any]:
        total_requests = self.successful_requests + self.failed_requests
        duration = self.end_time - self.start_time

        if not self.response_times:
            return {
                "total_requests": total_requests,
                "duration": duration,
                "throughput": 0,
                "success_rate": 0,
            }

        sorted_times = sorted(self.response_times)
        return {
            "total_requests": total_requests,
            "successful": self.successful_requests,
            "failed": self.failed_requests,
            "duration": duration,
            "throughput": total_requests / duration if duration > 0 else 0,
            "success_rate": (self.successful_requests / total_requests * 100) if total_requests > 0 else 0,
            "latency": {
                "min": min(self.response_times) * 1000,  # Convert to ms
                "max": max(self.response_times) * 1000,
                "mean": statistics.mean(self.response_times) * 1000,
                "median": statistics.median(self.response_times) * 1000,
                "p95": sorted_times[int(len(sorted_times) * 0.95)] * 1000 if len(sorted_times) > 0 else 0,
                "p99": sorted_times[int(len(sorted_times) * 0.99)] * 1000 if len(sorted_times) > 0 else 0,
            },
            "status_codes": self.status_codes,
            "errors": self.errors,
        }


def generate_random_location():
    """Generate random coordinates around Saigon center"""
    lat = SAIGON_CENTER_LAT + random.uniform(-COORDINATE_VARIANCE, COORDINATE_VARIANCE)
    lon = SAIGON_CENTER_LON + random.uniform(-COORDINATE_VARIANCE, COORDINATE_VARIANCE)
    return round(lat, 6), round(lon, 6)


async def update_location(session: aiohttp.ClientSession, result: LoadTestResult, request_delay: float = 0):
    """Send a single location update request"""
    if request_delay > 0:
        await asyncio.sleep(request_delay)

    latitude, longitude = generate_random_location()
    payload = {
        "latitude": latitude,
        "longitude": longitude,
    }

    headers = {
        "Authorization": f"Bearer {AUTH_TOKEN}",
        "Content-Type": "application/json",
    }

    start_time = time.time()

    try:
        async with session.post(
            LOCATION_ENDPOINT,
            json=payload,
            headers=headers,
            timeout=aiohttp.ClientTimeout(total=10),
        ) as response:
            response_time = time.time() - start_time
            status = response.status

            if status == 200 or status == 204:
                result.add_success(response_time, status)
            else:
                error_text = await response.text()
                result.add_failure(f"HTTP {status}", response_time, status)

    except asyncio.TimeoutError:
        response_time = time.time() - start_time
        result.add_failure("Timeout", response_time)
    except aiohttp.ClientError as e:
        response_time = time.time() - start_time
        result.add_failure(f"Client Error: {type(e).__name__}", response_time)
    except Exception as e:
        response_time = time.time() - start_time
        result.add_failure(f"Unexpected Error: {type(e).__name__}", response_time)


async def run_load_test(num_requests: int = 1000, rate_limit_safe: bool = False):
    """Run load test with specified number of concurrent requests"""
    result = LoadTestResult()

    # Calculate request delay for rate-limit-safe mode
    # API Gateway allows 1000 req/min = 16.67 req/s
    request_delay = 0.06 if rate_limit_safe else 0  # 60ms between requests = ~16 req/s

    connector = aiohttp.TCPConnector(limit=500, limit_per_host=500)
    timeout = aiohttp.ClientTimeout(total=30, connect=10)

    async with aiohttp.ClientSession(connector=connector, timeout=timeout) as session:
        print(f"\n🚀 Starting load test with {num_requests} requests...")
        print(f"Target: {LOCATION_ENDPOINT}")
        print(f"User: jane.smith@example.com (Driver, ID: 3)")
        print(f"Mode: {'Rate-limit safe (~16 req/s)' if rate_limit_safe else 'Full concurrent (may hit rate limits)'}\n")

        result.start_time = time.time()

        # Create all tasks
        tasks = [
            update_location(session, result, request_delay * i if rate_limit_safe else 0)
            for i in range(num_requests)
        ]

        # Execute all requests concurrently
        await asyncio.gather(*tasks)

        result.end_time = time.time()

    return result


def print_results(result: LoadTestResult):
    """Pretty print load test results"""
    stats = result.calculate_stats()

    print("\n" + "=" * 70)
    print("📊 LOAD TEST RESULTS - LOCATION UPDATE SERVICE")
    print("=" * 70)

    print(f"\n⏱️  Total Execution Time: {stats['duration']:.2f} seconds")
    print(f"📈 Throughput: {stats['throughput']:.2f} requests/second")

    print(f"\n✅ Successful Requests: {stats['successful']} ({stats['success_rate']:.1f}%)")
    print(f"❌ Failed Requests: {stats['failed']} ({100 - stats['success_rate']:.1f}%)")

    if stats['latency']['min'] > 0:
        print("\n📉 Latency Statistics (milliseconds):")
        print(f"   - Min:     {stats['latency']['min']:.2f} ms")
        print(f"   - Max:     {stats['latency']['max']:.2f} ms")
        print(f"   - Mean:    {stats['latency']['mean']:.2f} ms")
        print(f"   - Median:  {stats['latency']['median']:.2f} ms")
        print(f"   - P95:     {stats['latency']['p95']:.2f} ms")
        print(f"   - P99:     {stats['latency']['p99']:.2f} ms")

    if stats['status_codes']:
        print("\n📋 Status Code Distribution:")
        for code, count in sorted(stats['status_codes'].items()):
            percentage = (count / stats['total_requests'] * 100)
            print(f"   - {code}: {count} requests ({percentage:.1f}%)")

    if stats['errors']:
        print("\n⚠️  Error Types:")
        for error, count in sorted(stats['errors'].items(), key=lambda x: x[1], reverse=True):
            percentage = (count / stats['total_requests'] * 100)
            print(f"   - {error}: {count} ({percentage:.1f}%)")

    print("\n" + "=" * 70)

    # Performance assessment
    success_rate = stats['success_rate']
    p95_latency = stats['latency']['p95']

    print("\n🎯 Performance Assessment:")
    if success_rate >= 99.9 and p95_latency < 100:
        print("   ✅ EXCELLENT - Meets SLO targets")
    elif success_rate >= 99 and p95_latency < 200:
        print("   ✅ GOOD - Acceptable performance")
    elif success_rate >= 95 and p95_latency < 500:
        print("   ⚠️  FAIR - Performance degradation detected")
    else:
        print("   ❌ POOR - Performance issues detected")

    print("\n   SLO Targets:")
    print(f"   - Success Rate: {success_rate:.2f}% (Target: >99.9%)")
    print(f"   - P95 Latency:  {p95_latency:.2f}ms (Target: <100ms)")

    print("\n" + "=" * 70)

    print("\n💡 Tip: View real-time metrics in Grafana: http://localhost:3000")
    print("💡 Tip: Check Prometheus alerts: http://localhost:9090/alerts")
    print("💡 Tip: View traces in Jaeger: http://localhost:16686")


def main():
    parser = argparse.ArgumentParser(description="Load test for Location Update Service")
    parser.add_argument(
        "--requests",
        type=int,
        default=1000,
        help="Number of concurrent requests to send (default: 1000)",
    )
    parser.add_argument(
        "--safe",
        action="store_true",
        help="Use rate-limit-safe mode (~16 req/s to avoid HTTP 429)",
    )

    args = parser.parse_args()

    print("╔════════════════════════════════════════════════════════════════╗")
    print("║    UIT-Go Backend Load Test - Location Update Service         ║")
    print("╚════════════════════════════════════════════════════════════════╝")

    # Run the load test
    result = asyncio.run(run_load_test(args.requests, args.safe))

    # Print results
    print_results(result)


if __name__ == "__main__":
    main()
