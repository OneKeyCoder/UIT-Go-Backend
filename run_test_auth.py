#!/usr/bin/env python3
"""
Load Test Script for UIT-Go Backend Authentication Service
Sends 1000 concurrent requests to test performance and observability
"""

import asyncio
import aiohttp
import time
from typing import List, Dict
import statistics


async def send_auth_request(session: aiohttp.ClientSession, request_id: int) -> Dict:
    """
    Send authentication request to the API Gateway
    
    Args:
        session: aiohttp ClientSession for connection pooling
        request_id: Unique identifier for this request
    
    Returns:
        Dict with request result (success, latency, status_code)
    """
    url = "http://localhost:8080/grpc/auth"
    payload = {
        "email": "jane.smith@example.com",
        "password": "password123"
    }
    headers = {
        "Content-Type": "application/json"
    }
    
    start_time = time.time()
    
    try:
        async with session.post(url, json=payload, headers=headers, timeout=aiohttp.ClientTimeout(total=30)) as response:
            elapsed_time = (time.time() - start_time) * 100  # Convert to milliseconds
            status_code = response.status
            
            # Try to read response body
            try:
                response_body = await response.text()
            except:
                response_body = ""
            
            return {
                "request_id": request_id,
                "success": 200 <= status_code < 300,
                "status_code": status_code,
                "latency_ms": elapsed_time,
                "error": None if 200 <= status_code < 300 else f"HTTP {status_code}"
            }
    
    except asyncio.TimeoutError:
        elapsed_time = (time.time() - start_time) * 1000
        return {
            "request_id": request_id,
            "success": False,
            "status_code": 0,
            "latency_ms": elapsed_time,
            "error": "Timeout"
        }
    
    except Exception as e:
        elapsed_time = (time.time() - start_time) * 1000
        return {
            "request_id": request_id,
            "success": False,
            "status_code": 0,
            "latency_ms": elapsed_time,
            "error": str(e)
        }


async def run_load_test(num_requests: int = 1000, rate_limit_safe: bool = False) -> List[Dict]:
    """
    Run load test with concurrent requests
    
    Args:
        num_requests: Number of concurrent requests to send
        rate_limit_safe: If True, spread requests over time to avoid rate limiting
    
    Returns:
        List of results from all requests
    """
    print(f"🚀 Starting load test with {num_requests} requests...")
    print(f"Target: http://localhost:8080/grpc/auth")
    print(f"Payload: jane.smith@example.com / password123")
    if rate_limit_safe:
        print(f"Mode: Rate-limit safe (spread over ~60 seconds)")
    else:
        print(f"Mode: Full concurrent (may hit rate limits)")
    print()
    
    # Create connection pool with appropriate limits
    connector = aiohttp.TCPConnector(
        limit=500,  # Maximum number of connections
        limit_per_host=500,  # Maximum connections per host
        ttl_dns_cache=300  # DNS cache TTL
    )
    
    async with aiohttp.ClientSession(connector=connector) as session:
        start_time = time.time()
        
        if rate_limit_safe:
            # Spread requests over time to respect rate limiting (10 req/min)
            # Send in batches to avoid overwhelming the server
            batch_size = 50
            results = []
            for i in range(0, num_requests, batch_size):
                batch = [send_auth_request(session, j) for j in range(i, min(i + batch_size, num_requests))]
                batch_results = await asyncio.gather(*batch)
                results.extend(batch_results)
                # Small delay between batches (0.6s allows 10 req/min)
                if i + batch_size < num_requests:
                    await asyncio.sleep(0.3)
        else:
            # Execute all requests concurrently (may hit rate limits)
            tasks = [send_auth_request(session, i) for i in range(num_requests)]
            results = await asyncio.gather(*tasks)
        
        total_time = time.time() - start_time
    
    # Calculate statistics
    print_results(results, total_time)
    
    return results


