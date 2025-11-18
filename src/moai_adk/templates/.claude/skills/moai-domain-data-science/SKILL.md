---
name: moai-domain-data-science
version: 4.0.0
updated: '2025-11-18'
status: stable
description: "Production-grade data science specialist with TensorFlow 2.20.0, PyTorch\
  \ 2.9.0, \nScikit-learn 1.7.2 expertise. Master data processing, ML pipeline development,\
  \ \nmodel deployment, and statistical analysis. Build end-to-end data science \n\
  solutions with comprehensive experimentation and visualization.\n"
allowed-tools:
- Read
- Write
- Edit
- Bash
- Glob
- WebFetch
- WebSearch
tags:
- data-science
- tensorflow
- pytorch
- scikit-learn
- pandas
- numpy
- visualization
- statistics
stability: stable
---


# Data Science & Analytics

## Level 1: Quick Reference

### Core Capabilities
- **Data Processing**: Pandas 2.2.0, NumPy 1.26.0, Dask 2024.1.0
- **Machine Learning**: TensorFlow 2.20.0, PyTorch 2.9.0, Scikit-learn 1.7.2
- **Visualization**: Matplotlib 3.8.0, Seaborn 0.13.0, Plotly 5.17.0
- **Statistics**: SciPy 1.12.0, Statsmodels 0.14.0, Pingouin 0.8.0
- **Big Data**: Spark 3.5.0, Polars 0.20.0, Apache Arrow 14.0.0
- **Experimentation**: MLflow 2.9.0, Weights & Biases, Neptune

### Quick Setup Examples

```python
# Data science workflow starter
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import classification_report, confusion_matrix

# Load and explore data
df = pd.read_csv('data.csv')
print(f"Dataset shape: {df.shape}")
print(f"Missing values: {df.isnull().sum().sum()}")

# Basic preprocessing
X = df.drop('target', axis=1)
y = df['target']

# Split data
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# Train model
model = RandomForestClassifier(n_estimators=100, random_state=42)
model.fit(X_train, y_train)

# Evaluate
y_pred = model.predict(X_test)
print(classification_report(y_test, y_pred))
```

```python
# Modern data visualization with Plotly
import plotly.express as px
import plotly.graph_objects as go
from plotly.subplots import make_subplots

# Interactive scatter plot
fig = px.scatter(df, x='feature1', y='feature2', color='target',
                 size='feature3', hover_data=['feature4'])
fig.show()

# Multi-subplot dashboard
fig = make_subplots(
    rows=2, cols=2,
    subplot_titles=('Distribution', 'Correlation', 'Time Series', 'Feature Importance')
)
# Add traces...
fig.show()
```

## Level 2: Practical Implementation

### Data Processing & Analysis Pipeline

#### 1. Advanced Data Preprocessing

