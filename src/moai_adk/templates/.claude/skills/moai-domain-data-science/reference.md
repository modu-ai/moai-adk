# Data Science Domain - Official Documentation References

_Last updated: 2025-11-13_

## Core Libraries (November 2025 Stable)

### Deep Learning Frameworks

| Library | Version | Documentation | Key Usage |
| --- | --- | --- | --- |
| **TensorFlow** | 2.20.0 | https://www.tensorflow.org/api_docs | CNN, RNN, Transformers |
| **PyTorch** | 2.9.0 | https://pytorch.org/docs/stable/index.html | Dynamic computation, RNNs |
| **PyTorch Lightning** | 2.x | https://lightning.ai/docs/pytorch/stable/ | Simplified training loops |

### Machine Learning & Statistics

| Library | Version | Documentation | Key Usage |
| --- | --- | --- | --- |
| **scikit-learn** | 1.7.2 | https://scikit-learn.org/stable/ | Classical ML, pipelines, evaluation |
| **XGBoost** | Latest | https://xgboost.readthedocs.io/en/stable/ | Gradient boosting |
| **LightGBM** | Latest | https://lightgbm.readthedocs.io/en/latest/ | Fast gradient boosting |
| **scipy** | 1.x | https://docs.scipy.org/doc/scipy/ | Statistical tests, optimization |
| **statsmodels** | 0.14.x | https://www.statsmodels.org/stable/ | Time series, econometrics |

### Data Processing

| Library | Version | Documentation | Key Usage |
| --- | --- | --- | --- |
| **pandas** | 2.3.3 | https://pandas.pydata.org/docs/ | DataFrames, data manipulation |
| **polars** | 1.x | https://docs.pola.rs/ | High-performance large datasets |
| **NumPy** | 2.x | https://numpy.org/doc/stable/ | Numerical computing |
| **pandas-profiling** | Latest | https://pandas-profiling.ydata.ai/docs/ | EDA reports |

### Hyperparameter Optimization

| Library | Version | Documentation | Key Usage |
| --- | --- | --- | --- |
| **Optuna** | 3.x | https://optuna.readthedocs.io/en/stable/ | Hyperparameter search |
| **Hyperopt** | Latest | https://hyperopt.github.io/hyperopt/ | Bayesian optimization |
| **Ray Tune** | Latest | https://docs.ray.io/en/latest/tune/ | Distributed tuning |

### Visualization

| Library | Version | Documentation | Key Usage |
| --- | --- | --- | --- |
| **matplotlib** | 3.10.x | https://matplotlib.org/stable/contents.html | Static plots |
| **seaborn** | 0.13.x | https://seaborn.pydata.org/ | Statistical visualization |
| **plotly** | Latest | https://plotly.com/python/ | Interactive dashboards |
| **bokeh** | Latest | https://docs.bokeh.org/ | Interactive visualizations |

### Time Series

| Library | Version | Documentation | Key Usage |
| --- | --- | --- | --- |
| **Prophet** | Latest | https://facebook.github.io/prophet/ | Time series forecasting |
| **statsmodels TSA** | 0.14.x | https://www.statsmodels.org/stable/tsa.html | ARIMA, VAR, etc. |
| **pytorch-forecasting** | Latest | https://pytorch-forecasting.readthedocs.io/ | Neural network forecasting |

### MLOps & Experiment Tracking

| Library | Version | Documentation | Key Usage |
| --- | --- | --- | --- |
| **MLflow** | Latest | https://mlflow.org/docs/latest/index.html | Experiment tracking |
| **Weights & Biases** | Latest | https://docs.wandb.ai/ | ML monitoring |
| **DVC** | Latest | https://dvc.org/doc | Data & model versioning |

---

## Key Documentation Links by Topic

### Neural Networks & Deep Learning

- **PyTorch Neural Networks**: https://pytorch.org/docs/stable/nn.html
- **PyTorch CNN**: https://pytorch.org/vision/stable/models.html
- **PyTorch RNN/LSTM**: https://pytorch.org/docs/stable/generated/torch.nn.LSTM.html
- **TensorFlow Keras API**: https://www.tensorflow.org/api_docs/python/tf/keras
- **TensorFlow CNN Models**: https://www.tensorflow.org/api_docs/python/tf/keras/applications

### Model Evaluation

- **scikit-learn Metrics**: https://scikit-learn.org/stable/modules/model_evaluation.html
- **scikit-learn Cross-Validation**: https://scikit-learn.org/stable/modules/cross_validation.html
- **scikit-learn Grid Search**: https://scikit-learn.org/stable/modules/grid_search.html

### Feature Engineering

- **scikit-learn Preprocessing**: https://scikit-learn.org/stable/modules/preprocessing.html
- **scikit-learn Feature Selection**: https://scikit-learn.org/stable/modules/feature_selection.html
- **pandas String Operations**: https://pandas.pydata.org/docs/user_guide/text.html

