# moai_adk.resources.templates.scripts.check_licenses

MoAI-ADK License Compliance Checker v0.1.12
프로젝트 의존성 라이선스 검사 및 호환성 검증

이 스크립트는 프로젝트의 모든 의존성을 스캔하여:
- 라이선스 정보 수집 및 분석
- 제한적 라이선스 (GPL, AGPL 등) 감지
- 라이선스 정책 준수 확인
- 상세한 라이선스 리포트 생성

## Functions

### main

메인 실행 함수

```python
main()
```

### __init__

```python
__init__(self, project_root)
```

### init_license_database

라이선스 정보 데이터베이스 초기화

```python
init_license_database(self)
```

### load_license_policy

라이선스 정책 로드

```python
load_license_policy(self)
```

### scan_python_dependencies

Python 의존성 스캔

```python
scan_python_dependencies(self)
```

### scan_nodejs_dependencies

Node.js 의존성 스캔

```python
scan_nodejs_dependencies(self)
```

### get_package_license

Python 패키지의 라이선스 정보 조회

```python
get_package_license(self, package_name)
```

### get_npm_package_license

NPM 패키지의 라이선스 정보 조회

```python
get_npm_package_license(self, package_name)
```

### normalize_license_name

라이선스 이름 정규화

```python
normalize_license_name(self, license_text)
```

### evaluate_license_compliance

라이선스 컴플라이언스 평가

```python
evaluate_license_compliance(self, license_name)
```

### generate_report

라이선스 스캔 리포트 생성

```python
generate_report(self, scan_results)
```

### get_license_distribution

라이선스 분포 통계

```python
get_license_distribution(self, results)
```

### generate_recommendations

개선 권장사항 생성

```python
generate_recommendations(self, results)
```

### run_scan

전체 라이선스 스캔 실행

```python
run_scan(self)
```

## Classes

### LicenseInfo

라이선스 정보 구조

### PackageLicense

패키지 라이선스 정보

### LicenseChecker

라이선스 호환성 검사기