```python
import pandas as pd
import numpy as np
from sklearn.preprocessing import StandardScaler, LabelEncoder, RobustScaler
from sklearn.impute import SimpleImputer, KNNImputer
from sklearn.feature_selection import SelectKBest, f_classif, RFE
from sklearn.compose import ColumnTransformer
from sklearn.pipeline import Pipeline
from scipy import stats
import warnings
warnings.filterwarnings('ignore')

class DataPreprocessor:
    def __init__(self):
        self.numeric_features = []
        self.categorical_features = []
        self.preprocessor = None
        self.feature_selector = None
        self.outlier_detector = None
        
    def analyze_data(self, df: pd.DataFrame) -> dict:
        """Comprehensive data analysis"""
        analysis = {
            'shape': df.shape,
            'dtypes': df.dtypes.to_dict(),
            'missing_values': df.isnull().sum().to_dict(),
            'missing_percentage': (df.isnull().sum() / len(df) * 100).to_dict(),
            'duplicates': df.duplicated().sum(),
            'numeric_summary': df.describe().to_dict(),
            'categorical_summary': {},
            'correlation_matrix': None
        }
        
        # Categorical features analysis
        for col in df.select_dtypes(include=['object', 'category']).columns:
            analysis['categorical_summary'][col] = {
                'unique_count': df[col].nunique(),
                'unique_values': df[col].unique().tolist()[:10],  # First 10 values
                'value_counts': df[col].value_counts().head().to_dict()
            }
        
        # Correlation matrix for numeric features
        numeric_df = df.select_dtypes(include=[np.number])
        if len(numeric_df.columns) > 1:
            analysis['correlation_matrix'] = numeric_df.corr().to_dict()
        
        return analysis
    
    def detect_outliers(self, df: pd.DataFrame, method: str = 'iqr') -> pd.DataFrame:
        """Detect outliers using various methods"""
        numeric_df = df.select_dtypes(include=[np.number])
        outliers_info = {}
        
        for col in numeric_df.columns:
            if method == 'iqr':
                Q1 = df[col].quantile(0.25)
                Q3 = df[col].quantile(0.75)
                IQR = Q3 - Q1
                lower_bound = Q1 - 1.5 * IQR
                upper_bound = Q3 + 1.5 * IQR
                outliers = df[(df[col] < lower_bound) | (df[col] > upper_bound)]
                
            elif method == 'zscore':
                z_scores = np.abs(stats.zscore(df[col].dropna()))
                outliers = df[z_scores > 3]
                
            elif method == 'isolation_forest':
                from sklearn.ensemble import IsolationForest
                iso_forest = IsolationForest(contamination=0.1, random_state=42)
                outlier_labels = iso_forest.fit_predict(df[[col]].dropna())
                outliers = df[outlier_labels == -1]
            
            outliers_info[col] = {
                'count': len(outliers),
                'percentage': len(outliers) / len(df) * 100,
                'indices': outliers.index.tolist()
            }
        
        return outliers_info
    
    def build_preprocessing_pipeline(self, X: pd.DataFrame, 
                                   handle_outliers: bool = True,
                                   feature_selection: bool = True,
                                   selection_method: str = 'k_best') -> Pipeline:
        """Build comprehensive preprocessing pipeline"""
        
        # Identify feature types
        self.numeric_features = X.select_dtypes(include=[np.number]).columns.tolist()
        self.categorical_features = X.select_dtypes(include=['object', 'category']).columns.tolist()
        
        # Numeric preprocessing
        numeric_transformer = Pipeline(steps=[
            ('imputer', SimpleImputer(strategy='median')),
            ('scaler', RobustScaler())  # More robust to outliers than StandardScaler
        ])
        
        # Categorical preprocessing
        categorical_transformer = Pipeline(steps=[
            ('imputer', SimpleImputer(strategy='most_frequent')),
            ('encoder', LabelEncoder()) if len(self.categorical_features) == 1 else 
            ('onehot', SimpleImputer(strategy='most_frequent'))
        ])
        
        # Column transformer
        preprocessor = ColumnTransformer(
            transformers=[
                ('num', numeric_transformer, self.numeric_features),
                ('cat', categorical_transformer, self.categorical_features)
            ])
        
        pipeline_steps = [('preprocessor', preprocessor)]
        
        # Feature selection
        if feature_selection and selection_method == 'k_best':
            selector = SelectKBest(score_func=f_classif, k=10)
            pipeline_steps.append(('feature_selection', selector))
        elif feature_selection and selection_method == 'rfe':
            from sklearn.ensemble import RandomForestClassifier
            estimator = RandomForestClassifier(n_estimators=100, random_state=42)
            selector = RFE(estimator=estimator, n_features_to_select=10)
            pipeline_steps.append(('feature_selection', selector))
        
        self.preprocessor = Pipeline(pipeline_steps)
        return self.preprocessor
    
    def transform_data(self, X: pd.DataFrame, y: pd.Series = None) -> np.ndarray:
        """Transform data using preprocessing pipeline"""
        
        if self.preprocessor is None:
            self.build_preprocessing_pipeline(X)
        
        if y is not None:
            return self.preprocessor.fit_transform(X, y)
        else:
            return self.preprocessor.transform(X)
    
    def get_feature_importance(self, X: pd.DataFrame, y: pd.Series) -> dict:
        """Get feature importance using various methods"""
        
        # Prepare data
        X_processed = self.transform_data(X, y)
        
        feature_importance = {}
        
        # Random Forest importance
        rf = RandomForestClassifier(n_estimators=100, random_state=42)
        rf.fit(X_processed, y)
        
        # Get feature names (simplified for this example)
        feature_names = [f"feature_{i}" for i in range(X_processed.shape[1])]
        
        feature_importance['random_forest'] = dict(zip(feature_names, rf.feature_importances_))
        
        # Mutual information
        from sklearn.feature_selection import mutual_info_classif
        mi_scores = mutual_info_classif(X_processed, y)
        feature_importance['mutual_info'] = dict(zip(feature_names, mi_scores))
        
        # ANOVA F-test
        from sklearn.feature_selection import f_classif
        f_scores, _ = f_classif(X_processed, y)
        feature_importance['anova_f'] = dict(zip(feature_names, f_scores))
        
        return feature_importance
    
    def create_features(self, df: pd.DataFrame) -> pd.DataFrame:
        """Create engineered features"""
        
        df_engineered = df.copy()
        numeric_cols = df.select_dtypes(include=[np.number]).columns
        
        # Polynomial features for top correlated columns
        if len(numeric_cols) >= 2:
            # Interaction features
            for i, col1 in enumerate(numeric_cols[:3]):  # Limit to first 3 columns
                for col2 in numeric_cols[i+1:4]:  # Limit interactions
                    df_engineered[f'{col1}_x_{col2}'] = df[col1] * df[col2]
                    df_engineered[f'{col1}_div_{col2}'] = df[col1] / (df[col2] + 1e-6)
        
        # Polynomial features
        for col in numeric_cols[:3]:  # Limit to first 3 columns
            df_engineered[f'{col}_squared'] = df[col] ** 2
            df_engineered[f'{col}_cubed'] = df[col] ** 3
            df_engineered[f'{col}_log'] = np.log1p(np.abs(df[col]))
            df_engineered[f'{col}_sqrt'] = np.sqrt(np.abs(df[col]))
        
        # Statistical features
        if len(numeric_cols) >= 3:
            df_engineered['mean_of_numeric'] = df[numeric_cols].mean(axis=1)
            df_engineered['std_of_numeric'] = df[numeric_cols].std(axis=1)
            df_engineered['min_of_numeric'] = df[numeric_cols].min(axis=1)
            df_engineered['max_of_numeric'] = df[numeric_cols].max(axis=1)
            df_engineered['sum_of_numeric'] = df[numeric_cols].sum(axis=1)
        
        # Date/time features (if date columns exist)
        date_cols = df.select_dtypes(include=['datetime64']).columns
        for col in date_cols:
            df_engineered[f'{col}_year'] = df[col].dt.year
            df_engineered[f'{col}_month'] = df[col].dt.month
            df_engineered[f'{col}_day'] = df[col].dt.day
            df_engineered[f'{col}_dayofweek'] = df[col].dt.dayofweek
            df_engineered[f'{col}_quarter'] = df[col].dt.quarter
        
        return df_engineered

# Usage example
# preprocessor = DataPreprocessor()
# analysis = preprocessor.analyze_data(df)
# outliers = preprocessor.detect_outliers(df, method='iqr')
# pipeline = preprocessor.build_preprocessing_pipeline(X)
# X_processed = preprocessor.transform_data(X, y)
# feature_importance = preprocessor.get_feature_importance(X, y)
```

#### 2. Statistical Analysis & Hypothesis Testing

