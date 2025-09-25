# moai_adk.resources.templates.scripts.check_secrets

MoAI-ADK Secrets Scanner
소스 코드에서 시크릿 정보 탐지 및 보안 검사

## Functions

### print_report

보고서 출력

```python
print_report(report, verbose)
```

### main

```python
main()
```

### __init__

```python
__init__(self)
```

### scan_file

단일 파일에서 시크릿 스캔

```python
scan_file(self, file_path)
```

### determine_severity

시크릿의 심각도 결정

```python
determine_severity(self, line, secret_type, value)
```

### scan_directory

디렉토리 전체 스캔

```python
scan_directory(self, directory)
```

### generate_report

보고서 생성

```python
generate_report(self, secrets)
```

## Classes

### SecretsScanner
