# MLOps Enterprise Stack — Official Documentation Reference

**Version**: 4.0 Enterprise | **Updated**: 2025-11-12 | **Stack**: MLflow 3.6.0, DVC 3.x, Ray 2.51.1, Kubeflow 1.10, Seldon Core 2.9.1, Feast 0.56.0, Optuna 4.6.0, Evidently 0.2.2+

---

## Experiment Tracking & Model Registry

### MLflow 3.6.0 — Official Documentation
- **Official Site**: https://mlflow.org/
- **Latest Docs**: https://mlflow.org/docs/latest/
- **GitHub Releases**: https://github.com/mlflow/mlflow/releases
- **Tracking Server Guide**: https://mlflow.org/docs/latest/tracking/
- **Model Registry**: https://mlflow.org/docs/latest/model-registry/
- **MLflow Gateway**: https://mlflow.org/docs/latest/gateway/
- **Python API**: https://mlflow.org/docs/latest/python_api/
- **REST API**: https://mlflow.org/docs/latest/rest-api/
- **Deployment Guide**: https://mlflow.org/docs/latest/deployment/

### Key Resources
- Installation: `pip install mlflow==3.6.0`
- Quick Start: https://mlflow.org/docs/latest/quickstart/
- MLflow Tracking Tutorial: https://mlflow.org/docs/latest/tutorials/
- Production Deployment: https://mlflow.org/docs/latest/deployment/

---

## Data Versioning & Pipelines

### DVC (Data Version Control) — Official Documentation
- **Official Site**: https://dvc.org/
- **Documentation**: https://dvc.org/doc/
- **GitHub**: https://github.com/iterative/dvc
- **Latest Release**: https://github.com/iterative/dvc/releases
- **Installation**: `pip install dvc==3.63.0 dvc-s3`

### DVC Core Features
- **Data Versioning**: https://dvc.org/doc/user-guide/data-and-model-files
- **DVC Pipelines**: https://dvc.org/doc/user-guide/pipelines/
- **Remote Storage**: https://dvc.org/doc/user-guide/remote
- **Experiments**: https://dvc.org/doc/user-guide/experiments/
- **Model Registry**: https://dvc.org/doc/user-guide/model-registry/

### Integration Guides
- **S3 Remote**: https://dvc.org/doc/user-guide/remote/s3
- **Azure Blob**: https://dvc.org/doc/user-guide/remote/azure
- **Google Cloud Storage**: https://dvc.org/doc/user-guide/remote/gs
- **Upgrading to DVC 3.0**: https://dvc.org/doc/user-guide/upgrade

---

## ML Pipelines & Orchestration

### Kubeflow Pipelines 1.10 — Official Documentation
- **Official Site**: https://www.kubeflow.org/
- **Documentation**: https://www.kubeflow.org/docs/
- **GitHub**: https://github.com/kubeflow/pipelines
- **Release Notes**: https://github.com/kubeflow/pipelines/releases
- **KFP Python SDK**: https://kubeflow-pipelines.readthedocs.io/

### KFP v2 Resources
- **KFP v2 Overview**: https://www.kubeflow.org/docs/components/pipelines/
- **KFP v2 SDK**: https://kubeflow-pipelines.readthedocs.io/en/latest/
- **Components Guide**: https://www.kubeflow.org/docs/components/pipelines/user-guides/components/
- **Pipeline Tutorial**: https://www.kubeflow.org/docs/components/pipelines/user-guides/build-pipeline/
- **Deployment**: https://www.kubeflow.org/docs/components/pipelines/user-guides/hosted/

### Integration
- **Installation**: `pip install kfp==2.7.0 google-cloud-aiplatform`
- **Argo Workflows (backend)**: https://argoproj.github.io/workflows/
- **Model Registry**: https://www.kubeflow.org/docs/components/model-registry/

---

## Model Serving

### Ray Serve 2.51.1 — Official Documentation
- **Official Site**: https://www.ray.io/
- **Documentation**: https://docs.ray.io/en/latest/serve/index.html
- **GitHub**: https://github.com/ray-project/ray
- **Latest Release**: https://github.com/ray-project/ray/releases
- **Python API**: https://docs.ray.io/en/latest/serve/api.html

### Ray Serve Features
- **Quickstart**: https://docs.ray.io/en/latest/serve/getting-started.html
- **Deployments**: https://docs.ray.io/en/latest/serve/deployments.html
- **Model Multiplexing**: https://docs.ray.io/en/latest/serve/model-multiplexing.html
- **Scaling**: https://docs.ray.io/en/latest/serve/scaling-and-performance.html
- **Kubernetes**: https://docs.ray.io/en/latest/serve/deploy/index.html

### Installation
- `pip install ray[serve]==2.51.1`
- Cluster setup: https://docs.ray.io/en/latest/cluster/index.html

---

### Seldon Core 2.9.1 — Official Documentation
- **Official Site**: https://www.seldon.io/
- **Documentation**: https://docs.seldon.io/projects/seldon-core/en/latest/
- **GitHub**: https://github.com/SeldonIO/seldon-core
- **Release Notes**: https://github.com/SeldonIO/seldon-core/releases
- **Core 2.x Migration**: https://docs.seldon.io/projects/seldon-core/en/latest/migrating-to-core-v2.html

