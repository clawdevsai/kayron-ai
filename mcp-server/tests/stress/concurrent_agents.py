"""Stress test for concurrent agent access"""
import asyncio
import time
import sys
import os
from typing import List, Dict, Any
import random

# Add src to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../src"))

from operation_queue import OperationQueue
from callback_manager import CallbackManager
from db_schema import DatabaseSchema


class AgentSimulator:
    """Simulates agent behavior"""

    def __init__(
        self,
        agent_id: str,
        queue: OperationQueue,
        callback_mgr: CallbackManager,
        operations_per_agent: int = 10,
    ):
        self.agent_id = agent_id
        self.queue = queue
        self.callback_mgr = callback_mgr
        self.operations_per_agent = operations_per_agent
        self.completed = 0
        self.failed = 0
        self.latencies: List[float] = []

    async def run(self):
        """Run agent simulation"""
        # Create session and stream
        session_id = f"session_{self.agent_id}"
        stream_id = self.callback_mgr.create_stream(session_id, self.agent_id)
        stream = self.callback_mgr.get_stream(stream_id)

        operations = []

        # Submit operations
        for i in range(self.operations_per_agent):
            op_id = self.queue.enqueue(
                session_id,
                self.agent_id,
                "PlaceOrder",
                {
                    "symbol": random.choice(["EURUSD", "GBPUSD", "USDJPY"]),
                    "volume": random.uniform(0.1, 10.0),
                    "price": random.uniform(1.0, 1.2),
                },
            )

            operations.append(op_id)

            # Simulate operation execution
            stream.push_update(op_id, "QUEUED")
            await asyncio.sleep(0.001)  # Small delay between submissions

        # Process operations
        for op_id in operations:
            start = time.time()

            try:
                # Simulate execution
                self.queue.mark_executing(op_id)
                stream.push_update(op_id, "EXECUTING")

                # Simulate processing time
                await asyncio.sleep(random.uniform(0.01, 0.05))

                # Complete operation
                self.queue.update_status(op_id, "COMPLETED", {"order_id": 123})
                stream.push_update(op_id, "COMPLETED", {"order_id": 123})

                self.completed += 1

            except Exception as e:
                self.failed += 1
                self.queue.update_status(op_id, "FAILED", None, "ERROR", str(e))
                stream.push_update(
                    op_id, "FAILED", {"error_code": "ERROR", "error_message": str(e)}
                )

            latency = (time.time() - start) * 1000
            self.latencies.append(latency)

        # Close stream
        self.callback_mgr.close_stream(stream_id)

    def get_statistics(self) -> Dict[str, Any]:
        """Get agent statistics"""
        if self.latencies:
            import statistics

            return {
                "agent_id": self.agent_id,
                "completed": self.completed,
                "failed": self.failed,
                "total": self.operations_per_agent,
                "success_rate": self.completed / self.operations_per_agent * 100,
                "avg_latency_ms": round(statistics.mean(self.latencies), 3),
                "min_latency_ms": round(min(self.latencies), 3),
                "max_latency_ms": round(max(self.latencies), 3),
            }
        else:
            return {
                "agent_id": self.agent_id,
                "completed": self.completed,
                "failed": self.failed,
                "error": "No latencies recorded",
            }


