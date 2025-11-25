# Advanced Cloud Architecture Patterns

## Multi-Cloud Strategy

### Cloud Abstraction Layer

**Pattern**: Abstract cloud-specific APIs to enable portability.

```python
from abc import ABC, abstractmethod
from typing import List, Dict, Any

class CloudProvider(ABC):
    """Abstract interface for cloud providers"""

    @abstractmethod
    async def create_vm(self, config: Dict[str, Any]) -> str:
        """Create virtual machine. Returns instance ID."""
        pass

    @abstractmethod
    async def delete_vm(self, instance_id: str) -> bool:
        """Delete virtual machine"""
        pass

    @abstractmethod
    async def list_vms(self) -> List[str]:
        """List all virtual machine IDs"""
        pass

class AWSProvider(CloudProvider):
    def __init__(self, access_key: str, secret_key: str, region: str):
        self.ec2 = boto3.client('ec2', region_name=region)

    async def create_vm(self, config: Dict[str, Any]) -> str:
        response = self.ec2.run_instances(
            ImageId=config['image_id'],
            InstanceType=config['instance_type'],
            MinCount=1,
            MaxCount=1
        )
        return response['Instances'][0]['InstanceId']

class GCPProvider(CloudProvider):
    def __init__(self, project_id: str, zone: str):
        self.compute = googleapiclient.discovery.build('compute', 'v1')
        self.project_id = project_id
        self.zone = zone

    async def create_vm(self, config: Dict[str, Any]) -> str:
        operation = self.compute.instances().insert(
            project=self.project_id,
            zone=self.zone,
            body={
                'name': config['name'],
                'machineType': f'zones/{self.zone}/machineTypes/{config["machine_type"]}',
                'disks': [{
                    'boot': True,
                    'initializeParams': {
                        'sourceImage': config['image']
                    }
                }]
            }
        ).execute()
        return operation

class MultiCloudOrchestrator:
    def __init__(self):
        self.providers = {}

    def register_provider(self, name: str, provider: CloudProvider):
        self.providers[name] = provider

    async def create_vm_across_clouds(self, cloud: str, config: Dict) -> str:
        """Abstract cloud provider differences"""
        if cloud not in self.providers:
            raise ValueError(f"Cloud provider {cloud} not registered")
        return await self.providers[cloud].create_vm(config)

# Usage: Same code works with any cloud provider
orchestrator = MultiCloudOrchestrator()
orchestrator.register_provider('aws', AWSProvider(access_key, secret_key, 'us-east-1'))
orchestrator.register_provider('gcp', GCPProvider('my-project', 'us-central1-a'))

# Deploy to AWS
aws_vm_id = await orchestrator.create_vm_across_clouds('aws', config)

# Deploy to GCP (same config structure)
gcp_vm_id = await orchestrator.create_vm_across_clouds('gcp', config)
```

---

### Cross-Region Failover

**Pattern**: Automatically switch traffic to alternate regions on failure.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: failover-config
data:
  primary_region: us-east-1
  secondary_regions:
    - us-west-1
    - eu-west-1
  health_check_interval: 30s
  failover_threshold: 3

---
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: global-api
spec:
  hosts:
  - api.example.com
  ports:
  - number: 443
    name: https
    protocol: HTTPS
  location: MESH_EXTERNAL
  resolution: DNS
  endpoints:
  - address: api-us-east-1.example.com
    ports:
      https: 443
    region: us-east-1
    weight: 100  # Primary
  - address: api-us-west-1.example.com
    ports:
      https: 443
    region: us-west-1
    weight: 0   # Standby
  - address: api-eu-west-1.example.com
    ports:
      https: 443
    region: eu-west-1
    weight: 0   # Standby

---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: global-api
spec:
  host: api.example.com
  trafficPolicy:
    outlierDetection:
      consecutive5xxErrors: 3
      interval: 30s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
      minRequestVolume: 5
      splitExternalLocalOriginErrors: true
```

---

## Serverless Architecture

### Function Composition Pattern

**AWS Lambda composition without vendor lock-in**:

```python
from typing import Callable, Any
import json

class FunctionComposer:
    """Compose multiple serverless functions"""

    def __init__(self):
        self.functions = {}

    def register(self, name: str, func: Callable) -> None:
        self.functions[name] = func

    async def execute_pipeline(self, pipeline: List[str], input_data: Any) -> Any:
        """Execute sequence of functions"""
        result = input_data

        for func_name in pipeline:
            if func_name not in self.functions:
                raise ValueError(f"Function {func_name} not registered")

            result = await self.functions[func_name](result)

        return result

# Define individual functions
async def validate_input(data: Dict) -> Dict:
    """Validate and normalize input"""
    if 'email' not in data:
        raise ValueError("Missing email field")
    return {**data, 'email': data['email'].lower()}

async def enrich_user_data(data: Dict) -> Dict:
    """Add additional user information"""
    user_service = UserService()
    enriched = await user_service.get_full_profile(data['email'])
    return {**data, **enriched}