```python
import scipy.stats as stats
import statsmodels.api as sm
from statsmodels.stats.power import TTestPower, GofChisquarePower
from statsmodels.stats.multicomp import pairwise_tukeyhsd
import pingouin as pg
import numpy as np
import pandas as pd

class StatisticalAnalyzer:
    def __init__(self):
        self.results = {}
    
    def descriptive_statistics(self, data: pd.DataFrame, 
                             columns: list = None) -> dict:
        """Comprehensive descriptive statistics"""
        
        if columns is None:
            columns = data.select_dtypes(include=[np.number]).columns.tolist()
        
        stats_dict = {}
        
        for col in columns:
            series = data[col].dropna()
            
            stats_dict[col] = {
                'count': len(series),
                'mean': series.mean(),
                'median': series.median(),
                'std': series.std(),
                'var': series.var(),
                'min': series.min(),
                'max': series.max(),
                'range': series.max() - series.min(),
                'q25': series.quantile(0.25),
                'q75': series.quantile(0.75),
                'iqr': series.quantile(0.75) - series.quantile(0.25),
                'skewness': stats.skew(series),
                'kurtosis': stats.kurtosis(series),
                'cv': series.std() / series.mean() if series.mean() != 0 else np.nan
            }
            
            # Confidence intervals
            confidence_level = 0.95
            degrees_freedom = len(series) - 1
            sample_mean = series.mean()
            sample_std = series.std()
            standard_error = sample_std / np.sqrt(len(series))
            
            t_critical = stats.t.ppf((1 + confidence_level) / 2, degrees_freedom)
            margin_error = t_critical * standard_error
            
            stats_dict[col]['confidence_interval'] = {
                'lower': sample_mean - margin_error,
                'upper': sample_mean + margin_error,
                'level': confidence_level
            }
        
        return stats_dict
    
    def normality_tests(self, data: pd.Series, alpha: float = 0.05) -> dict:
        """Test for normality using multiple methods"""
        
        data_clean = data.dropna()
        
        # Shapiro-Wilk test (for n < 5000)
        shapiro_stat, shapiro_p = stats.shapiro(data_clean[:5000])
        
        # Kolmogorov-Smirnov test
        ks_stat, ks_p = stats.kstest(data_clean, 'norm', 
                                     args=(data_clean.mean(), data_clean.std()))
        
        # Anderson-Darling test
        ad_stat, ad_critical_values, ad_significance_levels = stats.anderson(data_clean, 'norm')
        
        # D'Agostino's K2 test
        dagostino_stat, dagostino_p = stats.normaltest(data_clean)
        
        results = {
            'shapiro_wilk': {
                'statistic': shapiro_stat,
                'p_value': shapiro_p,
                'is_normal': shapiro_p > alpha,
                'alpha': alpha
            },
            'kolmogorov_smirnov': {
                'statistic': ks_stat,
                'p_value': ks_p,
                'is_normal': ks_p > alpha,
                'alpha': alpha
            },
            'anderson_darling': {
                'statistic': ad_stat,
                'critical_values': dict(zip(ad_significance_levels, ad_critical_values)),
                'is_normal': ad_stat < ad_critical_values[2]  # 5% significance
            },
            'dagostino_k2': {
                'statistic': dagostino_stat,
                'p_value': dagostino_p,
                'is_normal': dagostino_p > alpha,
                'alpha': alpha
            }
        }
        
        return results
    
    def hypothesis_testing(self, data1: pd.Series, data2: pd.Series = None,
                          test_type: str = 'auto', alpha: float = 0.05,
                          alternative: str = 'two-sided') -> dict:
        """Comprehensive hypothesis testing"""
        
        if data2 is None:
            # One-sample tests
            if test_type == 'auto':
                # Determine appropriate test based on normality
                normality = self.normality_tests(data1, alpha)
                
                # Use t-test if data appears normal, otherwise Wilcoxon
                if normality['shapiro_wilk']['is_normal']:
                    return self._one_sample_t_test(data1, alpha, alternative)
                else:
                    return self._one_sample_wilcoxon_test(data1, alpha, alternative)
        
        else:
            # Two-sample tests
            if test_type == 'auto':
                # Check normality and equal variance assumptions
                normality1 = self.normality_tests(data1, alpha)
                normality2 = self.normality_tests(data2, alpha)
                
                # Levene's test for equal variances
                levene_stat, levene_p = stats.levene(data1.dropna(), data2.dropna())
                
                if (normality1['shapiro_wilk']['is_normal'] and 
                    normality2['shapiro_wilk']['is_normal'] and 
                    levene_p > alpha):
                    # Use t-test if assumptions are met
                    return self._two_sample_t_test(data1, data2, alpha, alternative)
                else:
                    # Use non-parametric test
                    return self._mann_whitney_test(data1, data2, alpha, alternative)
        
        return {'error': 'Invalid test configuration'}
    
    def _one_sample_t_test(self, data: pd.Series, alpha: float, alternative: str) -> dict:
        """One-sample t-test"""
        
        t_stat, p_value = stats.ttest_1samp(data.dropna(), 0, alternative=alternative)
        
        degrees_freedom = len(data.dropna()) - 1
        confidence_level = 1 - alpha
        
        # Effect size (Cohen's d)
        cohens_d = data.mean() / data.std()
        
        return {
            'test': 'One-sample t-test',
            'statistic': t_stat,
            'p_value': p_value,
            'degrees_freedom': degrees_freedom,
            'alpha': alpha,
            'is_significant': p_value < alpha,
            'alternative': alternative,
            'effect_size': cohens_d,
            'effect_size_interpretation': self._interpret_cohens_d(cohens_d)
        }
    
    def _one_sample_wilcoxon_test(self, data: pd.Series, alpha: float, alternative: str) -> dict:
        """One-sample Wilcoxon signed-rank test"""
        
        wilcoxon_stat, p_value = stats.wilcoxon(data.dropna(), alternative=alternative)
        
        return {
            'test': 'One-sample Wilcoxon signed-rank test',
            'statistic': wilcoxon_stat,
            'p_value': p_value,
            'alpha': alpha,
            'is_significant': p_value < alpha,
            'alternative': alternative,
            'note': 'Non-parametric test - does not assume normality'
        }
    
    def _two_sample_t_test(self, data1: pd.Series, data2: pd.Series, 
                          alpha: float, alternative: str) -> dict:
        """Two-sample t-test"""
        
        t_stat, p_value = stats.ttest_ind(data1.dropna(), data2.dropna(), alternative=alternative)
        
        # Effect size (Cohen's d)
        n1, n2 = len(data1.dropna()), len(data2.dropna())
        pooled_std = np.sqrt(((n1-1)*data1.var() + (n2-1)*data2.var()) / (n1+n2-2))
        cohens_d = (data1.mean() - data2.mean()) / pooled_std
        
        return {
            'test': 'Two-sample t-test',
            'statistic': t_stat,
            'p_value': p_value,
            'alpha': alpha,
            'is_significant': p_value < alpha,
            'alternative': alternative,
            'effect_size': cohens_d,
            'effect_size_interpretation': self._interpret_cohens_d(cohens_d),
            'sample_sizes': {'sample1': n1, 'sample2': n2}
        }
    
    def _mann_whitney_test(self, data1: pd.Series, data2: pd.Series, 
                          alpha: float, alternative: str) -> dict:
        """Mann-Whitney U test"""
        
        u_stat, p_value = stats.mannwhitneyu(data1.dropna(), data2.dropna(), 
                                          alternative=alternative)
        
        return {
            'test': 'Mann-Whitney U test',
            'statistic': u_stat,
            'p_value': p_value,
            'alpha': alpha,
            'is_significant': p_value < alpha,
            'alternative': alternative,
            'note': 'Non-parametric test - does not assume normality'
        }
    
    def _interpret_cohens_d(self, d: float) -> str:
        """Interpret Cohen's d effect size"""
        
        abs_d = abs(d)
        if abs_d < 0.2:
            return 'negligible'
        elif abs_d < 0.5:
            return 'small'
        elif abs_d < 0.8:
            return 'medium'
        else:
            return 'large'
    
    def correlation_analysis(self, data: pd.DataFrame, 
                           method: str = 'pearson') -> dict:
        """Comprehensive correlation analysis"""
        
        numeric_data = data.select_dtypes(include=[np.number])
        
        if method == 'pearson':
            corr_matrix, p_values = self._pearson_correlation(numeric_data)
        elif method == 'spearman':
            corr_matrix, p_values = self._spearman_correlation(numeric_data)
        elif method == 'kendall':
            corr_matrix, p_values = self._kendall_correlation(numeric_data)
        
        # Find significant correlations
        significant_correlations = []
        for i, col1 in enumerate(numeric_data.columns):
            for j, col2 in enumerate(numeric_data.columns):
                if i < j:  # Avoid duplicates
                    corr_val = corr_matrix.iloc[i, j]
                    p_val = p_values.iloc[i, j]
                    
                    if abs(corr_val) > 0.3 and p_val < 0.05:  # Thresholds
                        significant_correlations.append({
                            'variable1': col1,
                            'variable2': col2,
                            'correlation': corr_val,
                            'p_value': p_val,
                            'strength': self._interpret_correlation(corr_val)
                        })
        
        return {
            'correlation_matrix': corr_matrix.to_dict(),
            'p_value_matrix': p_values.to_dict(),
            'significant_correlations': significant_correlations,
            'method': method
        }
    
    def _pearson_correlation(self, data: pd.DataFrame) -> tuple:
        """Pearson correlation with p-values"""
        
        n = len(data)
        corr_matrix = data.corr(method='pearson')
        
        # Calculate p-values
        p_values = pd.DataFrame(index=data.columns, columns=data.columns)
        
        for i, col1 in enumerate(data.columns):
            for j, col2 in enumerate(data.columns):
                if i <= j:
                    corr, p_val = stats.pearsonr(data[col1].dropna(), data[col2].dropna())
                    p_values.iloc[i, j] = p_val
                    p_values.iloc[j, i] = p_val
                else:
                    p_values.iloc[i, j] = p_values.iloc[j, i]
        
        return corr_matrix, p_values
    
    def _interpret_correlation(self, r: float) -> str:
        """Interpret correlation coefficient magnitude"""
        
        abs_r = abs(r)
        if abs_r < 0.1:
            return 'negligible'
        elif abs_r < 0.3:
            return 'weak'
        elif abs_r < 0.5:
            return 'moderate'
        elif abs_r < 0.7:
            return 'strong'
        else:
            return 'very strong'
    
    def power_analysis(self, effect_size: float, alpha: float = 0.05, 
                      power: float = 0.8, sample_size: int = None,
                      test_type: str = 't-test') -> dict:
        """Statistical power analysis"""
        
        if test_type == 't-test':
            power_analysis = TTestPower()
            
            if sample_size is None:
                # Calculate required sample size
                required_n = power_analysis.solve_power(
                    effect_size=effect_size,
                    alpha=alpha,
                    power=power,
                    alternative='two-sided'
                )
                return {
                    'test_type': test_type,
                    'required_sample_size': required_n,
                    'effect_size': effect_size,
                    'alpha': alpha,
                    'desired_power': power
                }
            else:
                # Calculate achieved power
                achieved_power = power_analysis.power(
                    effect_size=effect_size,
                    nobs=sample_size,
                    alpha=alpha,
                    alternative='two-sided'
                )
                return {
                    'test_type': test_type,
                    'sample_size': sample_size,
                    'achieved_power': achieved_power,
                    'effect_size': effect_size,
                    'alpha': alpha
                }
        
        return {'error': 'Unsupported test type for power analysis'}

# Usage example
# analyzer = StatisticalAnalyzer()
# desc_stats = analyzer.descriptive_statistics(df)
# normality = analyzer.normality_tests(df['column'])
# hypothesis_result = analyzer.hypothesis_testing(df['group1'], df['group2'])
# corr_analysis = analyzer.correlation_analysis(df)
# power_result = analyzer.power_analysis(effect_size=0.5, sample_size=100)
```

