"""Performance benchmarks for MT5 gRPC service"""
import time
import statistics
import sys
import os
from typing import List, Dict, Any

# Add src to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../src"))

from operation_queue import OperationQueue
from db_schema import DatabaseSchema


class BenchmarkRunner:
    """Run performance benchmarks"""

    def __init__(self, db_path: str = "benchmark.db"):
        self.db_path = db_path
        self.queue = OperationQueue(db_path)

    def setup(self):
        """Setup benchmark database"""
        db = DatabaseSchema(self.db_path)
        db.connect()
        db.initialize()
        db.close()

    def teardown(self):
        """Cleanup benchmark database"""
        import os

        if os.path.exists(self.db_path):
            os.remove(self.db_path)

    def benchmark_operation_enqueue(self, iterations: int = 1000) -> Dict[str, Any]:
        """Benchmark operation enqueueing"""
        latencies = []

        for i in range(iterations):
            start = time.time() * 1000
            self.queue.enqueue(
                "session1",
                "agent1",
                "PlaceOrder",
                {"symbol": "EURUSD", "volume": 1.0},
            )
            end = time.time() * 1000
            latencies.append(end - start)

        return self._analyze_latencies("Operation Enqueue", latencies)

    def benchmark_operation_status_update(self, iterations: int = 1000) -> Dict[str, Any]:
        """Benchmark operation status updates"""
        # Create operations first
        op_ids = []
        for i in range(iterations):
            op_id = self.queue.enqueue(
                "session1",
                "agent1",
                "PlaceOrder",
                {"symbol": "EURUSD", "volume": 1.0},
            )
            op_ids.append(op_id)

        # Benchmark updates
        latencies = []

        for op_id in op_ids:
            start = time.time() * 1000
            self.queue.update_status(op_id, "COMPLETED", {"order_id": 123})
            end = time.time() * 1000
            latencies.append(end - start)

        return self._analyze_latencies("Operation Status Update", latencies)

    def benchmark_operation_retrieval(self, iterations: int = 1000) -> Dict[str, Any]:
        """Benchmark operation retrieval"""
        # Create operations first
        op_ids = []
        for i in range(iterations):
            op_id = self.queue.enqueue(
                "session1",
                "agent1",
                "PlaceOrder",
                {"symbol": "EURUSD"},
            )
            op_ids.append(op_id)

        # Benchmark retrieval
        latencies = []

        for op_id in op_ids:
            start = time.time() * 1000
            self.queue.get_operation(op_id)
            end = time.time() * 1000
            latencies.append(end - start)

        return self._analyze_latencies("Operation Retrieval", latencies)

    def benchmark_queue_fetch(self, iterations: int = 100) -> Dict[str, Any]:
        """Benchmark queue fetching"""
        # Create 1000 queued operations
        for i in range(1000):
            self.queue.enqueue(
                "session1",
                "agent1",
                "PlaceOrder",
                {"symbol": f"SYMBOL{i}", "volume": 1.0},
            )

        # Benchmark fetching
        latencies = []

        for i in range(iterations):
            start = time.time() * 1000
            self.queue.get_queued(limit=10)
            end = time.time() * 1000
            latencies.append(end - start)

        return self._analyze_latencies("Queue Fetch (10 items)", latencies)

    def _analyze_latencies(self, name: str, latencies: List[float]) -> Dict[str, Any]:
        """Analyze latency statistics"""
        if not latencies:
            return {"name": name, "error": "No latencies recorded"}

        return {
            "name": name,
            "count": len(latencies),
            "mean_ms": round(statistics.mean(latencies), 3),
            "median_ms": round(statistics.median(latencies), 3),
            "stdev_ms": round(statistics.stdev(latencies), 3) if len(latencies) > 1 else 0,
            "min_ms": round(min(latencies), 3),
            "max_ms": round(max(latencies), 3),
            "p95_ms": round(sorted(latencies)[int(len(latencies) * 0.95)], 3),
            "p99_ms": round(sorted(latencies)[int(len(latencies) * 0.99)], 3),
        }


def run_all_benchmarks():
    """Run all benchmarks"""
    runner = BenchmarkRunner()
    runner.setup()

    print("\n" + "=" * 60)
    print("MT5 gRPC Service Performance Benchmarks")
    print("=" * 60 + "\n")

    results = []

    # Run benchmarks
    print("Running operation enqueue benchmark...")
    results.append(runner.benchmark_operation_enqueue())

    print("Running operation status update benchmark...")
    results.append(runner.benchmark_operation_status_update())

    print("Running operation retrieval benchmark...")
    results.append(runner.benchmark_operation_retrieval())

    print("Running queue fetch benchmark...")
    results.append(runner.benchmark_queue_fetch())

    # Print results
    print("\n" + "=" * 60)
    print("Benchmark Results")
    print("=" * 60 + "\n")

    for result in results:
        print(f"Benchmark: {result['name']}")
        print(f"  Samples:   {result['count']}")
        print(f"  Mean:      {result['mean_ms']} ms")
        print(f"  Median:    {result['median_ms']} ms")
        print(f"  Stdev:     {result['stdev_ms']} ms")
        print(f"  Min:       {result['min_ms']} ms")
        print(f"  Max:       {result['max_ms']} ms")
        print(f"  P95:       {result['p95_ms']} ms")
        print(f"  P99:       {result['p99_ms']} ms")
        print()

    # Validate against requirements
    print("=" * 60)
    print("Requirement Validation (SC-001: P95 latency < 100ms)")
    print("=" * 60 + "\n")

    all_pass = True
    for result in results:
        if result['p95_ms'] < 100:
            status = "✓ PASS"
        else:
            status = "✗ FAIL"
            all_pass = False

        print(f"{status} - {result['name']}: P95={result['p95_ms']}ms")

    print()
    if all_pass:
        print("✓ All benchmarks pass requirement (P95 < 100ms)")
    else:
        print("✗ Some benchmarks fail requirement")

    runner.teardown()

    return all_pass


if __name__ == "__main__":
    success = run_all_benchmarks()
    sys.exit(0 if success else 1)