async def send_notification(data: Dict) -> Dict:
    """Send notification"""
    await NotificationService.send_email(
        data['email'],
        f"Welcome {data['name']}"
    )
    return data

# Compose pipeline
composer = FunctionComposer()
composer.register('validate', validate_input)
composer.register('enrich', enrich_user_data)
composer.register('notify', send_notification)

# Execute
result = await composer.execute_pipeline(
    ['validate', 'enrich', 'notify'],
    {'email': 'user@example.com'}
)
```

---

## Kubernetes Advanced Patterns

### Pod Autoscaling with Custom Metrics

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-autoscaler
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-service
  minReplicas: 2
  maxReplicas: 50
  metrics:
  # CPU-based scaling
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70

  # Memory-based scaling
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80

  # Custom metric scaling (requests/second)
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "1000"

  # External metric (from cloud provider)
  - type: External
    external:
      metric:
        name: queue_depth
        selector:
          matchLabels:
            queue: "job-queue"
      target:
        type: Value
        value: "30"

  behavior:
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 50  # Double size per scale-up
        periodSeconds: 30
      - type: Pods
        value: 4   # Add 4 pods per scale-up
        periodSeconds: 30
      selectPolicy: Max  # Use whichever adds more pods

    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50  # Reduce to 50% of current
        periodSeconds: 60
```

### Network Policies for Zero-Trust

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all-traffic
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-traffic
spec:
  podSelector:
    matchLabels:
      app: api
  policyTypes:
  - Ingress
  - Egress

  ingress:
  # Allow from Istio ingress gateway only
  - from:
    - namespaceSelector:
        matchLabels:
          name: istio-system
    - podSelector:
        matchLabels:
          app: istio-ingressgateway
    ports:
    - protocol: TCP
      port: 8080

  egress:
  # Allow DNS
  - to:
    - namespaceSelector: {}
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: UDP
      port: 53

  # Allow to database
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432

  # Allow external APIs
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443
```

---

## Infrastructure as Code

### Terraform Multi-Cloud Module

```hcl
# Multi-cloud compute module
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

module "compute_aws" {
  source = "./modules/compute"

  cloud_provider = "aws"
  instance_type  = "t3.medium"
  instance_count = var.aws_instance_count
  region        = var.aws_region

  tags = {
    Environment = var.environment
    Terraform   = "true"
  }
}

module "compute_gcp" {
  source = "./modules/compute"

  cloud_provider = "gcp"
  instance_type  = "n1-standard-1"
  instance_count = var.gcp_instance_count
  region        = var.gcp_region

  tags = {
    environment = var.environment
    terraform   = "true"
  }
}

output "aws_instances" {
  value = module.compute_aws.instance_ids
}

output "gcp_instances" {
  value = module.compute_gcp.instance_ids
}
```

---

## Disaster Recovery

### RTO/RPO Strategy

```python
from dataclasses import dataclass
from datetime import timedelta

@dataclass
class DisasterRecoveryPolicy:
    """Define RTO and RPO requirements"""
    name: str
    rto: timedelta  # Recovery Time Objective
    rpo: timedelta  # Recovery Point Objective

# Different tiers for different applications
policies = {
    'critical': DisasterRecoveryPolicy(
        name='Critical Services',
        rto=timedelta(minutes=15),   # 15 minutes downtime acceptable
        rpo=timedelta(minutes=5)     # 5 minutes data loss acceptable
    ),
    'important': DisasterRecoveryPolicy(
        name='Important Services',
        rto=timedelta(hours=1),
        rpo=timedelta(hours=1)
    ),
    'standard': DisasterRecoveryPolicy(
        name='Standard Services',
        rto=timedelta(hours=4),
        rpo=timedelta(hours=4)
    )
}

class RecoveryOrchestrator:
    """Implement DR strategy with backup/restore"""

    async def implement_rto_rpo(self, policy: DisasterRecoveryPolicy):
        """Configure infrastructure for RTO/RPO targets"""

        # For RTO < 1 hour: Active-Active or Active-Passive with fast failover
        if policy.rto < timedelta(hours=1):
            await self.setup_active_passive_failover(policy.rto)

        # For RTO > 1 hour: Automated backup and recovery
        else:
            await self.setup_backup_recovery(policy)

        # For RPO < 5 minutes: Continuous replication
        if policy.rpo < timedelta(minutes=5):
            await self.setup_continuous_replication()

        # For RPO > 1 hour: Scheduled backups
        else:
            await self.setup_scheduled_backups(policy.rpo)

    async def setup_active_passive_failover(self, rto: timedelta):
        """Setup active-passive with fast detection and failover"""
        health_check_interval = rto / 3  # Check health 3x during RTO window
        await self.configure_health_checks(interval=health_check_interval)
        await self.setup_automated_failover()

    async def setup_continuous_replication(self):
        """Real-time data replication to standby region"""
        await self.setup_database_replication(mode='synchronous')
        await self.setup_file_storage_replication(mode='continuous')
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