#### 3. Machine Learning Model Development

```python
from sklearn.model_selection import train_test_split, cross_val_score, GridSearchCV
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.svm import SVC
from sklearn.metrics import classification_report, confusion_matrix, roc_auc_score
from sklearn.preprocessing import StandardScaler
import xgboost as xgb
import lightgbm as lgb
import optuna

class MLModelBuilder:
    def __init__(self):
        self.models = {}
        self.best_model = None
        self.feature_importance = {}
        self.evaluation_results = {}
    
    def prepare_data(self, X: pd.DataFrame, y: pd.Series, 
                    test_size: float = 0.2, random_state: int = 42) -> tuple:
        """Prepare data for machine learning"""
        
        # Handle missing values
        X_clean = X.fillna(X.median(numeric_only=True))
        for col in X_clean.select_dtypes(include=['object', 'category']).columns:
            X_clean[col] = X_clean[col].fillna(X_clean[col].mode()[0] if not X_clean[col].mode().empty else 'unknown')
        
        # Encode categorical variables
        X_encoded = pd.get_dummies(X_clean, drop_first=True)
        
        # Split data
        X_train, X_test, y_train, y_test = train_test_split(
            X_encoded, y, test_size=test_size, random_state=random_state, stratify=y
        )
        
        return X_train, X_test, y_train, y_test
    
    def build_base_models(self, X_train, X_test, y_train, y_test) -> dict:
        """Build and evaluate multiple base models"""
        
        models = {
            'logistic_regression': LogisticRegression(max_iter=1000, random_state=42),
            'random_forest': RandomForestClassifier(n_estimators=100, random_state=42),
            'gradient_boosting': GradientBoostingClassifier(random_state=42),
            'xgboost': xgb.XGBClassifier(random_state=42, eval_metric='logloss'),
            'lightgbm': lgb.LGBMClassifier(random_state=42, verbose=-1)
        }
        
        results = {}
        
        for name, model in models.items():
            print(f"Training {name}...")
            
            # Train model
            model.fit(X_train, y_train)
            
            # Make predictions
            y_pred = model.predict(X_test)
            y_pred_proba = None
            
            if hasattr(model, 'predict_proba'):
                y_pred_proba = model.predict_proba(X_test)[:, 1]
            
            # Evaluate
            accuracy = model.score(X_test, y_test)
            
            if len(np.unique(y_test)) == 2 and y_pred_proba is not None:
                roc_auc = roc_auc_score(y_test, y_pred_proba)
            else:
                roc_auc = None
            
            # Cross-validation
            cv_scores = cross_val_score(model, X_train, y_train, cv=5, scoring='accuracy')
            
            results[name] = {
                'model': model,
                'accuracy': accuracy,
                'roc_auc': roc_auc,
                'cv_mean': cv_scores.mean(),
                'cv_std': cv_scores.std(),
                'predictions': y_pred,
                'probabilities': y_pred_proba
            }
            
            # Feature importance
            if hasattr(model, 'feature_importances_'):
                importance_dict = dict(zip(X_train.columns, model.feature_importances_))
                self.feature_importance[name] = importance_dict
        
        self.models = results
        return results
    
    def hyperparameter_tuning(self, X_train, y_train, model_type: str = 'random_forest',
                            n_trials: int = 100) -> dict:
        """Hyperparameter tuning using Optuna"""
        
        def objective(trial):
            if model_type == 'random_forest':
                params = {
                    'n_estimators': trial.suggest_int('n_estimators', 50, 300),
                    'max_depth': trial.suggest_int('max_depth', 3, 20),
                    'min_samples_split': trial.suggest_int('min_samples_split', 2, 20),
                    'min_samples_leaf': trial.suggest_int('min_samples_leaf', 1, 10),
                    'max_features': trial.suggest_categorical('max_features', ['sqrt', 'log2', None])
                }
                model = RandomForestClassifier(**params, random_state=42, n_jobs=-1)
                
            elif model_type == 'xgboost':
                params = {
                    'n_estimators': trial.suggest_int('n_estimators', 50, 300),
                    'max_depth': trial.suggest_int('max_depth', 3, 10),
                    'learning_rate': trial.suggest_float('learning_rate', 0.01, 0.3),
                    'subsample': trial.suggest_float('subsample', 0.6, 1.0),
                    'colsample_bytree': trial.suggest_float('colsample_bytree', 0.6, 1.0),
                    'gamma': trial.suggest_float('gamma', 0, 5),
                    'reg_alpha': trial.suggest_float('reg_alpha', 0, 5),
                    'reg_lambda': trial.suggest_float('reg_lambda', 0, 5)
                }
                model = xgb.XGBClassifier(**params, random_state=42, eval_metric='logloss')
                
            elif model_type == 'lightgbm':
                params = {
                    'n_estimators': trial.suggest_int('n_estimators', 50, 300),
                    'max_depth': trial.suggest_int('max_depth', 3, 15),
                    'learning_rate': trial.suggest_float('learning_rate', 0.01, 0.3),
                    'num_leaves': trial.suggest_int('num_leaves', 20, 150),
                    'feature_fraction': trial.suggest_float('feature_fraction', 0.6, 1.0),
                    'bagging_fraction': trial.suggest_float('bagging_fraction', 0.6, 1.0),
                    'bagging_freq': trial.suggest_int('bagging_freq', 1, 10),
                    'min_child_samples': trial.suggest_int('min_child_samples', 5, 100)
                }
                model = lgb.LGBMClassifier(**params, random_state=42, verbose=-1)
            
            # Cross-validation
            scores = cross_val_score(model, X_train, y_train, cv=5, scoring='accuracy')
            return scores.mean()
        
        # Create study
        study = optuna.create_study(direction='maximize')
        study.optimize(objective, n_trials=n_trials)
        
        # Train best model
        best_params = study.best_params
        
        if model_type == 'random_forest':
            best_model = RandomForestClassifier(**best_params, random_state=42, n_jobs=-1)
        elif model_type == 'xgboost':
            best_model = xgb.XGBClassifier(**best_params, random_state=42, eval_metric='logloss')
        elif model_type == 'lightgbm':
            best_model = lgb.LGBMClassifier(**best_params, random_state=42, verbose=-1)
        
        best_model.fit(X_train, y_train)
        
        return {
            'best_model': best_model,
            'best_params': best_params,
            'best_score': study.best_value,
            'study': study
        }
    
    def evaluate_model(self, model, X_test, y_test, model_name: str = "model") -> dict:
        """Comprehensive model evaluation"""
        
        # Make predictions
        y_pred = model.predict(X_test)
        y_pred_proba = None
        
        if hasattr(model, 'predict_proba'):
            y_pred_proba = model.predict_proba(X_test)
            if len(np.unique(y_test)) == 2:
                y_pred_proba_binary = y_pred_proba[:, 1]
            else:
                y_pred_proba_binary = None
        
        # Basic metrics
        accuracy = model.score(X_test, y_test)
        
        # Classification report
        class_report = classification_report(y_test, y_pred, output_dict=True)
        
        # Confusion matrix
        conf_matrix = confusion_matrix(y_test, y_pred)
        
        # ROC AUC for binary classification
        roc_auc = None
        if y_pred_proba_binary is not None:
            roc_auc = roc_auc_score(y_test, y_pred_proba_binary)
        
        # Feature importance
        feature_importance = None
        if hasattr(model, 'feature_importances_'):
            feature_importance = dict(zip(X_test.columns, model.feature_importances_))
        
        evaluation = {
            'model_name': model_name,
            'accuracy': accuracy,
            'classification_report': class_report,
            'confusion_matrix': conf_matrix.tolist(),
            'roc_auc': roc_auc,
            'feature_importance': feature_importance,
            'predictions': y_pred.tolist(),
            'probabilities': y_pred_proba.tolist() if y_pred_proba is not None else None
        }
        
        self.evaluation_results[model_name] = evaluation
        return evaluation
    
    def ensemble_models(self, X_train, X_test, y_train, y_test, 
                       models_to_ensemble: list = None) -> dict:
        """Create ensemble models"""
        
        if models_to_ensemble is None:
            models_to_ensemble = ['random_forest', 'xgboost', 'lightgbm']
        
        # Select models for ensemble
        ensemble_models = []
        for model_name in models_to_ensemble:
            if model_name in self.models:
                ensemble_models.append((model_name, self.models[model_name]['model']))
        
        if not ensemble_models:
            raise ValueError("No valid models found for ensemble")
        
        # Voting ensemble
        from sklearn.ensemble import VotingClassifier
        
        voting_ensemble = VotingClassifier(
            estimators=ensemble_models,
            voting='soft' if len(np.unique(y_train)) == 2 else 'hard'
        )
        
        voting_ensemble.fit(X_train, y_train)
        voting_evaluation = self.evaluate_model(voting_ensemble, X_test, y_test, "voting_ensemble")
        
        # Stacking ensemble
        from sklearn.ensemble import StackingClassifier
        
        base_estimators = [(name, model['model']) for name, model in self.models.items() 
                          if name in models_to_ensemble]
        
        stacking_ensemble = StackingClassifier(
            estimators=base_estimators,
            final_estimator=LogisticRegression(max_iter=1000, random_state=42),
            cv=5
        )
        
        stacking_ensemble.fit(X_train, y_train)
        stacking_evaluation = self.evaluate_model(stacking_ensemble, X_test, y_test, "stacking_ensemble")
        
        return {
            'voting_ensemble': voting_evaluation,
            'stacking_ensemble': stacking_evaluation,
            'voting_model': voting_ensemble,
            'stacking_model': stacking_ensemble
        }
    
    def model_explainability(self, model, X_test, feature_names: list = None) -> dict:
        """Model explainability using SHAP and LIME"""
        
        try:
            import shap
            import lime
            import lime.lime_tabular
            
            if feature_names is None:
                feature_names = X_test.columns.tolist()
            
            explainability_results = {}
            
            # SHAP explanations
            if hasattr(model, 'predict_proba'):
                explainer = shap.TreeExplainer(model)
                shap_values = explainer.shap_values(X_test)
                
                # For binary classification, use the positive class
                if isinstance(shap_values, list) and len(shap_values) == 2:
                    shap_values_positive = shap_values[1]
                else:
                    shap_values_positive = shap_values
                
                explainability_results['shap'] = {
                    'values': shap_values_positive.tolist(),
                    'base_values': explainer.expected_value.tolist() if hasattr(explainer.expected_value, '__iter__') else explainer.expected_value,
                    'feature_names': feature_names
                }
                
                # Feature importance from SHAP
                shap_importance = np.abs(shap_values_positive).mean(0)
                shap_importance_dict = dict(zip(feature_names, shap_importance))
                explainability_results['shap_importance'] = shap_importance_dict
            
            # LIME explanations
            lime_explainer = lime.lime_tabular.LimeTabularExplainer(
                X_test.values,
                feature_names=feature_names,
                mode='classification',
                discretize_continuous=True
            )
            
            # Explain a few instances
            lime_explanations = []
            for i in range(min(5, len(X_test))):
                exp = lime_explainer.explain_instance(
                    X_test.iloc[i].values,
                    model.predict_proba,
                    num_features=10
                )
                lime_explanations.append({
                    'instance_index': i,
                    'explanation': exp.as_list(),
                    'score': exp.score
                })
            
            explainability_results['lime'] = lime_explanations
            
        except ImportError:
            explainability_results = {'error': 'SHAP and LIME packages not installed'}
        
        return explainability_results

# Usage example
# ml_builder = MLModelBuilder()
# X_train, X_test, y_train, y_test = ml_builder.prepare_data(X, y)
# base_models = ml_builder.build_base_models(X_train, X_test, y_train, y_test)
# tuning_result = ml_builder.hyperparameter_tuning(X_train, y_train, 'xgboost')
# evaluation = ml_builder.evaluate_model(tuning_result['best_model'], X_test, y_test, 'tuned_xgboost')
# ensemble_results = ml_builder.ensemble_models(X_train, X_test, y_train, y_test)
# explainability = ml_builder.model_explainability(tuning_result['best_model'], X_test)
```

