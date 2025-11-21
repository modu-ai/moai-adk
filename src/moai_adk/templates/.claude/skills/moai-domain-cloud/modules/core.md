    services_secondary_range_name = "services"
  }
}

# Azure AKS Cluster
resource "azurerm_kubernetes_cluster" "main" {
  count = var.cloud_provider == "azure" ? 1 : 0
  
  name                = var.cluster_name
  location            = var.region
  resource_group_name = var.resource_group_name
  dns_prefix          = "${var.cluster_name}-dns"
  
  kubernetes_version = "1.34.0"
  
  default_node_pool {
    name       = "default"
    node_count = 1
    vm_size    = "Standard_D2s_v3"
  }
  
  identity {
    type = "SystemAssigned"
  }
}

# Output cluster connection details
output "cluster_endpoint" {
  value = var.cloud_provider == "aws" ? aws_eks_cluster.main[0].endpoint :
         var.cloud_provider == "gcp" ? google_container_cluster.main[0].endpoint :
         azurerm_kubernetes_cluster.main[0].fqdn
}

output "cluster_ca_certificate" {
  value = var.cloud_provider == "aws" ? aws_eks_cluster.main[0].certificate_authority[0].data :
         var.cloud_provider == "gcp" ? google_container_cluster.main[0].master_auth[0].cluster_ca_certificate :
         azurerm_kubernetes_cluster.main[0].kube_config[0].cluster_ca_certificate
}
```

**Kubernetes Deployment for Multi-Cloud:**

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  labels:
    app: webapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp
        image: nginx:1.27
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5

apiVersion: v1
kind: Service
metadata:
  name: webapp-service
spec:
  selector:
    app: webapp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: LoadBalancer
```


#### Pattern 3: Cloud-Native Database with AWS RDS PostgreSQL 17

**Problem**: Need scalable, highly available database with automated backups, monitoring, and security.

**Solution**: AWS RDS with PostgreSQL 17 and enhanced monitoring.

```python
# lib/database_stack.py
from aws_cdk import (
    Stack,
    aws_rds as rds,
    aws_ec2 as ec2,
    aws_secretsmanager as secretsmanager,
    RemovalPolicy
)
from constructs import Construct

class DatabaseStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, vpc, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)
        
        # Database security group
        db_security_group = ec2.SecurityGroup(
            self, "DatabaseSecurityGroup",
            vpc=vpc,
            description="Security group for RDS database",
            allow_all_outbound=False
        )
        
        # Database credentials secret
        db_secret = secretsmanager.Secret(
            self, "DatabaseSecret",
            secret_name="database-credentials",
            description="Database credentials for application"
        )
        
        # RDS PostgreSQL 17 instance
        database = rds.DatabaseInstance(
            self, "ApplicationDatabase",
            engine=rds.DatabaseInstanceEngine.postgres(
                version=rds.PostgresEngineVersion.VER_17
            ),
            instance_type=ec2.InstanceType("db.t3.micro"),
            vpc=vpc,
            vpc_subnets=ec2.SubnetSelection(
                subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS
            ),
            security_groups=[db_security_group],
            database_name="appdb",
            credentials=rds.Credentials.from_secret(db_secret),
            backup_retention=Duration.days(7),
            deletion_protection=False,
            removal_policy=RemovalPolicy.DESTROY,
            monitoring_interval=Duration.seconds(60),
            enable_performance_insights=True,
            performance_insight_retention=rds.PerformanceInsightRetention.DEFAULT
        )
        
        # Export database connection details
        self.database_secret = db_secret
        self.database_instance = database
```


### Level 3: Advanced Integration

#### Multi-Cloud Cost Optimization Strategy

```python
# cost_optimizer.py
import boto3
import google.cloud
from azure.mgmt.cost_management import CostManagementClient
from datetime import datetime, timedelta

class MultiCloudCostOptimizer:
    """Optimize costs across AWS, GCP, and Azure."""
    
    def __init__(self):
        self.aws_client = boto3.client('ce')
        self.gcp_client = google.cloud.billing.BudgetServiceClient()
        self.azure_client = CostManagementClient()
    
    def analyze_aws_costs(self, start_date, end_date):
        """Analyze AWS costs by service and region."""
        response = self.aws_client.get_cost_and_usage(
            TimePeriod={
                'Start': start_date,
                'End': end_date
            },
            Granularity='MONTHLY',
            Metrics=['BlendedCost'],
            GroupBy=[
                {'Type': 'DIMENSION', 'Key': 'SERVICE'},
                {'Type': 'DIMENSION', 'Key': 'REGION'}
            ]
        )
        
        return self._process_cost_data(response['ResultsByTime'])
    
    def optimize_aws_resources(self):
        """Provide AWS-specific cost optimization recommendations."""
        recommendations = []
        
        # Lambda optimization
        recommendations.append({
            'service': 'Lambda',
            'suggestion': 'Use provisioned concurrency for predictable workloads',
            'potential_savings': '20-30%'
        })
        
        # RDS optimization
        recommendations.append({
            'service': 'RDS',
            'suggestion': 'Enable serverless for bursty workloads',
            'potential_savings': '40-60%'
        })
        
        # EC2 optimization
        recommendations.append({
            'service': 'EC2',
            'suggestion': 'Use Spot instances for fault-tolerant workloads',
            'potential_savings': '70-90%'
        })
        
        return recommendations
```


## Implementation Guide




## Advanced Patterns




## Context7 Integration

### Related Libraries & Tools
- [AWS CDK](/aws/aws-cdk): Infrastructure as code framework for AWS
- [Terraform](/hashicorp/terraform): Infrastructure as Code to provision cloud resources
- [Kubernetes](/kubernetes/kubernetes): Container orchestration system for automating deployment
- [Pulumi](/pulumi/pulumi): Modern infrastructure as code platform
- [Docker](/docker/docker): Platform for developing, shipping, and running applications

### Official Documentation
- [AWS Documentation](https://docs.aws.amazon.com/)
- [Google Cloud Documentation](https://cloud.google.com/docs)
- [Azure Documentation](https://learn.microsoft.com/azure/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Terraform Documentation](https://www.terraform.io/docs)

### Version-Specific Guides
Latest stable version: AWS CDK 2.223.0, Terraform 1.9.8, Kubernetes 1.34
- [AWS CDK v2 Migration](https://docs.aws.amazon.com/cdk/v2/guide/migrating-v2.html)
- [Terraform 1.9 Upgrade Guide](https://developer.hashicorp.com/terraform/language/upgrade-guides)
- [Kubernetes 1.34 Release Notes](https://kubernetes.io/docs/setup/release/notes/)
- [Pulumi 3.x Migration](https://www.pulumi.com/docs/install/migrating-3.0/)

