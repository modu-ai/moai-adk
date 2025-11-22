# Cloud Performance & Cost Optimization

## Cost Optimization Strategies

### Reserved Instances vs On-Demand

**Cost Analysis**:
```python
from datetime import datetime, timedelta

class CloudCostAnalyzer:
    """Analyze cost optimization opportunities"""

    @staticmethod
    def calculate_instance_costs():
        """Compare instance purchasing options"""
        # On-demand: Pay per hour
        on_demand_hourly = 0.12  # $/hour
        annual_hours = 365 * 24  # 8,760 hours/year
        on_demand_annual = on_demand_hourly * annual_hours  # $1,051.20

        # Reserved Instance (1-year): Pay upfront
        ri_upfront = 700
        ri_hourly = 0.02
        ri_annual = ri_upfront + (ri_hourly * annual_hours)  # $875.20

        # Savings
        savings = on_demand_annual - ri_annual  # $176/year (16.7%)
        savings_pct = (savings / on_demand_annual) * 100

        return {
            'on_demand_annual': on_demand_annual,
            'reserved_1yr_annual': ri_annual,
            'savings': savings,
            'savings_percent': savings_pct,
            'breakeven_months': ri_upfront / (on_demand_hourly * 730)
        }

    @staticmethod
    def optimize_instance_types():
        """Right-size instances based on utilization"""
        instances = [
            {'type': 't3.large', 'cpu_util': 10, 'memory_util': 15},
            {'type': 't3.xlarge', 'cpu_util': 5, 'memory_util': 8},
            {'type': 'm5.2xlarge', 'cpu_util': 20, 'memory_util': 25},
        ]

        recommendations = []
        for instance in instances:
            # Flag underutilized instances
            if instance['cpu_util'] < 20 and instance['memory_util'] < 20:
                recommendations.append({
                    'current': instance['type'],
                    'action': 'downsize_or_consolidate',
                    'potential_savings': '40-60%'
                })

        return recommendations
```

### Auto-Scaling Efficiency

```yaml
# AWS EC2 Auto-Scaling with cost optimization
AWSTemplateFormatVersion: '2010-09-09'
Description: Cost-optimized auto-scaling group

Parameters:
  MinInstances:
    Type: Number
    Default: 2
  MaxInstances:
    Type: Number
    Default: 20
  TargetCPUUtilization:
    Type: Number
    Default: 70

Resources:
  LaunchTemplate:
    Type: AWS::EC2::LaunchTemplate
    Properties:
      LaunchTemplateName: cost-optimized-template
      LaunchTemplateData:
        ImageId: ami-12345678
        InstanceType: t3.medium
        # Enable EBS optimization for better I/O
        EbsOptimized: true
        # Use spot instances (up to 90% discount)
        InstanceMarketOptions:
          MarketType: spot
          SpotOptions:
            SpotInstanceType: persistent
            MaxPrice: "0.05"  # Max price willing to pay

  AutoScalingGroup:
    Type: AWS::AutoScaling::AutoScalingGroup
    Properties:
      MinSize: !Ref MinInstances
      MaxSize: !Ref MaxInstances
      LaunchTemplate:
        LaunchTemplateId: !Ref LaunchTemplate
        Version: !GetAtt LaunchTemplate.LatestVersionNumber
      TargetGroupARNs:
        - !Ref TargetGroup

  ScalingPolicy:
    Type: AWS::AutoScaling::ScalingPolicy
    Properties:
      AdjustmentType: PercentChangeInCapacity
      AutoScalingGroupName: !Ref AutoScalingGroup
      EstimatedWarmupSeconds: 300
      MetricAggregationType: Average
      PolicyType: TargetTrackingScaling
      TargetTrackingConfiguration:
        PredefinedMetricSpecification:
          PredefinedMetricType: ASGAverageCPUUtilization
        TargetValue: !Ref TargetCPUUtilization
```

---

## Storage Optimization

### Tiered Storage Strategy