## Level 3: Advanced Integration

### Production ML Systems

#### 4. ML Pipeline with Experiment Tracking

```python
import mlflow
import mlflow.sklearn
import mlflow.xgboost
import joblib
import json
from datetime import datetime
import os

class ProductionMLPipeline:
    def __init__(self, experiment_name: str, tracking_uri: str = None):
        self.experiment_name = experiment_name
        
        if tracking_uri:
            mlflow.set_tracking_uri(tracking_uri)
        
        mlflow.set_experiment(experiment_name)
        self.experiment_id = mlflow.get_experiment_by_name(experiment_name).experiment_id
        
        self.pipeline_stages = []
        self.run_id = None
    
    def run_complete_pipeline(self, data_path: str, target_column: str,
                            model_types: list = None, save_models: bool = True) -> dict:
        """Run complete ML pipeline with experiment tracking"""
        
        with mlflow.start_run(run_name=f"pipeline_run_{datetime.now().strftime('%Y%m%d_%H%M%S')}") as run:
            self.run_id = run.info.run_id
            
            # Log dataset information
            mlflow.log_param("data_path", data_path)
            mlflow.log_param("target_column", target_column)
            
            # Stage 1: Data Loading and Exploration
            print("Stage 1: Loading and exploring data...")
            data = self._load_data(data_path)
            data_info = self._explore_data(data)
            
            for key, value in data_info.items():
                if isinstance(value, (int, float)):
                    mlflow.log_metric(f"data_{key}", value)
                else:
                    mlflow.log_param(f"data_{key}", str(value))
            
            # Stage 2: Data Preprocessing
            print("Stage 2: Preprocessing data...")
            preprocessor = DataPreprocessor()
            processed_data = preprocessor.create_features(data)
            
            X = processed_data.drop(target_column, axis=1)
            y = processed_data[target_column]
            
            X_train, X_test, y_train, y_test = preprocessor.prepare_data(X, y)
            
            mlflow.log_param("train_size", len(X_train))
            mlflow.log_param("test_size", len(X_test))
            mlflow.log_param("feature_count", X_train.shape[1])
            
            # Stage 3: Model Training and Evaluation
            print("Stage 3: Training and evaluating models...")
            if model_types is None:
                model_types = ['random_forest', 'xgboost', 'lightgbm']
            
            ml_builder = MLModelBuilder()
            model_results = ml_builder.build_base_models(X_train, X_test, y_train, y_test)
            
            best_model_name = None
            best_score = 0
            best_model = None
            
            for model_name, result in model_results.items():
                if model_name in model_types:
                    # Log model metrics
                    mlflow.log_metric(f"{model_name}_accuracy", result['accuracy'])
                    mlflow.log_metric(f"{model_name}_cv_mean", result['cv_mean'])
                    
                    if result['roc_auc'] is not None:
                        mlflow.log_metric(f"{model_name}_roc_auc", result['roc_auc'])
                    
                    # Log model
                    try:
                        if 'xgboost' in model_name:
                            mlflow.xgboost.log_model(result['model'], f"{model_name}_model")
                        else:
                            mlflow.sklearn.log_model(result['model'], f"{model_name}_model")
                    except:
                        pass
                    
                    # Track best model
                    if result['accuracy'] > best_score:
                        best_score = result['accuracy']
                        best_model_name = model_name
                        best_model = result['model']
            
            # Stage 4: Hyperparameter Tuning for Best Model
            print("Stage 4: Hyperparameter tuning...")
            if best_model_name:
                tuning_result = ml_builder.hyperparameter_tuning(
                    X_train, y_train, 
                    model_type=best_model_name.replace('_tuned', ''),
                    n_trials=50
                )
                
                # Log best parameters
                for param, value in tuning_result['best_params'].items():
                    mlflow.log_param(f"best_{param}", value)
                
                mlflow.log_metric("best_tuned_score", tuning_result['best_score'])
                
                # Evaluate tuned model
                tuned_evaluation = ml_builder.evaluate_model(
                    tuning_result['best_model'], X_test, y_test, 
                    f"{best_model_name}_tuned"
                )
                
                mlflow.log_metric("tuned_model_accuracy", tuned_evaluation['accuracy'])
                if tuned_evaluation['roc_auc'] is not None:
                    mlflow.log_metric("tuned_model_roc_auc", tuned_evaluation['roc_auc'])
                
                # Update best model if tuning improved performance
                if tuned_evaluation['accuracy'] > best_score:
                    best_model = tuning_result['best_model']
                    best_score = tuned_evaluation['accuracy']
                
                # Log feature importance
                if tuned_evaluation['feature_importance']:
                    importance_json = json.dumps(tuned_evaluation['feature_importance'])
                    mlflow.log_text(importance_json, "feature_importance.json")
            
            # Stage 5: Ensemble Creation
            print("Stage 5: Creating ensemble models...")
            try:
                ensemble_results = ml_builder.ensemble_models(X_train, X_test, y_train, y_test)
                
                for ensemble_type, result in ensemble_results.items():
                    if 'evaluation' in ensemble_type:
                        mlflow.log_metric(f"{ensemble_type}_accuracy", result['accuracy'])
                        if result['roc_auc'] is not None:
                            mlflow.log_metric(f"{ensemble_type}_roc_auc", result['roc_auc'])
            except:
                print("Ensemble creation failed, continuing...")
            
            # Stage 6: Model Explainability
            print("Stage 6: Generating model explanations...")
            if best_model is not None:
                explainability = ml_builder.model_explainability(best_model, X_test)
                
                if 'shap_importance' in explainability:
                    shap_importance_json = json.dumps(explainability['shap_importance'])
                    mlflow.log_text(shap_importance_json, "shap_importance.json")
            
            # Save final model
            if save_models and best_model is not None:
                model_path = f"models/{best_model_name}_final"
                os.makedirs(model_path, exist_ok=True)
                joblib.dump(best_model, f"{model_path}/model.joblib")
                mlflow.log_artifact(f"{model_path}/model.joblib", "final_model")
            
            # Generate pipeline report
            pipeline_results = {
                'run_id': self.run_id,
                'experiment_id': self.experiment_id,
                'data_info': data_info,
                'best_model': best_model_name,
                'best_score': best_score,
                'model_results': {name: {'accuracy': result['accuracy'], 'cv_mean': result['cv_mean']} 
                                for name, result in model_results.items()},
                'timestamp': datetime.now().isoformat()
            }
            
            # Log pipeline results
            results_json = json.dumps(pipeline_results, indent=2)
            mlflow.log_text(results_json, "pipeline_results.json")
            
            return pipeline_results
    
    def _load_data(self, data_path: str) -> pd.DataFrame:
        """Load data from various formats"""
        
        if data_path.endswith('.csv'):
            return pd.read_csv(data_path)
        elif data_path.endswith('.parquet'):
            return pd.read_parquet(data_path)
        elif data_path.endswith('.json'):
            return pd.read_json(data_path)
        else:
            raise ValueError(f"Unsupported file format: {data_path}")
    
    def _explore_data(self, df: pd.DataFrame) -> dict:
        """Basic data exploration"""
        
        return {
            'shape': df.shape,
            'missing_values': df.isnull().sum().sum(),
            'duplicate_rows': df.duplicated().sum(),
            'numeric_columns': len(df.select_dtypes(include=[np.number]).columns),
            'categorical_columns': len(df.select_dtypes(include=['object', 'category']).columns),
            'memory_usage_mb': df.memory_usage(deep=True).sum() / 1024**2
        }
    
    def compare_experiments(self, metric: str = 'accuracy', top_k: int = 5) -> dict:
        """Compare multiple experiments"""
        
        from mlflow.tracking import MlflowClient
        
        client = MlflowClient()
        runs = client.search_runs(
            experiment_ids=[self.experiment_id],
            order_by=[f"metrics.{metric} DESC"]
        )
        
        comparison_results = []
        
        for run in runs[:top_k]:
            run_data = {
                'run_id': run.info.run_id,
                'status': run.info.status,
                'start_time': datetime.fromtimestamp(run.info.start_time / 1000),
                'params': dict(run.data.params),
                'metrics': dict(run.data.metrics)
            }
            comparison_results.append(run_data)
        
        return {
            'comparison_metric': metric,
            'top_runs': comparison_results,
            'total_runs': len(runs)
        }
    
    def deploy_model(self, run_id: str = None, model_name: str = "production_model") -> dict:
        """Prepare model for deployment"""
        
        if run_id is None:
            run_id = self.run_id
        
        if run_id is None:
            raise ValueError("No run ID provided and no current run found")
        
        # Load model from MLflow
        model_uri = f"runs:/{run_id}/final_model"
        model = mlflow.sklearn.load_model(model_uri)
        
        # Create deployment package
        deployment_info = {
            'model_uri': model_uri,
            'run_id': run_id,
            'model_name': model_name,
            'deployment_timestamp': datetime.now().isoformat(),
            'model_type': type(model).__name__
        }
        
        # Save deployment info
        os.makedirs('deployment', exist_ok=True)
        with open('deployment/deployment_info.json', 'w') as f:
            json.dump(deployment_info, f, indent=2)
        
        # Copy model to deployment directory
        joblib.dump(model, 'deployment/model.joblib')
        
        return deployment_info

# Usage example
# pipeline = ProductionMLPipeline("data_science_experiment")
# results = pipeline.run_complete_pipeline("data.csv", "target")
# comparison = pipeline.compare_experiments()
# deployment_info = pipeline.deploy_model()
```