### Core 2.x Features
- **Getting Started**: https://docs.seldon.io/projects/seldon-core/en/latest/getting-started.html
- **Deployment**: https://docs.seldon.io/projects/seldon-core/en/latest/deploying.html
- **Inference Graph**: https://docs.seldon.io/projects/seldon-core/en/latest/graph/inference-graph.html
- **Canary Deployments**: https://docs.seldon.io/projects/seldon-core/en/latest/deploying/#canary-deployments
- **Model Monitoring**: https://docs.seldon.io/projects/seldon-core/en/latest/monitoring/

### Installation
- Helm: https://docs.seldon.io/projects/seldon-core/en/latest/install/

---

## Feature Store

### Feast 0.56.0 — Official Documentation
- **Official Site**: https://feast.dev/
- **Documentation**: https://docs.feast.dev/
- **GitHub**: https://github.com/feast-dev/feast
- **Latest Release**: https://github.com/feast-dev/feast/releases
- **Python API**: https://api.docs.feast.dev/

### Feast Core Concepts
- **Feature Views**: https://docs.feast.dev/reference/feature-views
- **Entities**: https://docs.feast.dev/reference/entity
- **Data Sources**: https://docs.feast.dev/reference/data-sources
- **On-Demand Features**: https://docs.feast.dev/reference/on-demand-feature-views
- **Feature Store Setup**: https://docs.feast.dev/getting-started/

### Storage & Registries
- **Offline Store**: https://docs.feast.dev/reference/offline-stores/
- **Online Store**: https://docs.feast.dev/reference/online-stores/
- **Feature Registry**: https://docs.feast.dev/reference/registry

### Installation
- `pip install feast==0.56.0 feast[postgres]`

---

## Hyperparameter Optimization

### Optuna 4.6.0 — Official Documentation
- **Official Site**: https://optuna.org/
- **Documentation**: https://optuna.readthedocs.io/
- **GitHub**: https://github.com/optuna/optuna
- **Latest Release**: https://github.com/optuna/optuna/releases
- **PyPI**: https://pypi.org/project/optuna/

### Optuna Core Features
- **Quickstart**: https://optuna.readthedocs.io/en/stable/tutorial/index.html
- **Samplers**: https://optuna.readthedocs.io/en/stable/reference/samplers.html
- **Pruners**: https://optuna.readthedocs.io/en/stable/reference/pruners.html
- **Multi-Objective**: https://optuna.readthedocs.io/en/stable/tutorials/20_recipes/04_multi_objective.html
- **Distributed Optimization**: https://optuna.readthedocs.io/en/stable/tutorials/20_recipes/01_rdb_storage_mode.html

### Integration
- **Ray Tune**: https://optuna.readthedocs.io/en/stable/tutorials/20_recipes/10_ray_integration.html
- **Visualization**: https://optuna.readthedocs.io/en/stable/reference/visualization.html
- **Dashboard**: https://github.com/optuna/optuna-dashboard

### Installation
- `pip install optuna==4.6.0 optuna-dashboard`

---

## Model Monitoring & Observability

### Evidently AI — Official Documentation
- **Official Site**: https://www.evidentlyai.com/
- **Documentation**: https://docs.evidentlyai.com/
- **GitHub**: https://github.com/evidentlyai/evidently
- **Latest Release**: https://github.com/evidentlyai/evidently/releases

### Evidently Core Features
- **Data Drift Detection**: https://docs.evidentlyai.com/metrics/preset_data_drift
- **Model Performance**: https://docs.evidentlyai.com/metrics/model_performance_metrics
- **Report Generation**: https://docs.evidentlyai.com/user-guide/reports/
- **Test Suites**: https://docs.evidentlyai.com/user-guide/test-suites/
- **Dashboard**: https://docs.evidentlyai.com/features/dashboard/

### Installation
- `pip install evidently==0.4.24`

---

### Prometheus 3.7.3 — Official Documentation
- **Official Site**: https://prometheus.io/
- **Documentation**: https://prometheus.io/docs/
- **GitHub**: https://github.com/prometheus/prometheus
- **Download**: https://prometheus.io/download/

### Prometheus Features
- **Configuration**: https://prometheus.io/docs/prometheus/latest/configuration/
- **Querying**: https://prometheus.io/docs/prometheus/latest/querying/
- **Alerting**: https://prometheus.io/docs/prometheus/latest/alerting/overview/
- **Remote Write**: https://prometheus.io/docs/prometheus/latest/storage/#remote-write-tuning

### Client Libraries
- **Python**: https://github.com/prometheus/client_python

---

### Grafana 11.3+ — Official Documentation
- **Official Site**: https://grafana.com/
- **Documentation**: https://grafana.com/docs/grafana/latest/
- **GitHub**: https://github.com/grafana/grafana
- **Dashboard Library**: https://grafana.com/grafana/dashboards/