def print_results(results: List[Dict], total_time: float):
    """
    Print load test results with statistics
    
    Args:
        results: List of request results
        total_time: Total execution time in seconds
    """
    successful_requests = [r for r in results if r["success"]]
    failed_requests = [r for r in results if not r["success"]]
    
    # Extract latencies for successful requests
    latencies = [r["latency_ms"] for r in successful_requests]
    
    print("\n" + "="*70)
    print("📊 LOAD TEST RESULTS")
    print("="*70)
    
    print(f"\n⏱️  Total Execution Time: {total_time:.2f} seconds")
    print(f"📈 Throughput: {len(results) / total_time:.2f} requests/second")
    
    print(f"\n✅ Successful Requests: {len(successful_requests)} ({len(successful_requests)/len(results)*100:.1f}%)")
    print(f"❌ Failed Requests: {len(failed_requests)} ({len(failed_requests)/len(results)*100:.1f}%)")
    
    if latencies:
        print(f"\n📉 Latency Statistics (milliseconds):")
        print(f"   - Min:     {min(latencies):.2f} ms")
        print(f"   - Max:     {max(latencies):.2f} ms")
        print(f"   - Mean:    {statistics.mean(latencies):.2f} ms")
        print(f"   - Median:  {statistics.median(latencies):.2f} ms")
        print(f"   - P95:     {statistics.quantiles(latencies, n=20)[18]:.2f} ms")
        print(f"   - P99:     {statistics.quantiles(latencies, n=100)[98]:.2f} ms")
    
    # Status code distribution
    status_codes = {}
    for result in results:
        code = result["status_code"]
        status_codes[code] = status_codes.get(code, 0) + 1
    
    print(f"\n📋 Status Code Distribution:")
    for code in sorted(status_codes.keys()):
        count = status_codes[code]
        percentage = (count / len(results)) * 100
        print(f"   - {code}: {count} requests ({percentage:.1f}%)")
    
    # Error distribution
    if failed_requests:
        print(f"\n⚠️  Error Types:")
        error_types = {}
        for result in failed_requests:
            error = result["error"] or "Unknown"
            error_types[error] = error_types.get(error, 0) + 1
        
        for error, count in sorted(error_types.items(), key=lambda x: x[1], reverse=True):
            percentage = (count / len(failed_requests)) * 100
            print(f"   - {error}: {count} ({percentage:.1f}%)")
    
    print("\n" + "="*70)
    
    # Performance assessment
    if latencies:
        p95_latency = statistics.quantiles(latencies, n=20)[18]
        success_rate = len(successful_requests) / len(results)
        
        print("\n🎯 Performance Assessment:")
        if success_rate >= 0.999 and p95_latency < 100:
            print("   ✅ EXCELLENT - Meeting SLO targets!")
        elif success_rate >= 0.99 and p95_latency < 200:
            print("   ✅ GOOD - Performance acceptable")
        elif success_rate >= 0.95 and p95_latency < 500:
            print("   ⚠️  FAIR - Performance degraded")
        else:
            print("   ❌ POOR - Performance issues detected")
        
        print(f"\n   SLO Targets:")
        print(f"   - Success Rate: {success_rate*100:.2f}% (Target: >99.9%)")
        print(f"   - P95 Latency:  {p95_latency:.2f}ms (Target: <100ms)")
    
    print("\n" + "="*70)


async def main():
    """
    Main entry point
    """
    print("╔════════════════════════════════════════════════════════════════╗")
    print("║        UIT-Go Backend Load Test - Authentication Service       ║")
    print("╚════════════════════════════════════════════════════════════════╝\n")
    
    import sys
    
    # Check for rate-limit-safe flag
    rate_limit_safe = "--safe" in sys.argv or "-s" in sys.argv
    
    # Run load test with 1000 requests
    await run_load_test(num_requests=250, rate_limit_safe=rate_limit_safe)
    
    print("\n💡 Tip: View real-time metrics in Grafana: http://localhost:3000")
    print("💡 Tip: Check Prometheus alerts: http://localhost:9090/alerts")
    print("💡 Tip: View traces in Jaeger: http://localhost:16686\n")


if __name__ == "__main__":
    # Run the async main function
    asyncio.run(main())