### Data Processing

- **pandas User Guide**: https://pandas.pydata.org/docs/user_guide/index.html
- **pandas API Reference**: https://pandas.pydata.org/docs/reference/index.html
- **polars User Guide**: https://docs.pola.rs/user-guide/
- **NumPy Documentation**: https://numpy.org/doc/stable/

### Statistical Analysis

- **scipy.stats**: https://docs.scipy.org/doc/scipy/reference/stats.html
- **statsmodels TSA**: https://www.statsmodels.org/stable/tsa.html
- **statsmodels Regression**: https://www.statsmodels.org/stable/regression.html

### Visualization

- **matplotlib Gallery**: https://matplotlib.org/stable/gallery/
- **seaborn Gallery**: https://seaborn.pydata.org/examples.html
- **plotly Python**: https://plotly.com/python/
- **plotly Subplots**: https://plotly.com/python/subplots/

### Time Series Forecasting

- **statsmodels ARIMA**: https://www.statsmodels.org/stable/generated/statsmodels.tsa.arima.model.ARIMA.html
- **Prophet Documentation**: https://facebook.github.io/prophet/docs/quick_start.html
- **PyTorch Forecasting**: https://pytorch-forecasting.readthedocs.io/en/stable/

### Hyperparameter Optimization

- **Optuna Tutorial**: https://optuna.readthedocs.io/en/stable/tutorial/index.html
- **Optuna Samplers**: https://optuna.readthedocs.io/en/stable/reference/samplers.html
- **Optuna Pruners**: https://optuna.readthedocs.io/en/stable/reference/pruners.html
- **Ray Tune Tutorials**: https://docs.ray.io/en/latest/tune/tutorials/

### Distributed Training

- **PyTorch Distributed**: https://pytorch.org/docs/stable/distributed.html
- **PyTorch Lightning DDP**: https://lightning.ai/docs/pytorch/stable/accelerators/gpu_advanced.html
- **TensorFlow Distributed**: https://www.tensorflow.org/api_docs/python/tf/distribute

---

## Installation Commands

```bash
# Core ML libraries
pip install tensorflow==2.20.0
pip install torch==2.9.0 torchvision torchaudio
pip install pytorch-lightning==2.2.0
pip install scikit-learn==1.7.2
pip install pandas==2.3.3
pip install numpy==2.0.0
pip install polars==1.0.0

# Advanced tools
pip install optuna==3.0.0
pip install xgboost
pip install lightgbm
pip install scipy
pip install statsmodels

# Visualization
pip install matplotlib==3.10.0
pip install seaborn==0.13.0
pip install plotly
pip install bokeh

# Time series
pip install prophet
pip install pytorch-forecasting

# MLOps
pip install mlflow
pip install wandb
pip install dvc
```

---

## Quick Reference by Use Case

### Binary Classification
- Reference: https://scikit-learn.org/stable/modules/model_evaluation.html
- Metrics: accuracy, precision, recall, F1, ROC-AUC
- Tools: scikit-learn, XGBoost, PyTorch

### Time Series Forecasting
- ARIMA: https://www.statsmodels.org/stable/tsa.html
- Prophet: https://facebook.github.io/prophet/
- LSTM: https://pytorch.org/docs/stable/generated/torch.nn.LSTM.html

### Image Classification
- CNN Tutorial: https://pytorch.org/tutorials/beginner/basics/cnn_cifar10_tutorial.html
- Transfer Learning: https://pytorch.org/vision/stable/models.html
- TensorFlow: https://www.tensorflow.org/api_docs/python/tf/keras/applications

### Hyperparameter Tuning
- scikit-learn GridSearchCV: https://scikit-learn.org/stable/modules/grid_search.html
- Optuna: https://optuna.readthedocs.io/en/stable/
- Ray Tune: https://docs.ray.io/en/latest/tune/

### Feature Engineering
- scikit-learn Preprocessing: https://scikit-learn.org/stable/modules/preprocessing.html
- pandas Feature Creation: https://pandas.pydata.org/docs/user_guide/basics.html
- polars Expressions: https://docs.pola.rs/user-guide/expressions/

---

## Version Compatibility

| Python | TensorFlow | PyTorch | scikit-learn | pandas |
| --- | --- | --- | --- | --- |
| 3.10 | 2.20.0 | 2.9.0 | 1.7.2 | 2.3.3 |
| 3.11 | 2.20.0 | 2.9.0 | 1.7.2 | 2.3.3 |
| 3.12 | 2.20.0 | 2.9.0 | 1.7.2 | 2.3.3 |
| 3.13 | 2.20.0 | 2.9.0 | 1.7.2 | 2.3.3 |
| 3.14 | Latest | Latest | Latest | 2.3.3+ |

---

**Last Updated**: 2025-11-13  
**Coverage**: 20+ libraries with official docs  
**Tested**: All links verified for November 2025 releases