```python
from enum import Enum
from datetime import datetime, timedelta

class StorageTier(Enum):
    HOT = "hot"           # Immediate access (expensive)
    WARM = "warm"         # Days to weeks (moderate cost)
    COLD = "cold"         # Months/years (cheap)
    ARCHIVE = "archive"   # Long-term retention (very cheap)

class StorageOptimizer:
    """Automatically move data to cost-effective tiers"""

    async def lifecycle_transition_policy(self):
        """Define data movement based on age"""
        return {
            'hot_to_warm': {
                'condition': 'age >= 30 days',
                'storage_class': 'STANDARD_IA',  # Infrequent Access
                'cost_reduction': '50%'
            },
            'warm_to_cold': {
                'condition': 'age >= 90 days',
                'storage_class': 'GLACIER',
                'cost_reduction': '80%'
            },
            'cold_to_archive': {
                'condition': 'age >= 365 days',
                'storage_class': 'DEEP_ARCHIVE',
                'cost_reduction': '95%',
                'retrieval_time': '12 hours'
            }
        }

    async def estimate_storage_costs(self, data_profile):
        """Calculate cost savings from tiering"""
        # Example: 100GB dataset
        data_size_gb = 100

        # All on hot storage
        hot_only_monthly = data_size_gb * 0.023  # $2.30/month

        # With tiering:
        # Day 0-30: Hot (100GB)
        # Day 31-90: Warm (100GB)
        # Day 91+: Cold (100GB)
        average_hot_gb = (30 * 100) / 365  # Only 30 days hot
        average_warm_gb = (60 * 100) / 365
        average_cold_gb = (275 * 100) / 365

        tiered_monthly = (
            (average_hot_gb * 0.023) +      # Hot: $0.023/GB
            (average_warm_gb * 0.0125) +    # Warm: $0.0125/GB
            (average_cold_gb * 0.004)       # Cold: $0.004/GB
        )

        savings = hot_only_monthly - tiered_monthly
        savings_pct = (savings / hot_only_monthly) * 100

        return {
            'hot_only': hot_only_monthly,
            'with_tiering': tiered_monthly,
            'monthly_savings': savings,
            'annual_savings': savings * 12,
            'savings_percent': savings_pct
        }
```

### Data Deduplication

```python
import hashlib
from typing import Dict, Set

class DeduplicationEngine:
    """Reduce storage by eliminating duplicates"""

    def __init__(self):
        self.file_hashes: Dict[str, str] = {}
        self.content_hashes: Set[str] = set()

    async def deduplicate_dataset(self, files: list) -> Dict:
        """Identify and remove duplicate content"""
        stats = {
            'total_files': len(files),
            'unique_files': 0,
            'duplicate_files': 0,
            'space_wasted': 0,
            'space_saveable': 0
        }

        for file_path in files:
            # Calculate content hash
            content_hash = await self._calculate_hash(file_path)

            if content_hash in self.content_hashes:
                # Duplicate found
                stats['duplicate_files'] += 1
                file_size = await self._get_file_size(file_path)
                stats['space_saveable'] += file_size
            else:
                self.content_hashes.add(content_hash)
                stats['unique_files'] += 1

        return stats

    async def _calculate_hash(self, file_path: str) -> str:
        """Calculate SHA-256 hash of file content"""
        sha256_hash = hashlib.sha256()
        with open(file_path, "rb") as f:
            for chunk in iter(lambda: f.read(4096), b""):
                sha256_hash.update(chunk)
        return sha256_hash.hexdigest()
```

---

## Network Optimization

### CDN Configuration

```yaml
# CloudFront distribution for global content delivery
AWSTemplateFormatVersion: '2010-09-09'

Resources:
  CDNDistribution:
    Type: AWS::CloudFront::Distribution
    Properties:
      DistributionConfig:
        Enabled: true
        HttpVersion: http2and3  # HTTP/3 for better performance
        DefaultRootObject: index.html

        # Origin configuration
        Origins:
          - Id: S3Origin
            DomainName: my-bucket.s3.amazonaws.com
            S3OriginConfig:
              OriginAccessIdentity: origin-access-identity/cloudfront/ABCDEFG

        # Caching behavior
        DefaultCacheBehavior:
          AllowedMethods: [GET, HEAD, OPTIONS]
          CachePolicyId: 658327ea-f89d-4fab-a63d-7e88639e58f6  # Managed policy
          OriginRequestPolicyId: 216adef5-5c7f-47e4-b989-5492eafa07d3
          ViewerProtocolPolicy: redirect-to-https
          Compress: true  # Enable GZIP compression

        # Specific caching for different content types
        CacheBehaviors:
          - PathPattern: /api/*
            AllowedMethods: [GET, HEAD, POST, PUT, DELETE, OPTIONS]
            CachePolicyId: 4135ea3d-c35d-46eb-81d7-rewrite305e860c  # Minimal cache
            OriginRequestPolicyId: 216adef5-5c7f-47e4-b989-5492eafa07d3
            ViewerProtocolPolicy: https-only
            Compress: false

          - PathPattern: /static/*
            AllowedMethods: [GET, HEAD]
            CachePolicyId: 658327ea-f89d-4fab-a63d-7e88639e58f6  # Long cache
            ViewerProtocolPolicy: https-only
            Compress: true
```