## Related Skills

- **moai-domain-ml**: Advanced machine learning techniques
- **moai-domain-testing**: Model testing and validation
- **moai-document-processing**: Data processing and extraction
- **moai-essentials-refactor**: Code optimization and performance

## Quick Start Checklist

- [ ] Load and explore dataset structure
- [ ] Perform comprehensive data preprocessing
- [ ] Conduct statistical analysis and hypothesis testing
- [ ] Build baseline machine learning models
- [ ] Implement hyperparameter tuning
- [ ] Create ensemble models for better performance
- [ ] Generate model explanations and feature importance
- [ ] Set up experiment tracking with MLflow

## Data Science Best Practices

1. **Data Understanding**: Always explore and understand your data first
2. **Preprocessing**: Handle missing values, outliers, and feature engineering
3. **Statistical Validation**: Use appropriate statistical tests and validation
4. **Model Selection**: Try multiple algorithms and compare performance
5. **Cross-Validation**: Use proper cross-validation techniques
6. **Hyperparameter Tuning**: Optimize model parameters systematically
7. **Interpretability**: Use SHAP, LIME for model explanations
8. **Experiment Tracking**: Log all experiments for reproducibility

---

**Data Science & Analytics** - Build end-to-end data science solutions with comprehensive statistical analysis, machine learning, and experiment tracking capabilities.