### Grafana Features
- **Scenes (new)**: https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/manage-dashboards/
- **Queries**: https://grafana.com/docs/grafana/latest/panels-visualizations/query-transform/
- **Alerting**: https://grafana.com/docs/grafana/latest/alerting/alerting-rules/
- **Plugins**: https://grafana.com/grafana/plugins/

### Helm Installation
- `helm install grafana grafana/grafana`

---

## CI/CD for ML

### GitHub Actions — Official Documentation
- **Official Docs**: https://docs.github.com/en/actions
- **Workflows**: https://docs.github.com/en/actions/using-workflows
- **Docker**: https://docs.github.com/en/actions/publishing-packages/publishing-docker-images

### Useful Actions for ML
- **Checkout**: https://github.com/actions/checkout
- **Setup Python**: https://github.com/actions/setup-python
- **AWS Credentials**: https://github.com/aws-actions/configure-aws-credentials
- **Kubernetes**: https://github.com/steebchen/kubectl

### ML CI/CD Patterns
- Model training automation
- Data quality checks
- Model validation & testing
- Continuous deployment to production

---

## Container & Kubernetes

### Docker — Official Documentation
- **Official Site**: https://www.docker.com/
- **Documentation**: https://docs.docker.com/
- **Image Registry**: https://hub.docker.com/

### Kubernetes — Official Documentation
- **Official Site**: https://kubernetes.io/
- **Documentation**: https://kubernetes.io/docs/
- **kubectl**: https://kubernetes.io/docs/reference/kubectl/

### Helm — Official Documentation
- **Official Site**: https://helm.sh/
- **Documentation**: https://helm.sh/docs/
- **Chart Hub**: https://artifacthub.io/

---

## Database & Storage

### PostgreSQL (for MLflow Backend)
- **Official Site**: https://www.postgresql.org/
- **Documentation**: https://www.postgresql.org/docs/

### AWS S3 (for Artifact Storage)
- **Official Site**: https://aws.amazon.com/s3/
- **Documentation**: https://docs.aws.amazon.com/s3/

### Google Cloud Storage
- **Official Site**: https://cloud.google.com/storage
- **Documentation**: https://cloud.google.com/storage/docs

---

## Framework Integration

### PyTorch
- **Official Site**: https://pytorch.org/
- **Documentation**: https://pytorch.org/docs/

### TensorFlow / Keras
- **Official Site**: https://www.tensorflow.org/
- **Documentation**: https://www.tensorflow.org/api_docs

### Scikit-Learn
- **Official Site**: https://scikit-learn.org/
- **Documentation**: https://scikit-learn.org/stable/

### Transformers (Hugging Face)
- **Official Site**: https://huggingface.co/
- **Documentation**: https://huggingface.co/docs/transformers/

---

## Best Practices & Patterns

### MLOps Handbook
- https://ml-ops.systems/ (Community-driven MLOps best practices)

### Google Cloud AI/ML Best Practices
- https://cloud.google.com/architecture/mlops-continuous-delivery-and-automation-pipelines-in-machine-learning

### AWS MLOps Best Practices
- https://aws.amazon.com/blogs/machine-learning/

### Azure ML Best Practices
- https://learn.microsoft.com/en-us/azure/machine-learning/

---

## Community & Support

### GitHub Issues & Discussions
- MLflow: https://github.com/mlflow/mlflow/issues
- DVC: https://github.com/iterative/dvc/discussions
- Ray: https://github.com/ray-project/ray/issues
- Kubeflow: https://github.com/kubeflow/kubeflow/issues
- Seldon: https://github.com/SeldonIO/seldon-core/discussions
- Feast: https://github.com/feast-dev/feast/discussions
- Optuna: https://github.com/optuna/optuna/discussions
- Evidently: https://github.com/evidentlyai/evidently/issues

### Stack Overflow Tags
- `mlflow`, `dvc`, `ray-serve`, `kubeflow`, `seldon-core`, `feast`, `optuna`, `evidently`

### Community Slack/Discord
- Many projects offer community channels for support and discussions

---

## Version History & Changelog

| Component | Version | Release Date | Status |
| --------- | ------- | ------------ | ------ |
| MLflow | 3.6.0 | 2025-11-07 | Latest Stable |
| DVC | 3.63.0 | 2025-11-01 | Latest Stable |
| Ray | 2.51.1 | 2025-11-10 | Latest Stable |
| Kubeflow | 1.10 | 2025-04-07 | Latest Stable |
| Seldon Core | 2.9.1 | 2025-10-15 | Latest Stable |
| Feast | 0.56.0 | 2025-10-20 | Latest Stable |
| Optuna | 4.6.0 | 2025-11-10 | Latest Stable |
| Evidently | 0.4.24+ | 2025-11-05 | Latest Stable |
| Prometheus | 3.7.3 | 2025-10-29 | Latest Stable |
| Grafana | 11.3+ | 2025-11-01 | Latest Stable |

---

## Last Updated

2025-11-12 — All links verified, documentation current for November 2025 stable releases.