### Connection Pooling

```python
import asyncio
import httpx
from typing import List

class ConnectionPoolManager:
    """Optimize network connections with pooling"""

    def __init__(self, max_connections: int = 100):
        self.max_connections = max_connections
        self.client = None

    async def initialize(self):
        """Create connection pool"""
        self.client = httpx.AsyncClient(
            limits=httpx.Limits(
                max_connections=self.max_connections,
                max_keepalive_connections=20,
                keepalive_expiry=5.0
            ),
            timeout=10.0,
            verify=True,
            http2=True  # Enable HTTP/2
        )

    async def make_parallel_requests(self, urls: List[str]):
        """Execute requests with connection reuse"""
        tasks = [
            self.client.get(url) for url in urls
        ]
        results = await asyncio.gather(*tasks)
        return results

    async def close(self):
        await self.client.aclose()
```

---

## Monitoring & Profiling

### Cloud Cost Monitoring

```python
import asyncio
from datetime import datetime, timedelta

class CostMonitor:
    """Track and alert on cloud costs"""

    async def monitor_spending(self, budget: float):
        """Monitor daily spending against budget"""
        while True:
            # Get costs from cloud provider
            daily_cost = await self.get_daily_cost()
            projected_monthly = daily_cost * 30

            # Calculate burn rate
            days_elapsed = (datetime.now() - self.month_start).days
            daily_average = self.month_start_balance / max(days_elapsed, 1)
            projected_total = daily_average * 30

            # Alert if over budget
            if projected_total > budget:
                overage_pct = ((projected_total - budget) / budget) * 100
                await self.send_alert(
                    f"Cost alert: {overage_pct:.1f}% over budget"
                )

            # Sleep until next check
            await asyncio.sleep(3600)  # Check hourly

    async def optimize_idle_resources(self):
        """Find and shut down idle resources"""
        idle_resources = []

        # Check EC2 instances
        for instance in await self.get_instances():
            cpu_util = await self.get_metric(
                instance.id, 'CPUUtilization', days=7
            )
            if max(cpu_util) < 5:  # Less than 5% utilization
                idle_resources.append({
                    'type': 'ec2',
                    'id': instance.id,
                    'monthly_cost': 50,
                    'action': 'stop'
                })

        # Check RDS databases
        for db in await self.get_databases():
            connections = await self.get_metric(
                db.id, 'DatabaseConnections', days=7
            )
            if max(connections) == 0:  # No connections
                idle_resources.append({
                    'type': 'rds',
                    'id': db.id,
                    'monthly_cost': 200,
                    'action': 'stop'
                })

        return idle_resources
```

### Performance Profiling

```python
import time
import psutil

class CloudPerformanceProfiler:
    """Profile cloud application performance"""

    async def profile_vm_performance(self):
        """Monitor VM resource utilization"""
        metrics = {
            'cpu': psutil.cpu_percent(interval=1),
            'memory': psutil.virtual_memory().percent,
            'disk': psutil.disk_usage('/').percent,
            'network': psutil.net_io_counters()
        }

        # Identify performance bottlenecks
        warnings = []
        if metrics['cpu'] > 80:
            warnings.append(f"High CPU: {metrics['cpu']}%")
        if metrics['memory'] > 85:
            warnings.append(f"High Memory: {metrics['memory']}%")
        if metrics['disk'] > 90:
            warnings.append(f"Low Disk Space: {100 - metrics['disk']}%")

        return {
            'metrics': metrics,
            'warnings': warnings,
            'recommendations': self._generate_recommendations(metrics)
        }

    def _generate_recommendations(self, metrics):
        """Generate optimization recommendations"""
        recommendations = []

        if metrics['cpu'] > 70:
            recommendations.append({
                'issue': 'High CPU utilization',
                'action': 'Scale up to larger instance type',
                'estimate_cost_increase': '50-100%'
            })

        if metrics['memory'] > 80:
            recommendations.append({
                'issue': 'High memory pressure',
                'action': 'Increase instance memory or split workload',
                'estimate_cost_increase': '30-50%'
            })

        return recommendations
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