class StressTestRunner:
    """Runs stress tests with concurrent agents"""

    def __init__(
        self,
        num_agents: int = 50,
        operations_per_agent: int = 10,
        db_path: str = "stress.db",
    ):
        self.num_agents = num_agents
        self.operations_per_agent = operations_per_agent
        self.db_path = db_path
        self.queue = OperationQueue(db_path)
        self.callback_mgr = CallbackManager()
        self.agents: List[AgentSimulator] = []

    def setup(self):
        """Setup stress test"""
        db = DatabaseSchema(self.db_path)
        db.connect()
        db.initialize()
        db.close()

    def teardown(self):
        """Cleanup stress test"""
        import os

        if os.path.exists(self.db_path):
            os.remove(self.db_path)

    def create_agents(self):
        """Create simulated agents"""
        self.agents = [
            AgentSimulator(
                f"agent_{i}",
                self.queue,
                self.callback_mgr,
                self.operations_per_agent,
            )
            for i in range(self.num_agents)
        ]

    async def run_stress_test(self):
        """Run concurrent agents"""
        tasks = [agent.run() for agent in self.agents]
        await asyncio.gather(*tasks)

    def print_results(self):
        """Print test results"""
        print("\n" + "=" * 60)
        print(f"Stress Test Results ({self.num_agents} concurrent agents)")
        print("=" * 60 + "\n")

        total_completed = 0
        total_failed = 0
        all_latencies = []

        agent_stats = []
        for agent in self.agents:
            stats = agent.get_statistics()
            agent_stats.append(stats)
            total_completed += stats.get("completed", 0)
            total_failed += stats.get("failed", 0)
            all_latencies.extend(agent.latencies)

        # Summary statistics
        print(f"Total Agents:        {self.num_agents}")
        print(f"Operations/Agent:    {self.operations_per_agent}")
        print(f"Total Operations:    {self.num_agents * self.operations_per_agent}")
        print(f"Completed:           {total_completed}")
        print(f"Failed:              {total_failed}")
        print(f"Success Rate:        {total_completed / (self.num_agents * self.operations_per_agent) * 100:.1f}%")

        if all_latencies:
            import statistics

            print(f"\nLatency Statistics:")
            print(f"  Mean:              {statistics.mean(all_latencies):.3f} ms")
            print(f"  Median:            {statistics.median(all_latencies):.3f} ms")
            print(f"  Min:               {min(all_latencies):.3f} ms")
            print(f"  Max:               {max(all_latencies):.3f} ms")

            sorted_latencies = sorted(all_latencies)
            print(f"  P95:               {sorted_latencies[int(len(all_latencies) * 0.95)]:.3f} ms")
            print(f"  P99:               {sorted_latencies[int(len(all_latencies) * 0.99)]:.3f} ms")

        # Per-agent summary
        print(f"\nPer-Agent Summary (first 5 agents):")
        for stats in agent_stats[:5]:
            print(
                f"  {stats['agent_id']}: {stats['completed']}/{stats['total']} "
                f"({stats.get('success_rate', 0):.0f}%) "
                f"avg={stats.get('avg_latency_ms', 'N/A')}ms"
            )

        # Validation
        print("\n" + "=" * 60)
        print("Requirement Validation (SC-002: No degradation with 50 agents)")
        print("=" * 60 + "\n")

        # Check no degradation (99% success rate)
        success_rate = total_completed / (self.num_agents * self.operations_per_agent) * 100
        if success_rate >= 99.0:
            print(f"✓ PASS - Success rate: {success_rate:.1f}% (threshold: 99%)")
        else:
            print(f"✗ FAIL - Success rate: {success_rate:.1f}% (threshold: 99%)")

        # Check queues handled
        queued_stats = self.queue.get_queued(limit=1000)
        print(f"\nQueued Operations: {len(queued_stats)}")

        # Check callback streams
        callback_health = self.callback_mgr.health_check()
        print(f"Active Callback Streams: {callback_health['total_streams']}")


async def main():
    """Run stress test"""
    runner = StressTestRunner(num_agents=50, operations_per_agent=10)
    runner.setup()

    try:
        print("\nStarting stress test with 50 concurrent agents...")
        print("(This may take a minute...)\n")

        runner.create_agents()

        start_time = time.time()
        await runner.run_stress_test()
        elapsed = time.time() - start_time

        runner.print_results()

        print(f"\nTotal test time: {elapsed:.2f} seconds")
        print(f"Operations/sec: {(50 * 10) / elapsed:.1f}")

    finally:
        runner.teardown()


if __name__ == "__main__":
    asyncio.run(main())
